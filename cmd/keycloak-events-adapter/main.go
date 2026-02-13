package main

import (
	"context"
	"fmt"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/jessevdk/go-flags"
	tnt "github.com/tarantool/go-tarantool"
	tntqueue "github.com/tarantool/go-tarantool/queue"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"keycloak-events-adapter/internal"
	grpc_server "keycloak-events-adapter/internal/api/grpc"
	eventv1 "keycloak-events-adapter/internal/specs/gen/keycloak/event/v1"
	"keycloak-events-adapter/internal/tarantool"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var cfg Config
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal("Failed to parse config.", err)
	}

	logger, err := initLogger(cfg.LogLevel, cfg.LogJSON)
	if err != nil {
		log.Fatal("Failed to init logger.", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	defer func() {
		if msg := recover(); msg != nil {
			errR := fmt.Errorf("%s", msg)
			logger.Error("recovered from panic, but application will be terminated", zap.Error(errR))
		}
	}()

	osSigCh := make(chan os.Signal, 1)
	signal.Notify(
		osSigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		s := <-osSigCh
		if s == syscall.SIGINT ||
			s == syscall.SIGTERM ||
			s == syscall.SIGQUIT {
			logger.Info("Received signal! Process exited")
			cancelFunc()
		}
	}()

	tntConn, err := tnt.Connect(fmt.Sprintf("%s:%d", cfg.TntHost, cfg.TntPort), tnt.Opts{
		User:      cfg.TntUser,
		Pass:      cfg.TntPassword,
		Reconnect: 1 * time.Second,
		Timeout:   25 * time.Second,
	})
	if err != nil {
		logger.Fatal("can't connect tarantool", zap.Error(err))
	}

	tntQueue := tntqueue.New(tntConn, tarantool.EventsQueueName)
	ok, err := tntQueue.Exists()
	if err != nil {
		logger.Fatal("can't check queue existence", zap.Error(err))
	}
	if !ok {
		logger.Fatal("queue doesn't exist", zap.String("queue_name", tarantool.EventsQueueName))
	}

	tntAdminQueue := tntqueue.New(tntConn, tarantool.AdminEventsQueueName)
	ok, err = tntAdminQueue.Exists()
	if err != nil {
		logger.Fatal("can't check queue existence", zap.Error(err))
	}
	if !ok {
		logger.Fatal("queue doesn't exist", zap.String("queue_name", tarantool.AdminEventsQueueName))
	}

	adminEventNotify := internal.NewDummy[internal.AdminEvent](logger)
	eventNotify := internal.NewDummy[internal.Event](logger)
	adminEventStorage := tarantool.NewEvent[internal.AdminEvent](tntAdminQueue, adminEventNotify, logger)
	eventStorage := tarantool.NewEvent[internal.Event](tntQueue, eventNotify, logger)
	eventService := internal.NewEventService(adminEventStorage, eventStorage)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		errN := startGRPCServer(ctx, &cfg, eventService, logger)
		if errN != nil {
			logger.Error("can't start gRPC server or server return error while working", zap.Error(errN))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		eventService.Read(ctx, 5)
	}()

	wg.Wait()
	logger.Info("Application has been shutdown gracefully")
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(
	ctx context.Context,
	cfg *Config,
	eventService internal.EventProvider,
	logger *zap.Logger,
) error {
	logger.Info("gRPC started", zap.String("listen", cfg.GrpcListen))
	lis, err := net.Listen("tcp", cfg.GrpcListen)
	if err != nil {
		return fmt.Errorf("failed to listen GRPC server: %w", err)
	}

	recoverFromPanicHandler := func(p any) error {
		errN := fmt.Errorf("recovered from panic: %s", p)
		logger.Error("recovered from panic", zap.Error(errN))

		return errN
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoverFromPanicHandler),
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(opts...),
			grpc_validator.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(opts...),
			grpc_validator.StreamServerInterceptor(),
		),
	)

	eventv1.RegisterEventAPIServer(s, grpc_server.NewEventServer(eventService))

	reflection.Register(s)

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	return s.Serve(lis)
}

// initLogger создает и настраивает новый экземпляр логгера
func initLogger(logLevel string, isLogJSON bool) (*zap.Logger, error) {
	lvl := zap.InfoLevel
	err := lvl.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal log-level: %w", err)
	}

	opts := zap.NewProductionConfig()
	opts.Level = zap.NewAtomicLevelAt(lvl)
	opts.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if opts.InitialFields == nil {
		opts.InitialFields = map[string]any{}
	}

	if !isLogJSON {
		opts.Encoding = "console"
		opts.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return opts.Build()
}
