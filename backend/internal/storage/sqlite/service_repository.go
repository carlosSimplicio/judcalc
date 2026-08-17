package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type ServiceRepository struct {
	database *sql.DB
}

func NewServiceRepository(database *sql.DB) *ServiceRepository {
	return &ServiceRepository{database: database}
}

func (repository *ServiceRepository) ListServices(ctx context.Context, options domain.ListOptions) (domain.ListResult[domain.Service], error) {
	result := domain.ListResult[domain.Service]{Items: make([]domain.Service, 0)}
	search := buildFTSQuery(options.Query)

	countQuery := "SELECT COUNT(*) FROM services"
	listQuery := "SELECT id, area_id, name, amount_cents, percentage_min, percentage_max FROM services"
	orderBy := "name COLLATE NOCASE ASC, id ASC"
	countArguments := make([]any, 0, 1)
	listArguments := make([]any, 0, 3)
	if search != "" {
		countQuery = `SELECT COUNT(*) FROM services_fts
			JOIN services ON services.id = services_fts.rowid
			WHERE services_fts MATCH ?`
		listQuery = `SELECT services.id, services.area_id, services.name,
			services.amount_cents, services.percentage_min, services.percentage_max
			FROM services_fts
			JOIN services ON services.id = services_fts.rowid
			WHERE services_fts MATCH ?`
		orderBy = "services.name COLLATE NOCASE ASC, services.id ASC"
		countArguments = append(countArguments, search)
		listArguments = append(listArguments, search)
	}

	if err := repository.database.QueryRowContext(ctx, countQuery, countArguments...).Scan(&result.Total); err != nil {
		return result, fmt.Errorf("contar serviços: %w", err)
	}

	listQuery += " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"
	listArguments = append(listArguments, options.PageSize, options.Offset())
	rows, err := repository.database.QueryContext(ctx, listQuery, listArguments...)
	if err != nil {
		return result, fmt.Errorf("listar serviços: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var service domain.Service
		var amount sql.NullInt64
		var minimum sql.NullFloat64
		var maximum sql.NullFloat64
		if err := rows.Scan(
			&service.ID,
			&service.AreaID,
			&service.Name,
			&amount,
			&minimum,
			&maximum,
		); err != nil {
			return result, fmt.Errorf("ler serviço: %w", err)
		}
		service.AmountCents = nullableInt64(amount)
		service.PercentageMin = nullableFloat64(minimum)
		service.PercentageMax = nullableFloat64(maximum)
		result.Items = append(result.Items, service)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("percorrer serviços: %w", err)
	}
	return result, nil
}
