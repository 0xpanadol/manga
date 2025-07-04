package service_test

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/0xpanadol/manga/internal/domain"
// 	"github.com/0xpanadol/manga/internal/repository"
// 	"github.com/0xpanadol/manga/internal/repository/mocks"
// 	"github.com/0xpanadol/manga/internal/service"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestAuthService_Register(t *testing.T) {
// 	mockUserRepo := new(mocks.UserRepository)
// 	authService := service.NewAuthService(mockUserRepo, "access", "refresh", time.Minute, time.Hour)

// 	ctx := context.Background()
// 	testUsername := "testuser"
// 	testEmail := "test@example.com"
// 	testPassword := "password123"
// 	testRoleID := uuid.New()

// 	// Setup expectations
// 	mockUserRepo.On("FindDefaultUserRoleID", ctx).Return(testRoleID, nil)
// 	// We expect the Create method to be called with any context and a user object.
// 	// We can add more specific argument matching if needed.
// 	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
// 		// We can inspect the arguments passed to the mock call
// 		userArg := args.Get(1).(*domain.User)
// 		assert.Equal(t, testUsername, userArg.Username)
// 		assert.Equal(t, testEmail, userArg.Email)
// 		assert.NotEmpty(t, userArg.PasswordHash)
// 		assert.Equal(t, testRoleID, userArg.RoleID)
// 	})

// 	// Execute the service method
// 	user, err := authService.Register(ctx, testUsername, testEmail, testPassword)

// 	// Assert results
// 	assert.NoError(t, err)
// 	assert.NotNil(t, user)
// 	mockUserRepo.AssertExpectations(t) // Verify that all expected calls were made
// }

// func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
// 	mockUserRepo := new(mocks.UserRepository)
// 	authService := service.NewAuthService(mockUserRepo, "access", "refresh", time.Minute, time.Hour)

// 	ctx := context.Background()
// 	testRoleID := uuid.New()

// 	// Setup expectations
// 	mockUserRepo.On("FindDefaultUserRoleID", ctx).Return(testRoleID, nil)
// 	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(repository.ErrUserAlreadyExists)

// 	// Execute
// 	_, err := authService.Register(ctx, "testuser", "test@example.com", "password123")

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, repository.ErrUserAlreadyExists, err)
// 	mockUserRepo.AssertExpectations(t)
// }
