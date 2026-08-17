package sqlite

import (
	"context"
	"testing"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

func TestAreaRepositoryListsPaginatesAndSearches(t *testing.T) {
	repository, _, closeDatabase := newTestRepositories(t)
	defer closeDatabase()
	ctx := context.Background()

	areas, err := repository.ListAreas(ctx, domain.ListOptions{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if areas.Total != 3 || len(areas.Items) != 2 || areas.Items[0].Name != "Atividades de família" || areas.Items[1].Name != "Direito previdenciário" {
		t.Fatalf("unexpected areas result: %#v", areas)
	}

	secondPage, err := repository.ListAreas(ctx, domain.ListOptions{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].Name != "Tributário" {
		t.Fatalf("unexpected second page: %#v", secondPage)
	}

	search, err := repository.ListAreas(ctx, domain.ListOptions{Page: 1, PageSize: 20, Query: "famil"})
	if err != nil {
		t.Fatal(err)
	}
	if search.Total != 1 || search.Items[0].Name != "Atividades de família" {
		t.Fatalf("accent-insensitive prefix search failed: %#v", search)
	}
}
