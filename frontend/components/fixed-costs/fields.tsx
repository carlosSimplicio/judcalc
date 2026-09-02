import type { ReactNode } from "react";
import type { FixedCostKey } from "@/lib/fixed-costs/client";

export type CostFieldDefinition = {
  key: FixedCostKey;
  label: string;
  help?: string;
  wide?: boolean;
};

export type CostSectionDefinition = {
  id: string;
  title: string;
  description: string;
  tone: "blue" | "teal" | "orange";
  icon: ReactNode;
  fields: CostFieldDefinition[];
};

export const COST_SECTIONS: CostSectionDefinition[] = [
  {
    id: "professional-costs-title",
    title: "Estrutura profissional",
    description: "Obrigações e ferramentas essenciais para o escritório.",
    tone: "blue",
    icon: (
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        aria-hidden="true"
      >
        <path
          d="M9 6V4h6v2M5 7h14v13H5zM5 12h14M10 12v2h4v-2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    fields: [
      {
        key: "oab_annual_fee",
        label: "Anuidade da OAB",
        help: "Informe o valor anual. Consideraremos a média mensal.",
      },
      {
        key: "digital_certificate",
        label: "Certificado digital",
        help: "Valor mensal ou média do período contratado.",
      },
      { key: "accountant", label: "Contador" },
      { key: "legal_software", label: "Software jurídico" },
    ],
  },
  {
    id: "office-costs-title",
    title: "Escritório e operação",
    description: "Despesas para manter sua rotina de trabalho.",
    tone: "teal",
    icon: (
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        aria-hidden="true"
      >
        <path
          d="M3 21h18M5 21V7l7-4 7 4v14M9 10h.01M15 10h.01M9 14h.01M15 14h.01M10 21v-3h4v3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    fields: [
      { key: "internet", label: "Internet" },
      { key: "phone", label: "Telefone" },
      { key: "recurring_transport", label: "Transporte recorrente" },
      {
        key: "coworking_or_office_rent",
        label: "Coworking ou aluguel",
      },
      {
        key: "professional_domain_website_email",
        label: "Domínio, site e e-mail profissional",
      },
      { key: "marketing", label: "Marketing" },
    ],
  },
  {
    id: "other-costs-title",
    title: "Materiais e outros custos",
    description: "Complete o cadastro com as demais despesas mensais.",
    tone: "orange",
    icon: (
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        aria-hidden="true"
      >
        <path
          d="M4 7h16v13H4zM8 7V4h8v3M9 12h6M9 16h4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    fields: [
      { key: "office_supplies", label: "Materiais de escritório" },
      {
        key: "equipment_and_depreciation",
        label: "Equipamentos e depreciação",
      },
      {
        key: "other_costs",
        label: "Outros custos",
        help: "Use este campo para despesas recorrentes não listadas acima.",
        wide: true,
      },
    ],
  },
];
