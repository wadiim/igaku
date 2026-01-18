package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	"igaku/med-service/controllers"
	"igaku/med-service/dtos"
	"igaku/med-service/errors"
	"igaku/med-service/services"
	"igaku/med-service/tests/mocks"
)

func setupDrugRouter(mockAPI *mocks.MockRxClassAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)

	drugService := services.NewDrugService(mockAPI)
	drugController := controllers.NewDrugController(drugService)

	router := gin.Default()
	drugController.RegisterRoutes(router)

	return router
}

func TestDrugController_GetRecommendedDrugs_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	diseaseID := "D007251"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
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

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	diseaseID := "D007251"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
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

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	diseaseID := "D007251"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
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

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	diseaseID := "D007251"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "jdoe",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Patient,
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

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	diseaseID := "D007251"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "admin",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Admin,
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

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_InvalidPage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	// Zero page
	diseaseID := "D007251"
	page := 0
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?page=%d", diseaseID, page),
		nil,
	)
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

	expectedErrMsg := "Invalid page parameter. Must be a positive integer."
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid page number",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)

	// Negative page
	diseaseID = "D007251"
	page = -1
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?page=%d", diseaseID, page),
		nil,
	)
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

	expectedErrMsg = "Invalid page parameter. Must be a positive integer."
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid page number",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_InvalidPageSize(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	// Zero page
	diseaseID := "D007251"
	pageSize := 0
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?pageSize=%d", diseaseID, pageSize),
		nil,
	)
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

	expectedErrMsg := "Invalid pageSize parameter. Must be a positive integer."
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid page number",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)

	// Negative page
	diseaseID = "D007251"
	pageSize = -1
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?pageSize=%d", diseaseID, pageSize),
		nil,
	)
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

	expectedErrMsg = "Invalid pageSize parameter. Must be a positive integer."
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid page size",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_InvalidOrderBy(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	diseaseID := "D007251"
	orderBy := "invalid"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?orderBy=%s", diseaseID, orderBy),
		nil,
	)
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

	expectedErrMsg := "Invalid orderBy parameter. Must be `id`, `name` or `substance`"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid orderBy parameter",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_InvalidOrderMethod(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	diseaseID := "D007251"
	orderMethod := "invalid"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s?orderMethod=%s", diseaseID, orderMethod),
		nil,
	)
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

	expectedErrMsg := "Invalid orderMethod parameter. Must be `asc` or `desc`"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for invalid orderMethod parameter",
	)

	mockAPI.AssertNotCalled(t, "GetRecommendedDrugs", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_RxClassUnavailable(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	diseaseID := "D007251"
	mockAPI.On("GetSubstances", diseaseID).Return(
		nil,
		&errors.RxClassUnavailableError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusServiceUnavailable, rec.Code,
		"Expected HTTP status 503 Service Unavailable",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	log.Printf("%v", rec.Body)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "RxClass API Unavailable"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for RxClass API unavailable",
	)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_SubstanceNotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	diseaseID := "D007251"
	mockAPI.On("GetSubstances", diseaseID).Return(
		nil,
		&errors.SubstanceNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
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
	log.Printf("%v", rec.Body)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Substance not found"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for substance not found",
	)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_DrugsBySubstanceError(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substance2 := commonsModels.Substance{
		ID: "1598097",
		Name: "Non-existent",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1, substance2,
	}
	drugs := []commonsModels.Drug{
		{
			ID: "1115700",
			Name: "Oseltamivir",
			Substance: "baloxavir",
		},
	}
	diseaseID := "D007251"
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		nil,
		&errors.DrugNotFoundError{},
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance2).Return(
		drugs,
		nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/recommend/%s", diseaseID),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	expectedCount := 1
	assert.Equal(t, expectedCount, len(drugResponse))
	assert.Equal(t, drugs[0].ID, drugResponse[0].ID)
	assert.Equal(t, drugs[0].Name, drugResponse[0].Name)
	assert.Equal(t, drugs[0].Substance, drugResponse[0].Substance)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_OrderByID(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1,
	}
	drugs := []commonsModels.Drug{
		{
			ID: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: "1115700",
			Name: "AAA",
			Substance: "baloxavir",
		},
	}
	diseaseID := "D007251"

	// Asc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	orderMethod := "asc" 
	orderBy := "id"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?orderBy=%s&orderMethod=%s", diseaseID, orderBy, orderMethod,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	assert.Equal(t, "1115700", drugResponse[0].ID)
	assert.Equal(t, "1115701", drugResponse[1].ID)
	assert.Equal(t, "1115702", drugResponse[2].ID)

	mockAPI.AssertExpectations(t)

	// Desc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	orderMethod = "desc" 
	orderBy = "id"
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?orderBy=%s&orderMethod=%s", diseaseID, orderBy, orderMethod,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err = json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	assert.Equal(t, "1115702", drugResponse[0].ID)
	assert.Equal(t, "1115701", drugResponse[1].ID)
	assert.Equal(t, "1115700", drugResponse[2].ID)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_OrderByName(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1,
	}
	drugs := []commonsModels.Drug{
		{
			ID: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: "1115700",
			Name: "AAA",
			Substance: "baloxavir",
		},
	}
	diseaseID := "D007251"

	// Asc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	orderMethod := "asc" 
	orderBy := "name"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?orderBy=%s&orderMethod=%s", diseaseID, orderBy, orderMethod,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	assert.Equal(t, "AAA", drugResponse[0].Name)
	assert.Equal(t, "Oseltamivir", drugResponse[1].Name)
	assert.Equal(t, "ZZZ", drugResponse[2].Name)

	mockAPI.AssertExpectations(t)

	// Desc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	orderMethod = "desc" 
	orderBy = "name"
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?orderBy=%s&orderMethod=%s", diseaseID, orderBy, orderMethod,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err = json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	assert.Equal(t, "ZZZ", drugResponse[0].Name)
	assert.Equal(t, "Oseltamivir", drugResponse[1].Name)
	assert.Equal(t, "AAA", drugResponse[2].Name)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_SinglePage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1,
	}
	pageSize := 5
	count := 3
	var drugs []commonsModels.Drug

	for i := 0; i < count; i++ {
		drug := commonsModels.Drug{
			ID: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}
	diseaseID := "D007251"

	// Asc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?pageSize=%d", diseaseID, pageSize,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	assert.Equal(t, count, len(drugResponse))
	assert.Equal(t, "1115700", drugResponse[0].ID)
	assert.Equal(t, "Drug0", drugResponse[0].Name)
	assert.Equal(t, "Substance0", drugResponse[0].Substance)
	assert.Equal(t, "1115702", drugResponse[2].ID)
	assert.Equal(t, "Drug2", drugResponse[2].Name)
	assert.Equal(t, "Substance2", drugResponse[2].Substance)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_MultiplePages(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1,
	}
	pageSize := 5
	count := 6
	var drugs []commonsModels.Drug

	for i := 0; i < count; i++ {
		drug := commonsModels.Drug{
			ID: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}
	diseaseID := "D007251"

	page := 1
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?page=%d&pageSize=%d", diseaseID, page, pageSize,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	totalPages := 2
	totalCount := int64(count)
	assert.Equal(t, page, paginatedResponse.Page)
	assert.Equal(t, pageSize, paginatedResponse.PageSize)
	assert.Equal(t, totalPages, paginatedResponse.TotalPages)
	assert.Equal(t, totalCount, paginatedResponse.TotalCount)

	assert.Equal(t, pageSize, len(drugResponse))
	assert.Equal(t, "1115700", drugResponse[0].ID)
	assert.Equal(t, "Drug0", drugResponse[0].Name)
	assert.Equal(t, "Substance0", drugResponse[0].Substance)
	assert.Equal(t, "1115704", drugResponse[4].ID)
	assert.Equal(t, "Drug4", drugResponse[4].Name)
	assert.Equal(t, "Substance4", drugResponse[4].Substance)

	mockAPI.AssertExpectations(t)

	page = 2
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s?page=%d&pageSize=%d", diseaseID, page, pageSize,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err = json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	expectedCount := 1
	assert.Equal(t, page, paginatedResponse.Page)
	assert.Equal(t, expectedCount, len(drugResponse))
	assert.Equal(t, "1115705", drugResponse[0].ID)
	assert.Equal(t, "Drug5", drugResponse[0].Name)
	assert.Equal(t, "Substance5", drugResponse[0].Substance)

	mockAPI.AssertExpectations(t)
}

func TestDrugController_GetRecommendedDrugs_DefaultParams(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	router := setupDrugRouter(mockAPI)

	token := genDoctorToken(t)

	substance1 := commonsModels.Substance{
		ID: "1598096",
		Name: "baloxavir",
		Type: "IN",
	}
	substances := []commonsModels.Substance{
		substance1,
	}
	var drugs []commonsModels.Drug

	count := 5
	for i := 0; i < count; i++ {
		drug := commonsModels.Drug{
			ID: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}
	diseaseID := "D007251"

	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsBySubstance", substance1).Return(
		drugs,
		nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/recommend/%s", diseaseID,
		),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusOK, rec.Code,
		"Expected HTTP status 200 Success",
	)

	var paginatedResponse commonsDtos.PaginatedResponse
	err = json.Unmarshal(rec.Body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var drugResponse []dtos.DrugDetails
	err = json.Unmarshal(jsonData, &drugResponse)
	assert.NoError(t, err)

	page := 1
	pageSize := 5
	totalPages := 1
	totalCount := int64(count)
	assert.Equal(t, page, paginatedResponse.Page)
	assert.Equal(t, pageSize, paginatedResponse.PageSize)
	assert.Equal(t, totalPages, paginatedResponse.TotalPages)
	assert.Equal(t, totalCount, paginatedResponse.TotalCount)

	assert.Equal(t, count, len(drugResponse))
	assert.Equal(t, "1115700", drugResponse[0].ID)
	assert.Equal(t, "Drug0", drugResponse[0].Name)
	assert.Equal(t, "Substance0", drugResponse[0].Substance)
	assert.Equal(t, "1115704", drugResponse[4].ID)
	assert.Equal(t, "Drug4", drugResponse[4].Name)
	assert.Equal(t, "Substance4", drugResponse[4].Substance)

	mockAPI.AssertExpectations(t)
}
