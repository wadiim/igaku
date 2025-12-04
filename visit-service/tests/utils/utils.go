package utils

import (
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"igaku/visit-service/utils"
)

func SetupTestServices(
	ctx context.Context, nominatimURL string,
) (func(), error) {
	composeFilePath := "../../compose.yaml"

	stack, err := compose.NewDockerCompose(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %w", err)
	}

	env := map[string]string{
		"RABBITMQ_USER":	"rabbit",
		"RABBITMQ_PASS":	"tibbar",
		"USER_DB_NAME":		"userdb",
		"USER_DB_USER":		"user",
		"USER_DB_PASSWORD":	"P@ssw0rd!",
		"VISIT_DB_NAME":	"visitdb",
		"VISIT_DB_USER":	"visit",
		"VISIT_DB_PASSWORD":	"P@ssw0rd!",
		"CLIENT_PORT":		"8080",
		"STACK_VERSION":	"9.0.0",
		"GRAFANA_USER_ID":	"",
		"GRAFANA_TOKEN":	"",
		"GRAFANA_URL":		"",
		"NOMINATIM_URL":	nominatimURL,
	}

	err = stack.
		WithEnv(env).
		WaitForService("visit-db", wait.ForHealthCheck()).
		WaitForService("rabbitmq", wait.ForHealthCheck()).
		WaitForService(
			"nginx",
			wait.NewHTTPStrategy("/geo/health").
				WithPort("4000/tcp").
				WithStartupTimeout(200*time.Second).
				WithPollInterval(4*time.Second),
		).
		Up(ctx, compose.Wait(true))
	if err != nil {
		return nil, fmt.Errorf("failed to start stack: %w", err)
	}

	cleanup := func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
		)
		if err != nil {
			log.Printf("Failed to stop stack: %w", err)
		}
	}

	return cleanup, nil
}

// TODO: Move to `commons`
func SetupTestDatabase(ctx context.Context, t *testing.T) (db *gorm.DB, cleanup func()) {
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
