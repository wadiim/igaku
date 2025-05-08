package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/testcontainers/testcontainers-go/modules/compose"

	"context"
	"testing"
	"time"
)

func SetupTestServices(
	ctx context.Context,
	t *testing.T,
) (cleanup func()) {
	composeFilePath := "../../compose.yaml"

	stack, err := compose.NewDockerCompose(composeFilePath)
	require.NoError(t, err, "Failed to create stack")

	env := map[string]string{
		"POSTGRES_DB":		"igakudb",
		"POSTGRES_USER":	"igaku",
	}

	err = stack.
		WithEnv(env).
		WaitForService("db", wait.ForHealthCheck()).
		WaitForService(
			"api",
			wait.NewHTTPStrategy("/hello").
				WithPort("8080/tcp").
				WithStartupTimeout(180*time.Second).
				WithPollInterval(2*time.Second),
		).
		Up(ctx, compose.Wait(true))
	require.NoError(t, err, "Failed to start stack")

	cleanup = func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		assert.NoError(t, err, "Failed to stop stack")
	}

	return cleanup
}
