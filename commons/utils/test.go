package utils

import (
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type schemaMigrator func(*gorm.DB) error

func SetupTestDatabase(
	ctx context.Context, t *testing.T, migrateSchema schemaMigrator,
) (db *gorm.DB, cleanup func()) {
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
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database with GORM")

	err = migrateSchema(db)
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
