import {
  MONTHLY_COST_KEYS,
  type FixedCostsValues,
} from "@/lib/fixed-costs/client";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

export function formatCentsInput(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function parseMoneyInput(value: string): number {
  const digits = value.replace(/\D/g, "");
  return digits ? Number.parseInt(digits, 10) : 0;
}

export function normalizeMoneyInput(value: string): string {
  return formatCentsInput(parseMoneyInput(value));
}

export function formatCurrency(cents: number): string {
  return currencyFormatter.format(cents / 100);
}

export function monthlyTotal(values: FixedCostsValues): number {
  const oabMonthly =
    Math.floor(values.oab_annual_fee / 12) +
    (values.oab_annual_fee % 12 >= 6 ? 1 : 0);
  return MONTHLY_COST_KEYS.reduce(
    (total, key) => total + values[key],
    oabMonthly,
  );
}
