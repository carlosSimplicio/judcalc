package sqlite

import (
	"context"
	"testing"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

func TestServiceRepositorySearchesMultipleTerms(t *testing.T) {
	_, repository, closeDatabase := newTestRepositories(t)
	defer closeDatabase()

	result, err := repository.ListServices(context.Background(), domain.ListOptions{Page: 1, PageSize: 20, Query: "acao prev"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Items[0].Name != "Ação previdenciária comum" {
		t.Fatalf("multi-term search failed: %#v", result)
	}
}

func TestServiceRepositoryPreservesNullsAndReturnsEmptyPage(t *testing.T) {
	_, repository, closeDatabase := newTestRepositories(t)
	defer closeDatabase()
	ctx := context.Background()

	services, err := repository.ListServices(ctx, domain.ListOptions{Page: 1, PageSize: 20, Query: "consulta"})
	if err != nil {
		t.Fatal(err)
	}
	if services.Total != 1 || len(services.Items) != 1 {
		t.Fatalf("unexpected result: %#v", services)
	}
	service := services.Items[0]
	if service.AmountCents != nil || service.PercentageMin != nil || service.PercentageMax != nil {
		t.Fatalf("null values were not preserved: %#v", service)
	}

	empty, err := repository.ListServices(ctx, domain.ListOptions{Page: 99, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if empty.Total != 3 || len(empty.Items) != 0 || empty.Items == nil {
		t.Fatalf("unexpected empty page: %#v", empty)
	}
}

func TestServiceRepositoryEscapesFTSSyntax(t *testing.T) {
	_, repository, closeDatabase := newTestRepositories(t)
	defer closeDatabase()

	queries := []string{`"`, `*`, `acao OR`, `NEAR(acao)`}
	for _, query := range queries {
		if _, err := repository.ListServices(context.Background(), domain.ListOptions{Page: 1, PageSize: 20, Query: query}); err != nil {
			t.Fatalf("query %q caused an FTS error: %v", query, err)
		}
	}
}
