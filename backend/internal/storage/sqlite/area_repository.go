package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type AreaRepository struct {
	database *sql.DB
}

func NewAreaRepository(database *sql.DB) *AreaRepository {
	return &AreaRepository{database: database}
}

func (repository *AreaRepository) ListAreas(ctx context.Context, options domain.ListOptions) (domain.ListResult[domain.Area], error) {
	result := domain.ListResult[domain.Area]{Items: make([]domain.Area, 0)}
	search := buildFTSQuery(options.Query)

	countQuery := "SELECT COUNT(*) FROM areas"
	listQuery := "SELECT id, name FROM areas"
	orderBy := "name COLLATE NOCASE ASC, id ASC"
	countArguments := make([]any, 0, 1)
	listArguments := make([]any, 0, 3)
	if search != "" {
		countQuery = `SELECT COUNT(*) FROM areas_fts
			JOIN areas ON areas.id = areas_fts.rowid
			WHERE areas_fts MATCH ?`
		listQuery = `SELECT areas.id, areas.name FROM areas_fts
			JOIN areas ON areas.id = areas_fts.rowid
			WHERE areas_fts MATCH ?`
		orderBy = "areas.name COLLATE NOCASE ASC, areas.id ASC"
		countArguments = append(countArguments, search)
		listArguments = append(listArguments, search)
	}

	if err := repository.database.QueryRowContext(ctx, countQuery, countArguments...).Scan(&result.Total); err != nil {
		return result, fmt.Errorf("contar áreas: %w", err)
	}

	listQuery += " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"
	listArguments = append(listArguments, options.PageSize, options.Offset())
	rows, err := repository.database.QueryContext(ctx, listQuery, listArguments...)
	if err != nil {
		return result, fmt.Errorf("listar áreas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var area domain.Area
		if err := rows.Scan(&area.ID, &area.Name); err != nil {
			return result, fmt.Errorf("ler área: %w", err)
		}
		result.Items = append(result.Items, area)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("percorrer áreas: %w", err)
	}
	return result, nil
}
