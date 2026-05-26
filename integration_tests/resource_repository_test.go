package integration_tests

import (
	"context"
	"testing"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestResourceRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewResource(testPool)
	resource, err := repo.Create(ctx, "user_test123", concept.ID, domain.ConceptResource{
		Provider:   "google_docs",
		ExternalID: "abc123",
		URL:        "https://docs.google.com/document/d/abc123",
		Title:      "Go Concurrency Notes",
		Meta:       []byte(`{}`),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resource.ID)
	require.Equal(t, "user_test123", resource.UserID)
	require.Equal(t, concept.ID, resource.ConceptID)
	require.Equal(t, "google_docs", resource.Provider)
	require.Equal(t, "abc123", resource.ExternalID)
	require.Equal(t, "https://docs.google.com/document/d/abc123", resource.URL)
	require.Equal(t, "Go Concurrency Notes", resource.Title)
}

func TestResourceRepository_ListConceptResources(t *testing.T) {
	ctx := context.Background()

	t.Run("returns resources for concept", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewResource(testPool)
		r1, err := repo.Create(ctx, "user_test123", concept.ID, domain.ConceptResource{
			Provider: "google_docs", ExternalID: "id1",
			URL: "https://docs.google.com/1", Title: "Doc 1", Meta: []byte(`{}`),
		})
		require.NoError(t, err)

		r2, err := repo.Create(ctx, "user_test123", concept.ID, domain.ConceptResource{
			Provider: "google_docs", ExternalID: "id2",
			URL: "https://docs.google.com/2", Title: "Doc 2", Meta: []byte(`{}`),
		})
		require.NoError(t, err)

		resources, err := repo.ListConceptResources(ctx, "user_test123", concept.ID)
		require.NoError(t, err)
		require.Len(t, resources, 2)

		ids := []string{resources[0].ID, resources[1].ID}
		require.Contains(t, ids, r1.ID)
		require.Contains(t, ids, r2.ID)
	})

	t.Run("returns empty for concept with no resources", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewResource(testPool)
		resources, err := repo.ListConceptResources(ctx, "user_test123", concept.ID)
		require.NoError(t, err)
		require.Empty(t, resources)
	})
}

func TestResourceRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewResource(testPool)
	created, err := repo.Create(ctx, "user_test123", concept.ID, domain.ConceptResource{
		Provider: "google_docs", ExternalID: "id1",
		URL: "https://docs.google.com/old", Title: "Old Title", Meta: []byte(`{}`),
	})
	require.NoError(t, err)

	newURL := "https://docs.google.com/new"
	newTitle := "New Title"

	updated, err := repo.Update(ctx, "user_test123", created.ID, &newURL, &newTitle)
	require.NoError(t, err)
	require.Equal(t, "https://docs.google.com/new", updated.URL)
	require.Equal(t, "New Title", updated.Title)
	require.Equal(t, created.ID, updated.ID)
}

func TestResourceRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewResource(testPool)
	created, err := repo.Create(ctx, "user_test123", concept.ID, domain.ConceptResource{
		Provider: "google_docs", ExternalID: "id1",
		URL: "https://docs.google.com/1", Title: "Doc 1", Meta: []byte(`{}`),
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, "user_test123", created.ID)
	require.NoError(t, err)

	resources, err := repo.ListConceptResources(ctx, "user_test123", concept.ID)
	require.NoError(t, err)
	require.Empty(t, resources)
}
