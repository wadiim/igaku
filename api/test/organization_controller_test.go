package test

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"igaku/controllers"
	"igaku/models"
	"igaku/services"
)

type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) FindByID(id uuid.UUID) (*models.Organization, error) {
	args := m.Called(id)

	var r0 *models.Organization
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Organization)
	}

	r1 := args.Error(1)

	return r0, r1
}

func setupRouter(mockRepo *MockOrganizationRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	orgService := services.NewOrganizationService(mockRepo)
	orgController := controllers.NewOrganizationController(orgService)

	router := gin.Default()
	orgController.RegisterRoutes(router)
	return router
}

func TestOrganizationController_GetByID_Success(t *testing.T) {
	mockRepo := new(MockOrganizationRepository)
	router := setupRouter(mockRepo)

	testOrgID := uuid.New()
	expectedOrg := &models.Organization{
		ID: testOrgID,
		Name: "Test Org",
		Address: "42 Mock St",
	}

	mockRepo.On("FindByID", testOrgID).Return(expectedOrg, nil).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/organizations/%s", testOrgID.String()),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var responseOrg models.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &responseOrg)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrg.ID, responseOrg.ID)
	assert.Equal(t, expectedOrg.Name, responseOrg.Name)
	assert.Equal(t, expectedOrg.Address, responseOrg.Address)

	mockRepo.AssertExpectations(t)
}

func TestOrganizationController_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockOrganizationRepository)
	router := setupRouter(mockRepo)

	testOrgID := uuid.New()

	mockRepo.On("FindByID", testOrgID).Return(nil, gorm.ErrRecordNotFound).Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/organizations/%s", testOrgID.String()),
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
	assert.Equal(t, "Organization not found", responseBody["error"])

	mockRepo.AssertExpectations(t)
}

func TestOrganizationController_GetByID_InvalidUUID(t *testing.T) {
	mockRepo := new(MockOrganizationRepository)
	router := setupRouter(mockRepo)

	invalidUUID := "not-a-UUID"

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/organizations/%s", invalidUUID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid UUID format", responseBody["error"])

	mockRepo.AssertNotCalled(t, "FindByID", mock.Anything)
}
