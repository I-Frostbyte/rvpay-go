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

	model "github.com/I-Frostbyte/rvpay-go/integrations/config"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo"
	"github.com/I-Frostbyte/rvpay-go/integrations/integrations"
	"github.com/I-Frostbyte/rvpay-go/integrations/oauth"
	"github.com/I-Frostbyte/rvpay-go/integrations/webhook"
	"github.com/I-Frostbyte/rvpay-go/grpc/go/integrationsgrpc"
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
	logger.Info().Msg("starting grpc service...")

	config := model.Config{}

	err := config.LoadConfig()
	if err != nil {
		logger.Err(err).Msg("failed to load config")
		return err
	}

	logger.Info().Msg("successfully loaded configuration")

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	logger = logger.Level(logLevel)

	dbConnectionURL := getPostgresConnectionURL(config.DB)
	db, err := pgxpool.New(ctx, dbConnectionURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// FORCE AN EAGER CONNECTION TO DETECT INVALID PORT IN URL IMMEDIATELY
	err = db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to actually connect to database: %w", err)
	}

	// This line will now only print if the URL syntax and network route are 100% correct
	logger.Info().Msg("Successfully connected and pinged database!")

	defer db.Close()

	if config.RunMigrations {
		err = repo.Migrate(dbConnectionURL, config.MigrationPath, logger)
		if err != nil {
			logger.Err(err).Msg("Migration not successful...")
			return fmt.Errorf("failed to migrate: %w", err)
		}

		logger.Info().Msg("Migrations successful...")
	} else {
		logger.Info().Msg("database migrations are managed externally")
	}

	integrationsRepo := repo.NewIntegrationsRepo(db)
	logger.Info().Msg("Successfully initialized integrations repository...")

	integrationService := integrations.NewIntegrationService(integrationsRepo, logger)
	logger.Info().Msg("Successfully initialized integration service...")

	oauthService := oauth.NewService(
		integrationsRepo,
		logger,
		config.HighLevelClientID,
		config.HighLevelClientSecret,
		config.HighLevelRedirectURL,
		config.TokenEncryptionKey,
	)
	oauthHandler := oauth.NewHandler(oauthService)
	logger.Info().Msg("Successfully initialized oauth service...")

	webhookService := webhook.NewService(integrationsRepo, logger)
	webhookHandler := webhook.NewHandler(webhookService)
	logger.Info().Msg("Successfully initialized webhook service...")

	svrOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
		),
	}

	grpcServer := grpc.NewServer(svrOpts...)
	reflection.Register(grpcServer)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	integrationsgrpc.RegisterIntegrationServiceServer(grpcServer, integrationService)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	logger.Info().Msg("Successfully registered IntegrationServiceServer...")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.ListenPort))
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	gatewayMux := runtime.NewServeMux()
	if err := integrationsgrpc.RegisterIntegrationServiceHandlerServer(ctx, gatewayMux, integrationService); err != nil {
		return fmt.Errorf("register grpc-gateway handler: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	httpMux.HandleFunc("/oauth/callback", oauthHandler.Callback)
	httpMux.HandleFunc("/webhooks/highlevel", webhookHandler.HighLevel)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	logger.Info().Msgf(`grpc service is listening on port: %s`, listener.Addr().String())
	logger.Info().Msgf(`http gateway is listening on port: %s`, httpServer.Addr)

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

	logger.Info().Msgf(`gRPC server running on %s`, listener.Addr().String())
	logger.Info().Msgf(`HTTP gateway running on %s`, httpServer.Addr)

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

func getPostgresConnectionURL(config model.DBConfig) string {
	queryValues := url.Values{}
	if config.TLSDisabled {
		queryValues.Add("sslmode", "disable")
	} else {
		queryValues.Add("sslmode", "require")
	}

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(config.DBUser, config.DBPassword),
		Host:     fmt.Sprintf("%s:%d", config.DBHost, config.DBPort),
		Path:     config.DBName,
		RawQuery: queryValues.Encode(),
	}

	return dbURL.String()
}