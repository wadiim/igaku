package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"igaku/commons/models"
	"igaku/med-service/controllers"
	"igaku/med-service/errors"
	"igaku/med-service/services"
)

type MockDiseaseRepository struct {
	mock.Mock
}

func (m *MockDiseaseRepository) FindByName(name string) ([]*models.Disease, error) {
	args := m.Called(name)

	var r0 []*models.Disease
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*models.Disease)
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

func TestDiseaseController_GetByName_Success(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testDiseaseName := "Lupus"
	expectedDiseases := []*models.Disease{
		{ID: uuid.New(), RxNormID: "D000000", Name: "Lupus Vulgaris"},
		{ID: uuid.New(), RxNormID: "D000001", Name: "Lupus Nephritis"},
	}

	mockRepo.On("FindByName", testDiseaseName).Return(expectedDiseases, nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testDiseaseName),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var responseDiseases []*models.Disease
	err = json.Unmarshal(rec.Body.Bytes(), &responseDiseases)
	assert.NoError(t, err)
	assert.Equal(t, expectedDiseases[0].ID, responseDiseases[0].ID)
	assert.Equal(t, expectedDiseases[0].RxNormID, responseDiseases[0].RxNormID)
	assert.Equal(t, expectedDiseases[0].Name, responseDiseases[0].Name)

	assert.Equal(t, expectedDiseases[1].ID, responseDiseases[1].ID)
	assert.Equal(t, expectedDiseases[1].RxNormID, responseDiseases[1].RxNormID)
	assert.Equal(t, expectedDiseases[1].Name, responseDiseases[1].Name)

	mockRepo.AssertExpectations(t)
}

func TestDiseaseController_GetByName_NotFound(t *testing.T) {
	mockRepo := new(MockDiseaseRepository)
	router := setupRouter(mockRepo)

	testDiseaseName := "Wilson"

	mockRepo.On("FindByName", testDiseaseName).Return(
		nil, &errors.DiseaseNotFoundError{},
	).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/med/disease/%s", testDiseaseName),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Disease not found", responseBody["error"])

	mockRepo.AssertExpectations(t)
}
