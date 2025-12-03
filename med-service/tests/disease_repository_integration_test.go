//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"testing"

	"igaku/med-service/repositories"
	igakuErrors "igaku/med-service/errors"
	testUtils "igaku/med-service/tests/utils"
)

func TestGormDiseaseRepository(t *testing.T) {
	t.Run("FindByName_Success", func(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, cleanup := testUtils.SetupTestDatabase(ctx, t)
	defer cleanup()

	repo := repositories.NewGormDiseaseRepository(db)

	targetID1, err := uuid.Parse("32f5c8d5-9cb0-4b1a-b900-ad2aa78f3a19")
	require.NoError(t, err, "Failed to parse first target UUID")

	targetID2, err := uuid.Parse("ecdfed12-066d-4c45-837f-f159ca19ee22")
	require.NoError(t, err, "Failed to parse second target UUID")

	targetRxNormID1 := "D008177"
	targetRxNormID2 := "D008181"
	targetName1 := "Lupus Vulgaris"
	targetName2 := "Lupus Nephritis"
	targetResultsNum := 2

	name := "Lupus"
	diseases, err := repo.FindByName(name)

	assert.NoError(
		t, err, "Expected no error finding diseases",
	)
	assert.NotNil(t, diseases, "Expected diseases to be found")
	assert.Equal(
		t, targetResultsNum, len(diseases),
	)
	assert.Equal(
		t, targetID1, diseases[0].ID,
		"Expected first ID to match",
	)
	assert.Equal(
		t, targetRxNormID1, diseases[0].RxNormID,
		"Expected first RxNormID to match",
	)
	assert.Equal(
		t, targetName1, diseases[0].Name,
		"Expected first disease name to match",
	)

	assert.Equal(
		t, targetID2, diseases[1].ID,
		"Expected second ID to match",
	)
	assert.Equal(
		t, targetRxNormID2, diseases[1].RxNormID,
		"Expected second RxNormID to match",
	)
	assert.Equal(
		t, targetName2, diseases[1].Name,
		"Expected second disease name to match",
	)
	})

	t.Run("FindByName_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormDiseaseRepository(db)

		nonExistentName := "Wilson"
		diseases, err := repo.FindByName(nonExistentName)

		assert.Error(
			t, err, "Expected an error when finding non-existent disease",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.DiseaseNotFoundError{}),
			"Expected DiseaseNotFoundError",
		)
		assert.Nil(
			t, diseases,
			"Expected diseases to be nil when not found",
		)
	})
}
