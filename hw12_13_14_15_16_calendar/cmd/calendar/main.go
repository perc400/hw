package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalgrpc "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/grpc"     //nolint:depguard
	internalhttp "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/http"     //nolint:depguard
	memorystorage "github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
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

	storage := memorystorage.New()
	calendar := app.New(logg, storage)

	httpServer := internalhttp.NewServer(logg, calendar, cfg.Server.HTTPServer.Host, cfg.Server.HTTPServer.Port)
	grpcServer := internalgrpc.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpServer.Stop(ctxShutdown); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
		grpcServer.Stop()
	}()

	logg.Info("calendar is running...")

	go func() {
		if err := grpcServer.Start(cfg.Server.GRPCServer.Port); err != nil {
			logg.Error("failed to serve grpc server: " + err.Error())
		}
	}()

	if err := httpServer.Start(); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		os.Exit(1) //nolint:gocritic
	}
}
