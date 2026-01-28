package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"context"
	"path/filepath"
	"os"
	"testing"
	"time"

	"igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	"igaku/med-service/controllers"
	"igaku/med-service/services"
	"igaku/med-service/tests/mocks"
	"igaku/med-service/utils"
)

func SetupTestDatabase(ctx context.Context, t *testing.T) (db *gorm.DB, cleanup func()) {
	t.Helper()

	pgContainer, err := tcpostgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:latest"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	cleanup = func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate PostgreSQL container: %v", err)
		}
	}

	// Get the connection string.
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get container connection string")

	// Connect GORM to the test database.
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database with GORM")

	err = utils.MigrateSchema(db)
	require.NoError(t, err, "Failed to migrate the database schema")

	// Read and execute the initialization script.
	initScriptDir, err := filepath.Abs("../resources")
	require.NoError(t, err, "Failed to get absolute path for resources/")
	initScriptPath := filepath.Join(initScriptDir, "init_test.sql")
	sqlBytes, err := os.ReadFile(initScriptPath)
	require.NoError(t, err, "Failed to read db init script: %s", initScriptPath)
	sqlScript := string(sqlBytes)
	tx := db.Exec(sqlScript)
	require.NoError(t, tx.Error, "Failed to execute init script")

	return db, cleanup
}

func SetupRouter(
	mockAPI *mocks.MockRxClassAPI,
	mockUserClient *mocks.UserClient,
	mockRepo *mocks.MockMedRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	medService := services.NewMedService(mockAPI, mockUserClient, mockRepo)
	medController := controllers.NewMedController(medService)

	router := gin.Default()
	medController.RegisterRoutes(router)

	return router
}

func GenDoctorToken(t *testing.T) string {
	doctor := &models.User{
		ID: uuid.New(),
		Username: "ghouse",
		Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
		Role: models.Doctor,
	}

	token, err := commonsUtils.GenerateJWTToken(
		doctor,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	require.NoError(t, err)

	return token
}
