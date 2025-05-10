//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"testing"

	igakuErrors "igaku/user-service/errors"
	"igaku/user-service/repositories"
	testUtils "igaku/user-service/tests/utils"
)

func TestGormOrganizationRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormOrganizationRepository(db)

		targetID, err := uuid.Parse("86e6a1f3-d7aa-4e74-a20a-ea78bc13340b")
		require.NoError(t, err, "Failed to parse target UUID")

		org, err := repo.FindByID(targetID)

		assert.NoError(
			t, err, "Expected no error finding existing organization",
		)
		require.NotNil(t, org, "Expected organization to be found")
		assert.Equal(
			t, targetID, org.ID, "Expected organization ID to match",
		)
		assert.Equal(
			t, "The Lowell General Hospital", org.Name,
			"Expected organization name to match",
		)
		assert.Equal(
			t, "295 Varnum Ave", org.Address,
			"Expected organization address to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormOrganizationRepository(db)

		nonExistentID := uuid.New()

		org, err := repo.FindByID(nonExistentID)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent organization",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.OrganizationNotFoundError{}),
			"Expected OrganizationNotFoundError",
		)
		assert.Nil(
			t, org,
			"Expected organization to be nil when not found",
		)
	})
}
