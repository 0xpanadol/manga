package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xpanadol/manga/internal/config"
	"github.com/0xpanadol/manga/internal/service"
	"github.com/0xpanadol/manga/pkg/broker"
	"github.com/0xpanadol/manga/pkg/email"
	"github.com/0xpanadol/manga/pkg/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load or validate config: %v", err)
	}

	appLogger := logger.New(cfg.AppEnv)
	defer appLogger.Sync()

	appLogger.Info("Worker service starting...")

	messageBroker, err := broker.NewRabbitMQBroker(cfg.RabbitMQUrl)
	if err != nil {
		appLogger.Fatal("Could not initialize message broker", zap.Error(err))
	}
	defer messageBroker.Close()
	appLogger.Info("Message broker connected")

	// === Initialize Email Sender ===
	emailSender := email.NewSMTPSender(
		cfg.SmtpHost,
		cfg.SmtpPort,
		cfg.SmtpUsername,
		cfg.SmtpPassword,
		cfg.SmtpSender,
	)

	// Define the handler for user.registered messages
	userRegisteredHandler := func(d amqp091.Delivery) {
		appLogger.Info("Received a message", zap.String("body", string(d.Body)))

		var payload service.UserRegisteredPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			appLogger.Error("Failed to unmarshal message body", zap.Error(err))
			// We nack (negative-acknowledge) the message and tell RabbitMQ not to requeue it.
			// This moves it to a dead-letter queue if one is configured.
			d.Nack(false, false)
			return
		}

		// "Process" the job. In the future, this would be sending an email.
		appLogger.Info("Processing user.registered event",
			zap.String("user_id", payload.UserID),
			zap.String("email", payload.Email),
		)

		// Acknowledge the message to remove it from the queue.
		d.Ack(false)
	}

	// === Handler for password reset messages ===
	passwordResetHandler := func(d amqp091.Delivery) {
		appLogger.Info("Received a password.reset.requested message", zap.String("body", string(d.Body)))

		var payload service.PasswordResetRequestedPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			appLogger.Error("Failed to unmarshal message body", zap.Error(err))
			d.Nack(false, false)
			return
		}

		// Construct the email
		subject := "Your Password Reset Request"
		// In a real app, this link would point to your frontend.
		resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", payload.Token)
		body := fmt.Sprintf("Hi there,<br><br>Please use the following link to reset your password: <a href='%s'>%s</a><br><br>This link will expire in 15 minutes.", resetLink, resetLink)

		// Send the email
		if err := emailSender.SendEmail(payload.Email, subject, body); err != nil {
			appLogger.Error("Failed to send password reset email", zap.Error(err), zap.String("recipient", payload.Email))
			d.Nack(false, false) // Nack to indicate failure
			return
		}

		appLogger.Info("Successfully sent password reset email", zap.String("recipient", payload.Email))
		d.Ack(false)
	}

	// Start consumers for both queues
	if err := messageBroker.Consume("user.registered", userRegisteredHandler); err != nil {
		appLogger.Fatal("Failed to start user.registered consumer", zap.Error(err))
	}
	if err := messageBroker.Consume("password.reset.requested", passwordResetHandler); err != nil {
		appLogger.Fatal("Failed to start password.reset.requested consumer", zap.Error(err))
	}

	// Wait for termination signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Worker service shutting down...")
}
