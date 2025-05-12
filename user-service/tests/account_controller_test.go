package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"igaku/user-service/controllers"
	"igaku/user-service/dtos"
	"igaku/user-service/services"
	"igaku/user-service/tests/mocks"
	"igaku/user-service/utils"
	"igaku/commons/models"
	commonsDtos "igaku/commons/dtos"
	commonsUtils "igaku/commons/utils"
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

func genAdminToken(t *testing.T) string {
	admin := &models.User{
		ID: uuid.New(),
		Username: "admin",
		Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
		Role: models.Admin,
	}

	token, err := commonsUtils.GenerateJWTToken(
		admin,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	return token
}

func TestAccountController_GetSelf_NoToken(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/user/self", nil)
	require.NoError(t, err)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
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

	req, err := http.NewRequest(http.MethodGet, "/user/self", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "INVALID.TOKEN")

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
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

	req, err := http.NewRequest(http.MethodGet, "/user/self", nil)
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
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		issuedAt,
		expiresAt,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
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

	req, err := http.NewRequest(http.MethodGet, "/user/self", nil)
	require.NoError(t, err)

	token, err := commonsUtils.GenerateJWTToken(
		expectedUser,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var accDetails dtos.AccountDetails
	err = json.Unmarshal(w.Body.Bytes(), &accDetails)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, expectedUser.Username, accDetails.Username)
	assert.Equal(t, string(expectedUser.Role), accDetails.Role)
}

func TestAccountController_ListAccounts_NoToken(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusUnauthorized, w.Code,
		"Expected HTTP status 401 Unauthorized",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Authorization header required"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for missing header",
	)

	mockRepo.AssertNotCalled(t, "FindAll", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "CountAll")

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_Unauthorized_Patient(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)

	notAdmin := &models.User{
		ID: uuid.New(),
		Username: "notAdmin",
		Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
		Role: models.Patient,
	}

	token, err := commonsUtils.GenerateJWTToken(
		notAdmin,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errResponse.Message)

	mockRepo.AssertNotCalled(t, "FindAll", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "CountAll")

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_Unauthorized_Doctor(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)

	notAdmin := &models.User{
		ID: uuid.New(),
		Username: "notAdmin",
		Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
		Role: models.Doctor,
	}

	token, err := commonsUtils.GenerateJWTToken(
		notAdmin,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errResponse.Message)

	mockRepo.AssertNotCalled(t, "FindAll", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "CountAll")

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_InvalidPageParam(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	req1, err := http.NewRequest(http.MethodGet, "/user/list?page=abc", nil)
	require.NoError(t, err)
	req1.Header.Set("Authorization", adminToken)

	req2, err := http.NewRequest(http.MethodGet, "/user/list?page=0", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", adminToken)

	req3, err := http.NewRequest(http.MethodGet, "/user/list?page=-1", nil)
	require.NoError(t, err)
	req3.Header.Set("Authorization", adminToken)

	for i, req := range []*http.Request{req1, req2, req3} {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(
			t, http.StatusBadRequest, w.Code,
			"Test Case %d: Expected HTTP status 400 Bad Request",
			i+1,
		)

		var errResponse commonsDtos.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResponse)
		require.NoError(
			t, err,
			"Test Case %d: Failed to unmarshal error response body",
			i+1,
		)
		assert.Contains(
			t, errResponse.Message, "Invalid page parameter",
			"Test Case %d: Expected error message about page parameter",
			i+1,
		)
	}

	mockRepo.AssertNotCalled(
		t, "FindAll",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	)
	mockRepo.AssertNotCalled(t, "CountAll")
}

func TestAccountController_ListAccounts_InvalidPageSizeParam(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	req1, err := http.NewRequest(http.MethodGet, "/user/list?pageSize=xyz", nil)
	require.NoError(t, err)
	req1.Header.Set("Authorization", adminToken)

	req2, err := http.NewRequest(http.MethodGet, "/user/list?pageSize=0", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", adminToken)

	req3, err := http.NewRequest(http.MethodGet, "/user/list?pageSize=-5", nil)
	require.NoError(t, err)
	req3.Header.Set("Authorization", adminToken)

	for i, req := range []*http.Request{req1, req2, req3} {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(
			t, http.StatusBadRequest, w.Code,
			"Test Case %d: Expected HTTP status 400 Bad Request",
			i+1,
		)

		var errResponse commonsDtos.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResponse)
		require.NoError(
			t, err,
			"Test Case %d: Failed to unmarshal error response body",
			i+1,
		)
		assert.Contains(
			t, errResponse.Message, "Invalid pageSize parameter",
			"Test Case %d: Expected error message about pageSize parameter",
			i+1,
		)
	}

	mockRepo.AssertNotCalled(
		t, "FindAll",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	)
	mockRepo.AssertNotCalled(t, "CountAll")
}

func TestAccountController_ListAccounts_InvalidOrderByParam(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	req, err := http.NewRequest(
		http.MethodGet, "/user/list?orderBy=invalidField", nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusBadRequest,
		w.Code, "Expected HTTP status 400 Bad Request",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")
	assert.Contains(
		t, errResponse.Message, "Invalid orderBy parameter",
		"Expected error message about orderBy parameter",
	)

	mockRepo.AssertNotCalled(
		t, "FindAll",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	)
	mockRepo.AssertNotCalled(t, "CountAll")
}

func TestAccountController_ListAccounts_InvalidOrderMethodParam(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	req, err := http.NewRequest(
		http.MethodGet, "/user/list?orderMethod=invalidMethod", nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(
		t, http.StatusBadRequest,
		w.Code, "Expected HTTP status 400 Bad Request",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err, "Failed to unmarshal error response body")
	assert.Contains(
		t, errResponse.Message, "Invalid orderMethod parameter",
		"Expected error message about orderBy parameter",
	)

	mockRepo.AssertNotCalled(
		t, "FindAll",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	)
	mockRepo.AssertNotCalled(t, "CountAll")
}

func TestAccountController_ListAccounts_RepoError_FindAll(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	repoError := errors.New("database connection lost during find")
	expectedErrMsg := "Failed to retrieve accounts list"

	mockRepo.On("CountAll").Return(int64(5), nil).Once()
	mockRepo.On("FindAll", 0, 10, models.ID, utils.Asc).
		Return(nil, repoError).Once()

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedErrMsg, errResponse.Message)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_RepoError_CountAll(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	repoError := errors.New("database connection lost during count")
	expectedErrMsg := "Failed to retrieve accounts list"

	mockRepo.On("CountAll").Return(int64(0), repoError).Once()
	// FindAll should NOT be called if CountAll fails

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedErrMsg, errResponse.Message)

	mockRepo.AssertNotCalled(
		t, "FindAll",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_DefaultParams(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	mockUsers := make([]models.User, 15)
	for i := 0; i < 15; i++ {
		mockUsers[i] = models.User{
			ID: uuid.New(),
			Username: fmt.Sprintf("user%d", i),
			Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			Role: models.Patient,
		}
	}
	totalCount := int64(len(mockUsers))
	defaultPageSize := 10
	expectedPage := 1
	expectedTotalPages := 2

	mockRepo.On("CountAll").Return(totalCount, nil).Once()
	// The returned list won't probably be sorted by ID, but whatever.
	mockRepo.On("FindAll", 0, defaultPageSize, models.ID, utils.Asc).
		Return(mockUsers[:defaultPageSize], nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedPage, response.Page)
	assert.Equal(t, defaultPageSize, response.PageSize)
	assert.Equal(t, totalCount, response.TotalCount)
	assert.Equal(t, expectedTotalPages, response.TotalPages)

	require.NotNil(t, response.Data)
	var dataBytes []byte
	dataBytes, err = json.Marshal(response.Data)
	require.NoError(t, err)
	var actualDetails []dtos.AccountDetailsWithID
	err = json.Unmarshal(dataBytes, &actualDetails)
	require.NoError(t, err)

	assert.Len(t, actualDetails, defaultPageSize)
	for i := 0; i < defaultPageSize; i++ {
		assert.Equal(t, mockUsers[i].ID.String(), actualDetails[i].ID)
		assert.Equal(t, mockUsers[i].Username, actualDetails[i].Username)
		assert.Equal(t, string(mockUsers[i].Role), actualDetails[i].Role)
	}

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_WithParams(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	mockUsers := make([]models.User, 12)
	for i := 0; i < 12; i++ {
		mockUsers[i] = models.User{
			ID: uuid.New(),
			Username: fmt.Sprintf("user%d", i),
			Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			Role: models.Patient,
		}
	}
	totalCount := int64(len(mockUsers))

	page := 2
	pageSize := 5
	orderBy := "username"
	orderMethod := "desc"

	expectedOffset := 5
	expectedLimit := pageSize
	expectedOrderBy := models.Username
	expectedOrderMethod := utils.Desc
	expectedTotalPages := 3

	mockRepo.On("CountAll").Return(totalCount, nil).Once()
	mockRepo.On(
		"FindAll",
		expectedOffset,
		expectedLimit,
		expectedOrderBy,
		expectedOrderMethod,
	).Return(mockUsers[5:10], nil).Once()

	url := fmt.Sprintf(
		"/user/list?page=%d&pageSize=%d&orderBy=%s&orderMethod=%s",
		page, pageSize, orderBy, orderMethod,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, page, response.Page)
	assert.Equal(t, pageSize, response.PageSize)
	assert.Equal(t, totalCount, response.TotalCount)
	assert.Equal(t, expectedTotalPages, response.TotalPages)

	var dataBytes []byte
	dataBytes, err = json.Marshal(response.Data)
	require.NoError(t, err)
	var actualDetails []dtos.AccountDetailsWithID
	err = json.Unmarshal(dataBytes, &actualDetails)
	require.NoError(t, err)

	assert.Len(t, actualDetails, pageSize)

	for i := 0; i < pageSize; i++ {
		assert.Equal(
			t,
			mockUsers[expectedOffset+i].ID.String(),
			actualDetails[i].ID,
		)
		assert.Equal(
			t,
			mockUsers[expectedOffset+i].Username,
			actualDetails[i].Username,
		)
		assert.Equal(
			t,
			string(mockUsers[expectedOffset+i].Role),
			actualDetails[i].Role,
		)
	}

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_PageGreaterThanItemCount(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	totalCount := int64(15)
	page := 3
	pageSize := 10
	expectedOffset := 20
	expectedLimit := pageSize
	expectedTotalPages := 2

	mockRepo.On("CountAll").Return(totalCount, nil).Once()
	mockRepo.On(
		"FindAll",
		expectedOffset, expectedLimit, models.ID, utils.Asc,
	).Return([]models.User{}, nil).Once()

	url := fmt.Sprintf("/user/list?page=%d&pageSize=%d", page, pageSize)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, page, response.Page)
	assert.Equal(t, pageSize, response.PageSize)
	assert.Equal(t, totalCount, response.TotalCount)
	assert.Equal(t, expectedTotalPages, response.TotalPages)

	var dataBytes []byte
	dataBytes, err = json.Marshal(response.Data)
	require.NoError(t, err)
	var actualDetails []dtos.AccountDetailsWithID
	err = json.Unmarshal(dataBytes, &actualDetails)
	require.NoError(t, err)
	assert.Empty(t, actualDetails)
	assert.Len(t, actualDetails, 0)

	mockRepo.AssertExpectations(t)
}

func TestAccountController_ListAccounts_EmptyList(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	w, router := setupAccountRouter(t, mockRepo)
	adminToken := genAdminToken(t)

	totalCount := int64(0)
	defaultPageSize := 10
	expectedPage := 1
	expectedTotalPages := 0

	mockRepo.On("CountAll").Return(totalCount, nil).Once()
	mockRepo.On(
		"FindAll",
		0, defaultPageSize, models.ID, utils.Asc,
	).Return([]models.User{}, nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/user/list", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", adminToken)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dtos.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedPage, response.Page)
	assert.Equal(t, defaultPageSize, response.PageSize)
	assert.Equal(t, totalCount, response.TotalCount)
	assert.Equal(t, expectedTotalPages, response.TotalPages)

	var dataBytes []byte
	dataBytes, err = json.Marshal(response.Data)
	require.NoError(t, err)
	var actualDetails []dtos.AccountDetailsWithID
	err = json.Unmarshal(dataBytes, &actualDetails)
	require.NoError(t, err)
	assert.Empty(t, actualDetails)
	assert.Len(t, actualDetails, 0)

	mockRepo.AssertExpectations(t)
}
