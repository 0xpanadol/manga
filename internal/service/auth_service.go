package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/pkg/broker"
	"github.com/0xpanadol/manga/pkg/jwtauth"
	"github.com/0xpanadol/manga/pkg/password"
	"github.com/0xpanadol/manga/pkg/token"
)

type AuthService struct {
	userRepo         repository.UserRepository
	broker           *broker.RabbitMQBroker
	accessSecret     string
	refreshSecret    string
	accessExpiresIn  time.Duration
	refreshExpiresIn time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	broker *broker.RabbitMQBroker,
	accessSecret,
	refreshSecret string,
	accessExp,
	refreshExp time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		broker:           broker,
		accessSecret:     accessSecret,
		refreshSecret:    refreshSecret,
		accessExpiresIn:  accessExp,
		refreshExpiresIn: refreshExp,
	}
}

type UserRegisteredPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

func (s *AuthService) Register(ctx context.Context, username, email, plainPassword string) (*domain.User, error) {
	// Hash the password
	hashedPassword, err := password.Hash(plainPassword)
	if err != nil {
		return nil, err
	}

	// Get the default role ID for new users ("User")
	roleID, err := s.userRepo.FindDefaultUserRoleID(ctx)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     strings.TrimSpace(username),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hashedPassword,
		RoleID:       roleID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// === Publish event to RabbitMQ ===
	payload := UserRegisteredPayload{
		UserID:    user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	// We publish this asynchronously. If it fails, we log it but don't fail the registration.
	go func() {
		if err := s.broker.Publish(context.Background(), "user.registered", payload); err != nil {
			// In a real app, you'd use your structured logger here.
			log.Printf("Failed to publish user.registered event for user %s: %v", user.ID, err)
		}
	}()

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, plainPassword string) (*jwtauth.TokenDetails, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err // Can be ErrUserNotFound
	}

	// Verify password
	if !password.Verify(plainPassword, user.PasswordHash) {
		return nil, repository.ErrUserNotFound // Use the same error to prevent account enumeration
	}

	// Get user's role and permissions
	role, err := s.userRepo.GetRoleAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Generate JWTs
	tokens, err := jwtauth.GenerateTokens(user.ID, role.Name, role.Permissions, s.accessSecret, s.refreshSecret, s.accessExpiresIn, s.refreshExpiresIn)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

type PasswordResetRequestedPayload struct {
	Email     string `json:"email"`
	Token     string `json:"token"` // The plain, un-hashed token
	Timestamp string `json:"timestamp"`
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// IMPORTANT: Do not reveal if the user exists or not.
		// We return nil to indicate success to the user, even if the email is not found.
		// The actual work (email sending) is async, so the user sees a consistent response.
		return nil
	}

	// Generate a secure token for the user.
	resetToken, err := token.GenerateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Hash the token for database storage.
	tokenHash := token.HashToken(resetToken)

	// Set an expiry time for the token.
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store the hashed token in the database.
	if err := s.userRepo.CreatePasswordResetToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return err
	}

	// Publish an event to the message broker to send the email.
	payload := PasswordResetRequestedPayload{
		Email:     user.Email,
		Token:     resetToken, // Send the plain token to the user
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	go func() {
		if err := s.broker.Publish(context.Background(), "password.reset.requested", payload); err != nil {
			log.Printf("Failed to publish password.reset.requested event for user %s: %v", user.ID, err)
		}
	}()

	return nil
}
