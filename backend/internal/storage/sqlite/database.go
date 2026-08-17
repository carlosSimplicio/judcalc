package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenDatabase(ctx context.Context, path string) (*sql.DB, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("arquivo SQLite não encontrado em %q", path)
		}
		return nil, fmt.Errorf("verificar arquivo SQLite %q: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("o caminho SQLite %q é um diretório", path)
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolver caminho SQLite %q: %w", path, err)
	}

	dsn := "file:" + filepath.ToSlash(absolutePath) + "?mode=ro"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrir banco SQLite: %w", err)
	}
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(10)

	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("conectar ao banco SQLite: %w", err)
	}
	if err := validateSchema(ctx, database); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func validateSchema(ctx context.Context, database *sql.DB) error {
	queries := []string{
		"SELECT id, name FROM areas LIMIT 0",
		"SELECT id, area_id, name, amount_cents, percentage_min, percentage_max FROM services LIMIT 0",
		"SELECT rowid, name FROM areas_fts LIMIT 0",
		"SELECT rowid, name FROM services_fts LIMIT 0",
	}
	for _, query := range queries {
		rows, err := database.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("esquema SQLite incompatível: %w", err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("validar esquema SQLite: %w", err)
		}
	}
	return nil
}
