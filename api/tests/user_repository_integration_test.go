//go:build integration

package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"errors"
	"strings"
	"testing"

	igakuErrors "igaku/errors"
	"igaku/models"
	"igaku/repositories"
	"igaku/utils"
	testUtils "igaku/tests/utils"
)

func TestGormUserRepository(t *testing.T) {
	t.Run("FindByID_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failet to parse target UUID")

		usr, err := repo.FindByID(targetID)

		assert.NoError(
			t, err, "Expected no error finding existing user",
		)
		require.NotNil(t, usr, "Expected user to be found")
		assert.Equal(
			t, targetID, usr.ID, "Expected user ID to match",
		)
		assert.Equal(
			t, "jdoe", usr.Username,
			"Expected username to match",
		)
		assert.Equal(
			t,
			"$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			usr.Password,
			"Expected user password hash to match",
		)
		assert.Equal(
			t, models.Patient, usr.Role,
			"Expected user role to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		nonExistentID := uuid.New()

		usr, err := repo.FindByID(nonExistentID)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent user",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.UserNotFoundError{}),
			"Expected UserNotFoundError",
		)
		assert.Nil(
			t, usr,
			"Expected user to be nil when not found",
		)
	})

	t.Run("FindByUsername_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		targetUsername := "jdoe"

		usr, err := repo.FindByUsername(targetUsername)

		assert.NoError(
			t, err, "Expected no error finding existing user",
		)
		require.NotNil(t, usr, "Expected user to be found")
		assert.Equal(
			t, targetUsername, usr.Username,
			"Expected username to match",
		)
		targetID, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		require.NoError(t, err, "Failet to parse target UUID")
		assert.Equal(
			t, targetID, usr.ID,
			"Expected user ID to match",
		)
		assert.Equal(
			t,
			"$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			usr.Password,
			"Expected user password hash to match",
		)
		assert.Equal(
			t, models.Patient, usr.Role,
			"Expected user role to match",
		)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		nonExistentUsername := "nonExistentUserName"

		usr, err := repo.FindByUsername(nonExistentUsername)

		assert.Error(
			t, err,
			"Expected an error when finding non-existent user",
		)
		assert.True(
			t, errors.Is(err, &igakuErrors.UserNotFoundError{}),
			"Expected UserNotFoundError",
		)
		assert.Nil(
			t, usr,
			"Expected user to be nil when not found",
		)
	})

	t.Run("FindAll_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		userCount := 5
		list, err := repo.FindAll(
			0, userCount, models.Username, utils.Asc,
		)

		assert.NoError(
			t, err, "Expected no error finding users",
		)
		assert.NotNil(
			t, list, "Expected the user list not to be `nil`",
		)
		assert.Equal(
			t, userCount, len(list),
		)

		assert.Equal(t,
			list[0].ID.String(),
			"99ab51c4-a544-4352-a8df-4632ff8b105d",
		) // admin
		assert.Equal(t,
			list[1].ID.String(),
			"1f783647-4e06-4493-ade7-3b97d7e353dd",
		) // denji
		assert.Equal(t,
			list[2].ID.String(),
			"2a0de906-d3b2-4161-9672-bdfeab141c6c",
		) // fern
		assert.Equal(t,
			list[3].ID.String(),
			"91c1c531-2c0c-4f71-a6f7-ecd5377329fc",
		) // frieren
		assert.Equal(t,
			list[4].ID.String(),
			"e2c66717-12bb-4b6a-b7b6-3be939e170ad",
		) // ghouse
	})

	t.Run("CountAll_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		count, err := repo.CountAll()

		assert.NoError(
			t, err, "Expected no error counting users",
		)
		assert.Equal(
			t, int64(18), count,
		)

		// TODO: When `Persist()` will be implemented, test if
		// `CountAll()` still returns correct results when the count
		// changes, since currently `return 18, nil` would be enough
		// to pass.
	})

	t.Run("Persist_InvalidUser", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		id, err := uuid.Parse("0b6f13da-efb9-4221-9e89-e2729ae90030")
		assert.NoError(t, err)

		// Duplicated ID
		user := models.User{
			ID: id,
			Username: "validUsername",
			Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			Role: models.Patient,
		}
		err = repo.Persist(&user)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid user")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated id")

		id, err = uuid.Parse("358395da-2059-4768-ab9d-e41daf54af7d")
		assert.NoError(t, err)

		// Duplicated Username
		user = models.User{
			ID: id,
			Username: "jdoe",
			Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			Role: models.Patient,
		}
		err = repo.Persist(&user)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid user")
		assert.Contains(t, strings.ToLower(err.Error()), "duplicated username")
	})

	t.Run("Persist_Success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db, cleanup := testUtils.SetupTestDatabase(ctx, t)
		defer cleanup()

		repo := repositories.NewGormUserRepository(db)

		id, err := uuid.Parse("263bd7aa-46d1-4253-9232-fba5d68e161c")
		assert.NoError(t, err)

		user := models.User{
			ID: id,
			Username: "unique",
			Password: "$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6",
			Role: models.Patient,
		}
		err = repo.Persist(&user)
		assert.NoError(t, err)

		found_user, err := repo.FindByID(id)
		assert.NoError(t, err)

		assert.Equal(t, user.ID, found_user.ID)
		assert.Equal(t, user.Username, found_user.Username)
		assert.Equal(t, user.Role, found_user.Role)
	})
}
