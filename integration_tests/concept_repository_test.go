package integration_tests

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

func TestConceptRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("root concept", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDB(t, ctx)
		})
		repo := repository.New(testPool)
		concept := domain.Concept{
			ParentID:    nil,
			Name:        "test-concept",
			Description: new("test description"),
		}

		// create root concept
		c, err := repo.Create(ctx, "user_test123", concept)
		require.NoError(t, err)

		require.Equal(t, "test-concept", c.Name)
		require.Equal(t, "user_test123", c.UserID)
		require.Nil(t, c.ParentID)
		require.NotEmpty(t, c.ID)
		require.NotZero(t, c.CreatedAt)
		require.Equal(t, "test description", *c.Description)

		var path domain.ConceptPath
		err = testPool.QueryRow(ctx, `
		SELECT ancestor_id, descendant_id, depth
		FROM concept_paths
		WHERE descendant_id = $1
		`, c.ID).Scan(&path.AncestorID, &path.DescendantID, &path.Depth)
		require.NoError(t, err)

		// assert that values defined above are accurate once db has been queried
		require.Equal(t, c.ID, path.AncestorID)
		require.Equal(t, c.ID, path.DescendantID)
		require.Equal(t, int64(0), path.Depth)
	})

	t.Run("child concept", func(t *testing.T) {
		t.Cleanup(func() {
			restoreDB(t, ctx)
		})

		// create parent
		repo := repository.New(testPool)
		parent := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)

		// create child
		cconcept := domain.Concept{
			ParentID:    &parent.ID,
			Name:        "test-concept",
			Description: new("test description"),
		}

		child, err := repo.Create(ctx, "user_test123", cconcept)
		require.NoError(t, err)

		rows, err := testPool.Query(ctx, `
		SELECT ancestor_id, descendant_id, depth
		FROM concept_paths
		WHERE descendant_id = $1
		`, child.ID)
		require.NoError(t, err)
		defer rows.Close()

		var conceptPath []domain.ConceptPath
		for rows.Next() {
			var path domain.ConceptPath
			err := rows.Scan(&path.AncestorID, &path.DescendantID, &path.Depth)
			require.NoError(t, err)
			conceptPath = append(conceptPath, path)
		}

		require.NoError(t, rows.Err())

		sort.Slice(conceptPath, func(i, j int) bool {
			return conceptPath[i].Depth < conceptPath[j].Depth
		})
		require.Len(t, conceptPath, 2)

		// assert self-row exists
		require.Equal(t, child.ID, conceptPath[0].AncestorID)
		require.Equal(t, child.ID, conceptPath[0].DescendantID)
		require.Equal(t, int64(0), conceptPath[0].Depth)

		// assert parent -> child row exists
		require.Equal(t, parent.ID, conceptPath[1].AncestorID)
		require.Equal(t, child.ID, conceptPath[1].DescendantID)
		require.Equal(t, int64(1), conceptPath[1].Depth)
	})
}

// helper to create a root concept for use in other tests
func createTestConcept(t *testing.T, ctx context.Context, repo repository.ConceptRepository, userID string, name string, parentID *string) domain.Concept {
	t.Helper()
	c, err := repo.Create(ctx, userID, domain.Concept{
		Name:     name,
		ParentID: parentID,
	})
	require.NoError(t, err)
	return c
}

func TestConceptRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns concept for correct user", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		created := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)

		got, err := repo.GetByID(ctx, "user_test123", created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, "golang", got.Name)
		require.Equal(t, "user_test123", got.UserID)
	})

	t.Run("returns error for wrong user", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		created := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)

		_, err := repo.GetByID(ctx, "user_other", created.ID)
		require.Error(t, err)
	})
}

func TestConceptRepository_ListRoots(t *testing.T) {
	ctx := context.Background()

	t.Run("returns only root concepts for user", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root1 := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		root2 := createTestConcept(t, ctx, repo, "user_test123", "rust", nil)
		// child — should NOT appear in roots
		createTestConcept(t, ctx, repo, "user_test123", "concurrency", &root1.ID)
		// different user — should NOT appear
		createTestConcept(t, ctx, repo, "user_other", "python", nil)

		roots, err := repo.ListRoots(ctx, "user_test123")
		require.NoError(t, err)
		require.Len(t, roots, 2)

		ids := []string{roots[0].ID, roots[1].ID}
		require.Contains(t, ids, root1.ID)
		require.Contains(t, ids, root2.ID)
	})

	t.Run("returns empty slice when user has no concepts", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		roots, err := repo.ListRoots(ctx, "user_nobody")
		require.NoError(t, err)
		require.Empty(t, roots)
	})
}

func TestConceptRepository_GetChildren(t *testing.T) {
	ctx := context.Background()

	t.Run("returns only direct children, not deeper descendants", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		child := createTestConcept(t, ctx, repo, "user_test123", "concurrency", &root.ID)
		// grandchild — should NOT appear
		createTestConcept(t, ctx, repo, "user_test123", "goroutines", &child.ID)

		children, err := repo.GetChildren(ctx, "user_test123", root.ID)
		require.NoError(t, err)
		require.Len(t, children, 1)
		require.Equal(t, child.ID, children[0].ID)
	})

	t.Run("returns empty slice for leaf concept", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)

		children, err := repo.GetChildren(ctx, "user_test123", root.ID)
		require.NoError(t, err)
		require.Empty(t, children)
	})
}

