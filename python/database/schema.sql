PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS areas (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY,
    area_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    amount_cents INTEGER,
    percentage_min REAL,
    percentage_max REAL,
    FOREIGN KEY (area_id) REFERENCES areas(id) ON DELETE CASCADE,
    CHECK (amount_cents IS NULL OR amount_cents >= 0),
    CHECK (
        (percentage_min IS NULL AND percentage_max IS NULL)
        OR (
            percentage_min IS NOT NULL
            AND percentage_max IS NOT NULL
            AND percentage_min BETWEEN 0 AND 100
            AND percentage_max BETWEEN 0 AND 100
            AND percentage_min <= percentage_max
        )
    )
);

CREATE INDEX IF NOT EXISTS services_area_id_idx ON services(area_id);
CREATE UNIQUE INDEX IF NOT EXISTS services_fee_variant_idx ON services(
    area_id,
    name,
    COALESCE(amount_cents, -1),
    COALESCE(percentage_min, -1),
    COALESCE(percentage_max, -1)
);

CREATE VIRTUAL TABLE IF NOT EXISTS areas_fts USING fts5(
    name,
    content = 'areas',
    content_rowid = 'id',
    tokenize = 'unicode61 remove_diacritics 2'
);

CREATE VIRTUAL TABLE IF NOT EXISTS services_fts USING fts5(
    name,
    content = 'services',
    content_rowid = 'id',
    tokenize = 'unicode61 remove_diacritics 2'
);

CREATE TRIGGER IF NOT EXISTS areas_fts_insert
AFTER INSERT ON areas BEGIN
    INSERT INTO areas_fts(rowid, name) VALUES (new.id, new.name);
END;

CREATE TRIGGER IF NOT EXISTS areas_fts_delete
AFTER DELETE ON areas BEGIN
    INSERT INTO areas_fts(areas_fts, rowid, name)
    VALUES ('delete', old.id, old.name);
END;

CREATE TRIGGER IF NOT EXISTS areas_fts_update
AFTER UPDATE ON areas BEGIN
    INSERT INTO areas_fts(areas_fts, rowid, name)
    VALUES ('delete', old.id, old.name);
    INSERT INTO areas_fts(rowid, name) VALUES (new.id, new.name);
END;

CREATE TRIGGER IF NOT EXISTS services_fts_insert
AFTER INSERT ON services BEGIN
    INSERT INTO services_fts(rowid, name) VALUES (new.id, new.name);
END;

CREATE TRIGGER IF NOT EXISTS services_fts_delete
AFTER DELETE ON services BEGIN
    INSERT INTO services_fts(services_fts, rowid, name)
    VALUES ('delete', old.id, old.name);
END;

CREATE TRIGGER IF NOT EXISTS services_fts_update
AFTER UPDATE ON services BEGIN
    INSERT INTO services_fts(services_fts, rowid, name)
    VALUES ('delete', old.id, old.name);
    INSERT INTO services_fts(rowid, name) VALUES (new.id, new.name);
END;
