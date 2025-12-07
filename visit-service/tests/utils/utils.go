package utils

import (
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"context"
	"fmt"
	"log"
	"time"
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
		"NOMINATIM_TIMEOUT":	"2",
	}

	err = stack.
		WithEnv(env).
		WaitForService("visit-db", wait.ForHealthCheck()).
		WaitForService("rabbitmq", wait.ForHealthCheck()).
		WaitForService(
			"nginx",
			wait.NewHTTPStrategy("/geo/health").
				WithPort("4000/tcp").
				WithStartupTimeout(300*time.Second).
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
