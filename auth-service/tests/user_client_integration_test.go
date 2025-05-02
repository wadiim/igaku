//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"igaku/auth-service/clients"
	"igaku/commons/models"
	testUtils "igaku/auth-service/tests/utils"
)

func TestUserClient(t *testing.T) {
	t.Run("TestSetup", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		netName := "test-setup"
		network, netCleanup := testUtils.SetupTestNetwork(ctx, t, netName)
		defer netCleanup()

		db, dbCleanup := testUtils.SetupTestDatabase(ctx, t, network, netName)
		defer dbCleanup()

		client, clientCleanup := testUtils.SetupTestUserService(ctx, t, db, network, netName)
		defer clientCleanup()

		url := fmt.Sprintf("http://%s:%s/hello", client.Host, client.Port)
		res, err := http.Get(url)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("TestFindByUsername_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		netName := "findbyusername-notfound"
		network, netCleanup := testUtils.SetupTestNetwork(ctx, t, netName)
		defer netCleanup()

		db, dbCleanup := testUtils.SetupTestDatabase(ctx, t, network, netName)
		defer dbCleanup()

		client, clientCleanup := testUtils.SetupTestUserService(ctx, t, db, network, netName)
		defer clientCleanup()

		url := fmt.Sprintf("http://%s:%s", client.Host, client.Port)
		userClient := clients.NewUserClient(url)

		user, err := userClient.FindByUsername("nonexistinguser")
		assert.Nil(t, user, "Expected no user to be returned")
		require.NotNil(t, err, "Expected error when finding non-existend user")
		assert.Contains(t, strings.ToLower(err.Error()), "user not found")
	})

	t.Run("TestFindByUsername_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		netName := "findbyusername-success"
		network, netCleanup := testUtils.SetupTestNetwork(ctx, t, netName)
		defer netCleanup()

		db, dbCleanup := testUtils.SetupTestDatabase(ctx, t, network, netName)
		defer dbCleanup()

		client, clientCleanup := testUtils.SetupTestUserService(ctx, t, db, network, netName)
		defer clientCleanup()

		url := fmt.Sprintf("http://%s:%s", client.Host, client.Port)
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
		t.Parallel()
		ctx := context.Background()

		netName := "persist-success"
		network, netCleanup := testUtils.SetupTestNetwork(ctx, t, netName)
		defer netCleanup()

		db, dbCleanup := testUtils.SetupTestDatabase(ctx, t, network, netName)
		defer dbCleanup()

		client, clientCleanup := testUtils.SetupTestUserService(ctx, t, db, network, netName)
		defer clientCleanup()

		url := fmt.Sprintf("http://%s:%s", client.Host, client.Port)
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

// TODO: Reduce code duplication
