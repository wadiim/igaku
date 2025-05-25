package utils

import (
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/testcontainers/testcontainers-go/modules/compose"

	"context"
	"fmt"
	"log"
	"time"
)

func SetupTestServices(ctx context.Context) (func(), error) {
	composeFilePath := "../../compose.yaml"

	stack, err := compose.NewDockerCompose(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create stack: %w", err)
	}

	env := map[string]string{
		"RABBITMQ_USER":		"rabbit",
		"RABBITMQ_PASS":		"tibbar",
		"USER_DB_NAME":			"userdb",
		"USER_DB_USER":			"user",
		"USER_DB_PASSWORD":		"P@ssw0rd!",
		"ENCOUNTER_DB_NAME":		"encounterdb",
		"ENCOUNTER_DB_USER":		"encounter",
		"ENCOUNTER_DB_PASSWORD":	"P@ssw0rd!",
		"JWT_TOKEN_DURATION_IN_HOURS":	"1",
	}

	err = stack.
		WithEnv(env).
		WaitForService("user-db", wait.ForHealthCheck()).
		WaitForService("rabbitmq", wait.ForHealthCheck()).
		WaitForService(
			"user",
			wait.NewHTTPStrategy("/hello").
				WithPort("8080/tcp").
				WithStartupTimeout(180*time.Second).
				WithPollInterval(2*time.Second),
		).
		Up(ctx, compose.Wait(true))
	if err != nil {
		return nil, fmt.Errorf("Failed to start stack: %w", err)
	}

	cleanup := func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %w", err)
		}
	}

	return cleanup, nil
}
