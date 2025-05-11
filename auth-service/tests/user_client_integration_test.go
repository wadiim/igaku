//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"igaku/auth-service/clients"
	"igaku/commons/models"
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

func TestUserClient(t *testing.T) {
	t.Run("TestSetup", func(t *testing.T) {
		url := "http://localhost:8080/hello"

		res, err := http.Get(url)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("TestFindByUsername_NotFound", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		userClient := clients.NewUserClient(url)
		user, err := userClient.FindByUsername("nonexistinguser")
		assert.Nil(t, user, "Expected no user to be returned")
		require.NotNil(t, err, "Expected error when finding non-existend user")
		assert.Contains(t, strings.ToLower(err.Error()), "user not found")
	})

	t.Run("TestFindByUsername_Success", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		userClient := clients.NewUserClient(url)

		user, err := userClient.FindByUsername("jdoe")
		assert.Nil(t, err, "Expected no error when finding existing user")
		require.NotNil(t, user, "Expected the user to be found")
		assert.Equal(
			t, user.Username, "jdoe", "Expected username to match",
		)
		assert.Equal(
			t,
			user.ID.String(),
			"0b6f13da-efb9-4221-9e89-e2729ae90030",
			"Expected user ID to match",
		)
		assert.Equal(
			t,
			user.Password,
			"$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			"Expected user password hash to match",
		)
		assert.Equal(
			t, user.Role, models.Patient,
			"Expected user role to match",
		)
	})

	t.Run("TestPersist_Success", func(t *testing.T) {
		url := "amqp://rabbit:tibbar@localhost:5672/"

		userClient := clients.NewUserClient(url)

		id, err := uuid.Parse("263bd7aa-46d1-4253-9232-fba5d68e161c")
		require.NoError(t, err)
		username := "unique"
		password := "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6"

		expected_user := models.User{
			ID: id,
			Username: username,
			Password: password,
			Role: models.Patient,
		}

		err = userClient.Persist(&expected_user)
		require.NoError(t, err)

		user, err := userClient.FindByUsername(username)
		assert.Nil(t, err, "Expected no error when finding newly-created user")
		require.NotNil(t, user, "Expected the newly-created user to be found")

		assert.Equal(
			t, user.Username, username, "Expected username to match",
		)
		assert.Equal(
			t, user.ID, id, "Expected user ID to match",
		)
		assert.Equal(
			t, user.Password, password,
			"Expected user password hash to match",
		)
		assert.Equal(
			t, user.Role, models.Patient,
			"Expected user role to match",
		)
	})
}
