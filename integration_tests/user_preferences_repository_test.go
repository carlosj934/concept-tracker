package integration_tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"concept-tracker/internal/repository"
)

func TestUserPreferencesRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("creates preferences if not exists", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		repo := repository.NewUserPreference(testPool)
		prefs, err := repo.Update(ctx, "user_test123", "America/Los_Angeles")
		require.NoError(t, err)
		require.Equal(t, "user_test123", prefs.UserID)
		require.Equal(t, "America/Los_Angeles", prefs.Timezone)
		require.NotZero(t, prefs.UpdatedAt)
	})

	t.Run("updates existing preferences", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		repo := repository.NewUserPreference(testPool)

		_, err := repo.Update(ctx, "user_test123", "America/Los_Angeles")
		require.NoError(t, err)

		updated, err := repo.Update(ctx, "user_test123", "America/New_York")
		require.NoError(t, err)
		require.Equal(t, "America/New_York", updated.Timezone)
	})
}

func TestUserPreferencesRepository_GetUserPreferences(t *testing.T) {
	ctx := context.Background()

	t.Run("returns preferences for user", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		repo := repository.NewUserPreference(testPool)

		_, err := repo.Update(ctx, "user_test123", "America/Los_Angeles")
		require.NoError(t, err)

		prefs, err := repo.GetUserPreferences(ctx, "user_test123")
		require.NoError(t, err)
		require.Equal(t, "user_test123", prefs.UserID)
		require.Equal(t, "America/Los_Angeles", prefs.Timezone)
	})

	t.Run("returns error when preferences do not exist", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		repo := repository.NewUserPreference(testPool)

		_, err := repo.GetUserPreferences(ctx, "user_nobody")
		require.Error(t, err)
	})
}
