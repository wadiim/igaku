//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"testing"

	igakuErrors "igaku/errors"
	"igaku/models"
	"igaku/repositories"
	testUtils "igaku/tests/utils"
)

func TestGormUserRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failet to parse target UUID")

		usr, err := repo.FindByID(targetID)

		assert.NoError(
			t, err, "Expected no error finding existing user",
		)
		require.NotNil(t, usr, "Expected user to be found")
		assert.Equal(
			t, targetID, usr.ID, "Expected user ID to match",
		)
		assert.Equal(
			t, "jdoe", usr.Username,
			"Expected username to match",
		)
		assert.Equal(
			t,
			"$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			usr.Password,
			"Expected user password hash to match",
		)
		assert.Equal(
			t, models.Patient, usr.Role,
			"Expected user role to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		nonExistentID := uuid.New()

		usr, err := repo.FindByID(nonExistentID)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent user",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.UserNotFoundError{}),
			"Expected UserNotFoundError",
		)
		assert.Nil(
			t, usr,
			"Expected user to be nil when not found",
		)
	})

	t.Run("FindByUsername_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		targetUsername := "jdoe"

		usr, err := repo.FindByUsername(targetUsername)

		assert.NoError(
			t, err, "Expected no error finding existing user",
		)
		require.NotNil(t, usr, "Expected user to be found")
		assert.Equal(
			t, targetUsername, usr.Username,
			"Expected username to match",
		)
		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failet to parse target UUID")
		assert.Equal(
			t, targetID, usr.ID,
			"Expected user ID to match",
		)
		assert.Equal(
			t,
			"$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			usr.Password,
			"Expected user password hash to match",
		)
		assert.Equal(
			t, models.Patient, usr.Role,
			"Expected user role to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		nonExistentUsername := "nonExistentUserName"

		usr, err := repo.FindByUsername(nonExistentUsername)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent user",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.UserNotFoundError{}),
			"Expected UserNotFoundError",
		)
		assert.Nil(
			t, usr,
			"Expected user to be nil when not found",
		)
	})
}
