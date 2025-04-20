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

	"igaku/controllers"
	"igaku/dtos"
	"igaku/models"
	"igaku/services"
	"igaku/tests/mocks"
	"igaku/utils"
)

func setupAccountRouter(t *testing.T, mockRepo *mocks.UserRepository) (*httptest.ResponseRecorder, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	accountService := services.NewAccountService(mockRepo)
	accountController := controllers.NewAccountController(accountService)

	rec := httptest.NewRecorder()
	router := gin.Default()
	accountController.RegisterRoutes(router)

	return rec, router
}

func TestAccountController_GetSelf_NoBearerToken(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/accounts/self", nil)
	require.NoError(t, err)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse dtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Authorization header required"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_GetSelf_InvalidTokenFormat(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/accounts/self", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "INVALID.TOKEN")

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse dtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Unauthorized"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_GetSelf_ExpiredToken(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/accounts/self", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := models.User{
		ID: id,
		Username: "jdoe",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: models.Patient,
	}

	issuedAt, err := time.Parse(time.DateTime, "1998-06-07 08:00:00")
	require.NoError(t, err)
	expiresAt, err := time.Parse(time.DateTime, "1998-06-07 09:00:00")
	require.NoError(t, err)
	token, err := utils.GenerateJWTToken(
		&user,
		issuedAt,
		expiresAt,
	)
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse dtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Token has expired"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_GetSelf_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	testID := uuid.New()
	testUsername := "jdoe"
	hashedPassword := "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6"
	expectedUser := &models.User{
		ID: testID,
		Username: testUsername,
		Password: hashedPassword,
		Role: models.Patient,
	}
	mockRepo.On("FindByID", testID).Return(expectedUser, nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/accounts/self", nil)
	require.NoError(t, err)

	token, err := utils.GenerateJWTToken(
		expectedUser,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var accDetails dtos.AccountDetails
	err = json.Unmarshal(w.Body.Bytes(), &accDetails)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, expectedUser.Username, accDetails.Username)
	assert.Equal(t, string(expectedUser.Role), accDetails.Role)
}
