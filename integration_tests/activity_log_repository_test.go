package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

func TestActivityLogRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewActivityLog(testPool)
	dur := int64(30)
	notes := "covered goroutine lifecycle"
	loggedAt := time.Now().UTC().Truncate(time.Microsecond)

	log, err := repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
		ActivityType: "flashcards",
		DurationMins: &dur,
		Notes:        &notes,
		LoggedAt:     loggedAt,
	})
	require.NoError(t, err)
	require.NotEmpty(t, log.ID)
	require.Equal(t, "user_test123", log.UserID)
	require.Equal(t, concept.ID, log.ConceptID)
	require.Equal(t, "flashcards", log.ActivityType)
	require.Equal(t, int64(30), *log.DurationMins)
	require.Equal(t, "covered goroutine lifecycle", *log.Notes)
}

func TestActivityLogRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("returns logs newest first", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewActivityLog(testPool)
		now := time.Now().UTC()

		_, err := repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
			ActivityType: "reading",
			LoggedAt:     now.Add(-2 * time.Hour),
		})
		require.NoError(t, err)

		_, err = repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
			ActivityType: "practice",
			LoggedAt:     now.Add(-1 * time.Hour),
		})
		require.NoError(t, err)

		logs, err := repo.List(ctx, "user_test123", concept.ID, nil, 25)
		require.NoError(t, err)
		require.Len(t, logs, 2)
		// newest first
		require.Equal(t, "practice", logs[0].ActivityType)
		require.Equal(t, "reading", logs[1].ActivityType)
	})

	t.Run("cursor pagination returns next page", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewActivityLog(testPool)
		now := time.Now().UTC()

		for i := range 8 {
			_, err := repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
				ActivityType: "reading",
				LoggedAt:     now.Add(-time.Duration(i) * time.Hour),
			})
			require.NoError(t, err)
		}

		page1, err := repo.List(ctx, "user_test123", concept.ID, nil, 2)
		require.NoError(t, err)
		require.Len(t, page1, 3)

		// cursor built from last real result (index limit-1), not the peek item
		cursor := &domain.Cursor{
			LoggedAt: page1[1].LoggedAt,
			ID:       page1[1].ID,
		}

		page2, err := repo.List(ctx, "user_test123", concept.ID, cursor, 2)
		require.NoError(t, err)
		require.Len(t, page2, 3)

		// no overlap between pages
		page1IDs := map[string]bool{page1[0].ID: true, page1[1].ID: true}
		for _, l := range page2 {
			require.False(t, page1IDs[l.ID])
		}
	})
}

func TestActivityLogRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewActivityLog(testPool)
	created, err := repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
		ActivityType: "reading",
		LoggedAt:     time.Now().UTC(),
	})
	require.NoError(t, err)

	newType := "practice"
	newDur := int64(45)
	newNotes := "updated notes"

	updated, err := repo.Update(ctx, "user_test123", created.ID, &newType, &newDur, &newNotes, nil)
	require.NoError(t, err)
	require.Equal(t, "practice", updated.ActivityType)
	require.Equal(t, int64(45), *updated.DurationMins)
	require.Equal(t, "updated notes", *updated.Notes)
}

func TestActivityLogRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewActivityLog(testPool)
	created, err := repo.Create(ctx, "user_test123", concept.ID, domain.ActivityLog{
		ActivityType: "reading",
		LoggedAt:     time.Now().UTC(),
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, "user_test123", created.ID)
	require.NoError(t, err)

	logs, err := repo.List(ctx, "user_test123", concept.ID, nil, 25)
	require.NoError(t, err)
	require.Empty(t, logs)
}
