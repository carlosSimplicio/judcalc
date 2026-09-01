package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	result, err := database.Exec(`INSERT INTO users(email, name, password_hash, created_at) VALUES ('user@example.com', 'User', 'hash', 1700000000)`)
	if err != nil {
		t.Fatalf("writable database rejected a user write: %v", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO user_fixed_costs(user_id, internet_cents) VALUES (?, 1000)`, userID); err != nil {
		t.Fatalf("writable database rejected a write: %v", err)
	}
}

func TestOpenDatabaseRejectsSchemaWithoutAuthenticationTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	legacySchema := strings.Replace(testSchema, "CREATE TABLE users (", "CREATE TABLE users_missing (", 1)
	legacySchema = strings.Replace(legacySchema, "CREATE TABLE auth_tokens (", "CREATE TABLE auth_tokens_missing (", 1)
	if _, err := database.Exec(legacySchema); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if opened, err := OpenDatabase(context.Background(), path); err == nil {
		opened.Close()
		t.Fatal("OpenDatabase should reject a schema without authentication tables")
	}
}

func TestOpenDatabaseWaitsForConcurrentSQLiteWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent.db")
	seed, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Exec(testSchema); err != nil {
		seed.Close()
		t.Fatal(err)
	}
	if _, err := seed.Exec(`
		INSERT INTO users(id, email, name, password_hash, created_at)
		VALUES (1, 'user@example.com', 'User', 'hash', 1700000000)
	`); err != nil {
		seed.Close()
		t.Fatal(err)
	}
	seed.Close()

	database, err := OpenDatabase(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	connection, err := database.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}

	result := make(chan error, 1)
	go func() {
		result <- NewAuthRepository(database).CreateToken(
			context.Background(), 1, "token", time.Unix(1_700_000_000, 0), time.Unix(1_700_003_600, 0),
		)
	}()
	select {
	case err := <-result:
		connection.ExecContext(context.Background(), "ROLLBACK")
		t.Fatalf("concurrent write returned before lock release: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	if _, err := connection.ExecContext(context.Background(), "COMMIT"); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("concurrent write failed after lock release: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent write did not resume after lock release")
	}
}
