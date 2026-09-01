package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

func TestFixedCostsRepositoryGetsMissingRecordAndUpsertsPartialValues(t *testing.T) {
	repository, closeDatabase := newTestFixedCostsRepository(t)
	defer closeDatabase()
	ctx := context.Background()
	createFixedCostsUser(t, repository.database, 1, "user-1@example.com")
	createFixedCostsUser(t, repository.database, 2, "user-2@example.com")

	missing, exists, err := repository.GetFixedCosts(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if exists || missing.UserID != 1 || missing.InternetCents != 0 {
		t.Fatalf("unexpected missing result: %#v, exists=%v", missing, exists)
	}

	internet := int64(15000)
	created, err := repository.UpsertFixedCosts(ctx, domain.FixedCostsPatch{
		UserID: 1, InternetCents: &internet,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.InternetCents != internet || created.PhoneCents != 0 {
		t.Fatalf("unexpected created costs: %#v", created)
	}

	phone := int64(7000)
	updated, err := repository.UpsertFixedCosts(ctx, domain.FixedCostsPatch{
		UserID: 1, PhoneCents: &phone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.InternetCents != internet || updated.PhoneCents != phone {
		t.Fatalf("partial update did not preserve values: %#v", updated)
	}

	otherInternet := int64(9000)
	if _, err := repository.UpsertFixedCosts(ctx, domain.FixedCostsPatch{
		UserID: 2, InternetCents: &otherInternet,
	}); err != nil {
		t.Fatal(err)
	}
	first, exists, err := repository.GetFixedCosts(ctx, 1)
	if err != nil || !exists || first.InternetCents != internet {
		t.Fatalf("users were not isolated: %#v, exists=%v, err=%v", first, exists, err)
	}
}

func TestFixedCostsRepositoryRejectsNegativeValues(t *testing.T) {
	repository, closeDatabase := newTestFixedCostsRepository(t)
	defer closeDatabase()
	createFixedCostsUser(t, repository.database, 1, "user-1@example.com")
	negative := int64(-1)
	if _, err := repository.UpsertFixedCosts(context.Background(), domain.FixedCostsPatch{
		UserID: 1, InternetCents: &negative,
	}); err == nil {
		t.Fatal("negative cost should violate the database constraint")
	}
}

func createFixedCostsUser(t *testing.T, database *sql.DB, id int64, email string) {
	t.Helper()
	if _, err := database.Exec(`
		INSERT INTO users(id, email, name, password_hash, created_at)
		VALUES (?, ?, 'User', 'hash', 1700000000)
	`, id, email); err != nil {
		t.Fatal(err)
	}
}

func newTestFixedCostsRepository(t *testing.T) (*FixedCostsRepository, func()) {
	t.Helper()
	database, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "fixed-costs.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(testSchema); err != nil {
		database.Close()
		t.Fatal(err)
	}
	return NewFixedCostsRepository(database), func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	}
}
