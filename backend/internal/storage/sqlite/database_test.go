package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenDatabaseRejectsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := OpenDatabase(context.Background(), path); err == nil {
		t.Fatal("OpenDatabase should reject a missing file")
	}
}

func TestOpenDatabaseValidatesSchemaAndAllowsWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "writable.db")
	writable, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Exec(testSchema); err != nil {
		writable.Close()
		t.Fatal(err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}

	database, err := OpenDatabase(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`INSERT INTO user_fixed_costs(user_id, internet_cents) VALUES ('user-1', 1000)`); err != nil {
		t.Fatalf("writable database rejected a write: %v", err)
	}
}
