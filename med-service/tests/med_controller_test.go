package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonsDtos "igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	commonsUtils "igaku/commons/utils"
	commonsModels "igaku/commons/models"
	"igaku/med-service/dtos"
	"igaku/med-service/errors"
	"igaku/med-service/models"
	"igaku/med-service/tests/mocks"
	testUtils "igaku/med-service/tests/utils"
)

func unpackPaginatedResponse(t *testing.T, body *bytes.Buffer) (
	commonsDtos.PaginatedResponse,
	[]dtos.DiseaseDetails,
) {
	var paginatedResponse commonsDtos.PaginatedResponse
	err := json.Unmarshal(body.Bytes(), &paginatedResponse)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(paginatedResponse.Data)
	assert.NoError(t, err)

	var diseasesResponse []dtos.DiseaseDetails
	err = json.Unmarshal(jsonData, &diseasesResponse)
	assert.NoError(t, err)

	return paginatedResponse, diseasesResponse
}

func TestMedController_GetRecommendedDrugs_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetRecommendedDrugs_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetRecommendedDrugs_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetRecommendedDrugs_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetRecommendedDrugs_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetRecommendedDrugs_InvalidPage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_InvalidPageSize(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_InvalidOrderBy(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_InvalidOrderMethod(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_RxClassUnavailable(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_SubstanceNotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

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

func TestMedController_GetRecommendedDrugs_DrugsBySubstanceError(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id1 := uuid.New()
	id2 := uuid.New()
	substance1 := models.Substance{
		ID: id1,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substance2 := models.Substance{
		ID: id2,
		RXCUI: "1598097",
		Name: "Non-existent",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1, substance2,
	}
	id := uuid.New()
	drugs := []models.Drug{
		{
			ID: id,
			RXCUI: "1115700",
			Name: "Oseltamivir",
			Substance: "baloxavir",
		},
	}
	diseaseID := "D007251"
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
		nil,
		&errors.DrugNotFoundError{},
	).Once()
	mockAPI.On("GetDrugsByName", substance2.Name).Return(
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
	assert.Equal(t, drugs[0].RXCUI, drugResponse[0].ID)
	assert.Equal(t, drugs[0].Name, drugResponse[0].Name)
	assert.Equal(t, drugs[0].Substance, drugResponse[0].Substance)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetRecommendedDrugs_DrugNotFoundError(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	drugs := []models.Drug{}
	diseaseID := "D007251"
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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
		t, http.StatusNotFound, rec.Code,
		"Expected HTTP status 404 Not Found",
	)

	var errResponse commonsDtos.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errResponse)
	log.Printf("%v", rec.Body)
	require.NoError(t, err, "Failed to unmarshal error response body")

	expectedErrMsg := "Drug not found"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for drug not found",
	)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetRecommendedDrugs_OrderByID(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

func TestMedController_GetRecommendedDrugs_OrderByName(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

func TestMedController_GetRecommendedDrugs_OrderBySubstance(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

	assert.Equal(t, "baloxavir", drugResponse[0].Substance)
	assert.Equal(t, "caloxavir", drugResponse[1].Substance)
	assert.Equal(t, "daloxavir", drugResponse[2].Substance)

	mockAPI.AssertExpectations(t)

	// Desc
	mockAPI.On("GetSubstances", diseaseID).Return(
		substances,
		nil,
	).Once()
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

	assert.Equal(t, "daloxavir", drugResponse[0].Substance)
	assert.Equal(t, "caloxavir", drugResponse[1].Substance)
	assert.Equal(t, "baloxavir", drugResponse[2].Substance)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetRecommendedDrugs_SinglePage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	pageSize := 5
	count := 3
	var drugs []models.Drug

	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

func TestMedController_GetRecommendedDrugs_MultiplePages(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	pageSize := 5
	count := 6
	var drugs []models.Drug

	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

func TestMedController_GetRecommendedDrugs_DefaultParams(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	id := uuid.New()
	substance1 := models.Substance{
		ID: id,
		RXCUI: "1598096",
		Name: "baloxavir",
		TTY: "IN",
	}
	substances := []models.Substance{
		substance1,
	}
	var drugs []models.Drug

	count := 5
	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
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
	mockAPI.On("GetDrugsByName", substance1.Name).Return(
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

func TestMedController_GetDrugsByName_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	drugName := "morphine"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	drugName := "morphine"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	drugName := "morphine"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	drugName := "morphine"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	drugName := "morphine"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_InvalidPage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	// Zero page
	drugName := "morphine"
	page := 0
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?page=%d", drugName, page),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)

	// Negative page
	page = -1
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?page=%d", drugName, page),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_InvalidPageSize(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	// Zero page
	drugName := "morphine"
	pageSize := 0
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?pageSize=%d", drugName, pageSize),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)

	// Negative page
	pageSize = -1
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?pageSize=%d", drugName, pageSize),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_InvalidOrderBy(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugName := "morphine"
	orderBy := "invalid"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?orderBy=%s", drugName, orderBy),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_InvalidOrderMethod(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugName := "morphine"
	orderMethod := "invalid"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?orderMethod=%s", drugName, orderMethod),
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

	mockAPI.AssertNotCalled(t, "GetDrugsByName", mock.Anything)
	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_RxClassUnavailable(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugName := "morphine"
	mockAPI.On("GetDrugsByName", drugName).Return(nil, &errors.RxClassUnavailableError{}).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

func TestMedController_GetDrugsByName_DrugNotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugName := "morphine"
	mockAPI.On("GetDrugsByName", drugName).Return(nil, &errors.DrugNotFoundError{}).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	expectedErrMsg := "Drug not found"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for drug not found",
	)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_OrderByID(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
			Name: "AAA",
			Substance: "baloxavir",
		},
	}

	drugName := "morphine"
	// Asc
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod := "asc" 
	orderBy := "id"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod = "desc" 
	orderBy = "id"
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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

func TestMedController_GetDrugsByName_OrderByName(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
			Name: "AAA",
			Substance: "baloxavir",
		},
	}
	drugName := "morphine"

	// Asc
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod := "asc" 
	orderBy := "name"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod = "desc" 
	orderBy = "name"
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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

func TestMedController_GetDrugsByName_OrderBySubstance(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	drugID1 := uuid.New()
	drugID2 := uuid.New()
	drugID3 := uuid.New()
	drugs := []models.Drug{
		{
			ID: drugID1,
			RXCUI: "1115702",
			Name: "ZZZ",
			Substance: "daloxavir",
		},
		{
			ID: drugID2,
			RXCUI: "1115701",
			Name: "Oseltamivir",
			Substance: "caloxavir",
		},
		{
			ID: drugID3,
			RXCUI: "1115700",
			Name: "AAA",
			Substance: "baloxavir",
		},
	}
	drugName := "morphine"

	// Asc
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod := "asc" 
	orderBy := "name"
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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

	assert.Equal(t, "baloxavir", drugResponse[0].Substance)
	assert.Equal(t, "caloxavir", drugResponse[1].Substance)
	assert.Equal(t, "daloxavir", drugResponse[2].Substance)

	mockAPI.AssertExpectations(t)

	// Desc
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	orderMethod = "desc" 
	orderBy = "name"
	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/med/drug/%s?orderBy=%s&orderMethod=%s", drugName, orderBy, orderMethod,
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

	assert.Equal(t, "daloxavir", drugResponse[0].Substance)
	assert.Equal(t, "caloxavir", drugResponse[1].Substance)
	assert.Equal(t, "baloxavir", drugResponse[2].Substance)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_SinglePage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	count := 4
	var drugs []models.Drug

	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}

	drugName := "morphine"
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

	expectedPage := 1
	expectedPageSize := 5
	expectedTotalPages := 1
	expectedTotalCount := int64(4)

	assert.Equal(t, count, len(drugResponse))
	assert.Equal(t, "1115700", drugResponse[0].ID)
	assert.Equal(t, "Drug0", drugResponse[0].Name)
	assert.Equal(t, "Substance0", drugResponse[0].Substance)
	assert.Equal(t, "1115703", drugResponse[3].ID)
	assert.Equal(t, "Drug3", drugResponse[3].Name)
	assert.Equal(t, "Substance3", drugResponse[3].Substance)
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockAPI.AssertExpectations(t)
}

func TestMedController_GetDrugsByName_MultiplePages(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	count := 6
	var drugs []models.Drug

	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}
	drugName := "morphine"

	page := 1
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?page=%d", drugName, page),
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

	pageSize := 5
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
	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s?page=%d", drugName, page),
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

func TestMedController_GetDrugsByName_DefaultParams(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	token := testUtils.GenDoctorToken(t)

	var drugs []models.Drug

	count := 5
	for i := 0; i < count; i++ {
		id := uuid.New()
		drug := models.Drug{
			ID: id,
			RXCUI: fmt.Sprintf("111570%d", i),
			Name: fmt.Sprintf("Drug%d", i),
			Substance: fmt.Sprintf("Substance%d", i),
		}
		drugs = append(drugs, drug)
	}
	drugName := "morphine"

	mockAPI.On("GetDrugsByName", drugName).Return(drugs, nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/drug/%s", drugName),
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

func TestMedController_GetBySubstring_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/disease/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring", mock.Anything)

	mockRepo.AssertExpectations(t)
}
func TestMedController_GetBySubstring_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/disease/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetBySubstring_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/disease/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetBySubstring_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/disease/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindBySubstring", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetBySubstring_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/disease/test", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "lcuddy",
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

	mockRepo.AssertNotCalled(t, "FindBySubstring", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetBySubstring_DefaultParam(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

	mockRepo.On("FindBySubstring", testName, 0, 5).Return(expectedDiseases, nil).Once()
	mockRepo.On("CountBySubstring", testName).Return(int64(count), nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	paginatedResponse, diseasesResponse := unpackPaginatedResponse(t, rec.Body)

	expectedPage := 1
	expectedPageSize := 5
	expectedTotalPages := 1
	expectedTotalCount := int64(count)

	assert.Equal(t, count, len(diseasesResponse))
	assert.Equal(t, expectedPage, paginatedResponse.Page)
	assert.Equal(t, expectedPageSize, paginatedResponse.PageSize)
	assert.Equal(t, expectedTotalPages, paginatedResponse.TotalPages)
	assert.Equal(t, expectedTotalCount, paginatedResponse.TotalCount)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetBySubstring_WithParam(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

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

func TestMedController_GetBySubstring_CountMoreThanPageSize(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

	token := testUtils.GenDoctorToken(t)
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=%d&pageSize=%d", testName, page, pageSize),
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

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
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

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

func TestMedController_GetBySubstring_CountLessThanPageSize(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

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

func TestMedController_GetBySubstring_EmptyPage(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

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

func TestMedController_GetBySubstring_InvalidPageParam(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	testName := "Lupus"
	token := testUtils.GenDoctorToken(t)
	req1, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=-5", testName),
		nil,
	)
	req1.Header.Set("Authorization", token)
	require.NoError(t, err)
	req2, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=0", testName),
		nil,
	)
	req2.Header.Set("Authorization", token)
	require.NoError(t, err)
	req3, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?page=abc", testName),
		nil,
	)
	req3.Header.Set("Authorization", token)
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

func TestMedController_GetBySubstring_InvalidPageSizeParam(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	testName := "Lupus"
	token := testUtils.GenDoctorToken(t)
	req1, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=-5", testName),
		nil,
	)
	req1.Header.Set("Authorization", token)
	require.NoError(t, err)
	req2, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=0", testName),
		nil,
	)
	req2.Header.Set("Authorization", token)
	require.NoError(t, err)
	req3, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s?pageSize=abc", testName),
		nil,
	)
	req3.Header.Set("Authorization", token)
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

func TestMedController_GetBySubstring_DiseaseNotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	testName := "Wilson"

	offset := 0
	pageSize := 5

	mockRepo.On("FindBySubstring", testName, offset, pageSize).Return(
		nil, &errors.DiseaseNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

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

func TestMedController_GetBySubstring_RepoError(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	testName := "Test"

	offset := 0
	pageSize := 5

	mockRepo.On("FindBySubstring", testName, offset, pageSize).Return(
		nil, &commonsErrors.DatabaseError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testName),
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", testUtils.GenDoctorToken(t))

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

func TestMedController_GetByNationalID_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetMedicalHistoryItemByPatientID_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

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

func TestMedController_GetByNationalID_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestMedController_GetByNationalID_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "lcuddy",
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

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestMedController_GetByNationalID_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/med/patient/test", nil)
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

	mockRepo.AssertNotCalled(t, "FindByNationalID", mock.Anything)
	mockUserClient.AssertNotCalled(t, "FindByID", mock.Anything)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestMedController_GetByNationalID_InvalidNationalID(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
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

func TestMedController_GetByNationalID_NotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
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
	patient := &commonsModels.PatientRecord{
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

func TestMedController_GetByNationalID_Success(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0c0f5212-e90b-4d65-b4aa-60fa72c6565a")
	require.NoError(t, err)
	doctor := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
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
	patient := &commonsModels.PatientRecord{
		ID: id,
		NationalID: nationalID,
	}
	user := &commonsModels.User{
		ID: id,
		Username: "jdoe",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Patient,
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

func TestMedController_GetMedicalHistoryItemByPatientID_NoToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetByNationalID_InvalidTokenFormat(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetMedicalHistoryItemByPatientID_UnauthorizedPatient(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetMedicalHistoryItemByPatientID_UnauthorizedAdmin(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
		nil,
	)
	require.NoError(t, err)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "lcuddy",
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

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetMedicalHistoryItemByPatientID_ExpiredToken(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetMedicalHistoryItemByPatientID_NotFound(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	patientID, err := uuid.Parse(patientIDStr)
	require.NoError(t, err)

	mockRepo.On("GetMedicalHistoryItemByPatientID", patientID).Return(
		nil, &errors.MedicalHistoryItemNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	expectedErrMsg := "Medical history item not found"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for patient not found",
	)

	mockRepo.AssertExpectations(t)
}

func TestMedController_GetMedicalHistoryItemByPatientID_InvalidPatientID(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
	require.NoError(t, err)
	user := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&user,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	patientIDStr := "invalid uuid"

	req, err := http.NewRequest(
		http.MethodGet, 
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	expectedErrMsg := "Invalid patient ID"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for patient not found",
	)

	mockRepo.AssertNotCalled(t, "GetMedicalHistoryItemByPatientID", mock.Anything)
}

func TestMedController_GetMedicalHistoryItemByPatientID_Success(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("0c0f5212-e90b-4d65-b4aa-60fa72c6565a")
	require.NoError(t, err)
	doctor := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&doctor,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	patientIDStr := "0b6f13da-efb9-4221-9e89-e2729ae90030"
	patientID, err := uuid.Parse(patientIDStr)
	require.NoError(t, err)

	itemID, err := uuid.Parse("aa0e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)
	doctorID, err := uuid.Parse("880e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	layout := "2006-01-02 15:04:05"
	createdAt, err := time.Parse(layout, "2025-10-15 14:30:00")
	require.NoError(t, err)

	item := &models.MedicalHistoryItem{
		ID: itemID,
		PatientID: patientID,
		DoctorID: doctorID,
		CreatedAt: createdAt,
	}
	mockRepo.On("GetMedicalHistoryItemByPatientID", patientID).Return(
		item, nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/history/%s", patientIDStr),
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

	var response models.MedicalHistoryItem
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, itemID, response.ID)
	assert.Equal(t, patientID, response.PatientID)
	assert.Equal(t, doctorID, response.DoctorID)
	assert.Equal(t, createdAt, response.CreatedAt)

	mockRepo.AssertExpectations(t)
}

func TestMedController_CreatePrescription_MedicalHistoryItemInsertError(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("10b3d8eb-ebb9-4a0f-9280-54981a0bda9d")
	doctor := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&doctor,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	patientID, err := uuid.Parse("c2aa753e-ce76-43db-b855-399d1955ad66")
	require.NoError(t, err)
	patient := dtos.PatientDetails{
		ID: patientID,
		Username: "Irene Adler",
		Email: "iadler@mail.com",
		NationalID: "12345123451",
	}

	disease := dtos.DiseaseDetails{
		ID: "0d8209f8-a04d-493d-a162-50878a8ee5c0",
		RxNormID: "D018549",
		Name: "Cryptogenic Organizing Pneumonia",
	}

	drugs := []dtos.DrugDetails{
		{
			ID: "D007251",
			Name: "Tamiflu",
			Substance: "hydrocodone",
		},
	}
	prescription := &dtos.PrescriptionDetails{
		Patient: patient,
		Disease: disease,
		Drugs: drugs,
	}
	body, err := json.Marshal(prescription)
	require.NoError(t, err)

	mockRepo.On("AddMedicalHistoryItem", patientID, id).Return(
		nil, &errors.MedicalHistoryItemInsertError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodPost,
		"/med/prescribe",
		bytes.NewReader(body),
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

	expectedErrMsg := "Medical history item could not be inserted"
	assert.Equal(
		t, expectedErrMsg, errResponse.Message,
		"Expected specific error message for medical history item not inserted",
	)
}

func TestMedController_CreatePrescription_Success(t *testing.T) {
	mockAPI := new(mocks.MockRxClassAPI)
	mockUserClient := new(mocks.UserClient)
	mockRepo := new(mocks.MockMedRepository)
	router := testUtils.SetupRouter(mockAPI, mockUserClient, mockRepo)

	id, err := uuid.Parse("880e8400-e29b-41d4-a716-446655440001")
	doctor := commonsModels.User{
		ID: id,
		Username: "ghouse",
		Password: "$2a$12$OfvOLLULECgOzcUCzdCCCet8.9Ik7gwFipzQDDqU11rQngld5s8Nq",
		Role: commonsModels.Doctor,
	}
	token, err := commonsUtils.GenerateJWTToken(
		&doctor,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	patientID, err := uuid.Parse("c2aa753e-ce76-43db-b855-399d1955ad66")
	require.NoError(t, err)
	patient := dtos.PatientDetails{
		ID: patientID,
		Username: "Irene Adler",
		Email: "iadler@mail.com",
		NationalID: "12345123451",
	}

	disease := dtos.DiseaseDetails{
		ID: "0d8209f8-a04d-493d-a162-50878a8ee5c0",
		RxNormID: "D018549",
		Name: "Cryptogenic Organizing Pneumonia",
	}

	drugs := []dtos.DrugDetails{
		{
			ID: "D007251",
			Name: "Tamiflu",
			Substance: "hydrocodone",
		},
	}
	prescription := &dtos.PrescriptionDetails{
		Patient: patient,
		Disease: disease,
		Drugs: drugs,
	}
	body, err := json.Marshal(prescription)
	require.NoError(t, err)

	itemID, err := uuid.Parse("3f01c3e1-0c19-41b5-9a11-fcdfeb41864b")
	require.NoError(t, err)
	layout := "2006-01-02 15:04:05"
	createdAt, err := time.Parse(layout, "2025-10-15 14:30:00")
	require.NoError(t, err)
	item := &models.MedicalHistoryItem{
		ID: itemID,
		PatientID: patient.ID,
		DoctorID: id,
		CreatedAt: createdAt,
	}

	mockRepo.On("AddMedicalHistoryItem", patientID, id).Return(
		item, nil,
	).Once()

	req, err := http.NewRequest(
		http.MethodPost,
		"/med/prescribe",
		bytes.NewReader(body),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(
		t, http.StatusCreated, rec.Code,
		"Expected HTTP status 200 Created",
	)

	mockRepo.AssertExpectations(t)
}
