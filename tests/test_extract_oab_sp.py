from __future__ import annotations

import json
import unittest
from pathlib import Path

from scripts.extract_oab_sp import (
    extract_items,
    extract_pdf,
    parse_amount_cents,
    parse_percentages,
    serialize_items,
)


ROOT = Path(__file__).resolve().parents[1]
PDF_PATH = ROOT / "docs" / "tabela-oab-sp.pdf"


class ParserTests(unittest.TestCase):
    def test_parses_brazilian_currency(self) -> None:
        self.assertEqual(parse_amount_cents("R$ 6.256,51"), 625651)
        self.assertIsNone(parse_amount_cents("-"))

    def test_parses_simple_and_range_percentages(self) -> None:
        self.assertEqual(parse_percentages("20%"), (20, 20))
        self.assertEqual(parse_percentages("20% a 30%"), (20, 30))
        self.assertEqual(parse_percentages(""), (None, None))

    def test_associates_area_and_flattens_subitems(self) -> None:
        rows = [
            ["4", "ATIVIDADES EM MATÉRIA CÍVEL", "Valores mínimos", "Percentuais"],
            ["4.1", "Procedimento ordinário", "R$ 6.256,51", "20%"],
            ["4.2", "Adoção", "", ""],
            ["", "a) Por nacional", "R$ 8.689,58", ""],
            ["", "Procedimentos especiais", "Valores mínimos", "Percentuais"],
        ]

        self.assertEqual(
            extract_items(rows),
            [
                {
                    "area": "Atividades em matéria cível",
                    "service": "Procedimento ordinário",
                    "price": {
                        "amount_cents": 625651,
                        "percentage_min": 20,
                        "percentage_max": 20,
                    },
                },
                {
                    "area": "Atividades em matéria cível",
                    "service": "Adoção - a) Por nacional",
                    "price": {
                        "amount_cents": 868958,
                        "percentage_min": None,
                        "percentage_max": None,
                    },
                },
            ],
        )

    def test_output_contains_only_planned_fields(self) -> None:
        item = extract_items(
            [["1", "ÁREA", "", ""], ["1.1", "Serviço", "R$ 10,00", "5%"]]
        )[0]
        self.assertEqual(set(item), {"area", "service", "price"})
        self.assertEqual(
            set(item["price"]),
            {"amount_cents", "percentage_min", "percentage_max"},
        )

    def test_keeps_leaf_service_without_numeric_price(self) -> None:
        items = extract_items(
            [
                ["19", "ATENDIMENTO VIRTUAL / ELETRÔNICO", "", ""],
                ["19.1", "Mesmos honorários anteriormente previstos", "", ""],
            ]
        )
        self.assertEqual(len(items), 1)
        self.assertEqual(
            items[0]["price"],
            {
                "amount_cents": None,
                "percentage_min": None,
                "percentage_max": None,
            },
        )

    def test_does_not_emit_unpriced_parent_with_children(self) -> None:
        items = extract_items(
            [
                ["6", "FAMÍLIA", "", ""],
                ["6.1", "Divórcio judicial", "", ""],
                ["", "a) Consensual", "R$ 7.820,64", ""],
            ]
        )
        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["service"], "Divórcio judicial - a) Consensual")

    def test_supports_hyphenated_subitems(self) -> None:
        items = extract_items(
            [
                ["31", "PRIVACIDADE", "", ""],
                ["31.25", "Parecer / consultoria:", "", ""],
                ["", "- para agentes de pequeno porte", "R$ 4.987,87", ""],
            ]
        )
        self.assertEqual(len(items), 1)
        self.assertEqual(
            items[0]["service"],
            "Parecer / consultoria - para agentes de pequeno porte",
        )

    def test_ignores_unpriced_subheading_before_numbered_services(self) -> None:
        items = extract_items(
            [
                ["10", "CONSUMIDOR", "", ""],
                ["10.8", "Atuação em audiência", "R$ 2.433,09", ""],
                ["", "a) Representação em convenção coletiva", "", ""],
                ["10.9", "De entidade civil", "R$ 4.344,80", ""],
            ]
        )
        self.assertEqual([item["service"] for item in items], ["Atuação em audiência", "De entidade civil"])

    def test_serialization_is_deterministic(self) -> None:
        items = extract_items(
            [["1", "ÁREA", "", ""], ["1.1", "Serviço", "R$ 10,00", "5%"]]
        )
        self.assertEqual(serialize_items(items), serialize_items(items))
        self.assertEqual(json.loads(serialize_items(items)), items)


class PdfIntegrationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.items = extract_pdf(PDF_PATH)

    def test_extracts_expected_known_services(self) -> None:
        by_service = {item["service"]: item for item in self.items}
        civil = by_service["Procedimento ordinário: proposição ou defesa"]
        self.assertEqual(civil["area"], "Atividades em matéria cível")
        self.assertEqual(civil["price"]["amount_cents"], 625651)
        self.assertEqual(civil["price"]["percentage_min"], 20)

        previdenciario = by_service[
            "Concessão ou restabelecimento de aposentadoria, auxílio-acidente, "
            "pensão por morte e benefícios assistenciais (BPC)"
        ]
        self.assertEqual(previdenciario["price"]["percentage_min"], 20)
        self.assertEqual(previdenciario["price"]["percentage_max"], 30)

    def test_preserves_ambiguous_29_5_and_29_6_as_separate_services(self) -> None:
        services = [item["service"] for item in self.items]
        self.assertIn(
            '"Elaboração de contrato cível envolvendo publicidade comercial na internet/redes',
            services,
        )
        self.assertIn('sociais/plataformas digitais"', services)

    def test_does_not_emit_headers_or_metadata(self) -> None:
        services = {item["service"] for item in self.items}
        self.assertNotIn("Procedimentos Especiais:", services)
        self.assertNotIn("Considerações importantes", services)
        for item in self.items:
            self.assertEqual(set(item), {"area", "service", "price"})


if __name__ == "__main__":
    unittest.main()
