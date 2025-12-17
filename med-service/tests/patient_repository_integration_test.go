//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"igaku/commons/models"
	medErrors "igaku/med-service/errors"
	"igaku/med-service/repositories"
	testUtils "igaku/med-service/tests/utils"
)

func TestGormPatientRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormPatientRepository(db)

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

		repo := repositories.NewGormPatientRepository(db)

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
	t.Run("ValidateUniquePatient_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormPatientRepository(db)

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

		repo := repositories.NewGormPatientRepository(db)

		patientID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "44051401458"
		record := &models.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.Error(t, err, "Expected error when patient ID is duplicated")
		log.Println(err.Error())
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})
	t.Run("ValidateUniquePatient_DuplicatedNationalID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormPatientRepository(db)

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

		repo := repositories.NewGormPatientRepository(db)

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

		repo := repositories.NewGormPatientRepository(db)

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
