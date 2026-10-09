package main

import (
	partygrpc "Parties/internal/party/handler/grpc"
	partyhttp "Parties/internal/party/handler/http"
	"Parties/internal/party/repository"
	"Parties/internal/party/service"
	"Parties/internal/shared/config"
	"Parties/internal/shared/database"
	"Parties/internal/shared/logger"
	"Parties/internal/shared/migration"
	"context"
	"errors"
	"log/slog"
	"net"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	partyv1 "github.com/DoesNotNeedASword/proto-contracts/gen/go/party/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	if err := run(cfg, log); err != nil {
		log.Error("run server", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.RunMigrations {
		if err := migration.Run(cfg.DatabaseURL, log); err != nil {
			return err
		}
	}

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.New(pool)
	partyService := service.New(repo, log)
	handler := partyhttp.New(partyService)

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	partyHandler := partygrpc.New(partyService)
	partyv1.RegisterPartyServiceServer(grpcServer, partyHandler)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	server := &nethttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server started", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()
	go func() {
		log.Info("grpc server started", "addr", cfg.GRPCAddr)
		errCh <- grpcServer.Serve(grpcLis)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		grpcServer.GracefulStop()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, nethttp.ErrServerClosed) {
			return nil
		}
		return err
	}
}
