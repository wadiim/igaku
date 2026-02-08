//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	commonsModels "igaku/commons/models"
	medErrors "igaku/med-service/errors"
	"igaku/med-service/models"
	"igaku/med-service/repositories"
	testUtils "igaku/med-service/tests/utils"
)

func initRepo(t *testing.T) (repo repositories.MedRepository, cleanup func()) {
	t.Parallel()
	ctx := context.Background()
	db, cleanup := testUtils.SetupTestDatabase(ctx, t)

	return repositories.NewGormMedRepository(db), cleanup
}

func TestGormMedRepository(t *testing.T) {
	t.Run("FindBySubstring_LowercaseName", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "Pneumonia"
		offset := 5
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when not found")
		assert.Contains(t, strings.ToLower(err.Error()), "disease not found")
	})

	t.Run("FindBySubstring_NotFound", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "Wilson"
		offset := 5
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when not found")
		assert.Contains(t, strings.ToLower(err.Error()), "disease not found")
	})

	t.Run("FindBySubstring_OffsetNegative", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "Pneumonia"
		offset := -1
		limit := 3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when offset is negative")
		assert.Contains(t, strings.ToLower(err.Error()), "offset must be positive")
	})

	t.Run("FindBySubstring_LimitNegative", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "Pneumonia"
		offset := 1
		limit := -3
		resultDiseases, err := repo.FindBySubstring(name, offset, limit)

		assert.Nil(t, resultDiseases, "Expected result to be nil when limit is negative")
		assert.Contains(t, strings.ToLower(err.Error()), "limit must be positive")
	})

	t.Run("CountBySubstring_LowercaseName", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		expectedCount := int64(4)
		name := "pneumonia"
		count, err := repo.CountBySubstring(name)

		assert.NotNil(t, count, "Expected no error counting diseases")
		assert.NoError(t, err, "Expected no error counting diseases")
		assert.Equal(t, expectedCount, count)
	})

	t.Run("CountBySubstring_UppercaseName", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		expectedCount := int64(4)
		name := "Pneumonia"
		count, err := repo.CountBySubstring(name)

		assert.NotNil(t, count, "Expected no error counting diseases")
		assert.NoError(t, err, "Expected no error counting diseases")
		assert.Equal(t, expectedCount, count)
	})

	t.Run("CountBySubstring_NotFound", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

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
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("64814b64-d138-4209-b869-3b649db06ab1")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "44051401458"
		record := &commonsModels.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.NoError(t, err, "Expected no error validating patient's uniqueness")
	})

	t.Run("ValidateUniquePatient_DuplicatedID", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "44051401458"
		record := &commonsModels.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.Error(t, err, "Expected error when patient ID is duplicated")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("ValidateUniquePatient_DuplicatedNationalID", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("64814b64-d138-4209-b869-3b649db06ab1")
		require.NoError(t, err, "Failed to parse patient UUID")
		patientNationalID := "12345654321"
		record := &commonsModels.PatientRecord{
			ID: patientID,
			NationalID: patientNationalID,
		}

		err = repo.ValidateUniquePatient(record)

		assert.Error(t, err, "Expected error when patient NationalID is duplicated")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")
	})

	t.Run("AddPatient_InvalidPatient", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		// Duplicate ID
		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse target UUID")

		patient := commonsModels.PatientRecord{
			ID: targetID,
			NationalID: "11223311223",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")

		// Duplicate NationalID
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = commonsModels.PatientRecord{
			ID: targetID,
			NationalID: "12345123451",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")

		// Invalid NationalID (letters)
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = commonsModels.PatientRecord{
			ID: targetID,
			NationalID: "abc45123451",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid id")

		// Invalid NationalID (length)
		targetID, err = uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient = commonsModels.PatientRecord{
			ID: targetID,
			NationalID: "1234",
		}

		err = repo.AddPatient(&patient)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid id")
	})

	t.Run("AddPatient_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		targetID, err := uuid.Parse("b853ebce-828c-4ec6-a667-65e61c471877")
		require.NoError(t, err, "Failed to parse target UUID")

		patient := commonsModels.PatientRecord{
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

	t.Run("GetDiseaseByRxNormID_DiseaseNotFound", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "D000000"
		disease, err := repo.GetDiseaseByRxNormID(rxNormID)

		errMsg := "Disease not found"
		assert.Nil(t, disease, "Expected disease to be nil")
		assert.Error(t, err, "Expected error when disease not found")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("GetDiseaseByRxNormID_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "D011014"
		disease, err := repo.GetDiseaseByRxNormID(rxNormID)

		assert.NoError(t, err, "Expected no error finding disease")

		expectedID, err := uuid.Parse("ebb58b3c-4356-4564-bd01-ddd495927025")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRxNormID := "D011014"
		expectedName := "Pneumonia"
		assert.Equal(t, expectedID, disease.ID)
		assert.Equal(t, expectedRxNormID, disease.RxNormID)
		assert.Equal(t, expectedName, disease.Name)
	})

	t.Run("AddDisease_DiseaseInsertError", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := ""
		name := ""
		disease, err := repo.AddDisease(rxNormID, name)

		errMsg := "Disease could not be inserted"
		assert.Nil(t, disease, "Expected disease to be nil")
		assert.Error(t, err, "Expected error when disease cannot be inserted")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("AddDisease_NewDisease", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "D000000"
		name := "New disease"
		disease, err := repo.AddDisease(rxNormID, name)

		assert.NoError(t, err, "Expected no error adding disease")

		expectedRxNormID := "D000000"
		expectedName := "New disease"
		assert.Equal(t, expectedRxNormID, disease.RxNormID)
		assert.Equal(t, expectedName, disease.Name)
	})

	t.Run("AddDisease_ExistingDisease", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "D031249"
		name := "Erdheim-Chester Disease"
		disease, err := repo.AddDisease(rxNormID, name)

		assert.NoError(t, err, "Expected no error adding disease")

		expectedID, err := uuid.Parse("6288b3bd-959f-4b57-a26e-11688e26ce5c")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRxNormID := "D031249"
		expectedName := "Erdheim-Chester Disease"
		assert.Equal(t, expectedID, disease.ID)
		assert.Equal(t, expectedRxNormID, disease.RxNormID)
		assert.Equal(t, expectedName, disease.Name)
	})

	t.Run("GetSubstanceByName_SubstanceNotFound", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "non-existent substance"
		substance, err := repo.GetSubstanceByName(name)

		errMsg := "Substance not found"
		assert.Nil(t, substance, "Expected substance to be nil")
		assert.Error(t, err, "Expected error when substance not found")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("GetSubstanceByName_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		name := "hydrocodone"
		substance, err := repo.GetSubstanceByName(name)

		assert.NoError(t, err, "Expected no error finding substance")

		expectedID, err := uuid.Parse("a5504c0c-1eb4-4967-8185-2fd82b3295b4")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "1234567"
		expectedName := "hydrocodone"
		expectedType := "IN"
		assert.Equal(t, expectedID, substance.ID)
		assert.Equal(t, expectedRXCUI, substance.RXCUI)
		assert.Equal(t, expectedName, substance.Name)
		assert.Equal(t, expectedType, substance.TTY)
	})

	t.Run("AddSubstance_SubstanceInsertError", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxClassID := ""
		name := ""
		substanceType := ""
		substance, err := repo.AddSubstance(rxClassID, name, substanceType)

		errMsg := "Substance could not be inserted"
		assert.Nil(t, substance, "Expected substance to be nil")
		assert.Error(t, err, "Expected error when substance cannot be inserted")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("AddSubstance_NewSubstance", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxClassID := "7654321"
		name := "New Substance"
		substanceType := "PIN"
		substance, err := repo.AddSubstance(rxClassID, name, substanceType)

		assert.NoError(t, err, "Expected no error adding substance")

		expectedRXCUI := "7654321"
		expectedName := "New Substance"
		expectedTTY := "PIN"
		assert.Equal(t, expectedRXCUI, substance.RXCUI)
		assert.Equal(t, expectedName, substance.Name)
		assert.Equal(t, expectedTTY, substance.TTY)
	})

	t.Run("AddSubstance_ExistingSubstance", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxClassID := "1234567"
		name := "hydrocodone"
		substanceType := "IN"
		substance, err := repo.AddSubstance(rxClassID, name, substanceType)

		assert.NoError(t, err, "Expected no error adding substance")

		expectedID, err := uuid.Parse("a5504c0c-1eb4-4967-8185-2fd82b3295b4")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "1234567"
		expectedName := "hydrocodone"
		expectedTTY := "IN"
		assert.Equal(t, expectedID, substance.ID)
		assert.Equal(t, expectedRXCUI, substance.RXCUI)
		assert.Equal(t, expectedName, substance.Name)
		assert.Equal(t, expectedTTY, substance.TTY)
	})

	t.Run("GetDrugByRxNormID_DrugNotFoundError", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "non-existent substance"
		drug, err := repo.GetDrugByRXCUI(rxNormID)

		errMsg := "Drug not found"
		assert.Nil(t, drug, "Expected drug to be nil")
		assert.Error(t, err, "Expected error when drug not found")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("GetDrugByRxNormID_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "261315"
		drug, err := repo.GetDrugByRXCUI(rxNormID)

		assert.NoError(t, err, "Expected no error finding drug")

		expectedID, err := uuid.Parse("ee5fd388-c675-477b-9bc5-3f16cc359abe")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "261315"
		expectedName := "Tamiflu"
		expectedSubstance := "hydrocodone"
		assert.Equal(t, expectedID, drug.ID)
		assert.Equal(t, expectedRXCUI, drug.RXCUI)
		assert.Equal(t, expectedName, drug.Name)
		assert.Equal(t, expectedSubstance, drug.Substance)
	})

	t.Run("AddDrug_DrugInsertError", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := ""
		name := ""
		substance := ""
		drug, err := repo.AddDrug(rxNormID, name, substance)

		errMsg := "Drug could not be inserted"
		assert.Nil(t, drug, "Expected drug to be nil")
		assert.Error(t, err, "Expected error when drug cannot be inserted")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("AddDrug_NewDrug", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "111111"
		name := "New Drug"
		substance := "substance"
		drug, err := repo.AddDrug(rxNormID, name, substance)

		assert.NoError(t, err, "Expected no error adding drug")

		expectedRXCUI := "111111"
		expectedName := "New Drug"
		expectedSubstance := "substance"
		assert.Equal(t, expectedRXCUI, drug.RXCUI)
		assert.Equal(t, expectedName, drug.Name)
		assert.Equal(t, expectedSubstance, drug.Substance)
	})

	t.Run("AddDrug_ExistingDrug", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		rxNormID := "261315"
		name := "Tamiflu"
		substance:= "hydrocodone"
		drug, err := repo.AddDrug(rxNormID, name, substance)

		assert.NoError(t, err, "Expected no error adding drug")

		expectedID, err := uuid.Parse("ee5fd388-c675-477b-9bc5-3f16cc359abe")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "261315"
		expectedName := "Tamiflu"
		expectedSubstance:= "hydrocodone"
		assert.Equal(t, expectedID, drug.ID)
		assert.Equal(t, expectedRXCUI, drug.RXCUI)
		assert.Equal(t, expectedName, drug.Name)
		assert.Equal(t, expectedSubstance, drug.Substance)
	})

	t.Run("GetMedicalHistoryItem_MedicalHistoryItemNotFound", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("4a4e830a-36f8-4d32-a691-ff808bc36a56")
		require.NoError(t, err, "Failed to parse target UUID")
		item, err := repo.GetMedicalHistoryItemByPatientID(patientID)

		errMsg := "Medical history item not found"
		assert.Nil(t, item, "Expected medical history item to be nil")
		assert.Error(t, err, "Expected error when medical history item not found")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("GetMedicalHistoryItem_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse target UUID")

		item, err := repo.GetMedicalHistoryItemByPatientID(patientID)

		assert.NoError(t, err, "Expected no error finding medical history item")

		expectedID, err := uuid.Parse("aa0e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err, "Failed to parse target UUID")
		expectedPatientID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failed to parse target UUID")
		expectedDoctorID, err := uuid.Parse("880e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err, "Failed to parse target UUID")
		expectedDrugID, err := uuid.Parse("ee5fd388-c675-477b-9bc5-3f16cc359abe")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "261315"
		expectedName := "Tamiflu"
		expectedSubstance := "hydrocodone"

		assert.Equal(t, expectedID, item[0].ID)
		assert.Equal(t, expectedPatientID, item[0].PatientID)
		assert.Equal(t, expectedDoctorID, item[0].DoctorID)
		assert.Equal(t, expectedDrugID, item[0].Drugs[0].ID)
		assert.Equal(t, expectedRXCUI, item[0].Drugs[0].RXCUI)
		assert.Equal(t, expectedName, item[0].Drugs[0].Name)
		assert.Equal(t, expectedSubstance, item[0].Drugs[0].Substance)
	})

	t.Run("AddMedicalHistoryItem_MedicalHistoryItemInsertError", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("c2aa753e-ce76-43db-b855-399d1955ad66")
		require.NoError(t, err, "Failed to parse target UUID")
		doctorID, err := uuid.Parse("4f070b72-1b62-43d7-b085-b9af88d1eefb")
		require.NoError(t, err, "Failed to parse target UUID")

		drugs := []models.Drug{
			models.Drug{
				RXCUI:     "111111",
				Name:      "New drug",
				Substance: "Substance",
			},
		}

		item, err := repo.AddMedicalHistoryItem(patientID, doctorID, drugs)

		errMsg := "Medical history item could not be inserted"
		assert.Nil(t, item, "Expected medical history item to be nil")
		assert.Error(t, err, "Expected error when medical history item cannot be inserted")
		assert.Equal(t, err.Error(), errMsg)
	})

	t.Run("AddMedicalHistoryItem_Success", func(t *testing.T) {
		repo, cleanup := initRepo(t)
		defer cleanup()

		patientID, err := uuid.Parse("c2aa753e-ce76-43db-b855-399d1955ad66")
		require.NoError(t, err, "Failed to parse target UUID")
		doctorID, err := uuid.Parse("880e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err, "Failed to parse target UUID")
		drugs := []models.Drug{
			models.Drug{
				RXCUI:     "111111",
				Name:      "New drug",
				Substance: "Substance",
			},
		}

		item, err := repo.AddMedicalHistoryItem(patientID, doctorID, drugs)
		log.Printf("%v", item.Drugs)

		assert.NoError(t, err, "Expected no error adding medical history item")

		items, err := repo.GetMedicalHistoryItemByPatientID(patientID)
		assert.NoError(t, err, "Expected no error finding medical history items")

		expectedPatientID, err := uuid.Parse("c2aa753e-ce76-43db-b855-399d1955ad66")
		require.NoError(t, err, "Failed to parse target UUID")
		expectedDoctorID, err := uuid.Parse("880e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err, "Failed to parse target UUID")

		expectedRXCUI := "111111"
		expectedName := "New drug"
		expectedSubstance := "Substance"

		assert.Equal(t, expectedPatientID, items[0].PatientID)
		assert.Equal(t, expectedDoctorID, items[0].DoctorID)
		log.Printf("%v", items[0])
		assert.Equal(t, expectedRXCUI, items[0].Drugs[0].RXCUI)
		assert.Equal(t, expectedName, items[0].Drugs[0].Name)
		assert.Equal(t, expectedSubstance, items[0].Drugs[0].Substance)
	})

}
