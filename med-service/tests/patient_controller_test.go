package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonsDtos "igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	commonsUtils "igaku/commons/utils"
	"igaku/commons/models"
	"igaku/med-service/controllers"
	"igaku/med-service/dtos"
	"igaku/med-service/errors"
	"igaku/med-service/services"
	"igaku/med-service/tests/mocks"
)

func setupPatientRouter(
	t *testing.T,
	mockRepo *mocks.MockPatientRepository,
	mockUserClient *mocks.UserClient,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	patientService := services.NewPatientService(mockUserClient, mockRepo)
	patientController := controllers.NewPatientController(patientService)

	router := gin.Default()
	patientController.RegisterRoutes(router)

	return router
}

func TestPatientController_GetByNationalID_NoToken(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusUnauthorized, rec.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Authorization header required"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_InvalidTokenFormat(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "INVALID.TOKEN")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusUnauthorized, rec.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Unauthorized"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_UnauthorizedPatient(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "jdoe",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Patient,
	}

	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusForbidden, rec.Code,
		"Expected HTTP status 403 Forbidden",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Insufficient permissions"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_UnauthorizedAdmin(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "lcuddy",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Admin,
	}

	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusForbidden, rec.Code,
		"Expected HTTP status 403 Forbidden",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Insufficient permissions"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_ExpiredToken(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Doctor,
	}

	issuedAt, err := time.Parse(time.DateTime, "1998-06-07 08:00:00")
	require.NoError(t, err)
	expiresAt, err := time.Parse(time.DateTime, "1998-06-07 09:00:00")
	require.NoError(t, err)
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		issuedAt,
		expiresAt,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusUnauthorized, rec.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Token has expired"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_InvalidNationalID(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	// Letters
	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusBadRequest, rec.Code,
		"Expected HTTP status 400 Bad Request",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Invalid national_id parameter. Must be a 11-digit number"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	// Too short
	req, err = http.NewRequest(http.MethodGet, "/med/patient/1234", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusBadRequest, rec.Code,
		"Expected HTTP status 400 Bad Request",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg = "Invalid national_id parameter. Must be a 11-digit number"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	// Too long
	req, err = http.NewRequest(http.MethodGet, "/med/patient/123456789123", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusBadRequest, rec.Code,
		"Expected HTTP status 400 Bad Request",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg = "Invalid national_id parameter. Must be a 11-digit number"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)
}

func TestPatientController_GetByNationalID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	nationalID := "44051401458"

	// PatientRecord not found
	mockRepo.On("FindByNationalID", nationalID).Return(
		nil, &errors.PatientNotFoundError{},
	).Once()
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/patient/%s", nationalID),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusNotFound, rec.Code,
		"Expected HTTP status 404 Not Found",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Patient not found"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for patient not found",
	)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)

	// User not found
	id, err = uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	patient := &models.PatientRecord{
		ID: id,
		NationalID: nationalID,
	}
	mockRepo.On("FindByNationalID", nationalID).Return(
		patient, nil,
	).Once()
	mockUserClient.On("FindByID", id).Return(
		nil, &commonsErrors.UserNotFoundError{},
	)

	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/patient/%s", nationalID),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusNotFound, rec.Code,
		"Expected HTTP status 404 Not Found",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg = "User not found error"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for patient not found",
	)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestPatientController_GetByNationalID_Success(t *testing.T) {
	mockRepo := new(mocks.MockPatientRepository)
	mockUserClient := new(mocks.UserClient)
	router := setupPatientRouter(t, mockRepo, mockUserClient)

	id, err := uuid.Parse("0c0f5212-e90b-4d65-b4aa-60fa72c6565a")
	require.NoError(t, err)
	doctor := models.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&doctor,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	nationalID := "44051401458"

	id, err = uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	patient := &models.PatientRecord{
		ID: id,
		NationalID: nationalID,
	}
	user := &models.User{
		ID: id,
		Username: "jdoe",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Patient,
	}
	mockRepo.On("FindByNationalID", nationalID).Return(
		patient, nil,
	).Once()
	mockUserClient.On("FindByID", id).Return(
		user, nil,
	)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/patient/%s", nationalID),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 OK",
	)

	var response dtos.PatientDetails
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, user.Username, response.Username)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, patient.NationalID, response.NationalID)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}
