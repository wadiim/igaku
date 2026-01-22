//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"igaku/commons/models"
	"igaku/med-service/repositories"
	medErrors "igaku/med-service/errors"
	testUtils "igaku/med-service/tests/utils"
)

func TestGormMedRepository(t *testing.T) {
	t.Run("FindBySubstring_LowercaseName", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID1, err := uuid.Parse("ebb58b3c-4356-4564-bd01-ddd495927025")
		require.NoError(t, err, "Failed to parse first target UUID")

		targetID2, err := uuid.Parse("ff99edf6-b4c3-4134-9805-61defafa0b62")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetID3, err := uuid.Parse("0d8209f8-a04d-493d-a162-50878a8ee5c0")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetDiseases := []models.Disease {
			{ID: targetID1, RxNormID: "D011014", Name: "Pneumonia"},
			{ID: targetID2, RxNormID: "D011002", Name: "Pleuropneumonia"},
			{ID: targetID3, RxNormID: "D018549", Name: "Cryptogenic Organizing Pneumonia"},
		}

		name := "pneumonia"
		count := 3
		resultDiseases, err := repo.FindBySubstring(name, 0, count)

		assert.NoError(t, err, "Expected no error finding diseases")
		assert.NotNil(t, resultDiseases, "Expected diseases to be found")
		assert.Equal(t, count, len(resultDiseases))

		for i, disease := range targetDiseases {
			assert.Equal(
				t, resultDiseases[i].ID, disease.ID,
				fmt.Sprintf("Expected %d ID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].RxNormID, disease.RxNormID,
				fmt.Sprintf("Expected %d RxNormID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].Name, disease.Name,
				fmt.Sprintf("Expected %d disease name to match", i+1),
			)
		}
	})

	t.Run("FindBySubstring_UppercaseName", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID1, err := uuid.Parse("ebb58b3c-4356-4564-bd01-ddd495927025")
		require.NoError(t, err, "Failed to parse first target UUID")

		targetID2, err := uuid.Parse("ff99edf6-b4c3-4134-9805-61defafa0b62")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetID3, err := uuid.Parse("0d8209f8-a04d-493d-a162-50878a8ee5c0")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetDiseases := []models.Disease {
			{ID: targetID1, RxNormID: "D011014", Name: "Pneumonia"},
			{ID: targetID2, RxNormID: "D011002", Name: "Pleuropneumonia"},
			{ID: targetID3, RxNormID: "D018549", Name: "Cryptogenic Organizing Pneumonia"},
		}

		name := "Pneumonia"
		count := 3
		resultDiseases, err := repo.FindBySubstring(name, 0, count)

		assert.NoError(t, err, "Expected no error finding diseases")
		assert.NotNil(t, resultDiseases, "Expected diseases to be found")
		assert.Equal(t, count, len(resultDiseases))

		for i, disease := range targetDiseases {
			assert.Equal(
				t, resultDiseases[i].ID, disease.ID,
				fmt.Sprintf("Expected %d ID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].RxNormID, disease.RxNormID,
				fmt.Sprintf("Expected %d RxNormID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].Name, disease.Name,
				fmt.Sprintf("Expected %d disease name to match", i+1),
			)
		}
	})

	t.Run("FindBySubstring_CountLessThanLimit", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID1, err := uuid.Parse("32f5c8d5-9cb0-4b1a-b900-ad2aa78f3a19")
		require.NoError(t, err, "Failed to parse first target UUID")

		targetID2, err := uuid.Parse("4140b999-d05b-46eb-a83b-6ff1f06a9eda")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetDiseases := []models.Disease {
			{ID: targetID1, RxNormID: "D008177", Name: "Lupus Vulgaris"},
			{ID: targetID2, RxNormID: "D008179", Name: "Panniculitis, Lupus Erythematosus"},
		}

		name := "Lupus"
		count := 2
		limit := count + 1
		resultDiseases, err := repo.FindBySubstring(name, 0, limit)

		assert.NoError(t, err, "Expected no error finding diseases")
		assert.NotNil(t, resultDiseases, "Expected diseases to be found")
		assert.Equal(t, count, len(resultDiseases))

		for i, disease := range targetDiseases {
			assert.Equal(
				t, resultDiseases[i].ID, disease.ID,
				fmt.Sprintf("Expected %d ID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].RxNormID, disease.RxNormID,
				fmt.Sprintf("Expected %d RxNormID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].Name, disease.Name,
				fmt.Sprintf("Expected %d disease name to match", i+1),
			)
		}
	})

	t.Run("FindBySubstring_CountMoreThanLimit", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID1, err := uuid.Parse("ebb58b3c-4356-4564-bd01-ddd495927025")
		require.NoError(t, err, "Failed to parse first target UUID")

		targetID2, err := uuid.Parse("ff99edf6-b4c3-4134-9805-61defafa0b62")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetID3, err := uuid.Parse("0d8209f8-a04d-493d-a162-50878a8ee5c0")
		require.NoError(t, err, "Failed to parse second target UUID")

		targetDiseases := []models.Disease {
			{ID: targetID1, RxNormID: "D011014", Name: "Pneumonia"},
			{ID: targetID2, RxNormID: "D011002", Name: "Pleuropneumonia"},
			{ID: targetID3, RxNormID: "D018549", Name: "Cryptogenic Organizing Pneumonia"},
		}

		name := "Pneumonia"
		count := 3
		limit := count - 1
		resultDiseases, err := repo.FindBySubstring(name, 0, limit)

		assert.NoError(t, err, "Expected no error finding diseases")
		assert.NotNil(t, resultDiseases, "Expected diseases to be found")
		assert.Equal(t, limit, len(resultDiseases))

		for i, disease := range targetDiseases[:2] {
			assert.Equal(
				t, resultDiseases[i].ID, disease.ID,
				fmt.Sprintf("Expected %d ID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].RxNormID, disease.RxNormID,
				fmt.Sprintf("Expected %d RxNormID to match", i+1),
			)
			assert.Equal(
				t, resultDiseases[i].Name, disease.Name,
				fmt.Sprintf("Expected %d disease name to match", i+1),
			)
		}
	})

	t.Run("FindBySubstring_OffsetMoreThanCount", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		name := "Pneumonia"
		offset := 5
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when not found")
		assert.Contains(t, strings.ToLower(err.Error()), "disease not found")
	})

	t.Run("FindBySubstring_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		name := "Wilson"
		offset := 5
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when not found")
		assert.Contains(t, strings.ToLower(err.Error()), "disease not found")
	})

	t.Run("FindBySubstring_OffsetNegative", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		name := "Pneumonia"
		offset := -1
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when offset is negative")
		assert.Contains(t, strings.ToLower(err.Error()), "offset must be positive")
	})

	t.Run("FindBySubstring_LimitNegative", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		name := "Pneumonia"
		offset := 1
		limit := -3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when limit is negative")
		assert.Contains(t, strings.ToLower(err.Error()), "limit must be positive")
	})

	t.Run("CountBySubstring_LowercaseName", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		expectedCount := int64(4)
		name := "pneumonia"
		count, err := repo.CountBySubstring(name)

		assert.NotNil(t, count, "Expected no error counting diseases")
		assert.NoError(t, err, "Expected no error counting diseases")
		assert.Equal(t, expectedCount, count)
	})

	t.Run("CountBySubstring_UppercaseName", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		expectedCount := int64(4)
		name := "Pneumonia"
		count, err := repo.CountBySubstring(name)

		assert.NotNil(t, count, "Expected no error counting diseases")
		assert.NoError(t, err, "Expected no error counting diseases")
		assert.Equal(t, expectedCount, count)
	})

	t.Run("CountBySubstring_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		expectedCount := int64(0)
		name := "Wilson"
		count, err := repo.CountBySubstring(name)

		assert.Error(t, err, "Expected an error finding a non-existent disease")
		assert.Equal(t, expectedCount, count)
		assert.True(
			t, errors.Is(err, &medErrors.DiseaseNotFoundError{}),
			"Expected DiseaseNotFoundError",
		)
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failet to parse target UUID")

		patient, err := repo.FindByID(targetID)

		assert.NoError(
			t, err, "Expected no error finding existing patient",
		)
		assert.NotNil(t, patient, "Expected patient to be found")
		assert.Equal(
			t, targetID, patient.ID, "Expected patient ID to match",
		)
		assert.Equal(
			t, "12345123451", patient.NationalID, "Expected patient NationalID to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID, err := uuid.Parse("64814b64-d138-4209-b869-3b649db06ab1")
		require.NoError(t, err, "Failed to parse target UUID")

		patient, err := repo.FindByID(targetID)

		assert.Error(t, err, "Expected an error when finding non-existent patient")
		assert.True(
			t, errors.Is(err, &medErrors.PatientNotFoundError{}),
			"Expected PatientNotFoundError",
		)
		assert.Nil(
			t, patient,
			"Expected patient to be nil when not found",
		)
	})

	t.Run("FindByNationalID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse patient UUID")
		targetNationalID := "12345123451"

		patient, err := repo.FindByNationalID(targetNationalID)

		assert.NoError(t, err, "Expected no error finding patient")
		assert.NotNil(t, patient, "Expected patient to be found")
		assert.Equal(
			t, targetID, patient.ID, "Expected patient ID to match",
		)
		assert.Equal(
			t, targetNationalID, patient.NationalID, "Expected patient NationalID to match",
		)
	})

	t.Run("FindByNationalID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetNationalID := "44051401458"

		patient, err := repo.FindByNationalID(targetNationalID)

		assert.Error(t, err, "Expected an error when finding non-existent patient")
		assert.True(
			t, errors.Is(err, &medErrors.PatientNotFoundError{}),
			"Expected PatientNotFoundError",
		)
		assert.Nil(
			t, patient,
			"Expected patient to be nil when not found",
		)
	})

	t.Run("ValidateUniquePatient_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		patientID, err := uuid.Parse("64814b64-d138-4209-b869-3b649db06ab1")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "44051401458"
		record := &models.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.NoError(t, err, "Expected no error validating patient's uniqueness")
	})

	t.Run("ValidateUniquePatient_DuplicatedID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		patientID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "44051401458"
		record := &models.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.Error(t, err, "Expected error when patient ID is duplicated")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("ValidateUniquePatient_DuplicatedNationalID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		patientID, err := uuid.Parse("64814b64-d138-4209-b869-3b649db06ab1")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "12345654321"
		record := &models.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.Error(t, err, "Expected error when patient NationalID is duplicated")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("AddPatient_InvalidPatient", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		// Duplicate ID
		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse target UUID")

		patient := models.PatientRecord{
			ID: targetID,
			NationalID: "11223311223",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")

		// Duplicate NationalID
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = models.PatientRecord{
			ID: targetID,
			NationalID: "12345123451",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")

		// Invalid NationalID (letters)
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = models.PatientRecord{
			ID: targetID,
			NationalID: "abc45123451",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid id")

		// Invalid NationalID (length)
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = models.PatientRecord{
			ID: targetID,
			NationalID: "1234",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid id")
	})

	t.Run("AddPatient_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormMedRepository(db)

		targetID, err := uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient := models.PatientRecord{
			ID: targetID,
			NationalID: "12312312312",
		}

		err = repo.AddPatient(&patient)
		assert.NoError(t, err, "Expected no error adding patient")

		result, err := repo.FindByID(targetID)
		assert.NotNil(t, result, "Expected patient to be found")
		assert.NoError(t, err, "Expected no error finding patient")
		assert.Equal(
			t, patient.ID, result.ID, "Expected patient ID to match",
		)
		assert.Equal(
			t, patient.NationalID, result.NationalID, "Expected patient ID to match",
		)
	})
}
