//go:build integration

package test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"igaku/repositories"
	"igaku/models"
)

func setupTestDatabase(ctx context.Context, t *testing.T) (db *gorm.DB, cleanup func()) {
	t.Helper()

	// Configure the PostgreSQL container.
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
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database with GORM")

	// Migrate the database schema.
	err = gormDB.AutoMigrate(&models.Organization{})
	require.NoError(t, err, "Failed to migrate the database schema")

	// Read and execute the initialization script.
	initScriptDir, err := filepath.Abs("../resources")
	require.NoError(t, err, "Failed to get absolute path for resources/")
	initScriptPath := filepath.Join(initScriptDir, "init_test.sql")
	sqlBytes, err := os.ReadFile(initScriptPath)
	require.NoError(t, err, "Failed to read db init script: %s", initScriptPath)
	sqlScript := string(sqlBytes)
	tx := gormDB.Exec(sqlScript)
	require.NoError(t, tx.Error, "Failed to execute init script")

	return gormDB, cleanup
}

func TestGormOrganizationRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := setupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormOrganizationRepository(db)

		targetID, err := uuid.Parse("86e6a1f3-d7aa-4e74-a20a-ea78bc13340b")
		require.NoError(t, err, "Failed to parse target UUID")

		org, err := repo.FindByID(targetID)

		assert.NoError(
			t, err, "Expected no error finding existing organization",
		)
		require.NotNil(t, org, "Expected organization to be found")
		assert.Equal(
			t, targetID, org.ID, "Expected organization ID to match",
		)
		assert.Equal(
			t, "The Lowell General Hospital", org.Name,
			"Expected organization name to match",
		)
		assert.Equal(
			t, "295 Varnum Ave", org.Address,
			"Expected organization address to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := setupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormOrganizationRepository(db)

		nonExistentID := uuid.New()

		org, err := repo.FindByID(nonExistentID)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent organization",
		)
		assert.True(
			t, errors.Is(err, gorm.ErrRecordNotFound),
			"Expected gorm.ErrRecordNotFound",
		)
		assert.Nil(
			t, org,
			"Expected organization to be nil when not found",
		)
	})
}
