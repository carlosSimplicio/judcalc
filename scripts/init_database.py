from __future__ import annotations

import argparse
import json
import math
import sqlite3
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_INPUT = ROOT / "data" / "oab-sp.json"
DEFAULT_OUTPUT = ROOT / "data" / "app.db"
DEFAULT_SCHEMA = ROOT / "database" / "schema.sql"


def _optional_integer(value: Any, field: str, index: int) -> int | None:
    if value is None:
        return None
    if isinstance(value, bool) or not isinstance(value, int):
        raise ValueError(f"Item {index}: {field} deve ser inteiro ou null.")
    if value < 0:
        raise ValueError(f"Item {index}: {field} não pode ser negativo.")
    return value


def _optional_percentage(value: Any, field: str, index: int) -> float | int | None:
    if value is None:
        return None
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError(f"Item {index}: {field} deve ser numérico ou null.")
    if not math.isfinite(value) or not 0 <= value <= 100:
        raise ValueError(f"Item {index}: {field} deve estar entre 0 e 100.")
    return value


DatabaseItem = tuple[
    str, str, int | None, float | int | None, float | int | None
]


def load_items(json_path: Path) -> list[DatabaseItem]:
    try:
        raw = json.loads(json_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise ValueError(f"Não foi possível ler {json_path}: {exc}") from exc

    if not isinstance(raw, list):
        raise ValueError("A raiz do JSON deve ser uma lista.")

    items: list[DatabaseItem] = []
    seen: set[DatabaseItem] = set()

    for index, item in enumerate(raw, start=1):
        if not isinstance(item, dict):
            raise ValueError(f"Item {index}: deve ser um objeto.")

        area = item.get("area")
        service = item.get("service")
        price = item.get("price")
        if not isinstance(area, str) or not area.strip():
            raise ValueError(f"Item {index}: area deve ser texto não vazio.")
        if not isinstance(service, str) or not service.strip():
            raise ValueError(f"Item {index}: service deve ser texto não vazio.")
        if not isinstance(price, dict):
            raise ValueError(f"Item {index}: price deve ser um objeto.")

        amount_cents = _optional_integer(price.get("amount_cents"), "amount_cents", index)
        percentage_min = _optional_percentage(
            price.get("percentage_min"), "percentage_min", index
        )
        percentage_max = _optional_percentage(
            price.get("percentage_max"), "percentage_max", index
        )
        if (percentage_min is None) != (percentage_max is None):
            raise ValueError(
                f"Item {index}: percentage_min e percentage_max devem ser "
                "informados juntos."
            )
        if percentage_min is not None and percentage_min > percentage_max:
            raise ValueError(
                f"Item {index}: percentage_min não pode exceder percentage_max."
            )

        database_item = (
            area,
            service,
            amount_cents,
            percentage_min,
            percentage_max,
        )
        if database_item in seen:
            raise ValueError(f"Item {index}: serviço e preço duplicados: {service}")
        seen.add(database_item)
        items.append(database_item)

    return items


def initialize_database(
    json_path: Path = DEFAULT_INPUT,
    database_path: Path = DEFAULT_OUTPUT,
    schema_path: Path = DEFAULT_SCHEMA,
) -> tuple[int, int]:
    items = load_items(json_path)
    schema = schema_path.read_text(encoding="utf-8")
    database_path.parent.mkdir(parents=True, exist_ok=True)

    connection = sqlite3.connect(database_path)
    try:
        connection.execute("PRAGMA foreign_keys = ON")
        connection.executescript(schema)

        with connection:
            connection.execute("DELETE FROM areas")

            area_ids: dict[str, int] = {}
            for area, *_ in items:
                if area in area_ids:
                    continue
                cursor = connection.execute(
                    "INSERT INTO areas(name) VALUES (?)",
                    (area,),
                )
                area_ids[area] = cursor.lastrowid

            connection.executemany(
                """
                INSERT INTO services(
                    area_id, name, amount_cents, percentage_min, percentage_max
                ) VALUES (?, ?, ?, ?, ?)
                """,
                (
                    (area_ids[area], service, amount, minimum, maximum)
                    for area, service, amount, minimum, maximum in items
                ),
            )

        return len(area_ids), len(items)
    finally:
        connection.close()


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Cria e popula o banco SQLite com a tabela de honorários OAB-SP."
    )
    parser.add_argument("--input", type=Path, default=DEFAULT_INPUT)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--schema", type=Path, default=DEFAULT_SCHEMA)
    return parser


def main() -> int:
    args = build_parser().parse_args()
    try:
        area_count, service_count = initialize_database(
            args.input, args.output, args.schema
        )
    except (ValueError, OSError, sqlite3.Error) as exc:
        print(f"Erro ao inicializar banco: {exc}")
        return 1

    print(
        f"Banco criado em {args.output}: "
        f"{area_count} áreas e {service_count} serviços."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
