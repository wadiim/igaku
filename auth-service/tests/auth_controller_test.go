package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"igaku/auth-service/controllers"
	"igaku/auth-service/services"
	"igaku/auth-service/tests/mocks"
	"igaku/commons/errors"
	"igaku/commons/models"
	"igaku/commons/utils"
)

func setupAuthRouter(
	t *testing.T,
	mockUserClient *mocks.UserClient,
	mockMailClient *mocks.MailClient,
	mockPatientClient *mocks.PatientClient,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	authService, err := services.NewAuthService(
		mockUserClient, mockMailClient, mockPatientClient, 1, "support@igaku.com",
	)
	require.NoError(t, err)
	authController := controllers.NewAuthController(authService)

	router := gin.Default()
	authController.RegisterRoutes(router)
	return router
}

func TestAuthController_Login_NoPasswordField(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

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

	mockUserClient.AssertNotCalled(t, "FindByUsername", mock.Anything)
}

func TestAuthController_Login_InvalidUsername(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	invalidUsername := "invalidUsername"
	mockUserClient.On("FindByUsername", invalidUsername).
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
	assert.Equal(t, "Invalid username or password", responseBody["error"])
}

func TestAuthController_Login_InvalidPassword(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	testUsername := "jdoe"
	expectedUser := &models.User{
		ID: uuid.New(),
		Username: testUsername,
		Password: "P@ssw0rd!",
	}
	mockUserClient.On("FindByUsername", testUsername).Return(expectedUser, nil).Once()

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
	assert.Equal(t, "Invalid username or password", responseBody["error"])
}

func TestAuthController_Login_Success(t *testing.T) {
	jwtSecretKey := []byte(os.Getenv("SECRET_KEY"))

	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

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
	mockUserClient.On("FindByUsername", testUsername).
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

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_InvalidParams(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

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

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_DuplicatedUsername(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	usrID := uuid.New()
	dupName := "jdoe"
	email := "unique@mail.com"
	nationalID := "44051401458"
	hashedPassword := "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6"

	existingUser := &models.User{
		ID: usrID,
		Username: dupName,
		Email: email,
		Password: hashedPassword,
		Role: models.Patient,
	}
	mockUserClient.On("FindByUsername", dupName).Return(existingUser, nil).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		dupName,
		email,
		nationalID,
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
	assert.Contains(
		t, responseBody["error"],
		fmt.Sprintf("Username '%s' already taken", existingUser.Username),
	)

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_DuplicatedPatientID(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	usrID := uuid.New()
	dupName := "jdoe"
	email := "unique@mail.com"
	nationalID := "44051401458"

	mockUserClient.On("FindByUsername", dupName).Return(
		nil, &errors.UserNotFoundError{},
	).Once()
	mockPatientClient.On("ValidateUniquePatient", mock.Anything).Return(
		&errors.DuplicatedIDError{ID: usrID},
	).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		dupName,
		email,
		nationalID,
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
	assert.Contains(
		t, strings.ToLower(responseBody["error"]),
		"duplicated id",
	)

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_DuplicatedPatientNationalID(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	dupName := "jdoe"
	email := "unique@mail.com"
	nationalID := "44051401458"

	mockUserClient.On("FindByUsername", dupName).Return(
		nil, &errors.UserNotFoundError{},
	).Once()
	mockPatientClient.On("ValidateUniquePatient", mock.Anything).Return(
		&errors.DuplicatedNationalIDError{NationalID: nationalID},
	).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		dupName,
		email,
		nationalID,
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
	log.Printf("%v", responseBody["error"])
	assert.Contains(
		t, strings.ToLower(responseBody["error"]),
		"duplicated id",
	)

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_InvalidNationalID(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	dupName := "jdoe"
	email := "unique@mail.com"
	nationalID := "44051401458"

	mockUserClient.On("FindByUsername", dupName).Return(
		nil, &errors.UserNotFoundError{},
	).Once()
	mockPatientClient.On("ValidateUniquePatient", mock.Anything).Return(
		&errors.InvalidNationalIDError{NationalID: nationalID},
	).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		dupName,
		email,
		nationalID,
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
	assert.Contains(
		t, strings.ToLower(responseBody["error"]),
		"invalid id",
	)

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_InvalidEmail(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	dupName := "jdoe"
	email := "duplicate@mail.com"
	nationalID := "44051401458"

	mockUserClient.On("FindByUsername", dupName).Return(
		nil, &errors.UserNotFoundError{},
	).Once()
	mockPatientClient.On("ValidateUniquePatient", mock.Anything).Return(
		nil,
	).Once()
	mockUserClient.On("Persist", mock.Anything).Return(
		&errors.EmailAlreadyTakenError{Email: email},
	).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		dupName,
		email,
		nationalID,
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
	assert.Contains(
		t, strings.ToLower(responseBody["error"]),
		fmt.Sprintf("email '%s' already taken", email),
	)

	mockUserClient.AssertExpectations(t)
}

func TestAuthController_Registration_Success(t *testing.T) {
	mockUserClient := new(mocks.UserClient)
	mockMailClient := new(mocks.MailClient)
	mockPatientClient := new(mocks.PatientClient)
	router := setupAuthRouter(t, mockUserClient, mockMailClient, mockPatientClient)

	username := "jdoe"
	email := "unique@mail.com"
	nationalID := "44051401458"

	mockUserClient.On("FindByUsername", username).Return(
		nil, &errors.UserNotFoundError{},
	).Once()
	mockPatientClient.On("ValidateUniquePatient", mock.Anything).Return(
		nil,
	).Once()
	mockUserClient.On("Persist", mock.Anything).Return(
		nil,
	).Once()
	mockPatientClient.On("AddPatientRecord", mock.Anything).Return(
		nil,
	).Once()
	mockMailClient.On("SendMail", []string{email}, mock.Anything).Return(
		nil,
	).Once()

	body := []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "national_id":"%s", "password":"%s"}`,
		username,
		email,
		nationalID,
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

	var responseBody map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.Contains(t, strings.ToLower(responseBody["error"]), "")

	mockUserClient.AssertExpectations(t)
}
