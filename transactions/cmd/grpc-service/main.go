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

	"github.com/I-Frostbyte/rvpay-go/transactions/config"
	"github.com/I-Frostbyte/rvpay-go/transactions/customers"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/deposits"
	"github.com/I-Frostbyte/rvpay-go/transactions/merchants"
	"github.com/I-Frostbyte/rvpay-go/transactions/payouts"
	"github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	commondatabase "github.com/I-Frostbyte/rvpay-go/shared/database"
	commonlogger "github.com/I-Frostbyte/rvpay-go/shared/logger"
	commonobservability "github.com/I-Frostbyte/rvpay-go/shared/observability"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
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
	logger.Info().Msg("starting grpc service...")

	config := model.Config{}

	err := config.LoadConfig()
	if err != nil {
		logger.Err(err).Msg("failed to load config")
		return err
	}

	logger.Info().Msg("successfully loaded configuration")

	logger, err = commonlogger.New(config.LogLevel, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	dbConnectionURL := getPostgresConnectionURL(config.DB)
	db, err := commondatabase.Connect(ctx, dbConnectionURL)
	if err != nil {
		return err
	}

	// This line will now only print if the URL syntax and network route are 100% correct
	logger.Info().Msg("Successfully connected and pinged database!")

	defer db.Close()

	if config.RunMigrations {
		err = commondatabase.Migrate(dbConnectionURL, config.MigrationPath, logger)
		if err != nil {
			logger.Err(err).Msg("Migration not successful...")
			return fmt.Errorf("failed to migrate: %w", err)
		}

		logger.Info().Msg("Migrations successful...")
	} else {
		logger.Info().Msg("database migrations are managed externally")
	}

	transactionsRepo := repo.NewTransactionsRepo(db)
	queries := transactionsRepo.Do()

	merchantRepo := repo.NewMerchantRepo(queries)
	customerRepo := repo.NewCustomerRepo(queries)
	depositRepo := repo.NewDepositRepo(queries)
	payoutRepo := repo.NewPayoutRepo(queries)

	merchantService := merchants.NewMerchantService(merchantRepo, logger)
	customerService := customers.NewCustomerService(customerRepo, logger)
	depositService := deposits.NewDepositService(depositRepo, customerRepo, logger)
	payoutService := payouts.NewPayoutService(payoutRepo, logger)

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

	transactionsgrpc.RegisterMerchantServiceServer(grpcServer, merchantService)
	transactionsgrpc.RegisterCustomerServiceServer(grpcServer, customerService)
	transactionsgrpc.RegisterDepositServiceServer(grpcServer, depositService)
	transactionsgrpc.RegisterPayoutServiceServer(grpcServer, payoutService)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	logger.Info().Msg("Successfully registered Transactions services...")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.ListenPort))
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	gatewayMux := runtime.NewServeMux()
	if err := transactionsgrpc.RegisterMerchantServiceHandlerServer(ctx, gatewayMux, merchantService); err != nil {
		return fmt.Errorf("register grpc-gateway merchant handler: %w", err)
	}
	if err := transactionsgrpc.RegisterCustomerServiceHandlerServer(ctx, gatewayMux, customerService); err != nil {
		return fmt.Errorf("register grpc-gateway customer handler: %w", err)
	}
	if err := transactionsgrpc.RegisterDepositServiceHandlerServer(ctx, gatewayMux, depositService); err != nil {
		return fmt.Errorf("register grpc-gateway deposit handler: %w", err)
	}
	if err := transactionsgrpc.RegisterPayoutServiceHandlerServer(ctx, gatewayMux, payoutService); err != nil {
		return fmt.Errorf("register grpc-gateway payout handler: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	return commondatabase.PostgresURL(config.DBUser, config.DBPassword, int(config.DBPort), config.DBHost, config.DBName, config.TLSDisabled)
}
