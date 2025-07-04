package postgres

import (
	"context"
	"testing"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	// The testPool is initialized in main_test.go
	repo := NewPostgresUserRepository(testPool)
	ctx := context.Background()

	// Get the "User" role ID that was seeded by the migrations
	defaultRoleID, err := repo.FindDefaultUserRoleID(ctx)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, defaultRoleID)

	user := &domain.User{
		Username:     "integration_test_user",
		Email:        "integration@test.com",
		PasswordHash: "hashed_password",
		RoleID:       defaultRoleID,
	}

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify the user was created
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())

	// Test for duplicate creation
	err = repo.Create(ctx, user)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrUserAlreadyExists, err)
}

func TestPostgresUserRepository_FindByEmail(t *testing.T) {
	repo := NewPostgresUserRepository(testPool)
	ctx := context.Background()

	// First, create a user to find
	defaultRoleID, err := repo.FindDefaultUserRoleID(ctx)
	require.NoError(t, err)
	userToCreate := &domain.User{
		Username:     "find_me",
		Email:        "findme@test.com",
		PasswordHash: "some_hash",
		RoleID:       defaultRoleID,
	}
	err = repo.Create(ctx, userToCreate)
	require.NoError(t, err)

	// Test finding the user
	foundUser, err := repo.FindByEmail(ctx, "findme@test.com")
	require.NoError(t, err)
	assert.Equal(t, userToCreate.ID, foundUser.ID)
	assert.Equal(t, "find_me", foundUser.Username)

	// Test finding a non-existent user
	_, err = repo.FindByEmail(ctx, "nonexistent@test.com")
	assert.Error(t, err)
	assert.Equal(t, repository.ErrUserNotFound, err)
}
