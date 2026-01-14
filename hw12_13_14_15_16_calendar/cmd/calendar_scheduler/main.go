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
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/queue/rabbitmq"               //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/scheduler"                    //nolint:depguard
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
		logg.Warn("scheduler is running with memory storage - events will be lost between restarts")
	case "sql":
		storage, err = sqlstorage.New(cfg.SQL.DSN)
		if err != nil {
			logg.Error("failed to init sql storage: " + err.Error())
			os.Exit(1)
		}
	}

	publisher, err := rabbitmq.NewClient(cfg.AMQPClient.Queue.Name, cfg.AMQPClient.URL, logg)
	if err != nil {
		logg.Info(fmt.Sprintf("%+v", cfg))
		logg.Error("failed to init RabbitMQ client: " + err.Error())
		os.Exit(1)
	}
	calendar := app.New(logg, storage)
	sched := scheduler.NewScheduler(logg, calendar, publisher, time.Duration(cfg.AMQPClient.PollInterval)*time.Second)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	go sched.Start(ctx)

	<-ctx.Done()
	logg.Info("shutting down scheduler...")

	if err := publisher.Close(); err != nil {
		logg.Error("failed to close rabbitmq: " + err.Error())
	}

	if closer, ok := storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logg.Error("failed to close storage: " + err.Error())
		}
	}
}
