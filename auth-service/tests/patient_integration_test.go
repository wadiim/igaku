//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// "context"
	// "log"
	// "os"
	"errors"
	// "log"
	"strings"
	"testing"
	// "time"
	// "fmt"

	"igaku/auth-service/clients"
	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
	// testUtils "igaku/auth-service/tests/utils"
)

// func TestMain(m *testing.M) {
// 	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Minute)
// 	defer cancelCtx()

// 	cleanup, err := testUtils.SetupTestServices(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to setup test environment: %v", err)
// 	}

// 	exitCode := m.Run()

// 	if cleanup != nil {
// 		cleanup()
// 	}

// 	os.Exit(exitCode)
// }

func TestPatientClient(t *testing.T) {
	t.Run("TestAddPatientRecord_DuplicatedPatientID", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		patientClient, err := clients.NewPatientClient(url)
		require.NoError(t, err)
		defer patientClient.Shutdown()

		id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err)
		nationalID := "44051401458"
		record := &models.PatientRecord{
			ID: id,
			NationalID: nationalID,
		}

		err = patientClient.AddPatientRecord(record)
		var dupErr *commonsErrors.DuplicatedIDError
		assert.True(
			t, errors.As(err, &dupErr),
			"Expected DuplicatedIDError",
		)
		assert.Equal(t, id, dupErr.ID)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("TestAddPatientRecord_DuplicatedPatientNationalID", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		patientClient, err := clients.NewPatientClient(url)
		require.NoError(t, err)
		defer patientClient.Shutdown()

		id, err := uuid.Parse("d93488f1-789d-4e2c-a962-5cb8b5d8e370")
		require.NoError(t, err)
		nationalID := "44051401458"
		record := &models.PatientRecord{
			ID: id,
			NationalID: nationalID,
		}

		err = patientClient.AddPatientRecord(record)
		var dupErr *commonsErrors.DuplicatedNationalIDError
		assert.True(
			t, errors.As(err, &dupErr),
			"Expected DuplicatedNationalIDError",
		)
		assert.Equal(t, nationalID, dupErr.NationalID)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("TestAddPatientRecord_InvalidPatientNationalID", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		patientClient, err := clients.NewPatientClient(url)
		require.NoError(t, err)
		defer patientClient.Shutdown()

		id, err := uuid.Parse("d93488f1-789d-4e2c-a962-5cb8b5d8e370")
		require.NoError(t, err)
		nationalID := "111"
		record := &models.PatientRecord{
			ID: id,
			NationalID: nationalID,
		}

		err = patientClient.AddPatientRecord(record)
		var dupErr *commonsErrors.InvalidNationalIDError
		assert.True(
			t, errors.As(err, &dupErr),
			"Expected InvalidNationalIDError",
		)
		assert.Equal(t, nationalID, dupErr.NationalID)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid id")
	})

	t.Run("TestValidateUniquePatient_DuplicatedPatientID", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		patientClient, err := clients.NewPatientClient(url)
		require.NoError(t, err)
		defer patientClient.Shutdown()

		id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err)
		nationalID := "44051401458"
		record := &models.PatientRecord{
			ID: id,
			NationalID: nationalID,
		}

		err = patientClient.ValidateUniquePatient(record)
		var dupErr *commonsErrors.DuplicatedIDError
		assert.True(
			t, errors.As(err, &dupErr),
			"Expected DuplicatedIDError",
		)
		assert.Equal(t, id, dupErr.ID)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("TestValidateUniquePatient_DuplicatedPatientNationalID", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		patientClient, err := clients.NewPatientClient(url)
		require.NoError(t, err)
		defer patientClient.Shutdown()

		id, err := uuid.Parse("d93488f1-789d-4e2c-a962-5cb8b5d8e370")
		require.NoError(t, err)
		nationalID := "44051401458"
		record := &models.PatientRecord{
			ID: id,
			NationalID: nationalID,
		}

		err = patientClient.ValidateUniquePatient(record)
		var dupErr *commonsErrors.DuplicatedNationalIDError
		assert.True(
			t, errors.As(err, &dupErr),
			"Expected DuplicatedNationalIDError",
		)
		assert.Equal(t, nationalID, dupErr.NationalID)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})
}
