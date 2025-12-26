//go:build integration

package tests

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	testUtils "igaku/auth-service/tests/utils"
)

func TestMain(m *testing.M) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelCtx()

	cleanup, err := testUtils.SetupTestServices(ctx)
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	exitCode := m.Run()

	if cleanup != nil {
		cleanup()
	}

	os.Exit(exitCode)
}
