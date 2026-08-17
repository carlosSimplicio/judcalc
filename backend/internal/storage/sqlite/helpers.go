package sqlite

import (
	"database/sql"
	"strings"
)

func buildFTSQuery(query string) string {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return ""
	}
	for index, term := range terms {
		terms[index] = `"` + strings.ReplaceAll(term, `"`, `""`) + `"*`
	}
	return strings.Join(terms, " AND ")
}

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func nullableFloat64(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}
