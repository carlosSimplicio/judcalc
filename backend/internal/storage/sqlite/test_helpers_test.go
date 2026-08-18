package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

const testSchema = `
CREATE TABLE areas (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE);
CREATE TABLE services (
    id INTEGER PRIMARY KEY,
    area_id INTEGER NOT NULL REFERENCES areas(id),
    name TEXT NOT NULL,
    amount_cents INTEGER,
    percentage_min REAL,
    percentage_max REAL
);
CREATE TABLE user_fixed_costs (
    user_id TEXT PRIMARY KEY CHECK (length(trim(user_id)) > 0),
    oab_annual_fee_cents INTEGER NOT NULL DEFAULT 0 CHECK (oab_annual_fee_cents >= 0),
    digital_certificate_cents INTEGER NOT NULL DEFAULT 0 CHECK (digital_certificate_cents >= 0),
    accountant_cents INTEGER NOT NULL DEFAULT 0 CHECK (accountant_cents >= 0),
    legal_software_cents INTEGER NOT NULL DEFAULT 0 CHECK (legal_software_cents >= 0),
    internet_cents INTEGER NOT NULL DEFAULT 0 CHECK (internet_cents >= 0),
    phone_cents INTEGER NOT NULL DEFAULT 0 CHECK (phone_cents >= 0),
    recurring_transport_cents INTEGER NOT NULL DEFAULT 0 CHECK (recurring_transport_cents >= 0),
    coworking_or_office_rent_cents INTEGER NOT NULL DEFAULT 0 CHECK (coworking_or_office_rent_cents >= 0),
    professional_domain_website_email_cents INTEGER NOT NULL DEFAULT 0 CHECK (professional_domain_website_email_cents >= 0),
    marketing_cents INTEGER NOT NULL DEFAULT 0 CHECK (marketing_cents >= 0),
    office_supplies_cents INTEGER NOT NULL DEFAULT 0 CHECK (office_supplies_cents >= 0),
    equipment_and_depreciation_cents INTEGER NOT NULL DEFAULT 0 CHECK (equipment_and_depreciation_cents >= 0),
    other_costs_cents INTEGER NOT NULL DEFAULT 0 CHECK (other_costs_cents >= 0)
);
CREATE VIRTUAL TABLE areas_fts USING fts5(
    name, content='areas', content_rowid='id',
    tokenize='unicode61 remove_diacritics 2'
);
CREATE VIRTUAL TABLE services_fts USING fts5(
    name, content='services', content_rowid='id',
    tokenize='unicode61 remove_diacritics 2'
);
CREATE TRIGGER areas_fts_insert AFTER INSERT ON areas BEGIN
    INSERT INTO areas_fts(rowid, name) VALUES (new.id, new.name);
END;
CREATE TRIGGER services_fts_insert AFTER INSERT ON services BEGIN
    INSERT INTO services_fts(rowid, name) VALUES (new.id, new.name);
END;
`

func newTestRepositories(t *testing.T) (*AreaRepository, *ServiceRepository, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(testSchema); err != nil {
		database.Close()
		t.Fatal(err)
	}
	statements := []string{
		`INSERT INTO areas(id, name) VALUES (1, 'Tributário'), (2, 'Atividades de família'), (3, 'Direito previdenciário')`,
		`INSERT INTO services(id, area_id, name, amount_cents, percentage_min, percentage_max) VALUES
			(1, 3, 'Ação previdenciária comum', 10000, 10, 20),
			(2, 2, 'Consulta familiar', NULL, NULL, NULL),
			(3, 1, 'Defesa tributária', 25000, NULL, NULL)`,
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			database.Close()
			t.Fatal(err)
		}
	}
	return NewAreaRepository(database), NewServiceRepository(database), func() {
		if err := database.Close(); err != nil && !os.IsNotExist(err) {
			t.Errorf("close database: %v", err)
		}
	}
}
