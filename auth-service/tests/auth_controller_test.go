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

	"igaku/auth-service/controllers"
	"igaku/auth-service/services"
	"igaku/auth-service/tests/mocks"
	"igaku/commons/errors"
	"igaku/commons/models"
	"igaku/commons/utils"
)

func setupAuthRouter(mockClient *mocks.UserClient) *gin.Engine {
	gin.SetMode(gin.TestMode)

	authService := services.NewAuthService(mockClient)
	authController := controllers.NewAuthController(authService)

	router := gin.Default()
	authController.RegisterRoutes(router)
	return router
}

func TestAuthController_Login_NoPasswordField(t *testing.T) {
	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	body := []byte(`{"username":"jdoe"}`)
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/login",
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

	mockClient.AssertNotCalled(t, "FindByUsername", mock.Anything)
}

func TestAuthController_Login_InvalidUsername(t *testing.T) {
	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	invalidUsername := "invalidUsername"
	mockClient.On("FindByUsername", invalidUsername).
		Return(nil, &errors.UserNotFoundError{}).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"P@ssw0rd!"}`,
		invalidUsername,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/login",
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
	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	testUsername := "jdoe"
	expectedUser := &models.User{
		ID: uuid.New(),
		Username: testUsername,
		Password: "P@ssw0rd!",
	}
	mockClient.On("FindByUsername", testUsername).Return(expectedUser, nil).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"invalidPassword"}`,
		testUsername,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/login",
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

	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

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
	mockClient.On("FindByUsername", testUsername).
		Return(expectedUser, nil).Once()

	body := []byte(fmt.Sprintf(
		`{"username":"%s", "password":"%s"}`,
		testUsername,
		testPassword,
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/login",
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

	mockClient.AssertExpectations(t)
}

func TestAuthController_Registration_InvalidParams(t *testing.T) {
	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	body := []byte(`{"foo":"bar"}`)
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/register",
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

	mockClient.AssertExpectations(t)
}

func TestAuthController_Registration_DuplicatedUsername(t *testing.T) {
	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	usrID := uuid.New()
	dupName := "jdoe"
	email := "unique@mail.com"
	hashedPassword := "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6"

	existingUser := &models.User{
		ID: usrID,
		Username: dupName,
		Email: email,
		Password: hashedPassword,
		Role: models.Patient,
	}
	mockClient.On("FindByUsername", dupName).Return(existingUser, nil).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "password":"%s"}`,
		dupName,
		email,
		"P@ssw0rd!",
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/register",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "Username already taken")

	mockClient.AssertExpectations(t)
}

func TestAuthController_Registration_Success(t *testing.T) {
	jwtSecretKey := []byte(os.Getenv("SECRET_KEY"))

	mockClient := new(mocks.UserClient)
	router := setupAuthRouter(mockClient)

	usrName := "newuser"
	usrEmail := "newuser@mail.com"

	mockClient.On("FindByUsername", usrName).
		Return(nil, &errors.UserNotFoundError{}).Once()
	mockClient.On("Persist", mock.Anything).Return(nil).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "password":"%s"}`,
		usrName,
		usrEmail,
		"P@ssw0rd!",
	))
	req, err := http.NewRequest(
		http.MethodPost,
		"/auth/register",
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
	assert.Equal(t, models.Patient, claims.Role)
	assert.Equal(t, "igaku", claims.Issuer)
	// We do not know the generated ID

	mockClient.AssertExpectations(t)
}
