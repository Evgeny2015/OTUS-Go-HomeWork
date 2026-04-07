package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/app"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/config"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func createStorage(conf config.StorageConf, logg *logger.Logger) (app.Storage, error) {
	switch conf.Type {
	case "memory":
		return memorystorage.New(), nil
	case "sql":
		if conf.DSN == "" {
			// Use default DSN from sqlstorage.Connect (for backward compatibility)
			// but we can also error out.
			logg.Warning("DSN not provided, using default PostgreSQL connection")
			storage := sqlstorage.New()
			ctx := context.Background()
			if err := storage.Connect(ctx); err != nil {
				return nil, fmt.Errorf("failed to connect to database: %w", err)
			}
			return storage, nil
		}
		return sqlstorage.NewWithDSN(conf.DSN)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", conf.Type)
	}
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg := config.LoadConfigFromDefault(configFile)
	logg := logger.NewFromConfig(logger.Config{
		Level:  cfg.Logger.Level,
		Output: cfg.Logger.Output,
		Format: cfg.Logger.Format,
	})

	storage, err := createStorage(cfg.Storage, logg)
	if err != nil {
		logg.Error("failed to create storage: " + err.Error())
		os.Exit(1)
	}
	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar, cfg.HTTP.Host, cfg.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
