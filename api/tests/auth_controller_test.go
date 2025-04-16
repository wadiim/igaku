package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"igaku/controllers"
	"igaku/errors"
	"igaku/models"
	"igaku/services"
	"igaku/utils"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)

	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}

	r1 := args.Error(1)

	return r0, r1
}

func setupAuthRouter(mockRepo *MockUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	authService := services.NewAuthService(mockRepo)
	authController := controllers.NewAuthController(authService)

	router := gin.Default()
	authController.RegisterRoutes(router)
	return router
}

func TestAuthController_Login_NoPasswordField(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupAuthRouter(mockRepo)

	body := []byte(`{"username":"jdoe"}`)
	req, err := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
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
	assert.Equal(t, "Invalid request payload", responseBody["error"])

	mockRepo.AssertNotCalled(t, "FindByUsername", mock.Anything)
}

func TestAuthController_Login_InvalidUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupAuthRouter(mockRepo)

	invalidUsername := "invalidUsername"
	mockRepo.On("FindByUsername", invalidUsername).
		Return(nil, &errors.UserNotFoundError{}).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"P@ssw0rd!"}`,
		invalidUsername,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid login or password", responseBody["error"])
}

func TestAuthController_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupAuthRouter(mockRepo)

	testUsername := "jdoe"
	expectedUser := &models.User{
		ID: uuid.New(),
		Username: testUsername,
		Password: "P@ssw0rd!",
	}
	mockRepo.On("FindByUsername", testUsername).Return(expectedUser, nil).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"invalidPassword"}`,
		testUsername,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid login or password", responseBody["error"])
}

func TestAuthController_Login_Success(t *testing.T) {
	jwtSecretKey := []byte(os.Getenv("SECRET_KEY"))

	mockRepo := new(MockUserRepository)
	router := setupAuthRouter(mockRepo)

	testID := uuid.New()
	testUsername := "jdoe"
	testPassword := "P@ssw0rd!"
	hashedPassword := "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6"
	expectedUser := &models.User{
		ID: testID,
		Username: testUsername,
		Password: hashedPassword,
		Role: models.Patient,
	}
	mockRepo.On("FindByUsername", testUsername).
		Return(expectedUser, nil).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"%s"}`,
		testUsername,
		testPassword,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	responseToken := string(rec.Body.Bytes()[:])
	claims := utils.Claims{}
	_, err = jwt.ParseWithClaims(
		responseToken,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Role, claims.Role)
	assert.Equal(t, "igaku", claims.Issuer)
	assert.Equal(t, expectedUser.ID.String(), claims.Subject)

	mockRepo.AssertExpectations(t)
}
