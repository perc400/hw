package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/logger"         //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/queue/rabbitmq" //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/sender"         //nolint:depguard
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

	consumer, err := rabbitmq.NewClient(cfg.AMQPClient.Queue.Name, cfg.AMQPClient.URL, logg)
	if err != nil {
		logg.Info(fmt.Sprintf("%+v", cfg))
		logg.Error("failed to init RabbitMQ client: " + err.Error())
		os.Exit(1)
	}

	snd := sender.NewSender(consumer, logg)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		errCh <- snd.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		logg.Info("shutting down...")
	case err := <-errCh:
		logg.Error("consumer error: " + err.Error())
	}

	if err := consumer.Close(); err != nil {
		logg.Error("failed to close rabbitmq: " + err.Error())
	}
}
