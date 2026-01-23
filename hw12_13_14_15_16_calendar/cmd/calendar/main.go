package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalgrpc "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/grpc"     //nolint:depguard
	internalhttp "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/http"     //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage"                      //nolint:depguard
	memorystorage "github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
	sqlstorage "github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage/sql"       //nolint:depguard
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to unmarshal configuration: %v", err)
		os.Exit(1)
	}

	logg, err := logger.New(cfg.Logger.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to init logger: %v", err)
		os.Exit(1)
	}

	var storage storage.Storage
	switch cfg.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "sql":
		storage, err = sqlstorage.New(cfg.SQL.DSN)
		if err != nil {
			logg.Error("failed to init sql storage: " + err.Error())
			os.Exit(1)
		}
	}

	calendar := app.New(logg, storage)

	httpServer := internalhttp.NewServer(logg, calendar, cfg.Server.HTTPServer.Host, cfg.Server.HTTPServer.Port)
	grpcServer := internalgrpc.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	errCh := make(chan error, 2)

	logg.Info("calendar is running...")

	go func() {
		errCh <- grpcServer.Start(net.JoinHostPort("", cfg.Server.GRPCServer.Port))
	}()

	go func() {
		errCh <- httpServer.Start()
	}()

	select {
	case <-ctx.Done():
		logg.Info("shutdown signal received")
	case err := <-errCh:
		logg.Error("server error: " + err.Error())
		cancel()
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := httpServer.Stop(ctxShutdown); err != nil {
		logg.Error("failed to stop http server: " + err.Error())
	}

	grpcServer.Stop()

	if closer, ok := storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logg.Error("failed to close storage: " + err.Error())
		}
	}
}
