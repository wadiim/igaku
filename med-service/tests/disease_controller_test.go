package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	commonsDtos "igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
	"igaku/med-service/controllers"
	"igaku/med-service/dtos"
	"igaku/med-service/errors"
	"igaku/med-service/services"
)

type MockDiseaseRepository struct {
	mock.Mock
}

func (m *MockDiseaseRepository) FindBySubstring(
	name string,
	offset int,
	limit int,
) ([]*models.Disease, error) {
	args := m.Called(name, offset, limit)

	var r0 []*models.Disease
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*models.Disease)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockDiseaseRepository) CountBySubstring(name string) (int64, error) {
	args := m.Called(name)

	var r0 int64
	if args.Get(0) != nil {
		r0 = args.Get(0).(int64)
	}

	r1 := args.Error(1)

	return r0, r1
}

func setupRouter(mockRepo *MockDiseaseRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	diseaseService := services.NewDiseaseService(mockRepo)
	diseaseController := controllers.NewDiseaseController(diseaseService)

	router := gin.Default()
	diseaseController.RegisterRoutes(router)

	return router
}

func unpackPaginatedResponse(t *testing.T, body *bytes.Buffer) (
	dtos.PaginatedResponse,
	[]dtos.DiseaseDetails,
) {
	var paginatedResponse dtos.PaginatedResponse
	err := json.Unmarshal(body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var diseasesResponse []dtos.DiseaseDetails
	err = json.Unmarshal(jsonData, &diseasesResponse)
	assert.NoError(t, err)

	return paginatedResponse, diseasesResponse
}

func TestDiseaseController_GetBySubstring_DefaultParam(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"
	// expectedDiseases := []*models.Disease{
	// 	{ID: uuid.New(), RxNormID: "D000000", Name: "Lupus Vulgaris"},
	// 	{ID: uuid.New(), RxNormID: "D000001", Name: "Lupus Nephritis"},
	// }
	count := 9
	expectedDiseases := make([]*models.Disease, count)
	for i := range count {
		expectedDiseases[i] = &models.Disease{
			ID: uuid.New(),
			RxNormID: fmt.Sprintf("D000%d", i),
			Name: fmt.Sprintf("Lupus %d", i),
		}
	}

	mockRepo.On("FindBySubstring", testName, 0, 10).Return(expectedDiseases, nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse := unpackPaginatedResponse(t, rec.Body)

	expectedPage := 1
	expectedPageSize := 10
	expectedTotalPages := 1
	expectedTotalCount := int64(count)

	assert.Equal(t, count, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetBySubstring_WithParam(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"

	count := 6
	expectedDiseases := make([]*models.Disease, count)
	for i := range count {
		expectedDiseases[i] = &models.Disease{
			ID: uuid.New(),
			RxNormID: fmt.Sprintf("D000%d", i),
			Name: fmt.Sprintf("Lupus %d", i),
		}
	}

	mockRepo.On("FindBySubstring", testName, 0, 5).Return(expectedDiseases, nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	page := 1
	pageSize := 5
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse := unpackPaginatedResponse(t, rec.Body)

	expectedPage := 1
	expectedPageSize := 5
	expectedTotalPages := 2
	expectedTotalCount := int64(count)

	assert.Equal(t, count, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetBySubstring_CountMoreThanPageSize(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"

	count := 6
	expectedDiseases := make([]*models.Disease, count)
	for i := range count {
		expectedDiseases[i] = &models.Disease{
			ID: uuid.New(),
			RxNormID: fmt.Sprintf("D000%d", i),
			Name: fmt.Sprintf("Lupus %d", i),
		}
	}

	page := 1
	pageSize := 5

	mockRepo.On("FindBySubstring", testName, 0, pageSize).Return(expectedDiseases[0:pageSize], nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse := unpackPaginatedResponse(t, rec.Body)

	expectedCount := 5
	expectedPage := 1
	expectedPageSize := 5
	expectedTotalPages := 2
	expectedTotalCount := int64(count)

	assert.Equal(t, expectedCount, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockRepo.AssertExpectations(t)

	page = 2
	mockRepo.On("FindBySubstring", testName, 5, 5).Return(expectedDiseases[pageSize:count], nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse = unpackPaginatedResponse(t, rec.Body)

	expectedCount = 1
	expectedPage = 2
	assert.Equal(t, expectedCount, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)
}

func TestDiseaseController_GetBySubstring_CountLessThanPageSize(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"

	count := 4
	expectedDiseases := make([]*models.Disease, count)
	for i := range count {
		expectedDiseases[i] = &models.Disease{
			ID: uuid.New(),
			RxNormID: fmt.Sprintf("D000%d", i),
			Name: fmt.Sprintf("Lupus %d", i),
		}
	}

	page := 1
	pageSize := 5

	mockRepo.On("FindBySubstring", testName, 0, pageSize).Return(expectedDiseases, nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse := unpackPaginatedResponse(t, rec.Body)

	expectedCount := count
	expectedPage := 1
	expectedPageSize := pageSize
	expectedTotalPages := 1
	expectedTotalCount := int64(count)

	assert.Equal(t, expectedCount, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetBySubstring_EmptyPage(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"

	count := 5
	expectedDiseases := make([]*models.Disease, count)
	for i := range count {
		expectedDiseases[i] = &models.Disease{
			ID: uuid.New(),
			RxNormID: fmt.Sprintf("D000%d", i),
			Name: fmt.Sprintf("Lupus %d", i),
		}
	}

	page := 2
	pageSize := 5
	offset := (page - 1) * pageSize

	mockRepo.On("FindBySubstring", testName, offset, pageSize).Return(
		nil, &errors.DiseaseNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	expectedErrMsg := "Disease not found"
	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedErrMsg, errResponse.Message)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetBySubstring_InvalidPageParam(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"
	req1, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=-5", testName),
		nil,
	)
	require.NoError(t, err)
	req2, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=0", testName),
		nil,
	)
	require.NoError(t, err)
	req3, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=abc", testName),
		nil,
	)
	require.NoError(t, err)
	for i, req := range []*http.Request{req1, req2, req3} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(
			t, http.StatusBadRequest, rec.Code,
			"Test Case %d: Expected HTTP status 400 Bad Request",
			i+1,
		)

		var errResponse commonsDtos.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring")
	mockRepo.AssertNotCalled(t, "CountBySubstring")
}

func TestDiseaseController_GetBySubstring_InvalidPageSizeParam(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Lupus"
	req1, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=-5", testName),
		nil,
	)
	require.NoError(t, err)
	req2, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=0", testName),
		nil,
	)
	require.NoError(t, err)
	req3, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=abc", testName),
		nil,
	)
	require.NoError(t, err)
	for i, req := range []*http.Request{req1, req2, req3} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(
			t, http.StatusBadRequest, rec.Code,
			"Test Case %d: Expected HTTP status 400 Bad Request",
			i+1,
		)

		var errResponse commonsDtos.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring")
	mockRepo.AssertNotCalled(t, "CountBySubstring")
}

func TestDiseaseController_GetBySubstring_DiseaseNotFound(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Wilson"

	offset := 0
	pageSize := 10

	mockRepo.On("FindBySubstring", testName, offset, pageSize).Return(
		nil, &errors.DiseaseNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	expectedErrMsg := "Disease not found"
	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedErrMsg, errResponse.Message)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetBySubstring_RepoError(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testName := "Test"

	offset := 0
	pageSize := 10

	mockRepo.On("FindBySubstring", testName, offset, pageSize).Return(
		nil, &commonsErrors.DatabaseError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	expectedErrMsg := "Failed to retrieve disease details"
	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedErrMsg, errResponse.Message)

	mockRepo.AssertExpectations(t)
}
