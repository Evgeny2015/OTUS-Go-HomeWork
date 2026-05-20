package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/app"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/config"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/logger"
	memorystorage "github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/rmq"
	sqlstorage "github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.yml", "Path to configuration file")
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadSchedulerConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logg := logger.NewFromConfig(logger.Config{
		Level:  cfg.Logger.Level,
		Output: cfg.Logger.Output,
		Format: cfg.Logger.Format,
	})

	// Initialize storage
	var storage app.Storage
	switch cfg.Storage.Type {
	case "memory":
		storage = memorystorage.New()
		logg.Info("Using memory storage")
	case "sql":
		sqlStorage, err := sqlstorage.NewWithDSN(cfg.Storage.DSN)
		if err != nil {
			logg.Error(fmt.Sprintf("Failed to create SQL storage: %v", err))
			os.Exit(1)
		}
		storage = sqlStorage
		logg.Info("Using SQL storage")
	default:
		logg.Error(fmt.Sprintf("Unknown storage type: %s", cfg.Storage.Type))
		os.Exit(1)
	}

	// Initialize RabbitMQ
	queue := rmq.NewRabbitMQFromConfig(&cfg.RabbitMQ)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to RabbitMQ
	if err := queue.Connect(ctx); err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		os.Exit(1)
	}
	defer queue.Close()

	logg.Info("Scheduler started")

	// Parse durations from config
	interval, err := time.ParseDuration(cfg.Scheduler.Interval)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to parse interval duration: %v", err))
		os.Exit(1)
	}

	lookAhead, err := time.ParseDuration(cfg.Scheduler.LookAhead)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to parse look ahead duration: %v", err))
		os.Exit(1)
	}

	cleanupOlder, err := time.ParseDuration(cfg.Scheduler.CleanupOlder)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to parse cleanup older duration: %v", err))
		os.Exit(1)
	}

	// Create ticker for periodic execution
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run initial check
	if err := processNotifications(ctx, storage, queue, lookAhead, cleanupOlder, logg); err != nil {
		logg.Error(fmt.Sprintf("Initial check failed: %v", err))
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			if err := processNotifications(ctx, storage, queue, lookAhead, cleanupOlder, logg); err != nil {
				logg.Error(fmt.Sprintf("Processing failed: %v", err))
			}
		case sig := <-sigChan:
			logg.Info(fmt.Sprintf("Received signal %v, shutting down", sig))
			return
		}
	}
}

func processNotifications(
	ctx context.Context,
	storage app.Storage,
	queue rmq.Queue,
	lookAhead time.Duration,
	cleanupOlder time.Duration,
	logg app.Logger,
) error {
	now := time.Now()
	fromTime := now
	toTime := now.Add(lookAhead)

	// Get events that need notifications
	events, err := storage.GetEventsForNotification(ctx, fromTime, toTime)
	if err != nil {
		return fmt.Errorf("failed to get events for notification: %w", err)
	}

	logg.Info(fmt.Sprintf("Found %d events needing notifications between %s and %s",
		len(events), fromTime.Format(time.RFC3339), toTime.Format(time.RFC3339)))

	// Create and publish notifications
	for _, event := range events {
		notification := internal.NewNotification(
			event.ID,
			event.Title,
			event.DateTime,
			event.UserID,
		)

		if err := queue.PublishNotification(ctx, notification); err != nil {
			logg.Error(fmt.Sprintf("Failed to publish notification for event %s: %v", event.ID, err))
			continue
		}
	}

	// Clean up old events (older than 1 year)
	olderThan := now.Add(-cleanupOlder)
	if err := storage.DeleteOldEvents(ctx, olderThan); err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	logg.Info(fmt.Sprintf("Cleaned up events older than %s", olderThan.Format(time.RFC3339)))
	return nil
}
