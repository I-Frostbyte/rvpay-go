package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/config"
	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	clientshttp "github.com/I-Frostbyte/rvpay-go/clients/http"
	"github.com/I-Frostbyte/rvpay-go/clients/oauth"
	"github.com/I-Frostbyte/rvpay-go/clients/payments"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/I-Frostbyte/rvpay-go/clients/service"
	"github.com/I-Frostbyte/rvpay-go/clients/webhooks"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	commondatabase "github.com/I-Frostbyte/rvpay-go/shared/database"
	commonlogger "github.com/I-Frostbyte/rvpay-go/shared/logger"
	commonobservability "github.com/I-Frostbyte/rvpay-go/shared/observability"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := commonlogger.New("", os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	err = run(ctx, logger)
	if err != nil {
		logger.Err(err).Msg("failed to run grpc service")
		os.Exit(1)
	}
}

func run(ctx context.Context, logger zerolog.Logger) error {
	logger.Info().Msg("starting clients grpc service...")

	cfg := config.Config{}
	err := cfg.LoadConfig()
	if err != nil {
		logger.Err(err).Msg("failed to load config")
		return err
	}
	logger.Info().Msg("successfully loaded configuration")

	logger, err = commonlogger.New(cfg.LogLevel, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	dbConnectionURL := getPostgresConnectionURL(cfg.DB)
	db, err := commondatabase.Connect(ctx, dbConnectionURL)
	if err != nil {
		return err
	}
	logger.Info().Msg("Successfully connected to database")
	defer db.Close()

	if cfg.RunMigrations {
		err = commondatabase.Migrate(dbConnectionURL, cfg.MigrationPath, logger)
		if err != nil {
			logger.Err(err).Msg("Migration not successful...")
			return fmt.Errorf("failed to migrate: %w", err)
		}
		logger.Info().Msg("Migrations successful...")
	} else {
		logger.Info().Msg("database migrations are managed externally")
	}

	clientsRepo := repo.NewClientsRepo(db)
	clientRepo := repo.NewClientRepo(clientsRepo.Do())
	platformRepo := repo.NewPlatformRepo(clientsRepo.Do())
	integrationRepo := repo.NewIntegrationRepo(clientsRepo.Do())
	oauthTokenRepo := repo.NewOAuthTokenRepo(clientsRepo.Do())
	webhookSubscriptionRepo := repo.NewWebhookSubscriptionRepo(clientsRepo.Do())
	oauthStateRepo := repo.NewOAuthStateRepo(clientsRepo.Do())
	webhookEventRepo := repo.NewWebhookEventRepo(clientsRepo.Do())
	paymentProviderConfigRepo := repo.NewPaymentProviderConfigRepo(clientsRepo.Do())

	providerRegistry := providers.NewProviderRegistry()
	// The HighLevel Custom Payment Provider client makes authenticated outbound
	// calls to HighLevel for provider registration/configuration. The base URL
	// comes from configuration (HIGHLEVEL_API_BASE_URL); it is never hard-coded.
	highLevelPaymentProvider := providers.NewHighLevelPaymentProviderClient(cfg.HighLevel.APIBaseURL, nil)
	highLevelProvider := providers.NewHighLevelProvider(cfg.HighLevel.ClientID, cfg.HighLevel.ClientSecret, cfg.HighLevel.RedirectURI, cfg.HighLevel.WebhookPublicKey, highLevelPaymentProvider)
	providerRegistry.Register(highLevelProvider)
	logger.Info().Msg("providers registered successfully")

	clientsService := service.NewClientsServiceImpl(clientRepo, logger)
	platformsService := service.NewPlatformsServiceImpl(platformRepo, logger)
	integrationsService := service.NewIntegrationsServiceImpl(integrationRepo, clientRepo, platformRepo, oauthTokenRepo, webhookSubscriptionRepo, logger)

	oauthService := oauth.NewService(
		integrationRepo,
		oauthTokenRepo,
		clientRepo,
		platformRepo,
		oauthStateRepo,
		paymentProviderConfigRepo,
		providerRegistry,
		cfg.HighLevel.RedirectURI,
		oauth.ProviderConfigSettings{
			Name:        cfg.HighLevel.ProviderName,
			Description: cfg.HighLevel.ProviderDescription,
			ImageURL:    cfg.HighLevel.ProviderImageURL,
			PaymentsURL: cfg.HighLevel.PaymentURL,
			QueryURL:    cfg.HighLevel.QueryURL,
		},
		logger,
	)
	webhookService := webhooks.NewService(integrationRepo, webhookSubscriptionRepo, webhookEventRepo, platformRepo, providerRegistry, logger)
	oauthHandler := clientshttp.NewOAuthHandler(oauthService, logger)
	webhookHandler := clientshttp.NewWebhookHandler(webhookService, logger)

	// GHL Custom Payment Provider integration. The Clients service owns the
	// GHL-facing payment query and webhook endpoints; it correlates HighLevel
	// transactions with RVPay deposits by calling the Transactions service via
	// gRPC. The Transactions gRPC address comes from configuration
	// (TRANSACTIONS_GRPC_ADDR); it is never hard-coded.
	transactionsAddr := os.Getenv("TRANSACTIONS_GRPC_ADDR")
	if transactionsAddr == "" {
		transactionsAddr = "localhost:50052"
	}
	transactionsConn, err := grpc.NewClient(transactionsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect to transactions service: %w", err)
	}
	defer transactionsConn.Close()
	transactionsClient := transactionsgrpc.NewPaymentServiceClient(transactionsConn)

	paymentService := payments.NewService(paymentProviderConfigRepo, integrationRepo, webhookEventRepo, transactionsClient, logger)
	paymentQueryHandler := clientshttp.NewPaymentQueryHandler(paymentService, logger)
	paymentWebhookHandler := clientshttp.NewPaymentWebhookHandler(paymentService, logger)

	svrOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			commonobservability.UnaryServerInterceptor(logger),
		),
	}

	grpcServer := grpc.NewServer(svrOpts...)
	reflection.Register(grpcServer)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	clientsgrpc.RegisterClientsServiceServer(grpcServer, clientsService)
	clientsgrpc.RegisterPlatformsServiceServer(grpcServer, platformsService)
	clientsgrpc.RegisterIntegrationsServiceServer(grpcServer, integrationsService)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	logger.Info().Msg("Successfully registered gRPC services...")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.ListenPort))
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	gatewayMux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := clientsgrpc.RegisterClientsServiceHandlerServer(ctx, gatewayMux, clientsService); err != nil {
		return fmt.Errorf("register clients grpc-gateway handler: %w", err)
	}
	if err := clientsgrpc.RegisterPlatformsServiceHandlerServer(ctx, gatewayMux, platformsService); err != nil {
		return fmt.Errorf("register platforms grpc-gateway handler: %w", err)
	}
	if err := clientsgrpc.RegisterIntegrationsServiceHandlerServer(ctx, gatewayMux, integrationsService); err != nil {
		return fmt.Errorf("register integrations grpc-gateway handler: %w", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", commonobservability.AccessLog(logger)(gatewayMux))
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if ctx.Err() != nil {
			http.Error(w, "shutting down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// GoHighLevel OAuth callback and webhook delivery endpoints. These are
	// registered as direct HTTP handlers (not grpc-gateway RPCs) because they
	// are external provider/browser-facing endpoints, consistent with the
	// project's protobuf transport strategy.
	httpMux.HandleFunc("/oauth/callback", oauthHandler.Callback)
	httpMux.HandleFunc("/webhooks/highlevel", webhookHandler.HighLevel)
	// GHL Custom Payment Provider endpoints. These are distinct from the
	// Marketplace OAuth callback and webhook; they handle payment query
	// operations and payment-provider webhook events.
	httpMux.HandleFunc("/payments/custom-provider/query", paymentQueryHandler.Query)
	httpMux.HandleFunc("/payments/custom-provider/webhook", paymentWebhookHandler.Payment)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	logger.Info().Msgf("gRPC service is listening on port: %d", cfg.ListenPort)
	logger.Info().Msgf("HTTP gateway is listening on port: %s", httpServer.Addr)

	var startupErr error
	startupErrCh := make(chan error, 2)
	var startupErrOnce sync.Once
	reportStartupErr := func(err error) {
		if err == nil {
			return
		}
		startupErrOnce.Do(func() {
			startupErrCh <- err
			cancel()
		})
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(listener)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			reportStartupErr(fmt.Errorf("grpcServer.Serve: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			reportStartupErr(fmt.Errorf("httpServer.ListenAndServe: %w", err))
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info().Msg("Shutting down servers...")
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			reportStartupErr(fmt.Errorf("httpServer.Shutdown: %w", err))
		}
		grpcServer.GracefulStop()
		logger.Info().Msg("Servers stopped.")
	}()

	logger.Info().Msgf("gRPC server running on %s", listener.Addr().String())
	logger.Info().Msgf("HTTP gateway running on %s", httpServer.Addr)

	wg.Wait()

	select {
	case startupErr = <-startupErrCh:
	default:
	}

	if startupErr != nil {
		return startupErr
	}

	logger.Info().Msg("servers have shut down gracefully...")
	return nil
}

func getPostgresConnectionURL(cfg config.DBConfig) string {
	return commondatabase.PostgresURL(cfg.DBUser, cfg.DBPassword, cfg.DBPort, cfg.DBHost, cfg.DBName, cfg.TLSDisabled)
}
