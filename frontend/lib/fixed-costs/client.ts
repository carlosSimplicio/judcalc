import { apiRequest, isRecord, unexpectedResponse } from "@/lib/api/client";

export const MONTHLY_COST_KEYS = [
  "digital_certificate",
  "accountant",
  "legal_software",
  "internet",
  "phone",
  "recurring_transport",
  "coworking_or_office_rent",
  "professional_domain_website_email",
  "marketing",
  "office_supplies",
  "equipment_and_depreciation",
  "other_costs",
] as const;

export type MonthlyCostKey = (typeof MONTHLY_COST_KEYS)[number];
export type FixedCostKey = "oab_annual_fee" | MonthlyCostKey;

export type FixedCostsValues = Record<FixedCostKey, number>;

export type FixedCosts = {
  userId: number;
  values: FixedCostsValues;
};

export async function getFixedCosts(token: string): Promise<FixedCosts> {
  const payload = await apiRequest("/fixed-costs", { token });
  return parseFixedCostsEnvelope(payload, 200);
}

export async function saveFixedCosts(
  token: string,
  values: FixedCostsValues,
): Promise<FixedCosts> {
  const monthlyCosts = Object.fromEntries(
    MONTHLY_COST_KEYS.map((key) => [
      key,
      { monthly_amount_cents: values[key] },
    ]),
  );
  const payload = await apiRequest("/fixed-costs", {
    method: "PATCH",
    token,
    body: {
      costs: {
        oab_annual_fee: {
          annual_amount_cents: values.oab_annual_fee,
        },
        ...monthlyCosts,
      },
    },
  });
  return parseFixedCostsEnvelope(payload, 200);
}

function parseFixedCostsEnvelope(
  value: unknown,
  status: number,
): FixedCosts {
  if (!isRecord(value) || !isRecord(value.data)) {
    throw unexpectedResponse(status);
  }

  const data = value.data;
  if (!isNonNegativeInteger(data.user_id) || !isRecord(data.costs)) {
    throw unexpectedResponse(status);
  }

  const costs = data.costs;
  const oab = costs.oab_annual_fee;
  if (
    !isRecord(oab) ||
    !isNonNegativeInteger(oab.annual_amount_cents) ||
    !isNonNegativeInteger(oab.monthly_amount_cents)
  ) {
    throw unexpectedResponse(status);
  }

  const values = {
    oab_annual_fee: oab.annual_amount_cents,
  } as FixedCostsValues;

  for (const key of MONTHLY_COST_KEYS) {
    const cost = costs[key];
    if (!isRecord(cost) || !isNonNegativeInteger(cost.monthly_amount_cents)) {
      throw unexpectedResponse(status);
    }
    values[key] = cost.monthly_amount_cents;
  }

  return { userId: data.user_id, values };
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}
