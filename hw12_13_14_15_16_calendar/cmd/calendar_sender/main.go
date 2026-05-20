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
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/config"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/logger"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage/rmq"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/sender_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := config.LoadSenderConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logg := logger.New(cfg.Logger.Level)

	logg.Info("Starting calendar sender service")
	logg.Info(fmt.Sprintf("Config: %+v", cfg))

	// Create RabbitMQ client
	queue := rmq.NewRabbitMQFromConfig(&cfg.RabbitMQ)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to RabbitMQ
	if err := queue.Connect(ctx); err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		os.Exit(1)
	}
	defer queue.Close()

	logg.Info("Connected to RabbitMQ")

	// Start consuming notifications
	notifications, err := queue.ConsumeNotifications(ctx)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to start consuming notifications: %v", err))
		os.Exit(1)
	}

	logg.Info("Started consuming notifications from queue")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Process notifications
	for {
		select {
		case <-sigChan:
			logg.Info("Received shutdown signal")
			cancel()
			time.Sleep(1 * time.Second) // Give some time for cleanup
			logg.Info("Sender service stopped")
			return
		case notification, ok := <-notifications:
			if !ok {
				logg.Info("Notification channel closed")
				return
			}
			// Log the notification (instead of actually sending)
			logg.Info(fmt.Sprintf("Notification processed successfully: EventID=%s, Title='%s', Date=%s, UserID=%s",
				notification.EventID,
				notification.EventTitle,
				notification.EventDate.Format(time.RFC3339),
				notification.UserID))

			// Publish notification status
			status := internal.NewNotificationStatus(
				notification.EventID,
				"processed",
				notification.UserID,
				time.Now(),
			).WithEventDetails(notification.EventTitle, notification.EventDate)
			if err := queue.PublishNotificationStatus(ctx, status); err != nil {
				logg.Error(fmt.Sprintf("Failed to publish notification status: %v", err))
			}
		}
	}
}
