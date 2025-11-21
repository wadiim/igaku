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
		"MAIL_ENABLED":			"",
		"CLIENT_PORT":			"8080",
		"RABBITMQ_USER":		"rabbit",
		"RABBITMQ_PASS":		"tibbar",
		"USER_DB_NAME":			"userdb",
		"USER_DB_USER":			"user",
		"USER_DB_PASSWORD":		"P@ssw0rd!",
		"ENCOUNTER_DB_NAME":		"visitdb",
		"ENCOUNTER_DB_USER":		"visit",
		"ENCOUNTER_DB_PASSWORD":	"P@ssw0rd!",
		"JWT_TOKEN_DURATION_IN_HOURS":	"1",
	}

	err = stack.
		WithEnv(env).
		WaitForService("user-db", wait.ForHealthCheck()).
		WaitForService("rabbitmq", wait.ForHealthCheck()).
		WaitForService(
			"nginx",
			wait.NewHTTPStrategy("/user/health").
				WithPort("4000/tcp").
				WithStartupTimeout(200*time.Second).
				WithPollInterval(4*time.Second),
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
