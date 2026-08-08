package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/config"
	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/oauth"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/I-Frostbyte/rvpay-go/clients/service"
	"github.com/I-Frostbyte/rvpay-go/clients/webhooks"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()

	err := run(ctx, logger)
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

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	logger = logger.Level(logLevel)

	dbConnectionURL := getPostgresConnectionURL(cfg.DB)
	db, err := pgxpool.New(ctx, dbConnectionURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info().Msg("Successfully connected to database")
	defer db.Close()

	if cfg.RunMigrations {
		err = repo.Migrate(dbConnectionURL, cfg.MigrationPath, logger)
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

	providerRegistry := providers.NewProviderRegistry()
	highLevelProvider := providers.NewHighLevelProvider(cfg.HighLevel.ClientID, cfg.HighLevel.ClientSecret, cfg.HighLevel.RedirectURI)
	providerRegistry.Register(highLevelProvider)
	logger.Info().Msg("providers registered successfully")

	clientsService := service.NewClientsServiceImpl(clientRepo, logger)
	platformsService := service.NewPlatformsServiceImpl(platformRepo, logger)
	integrationsService := service.NewIntegrationsServiceImpl(integrationRepo, clientRepo, platformRepo, oauthTokenRepo, webhookSubscriptionRepo, logger)

	oauth.NewService(integrationRepo, oauthTokenRepo, clientRepo, platformRepo, providerRegistry, logger)
	webhooks.NewService(integrationRepo, webhookSubscriptionRepo, platformRepo, providerRegistry, logger)

	svrOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
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
	httpMux.Handle("/", gatewayMux)
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
	queryValues := url.Values{}
	if cfg.TLSDisabled {
		queryValues.Add("sslmode", "disable")
	} else {
		queryValues.Add("sslmode", "require")
	}

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     fmt.Sprintf("%s:%d", cfg.DBHost, cfg.DBPort),
		Path:     cfg.DBName,
		RawQuery: queryValues.Encode(),
	}

	return dbURL.String()
}