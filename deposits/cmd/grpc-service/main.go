package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/I-Frostbyte/rvpay-go/deposits/config"
	"github.com/I-Frostbyte/rvpay-go/deposits/deposits"
	"github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
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

	logger.Info().Msgf("Successfully loaded config...: %+v", config)

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	logger = logger.Level(logLevel)

	/*
		dbConnectionURL := getPostgresConnectionURL(config.DB)
		db, err := pgxpool.New(ctx, dbConnectionURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		defer db.Close()
	*/

	svrOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
		),
	}

	grpcServer := grpc.NewServer(svrOpts...)
	reflection.Register(grpcServer)

	depositsgrpc.RegisterDepositsServiceServer(grpcServer, deposits.NewDepositsService(logger))
	logger.Info().Msg("Successfully registered DepositsServiceServer...")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.ListenPort))
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	logger.Info().Msgf(`grpc service is listening on port: %s`, listener.Addr().String())

	var startupErr error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = grpcServer.Serve(listener)
		if err != nil {
			startupErr = fmt.Errorf("grpcServer.Serve: %w", err)
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info().Msg("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		logger.Info().Msg("gRPC server stopped.")
	}()

	logger.Info().Msgf(`HTTP server running on %s`, listener.Addr().String())

	wg.Wait()
	logger.Info().Msg("gRPC server has shut down gracefully...")

	return startupErr
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