func TestConceptRepository_GetSubtree(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all descendants excluding root itself", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		child := createTestConcept(t, ctx, repo, "user_test123", "concurrency", &root.ID)
		grandchild := createTestConcept(t, ctx, repo, "user_test123", "goroutines", &child.ID)

		subtree, err := repo.GetSubtree(ctx, "user_test123", root.ID)
		require.NoError(t, err)
		require.Len(t, subtree, 2)

		ids := []string{subtree[0].ID, subtree[1].ID}
		require.Contains(t, ids, child.ID)
		require.Contains(t, ids, grandchild.ID)
		require.NotContains(t, ids, root.ID)
	})
}

func TestConceptRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("updates name and description", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		created := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		newDesc := "updated description"

		updated, err := repo.Update(ctx, "user_test123", created.ID, "go language", &newDesc)
		require.NoError(t, err)
		require.Equal(t, "go language", updated.Name)
		require.Equal(t, "updated description", *updated.Description)
		require.Equal(t, created.ID, updated.ID)
	})

	t.Run("returns error for wrong user", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		created := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		newDesc := "updated description"

		_, err := repo.Update(ctx, "user_other", created.ID, "go language", &newDesc)
		require.Error(t, err)
	})
}

func TestConceptRepository_Move(t *testing.T) {
	ctx := context.Background()

	t.Run("reparents concept and rebuilds closure table", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		// original tree: root1 -> child, root2 (standalone)
		root1 := createTestConcept(t, ctx, repo, "user_test123", "root1", nil)
		root2 := createTestConcept(t, ctx, repo, "user_test123", "root2", nil)
		child := createTestConcept(t, ctx, repo, "user_test123", "child", &root1.ID)

		// move child from root1 to root2
		err := repo.Move(ctx, "user_test123", child.ID, &root2.ID)
		require.NoError(t, err)

		// verify parent_id updated on concept
		moved, err := repo.GetByID(ctx, "user_test123", child.ID)
		require.NoError(t, err)
		require.Equal(t, root2.ID, *moved.ParentID)

		// verify closure table: child should now be connected to root2, not root1
		rows, err := testPool.Query(ctx, `
		SELECT ancestor_id, descendant_id, depth
		FROM concept_paths
		WHERE descendant_id = $1
		`, child.ID)
		require.NoError(t, err)
		defer rows.Close()

		var paths []domain.ConceptPath
		for rows.Next() {
			var p domain.ConceptPath
			err := rows.Scan(&p.AncestorID, &p.DescendantID, &p.Depth)
			require.NoError(t, err)
			paths = append(paths, p)
		}
		require.NoError(t, rows.Err())
		require.Len(t, paths, 2)

		ancestorIDs := []string{paths[0].AncestorID, paths[1].AncestorID}
		require.Contains(t, ancestorIDs, child.ID)    // self-row
		require.Contains(t, ancestorIDs, root2.ID)    // new parent row
		require.NotContains(t, ancestorIDs, root1.ID) // old parent gone
	})

	t.Run("move to root removes parent connection", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root := createTestConcept(t, ctx, repo, "user_test123", "root", nil)
		child := createTestConcept(t, ctx, repo, "user_test123", "child", &root.ID)

		// move child to root level (no parent)
		err := repo.Move(ctx, "user_test123", child.ID, nil)
		require.NoError(t, err)

		moved, err := repo.GetByID(ctx, "user_test123", child.ID)
		require.NoError(t, err)
		require.Nil(t, moved.ParentID)

		// only self-row should remain in concept_paths
		var count int
		err = testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM concept_paths WHERE descendant_id = $1
		`, child.ID).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}

func TestConceptRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes concept and entire subtree", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		root := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)
		child := createTestConcept(t, ctx, repo, "user_test123", "concurrency", &root.ID)
		grandchild := createTestConcept(t, ctx, repo, "user_test123", "goroutines", &child.ID)

		err := repo.Delete(ctx, "user_test123", root.ID)
		require.NoError(t, err)

		// all three should be gone
		_, err = repo.GetByID(ctx, "user_test123", root.ID)
		require.Error(t, err)

		_, err = repo.GetByID(ctx, "user_test123", child.ID)
		require.Error(t, err)

		_, err = repo.GetByID(ctx, "user_test123", grandchild.ID)
		require.Error(t, err)
	})

	t.Run("does not delete concepts belonging to other users", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })
		repo := repository.New(testPool)

		c := createTestConcept(t, ctx, repo, "user_test123", "golang", nil)

		err := repo.Delete(ctx, "user_other", c.ID)
		require.NoError(t, err)

		// concept should still exist for original user
		got, err := repo.GetByID(ctx, "user_test123", c.ID)
		require.NoError(t, err)
		require.Equal(t, c.ID, got.ID)
	})
}
