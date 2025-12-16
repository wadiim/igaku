//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"testing"

	"igaku/visit-service/repositories"
	"igaku/visit-service/utils"
	igakuErrors "igaku/visit-service/errors"
	testUtils "igaku/commons/utils"
)

func TestGormOrganizationRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(
			ctx, t, utils.MigrateSchema,
		)
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
			t, "Massachusetts General Hospital", org.Name,
			"Expected organization name to match",
		)
		assert.Equal(
			t, int64(117853077), org.Location.ID,
			"Expected organiztion's location ID to match",
		)
		assert.Equal(
			t, "42.3628605", org.Location.Lat,
			"Expected organiztion's location latitude to match",
		)
		assert.Equal(
			t, "-71.0687530", org.Location.Lon,
			"Expected organiztion's location longitude to match",
		)
		assert.Equal(
			t,
			"Massachusetts General Hospital, 55, Fruit Street, West End, Boston, Suffolk County, Massachusetts, 02114, United States",
			org.Location.Name,
			"Expected organiztion's location name to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(
			ctx, t, utils.MigrateSchema,
		)
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
