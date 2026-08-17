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

func TestOpenDatabaseValidatesSchemaAndEnforcesReadOnlyMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read-only.db")
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

	readOnly, err := OpenDatabase(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if _, err := readOnly.Exec(`INSERT INTO areas(name) VALUES ('Proibida')`); err == nil {
		t.Fatal("read-only database accepted a write")
	}
}
