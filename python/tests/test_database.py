from __future__ import annotations

import json
import sqlite3
import tempfile
import unittest
from pathlib import Path

from python.scripts.init_database import DEFAULT_INPUT, initialize_database


class DatabaseTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.database_path = Path(self.temporary_directory.name) / "test.db"
        initialize_database(DEFAULT_INPUT, self.database_path)
        self.connection = sqlite3.connect(self.database_path)
        self.connection.execute("PRAGMA foreign_keys = ON")

    def tearDown(self) -> None:
        self.connection.close()
        self.temporary_directory.cleanup()

    def test_loads_expected_counts_and_nullable_prices(self) -> None:
        area_count = self.connection.execute("SELECT count(*) FROM areas").fetchone()[0]
        service_count = self.connection.execute(
            "SELECT count(*) FROM services"
        ).fetchone()[0]
        null_amounts = self.connection.execute(
            "SELECT count(*) FROM services WHERE amount_cents IS NULL"
        ).fetchone()[0]

        self.assertEqual(area_count, 31)
        self.assertEqual(service_count, 506)
        self.assertEqual(null_amounts, 22)

    def test_reload_is_idempotent(self) -> None:
        initialize_database(DEFAULT_INPUT, self.database_path)

        self.assertEqual(
            self.connection.execute("SELECT count(*) FROM areas").fetchone()[0], 31
        )
        self.assertEqual(
            self.connection.execute("SELECT count(*) FROM services").fetchone()[0],
            506,
        )

    def test_enforces_relationships_uniqueness_and_price_checks(self) -> None:
        with self.assertRaises(sqlite3.IntegrityError):
            self.connection.execute(
                "INSERT INTO services(area_id, name) VALUES (?, ?)",
                (999_999, "Serviço sem área"),
            )

        area_id = self.connection.execute("SELECT id FROM areas LIMIT 1").fetchone()[0]
        with self.assertRaises(sqlite3.IntegrityError):
            self.connection.execute(
                """
                INSERT INTO services(
                    area_id, name, percentage_min, percentage_max
                ) VALUES (?, ?, ?, ?)
                """,
                (area_id, "Percentual inválido", 30, 20),
            )

        existing = self.connection.execute(
            """
            SELECT name, amount_cents, percentage_min, percentage_max
            FROM services
            WHERE area_id = ?
            LIMIT 1
            """,
            (area_id,),
        ).fetchone()
        with self.assertRaises(sqlite3.IntegrityError):
            self.connection.execute(
                """
                INSERT INTO services(
                    area_id, name, amount_cents, percentage_min, percentage_max
                ) VALUES (?, ?, ?, ?, ?)
                """,
                (area_id, *existing),
            )

        self.connection.execute(
            "INSERT INTO services(area_id, name, amount_cents) VALUES (?, ?, ?)",
            (area_id, existing[0], 1),
        )

    def test_fts_search_ignores_accents_and_supports_prefixes(self) -> None:
        area_names = {
            row[0]
            for row in self.connection.execute(
                """
                SELECT areas.name
                FROM areas_fts
                JOIN areas ON areas.id = areas_fts.rowid
                WHERE areas_fts MATCH 'familia'
                """
            )
        }
        action_count = self.connection.execute(
            "SELECT count(*) FROM services_fts WHERE services_fts MATCH 'acao'"
        ).fetchone()[0]
        prefix_count = self.connection.execute(
            "SELECT count(*) FROM services_fts WHERE services_fts MATCH 'previdenci*'"
        ).fetchone()[0]

        self.assertIn("Atividades em matéria de família e sucessões", area_names)
        self.assertGreater(action_count, 0)
        self.assertGreater(prefix_count, 0)

    def test_invalid_json_preserves_existing_database(self) -> None:
        invalid_json = Path(self.temporary_directory.name) / "invalid.json"
        invalid_json.write_text(
            json.dumps(
                [
                    {
                        "area": "Área inválida",
                        "service": "Serviço inválido",
                        "price": {
                            "amount_cents": -1,
                            "percentage_min": None,
                            "percentage_max": None,
                        },
                    }
                ]
            ),
            encoding="utf-8",
        )

        with self.assertRaises(ValueError):
            initialize_database(invalid_json, self.database_path)

        self.assertEqual(
            self.connection.execute("SELECT count(*) FROM areas").fetchone()[0], 31
        )
        self.assertEqual(
            self.connection.execute("SELECT count(*) FROM services").fetchone()[0],
            506,
        )

    def test_fts_triggers_track_changes(self) -> None:
        cursor = self.connection.execute(
            "INSERT INTO areas(name) VALUES ('Área temporária')"
        )
        area_id = cursor.lastrowid
        self.assertEqual(
            self.connection.execute(
                "SELECT count(*) FROM areas_fts WHERE areas_fts MATCH 'temporaria'"
            ).fetchone()[0],
            1,
        )

        self.connection.execute(
            "UPDATE areas SET name = 'Setor provisório' WHERE id = ?", (area_id,)
        )
        self.assertEqual(
            self.connection.execute(
                "SELECT count(*) FROM areas_fts WHERE areas_fts MATCH 'temporaria'"
            ).fetchone()[0],
            0,
        )
        self.assertEqual(
            self.connection.execute(
                "SELECT count(*) FROM areas_fts WHERE areas_fts MATCH 'provisorio'"
            ).fetchone()[0],
            1,
        )

        service_cursor = self.connection.execute(
            "INSERT INTO services(area_id, name) VALUES (?, 'Serviço efêmero')",
            (area_id,),
        )
        service_id = service_cursor.lastrowid
        self.assertEqual(
            self.connection.execute(
                "SELECT count(*) FROM services_fts WHERE services_fts MATCH 'efemero'"
            ).fetchone()[0],
            1,
        )
        self.connection.execute("DELETE FROM services WHERE id = ?", (service_id,))
        self.assertEqual(
            self.connection.execute(
                "SELECT count(*) FROM services_fts WHERE services_fts MATCH 'efemero'"
            ).fetchone()[0],
            0,
        )

    def test_database_integrity(self) -> None:
        self.assertEqual(
            self.connection.execute("PRAGMA foreign_key_check").fetchall(), []
        )
        self.assertEqual(
            self.connection.execute("PRAGMA integrity_check").fetchone()[0], "ok"
        )


if __name__ == "__main__":
    unittest.main()
