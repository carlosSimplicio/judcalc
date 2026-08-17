from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Iterable, Iterator, Sequence

import pdfplumber


ROOT = Path(__file__).resolve().parents[1]

FIRST_TABLE_PAGE = 5
LAST_TABLE_PAGE = 23

AREA_CODE_RE = re.compile(r"^\d+$")
SERVICE_CODE_RE = re.compile(r"^\d+(?:\.\d+)+$")
SUBITEM_RE = re.compile(r"^(?:[a-z]\)|-\s+)", re.IGNORECASE)
MONEY_RE = re.compile(r"R\$\s*([\d.]+,\d{2})")
PERCENT_RE = re.compile(r"(\d+(?:[.,]\d+)?)\s*%")


def normalize_text(value: object) -> str:
    if value is None:
        return ""
    return " ".join(str(value).replace("\u00a0", " ").split())


def normalize_area(value: str) -> str:
    return normalize_text(value).capitalize()


def parse_amount_cents(*values: str) -> int | None:
    for value in values:
        match = MONEY_RE.search(normalize_text(value))
        if match:
            integer, decimal = match.group(1).replace(".", "").split(",")
            return int(integer) * 100 + int(decimal)
    return None


def parse_percentages(*values: str) -> tuple[float | int | None, float | int | None]:
    percentages: list[float | int] = []
    for value in values:
        for raw in PERCENT_RE.findall(normalize_text(value)):
            parsed = float(raw.replace(",", "."))
            normalized = int(parsed) if parsed.is_integer() else parsed
            if normalized not in percentages:
                percentages.append(normalized)

    if not percentages:
        return None, None
    if len(percentages) == 1:
        return percentages[0], percentages[0]
    return percentages[0], percentages[1]


def make_price(description: str, amount: str, percentage: str) -> dict[str, object]:
    percentage_min, percentage_max = parse_percentages(
        percentage, amount, description
    )
    return {
        "amount_cents": parse_amount_cents(amount, percentage, description),
        "percentage_min": percentage_min,
        "percentage_max": percentage_max,
    }


def has_price(description: str, amount: str, percentage: str) -> bool:
    price = make_price(description, amount, percentage)
    return any(value is not None for value in price.values())


def _row_cells(row: Sequence[object]) -> tuple[str, str, str, str]:
    cells = [normalize_text(cell) for cell in row]
    cells.extend([""] * (4 - len(cells)))
    return cells[0], cells[1], cells[2], cells[3]


def _is_column_header(description: str, amount: str, percentage: str) -> bool:
    combined = " ".join((description, amount, percentage)).casefold()
    return (
        "valores mínimos" in combined
        or "valor sugerido" in combined
        or combined.strip() == "percentuais"
    )


def _has_children(
    rows: Sequence[tuple[str, str, str, str]], index: int, code: str
) -> bool:
    for next_code, next_description, _, _ in rows[index + 1 :]:
        if AREA_CODE_RE.fullmatch(next_code):
            return False
        if SERVICE_CODE_RE.fullmatch(next_code):
            return next_code.startswith(f"{code}.")
        if not next_code and SUBITEM_RE.match(next_description):
            return True
    return False


def _is_unpriced_subheading(
    rows: Sequence[tuple[str, str, str, str]], index: int, description: str
) -> bool:
    pricing_words = ("valor", "índice", "mesmo", "honorário", "percentual", "pensão")
    if any(word in description.casefold() for word in pricing_words):
        return False

    for next_code, next_description, _, _ in rows[index + 1 :]:
        if not next_description:
            continue
        return bool(
            AREA_CODE_RE.fullmatch(next_code)
            or SERVICE_CODE_RE.fullmatch(next_code)
        )
    return False


def extract_items(rows: Iterable[Sequence[object]]) -> list[dict[str, object]]:
    normalized_rows = [_row_cells(row) for row in rows]
    items: list[dict[str, object]] = []
    area: str | None = None
    parent_description: str | None = None

    for index, (code, description, amount, percentage) in enumerate(normalized_rows):
        if not description:
            continue

        if AREA_CODE_RE.fullmatch(code):
            area = normalize_area(description)
            parent_description = None
            continue

        if area is None:
            continue

        if SERVICE_CODE_RE.fullmatch(code):
            parent_description = description
            if has_price(description, amount, percentage) or not _has_children(
                normalized_rows, index, code
            ):
                items.append(
                    {
                        "area": area,
                        "service": description,
                        "price": make_price(description, amount, percentage),
                    }
                )
            continue

        if not code and SUBITEM_RE.match(description) and parent_description:
            if not has_price(description, amount, percentage) and _is_unpriced_subheading(
                normalized_rows, index, description
            ):
                continue
            subitem_description = (
                description[2:] if description.startswith("- ") else description
            )
            service = f"{parent_description.rstrip(':')} - {subitem_description}"
            items.append(
                {
                    "area": area,
                    "service": service,
                    "price": make_price(description, amount, percentage),
                }
            )
            continue

        if not code and _is_column_header(description, amount, percentage):
            continue

    return items


def iter_pdf_rows(pdf_path: Path) -> Iterator[Sequence[object]]:
    with pdfplumber.open(pdf_path) as pdf:
        if len(pdf.pages) < LAST_TABLE_PAGE:
            raise ValueError(
                f"PDF possui {len(pdf.pages)} páginas; esperadas ao menos "
                f"{LAST_TABLE_PAGE}."
            )

        for page_number in range(FIRST_TABLE_PAGE, LAST_TABLE_PAGE + 1):
            page = pdf.pages[page_number - 1]
            for table in page.extract_tables():
                yield from table


def extract_pdf(pdf_path: Path) -> list[dict[str, object]]:
    return extract_items(iter_pdf_rows(pdf_path))


def serialize_items(items: list[dict[str, object]]) -> str:
    return json.dumps(items, ensure_ascii=False, indent=2) + "\n"


def write_json(items: list[dict[str, object]], output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(serialize_items(items), encoding="utf-8")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Extrai área, serviço e preço da tabela OAB-SP."
    )
    parser.add_argument(
        "--input",
        type=Path,
        default=ROOT / "data" / "tabela-oab-sp.pdf",
        help="PDF de origem.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=ROOT / "data" / "oab-sp.json",
        help="Arquivo JSON de destino.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Verifica se o JSON existente corresponde à extração atual.",
    )
    return parser


def main() -> int:
    args = build_parser().parse_args()
    serialized = serialize_items(extract_pdf(args.input))

    if args.check:
        if not args.output.exists():
            print(f"Arquivo não encontrado: {args.output}")
            return 1
        if args.output.read_text(encoding="utf-8") != serialized:
            print(f"Arquivo desatualizado: {args.output}")
            return 1
        print(f"Arquivo atualizado: {args.output}")
        return 0

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(serialized, encoding="utf-8")
    print(f"{len(json.loads(serialized))} serviços gravados em {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
