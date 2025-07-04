package service

import (
	"context"
	"strings"
	"time"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/pkg/jwtauth"
	"github.com/0xpanadol/manga/pkg/password"
)

type AuthService struct {
	userRepo         repository.UserRepository
	accessSecret     string
	refreshSecret    string
	accessExpiresIn  time.Duration
	refreshExpiresIn time.Duration
}

func NewAuthService(userRepo repository.UserRepository, accessSecret, refreshSecret string, accessExp, refreshExp time.Duration) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		accessSecret:     accessSecret,
		refreshSecret:    refreshSecret,
		accessExpiresIn:  accessExp,
		refreshExpiresIn: refreshExp,
	}
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
