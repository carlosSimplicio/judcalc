import { apiRequest, isRecord, unexpectedResponse } from "@/lib/api/client";

export type FeeLevel = "low" | "medium" | "high";

export type Area = {
  id: number;
  name: string;
};

export type Service = {
  id: number;
  areaId: number;
  name: string;
  amountCents: number | null;
  percentageMin: number | null;
  percentageMax: number | null;
};

export type PageMetadata = {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type ServicesPage = {
  items: Service[];
  meta: PageMetadata;
};

export type FeeCalculationInput = {
  serviceId: number;
  estimatedHours: number;
  billableHoursPerMonth: number;
  complexity: FeeLevel;
  risk: FeeLevel;
};

export type FeeCalculationResult = {
  service: Pick<Service, "id" | "areaId" | "name">;
  oabReference: {
    amountCents: number | null;
    percentageMin: number | null;
    percentageMax: number | null;
  };
  inputs: {
    estimatedHours: number;
    billableHoursPerMonth: number;
    complexity: FeeLevel;
    complexityFactor: number;
    risk: FeeLevel;
    riskFactor: number;
  };
  calculation: {
    monthlyFixedCostsCents: number;
    operationalHourCostCents: number | null;
    minimumSustainableCostCents: number | null;
    technicalEstimateCents: number | null;
  };
  warnings: Array<{ code: string; message: string }>;
};

export async function listServices(
  token: string,
  page: number,
  query: string,
  pageSize = 20,
): Promise<ServicesPage> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (query) params.set("q", query);

  const payload = await apiRequest(`/services?${params}`, { token });
  if (!isRecord(payload) || !Array.isArray(payload.data)) {
    throw unexpectedResponse(200);
  }

  return {
    items: payload.data.map(parseService),
    meta: parseMetadata(payload.meta),
  };
}

export async function listAllAreas(token: string): Promise<Area[]> {
  const firstPage = await listAreasPage(token, 1);
  if (firstPage.meta.totalPages <= 1) return firstPage.items;

  const remainingPages = await Promise.all(
    Array.from({ length: firstPage.meta.totalPages - 1 }, (_, index) =>
      listAreasPage(token, index + 2),
    ),
  );
  return [firstPage, ...remainingPages].flatMap((page) => page.items);
}

export async function calculateFee(
  token: string,
  input: FeeCalculationInput,
): Promise<FeeCalculationResult> {
  const payload = await apiRequest("/services/fee-calculation", {
    method: "POST",
    token,
    body: {
      service_id: input.serviceId,
      estimated_hours: input.estimatedHours,
      billable_hours_per_month: input.billableHoursPerMonth,
      complexity: input.complexity,
      risk: input.risk,
    },
  });

  if (!isRecord(payload) || !isRecord(payload.data)) {
    throw unexpectedResponse(200);
  }
  return parseCalculation(payload.data);
}

async function listAreasPage(
  token: string,
  page: number,
): Promise<{ items: Area[]; meta: PageMetadata }> {
  const payload = await apiRequest(`/areas?page=${page}&page_size=100`, { token });
  if (!isRecord(payload) || !Array.isArray(payload.data)) {
    throw unexpectedResponse(200);
  }
  return {
    items: payload.data.map(parseArea),
    meta: parseMetadata(payload.meta),
  };
}

function parseService(value: unknown): Service {
  if (
    !isRecord(value) ||
    !isPositiveInteger(value.id) ||
    !isPositiveInteger(value.area_id) ||
    typeof value.name !== "string" ||
    !isNullableNonNegativeInteger(value.amount_cents) ||
    !isNullableNumber(value.percentage_min) ||
    !isNullableNumber(value.percentage_max)
  ) {
    throw unexpectedResponse(200);
  }
  return {
    id: value.id,
    areaId: value.area_id,
    name: value.name,
    amountCents: value.amount_cents,
    percentageMin: value.percentage_min,
    percentageMax: value.percentage_max,
  };
}

function parseArea(value: unknown): Area {
  if (!isRecord(value) || !isPositiveInteger(value.id) || typeof value.name !== "string") {
    throw unexpectedResponse(200);
  }
  return { id: value.id, name: value.name };
}

function parseMetadata(value: unknown): PageMetadata {
  if (
    !isRecord(value) ||
    !isPositiveInteger(value.page) ||
    !isPositiveInteger(value.page_size) ||
    !isNonNegativeInteger(value.total) ||
    !isNonNegativeInteger(value.total_pages)
  ) {
    throw unexpectedResponse(200);
  }
  return {
    page: value.page,
    pageSize: value.page_size,
    total: value.total,
    totalPages: value.total_pages,
  };
}

function parseCalculation(value: Record<string, unknown>): FeeCalculationResult {
  const service = value.service;
  const reference = value.oab_reference;
  const inputs = value.inputs;
  const calculation = value.calculation;
  const warnings = value.warnings;
  if (
    !isRecord(service) ||
    !isPositiveInteger(service.id) ||
    !isPositiveInteger(service.area_id) ||
    typeof service.name !== "string" ||
    !isRecord(reference) ||
    !isNullableNonNegativeInteger(reference.amount_cents) ||
    !isNullableNumber(reference.percentage_min) ||
    !isNullableNumber(reference.percentage_max) ||
    !isRecord(inputs) ||
    !isPositiveNumber(inputs.estimated_hours) ||
    !isPositiveNumber(inputs.billable_hours_per_month) ||
    !isFeeLevel(inputs.complexity) ||
    !isPositiveNumber(inputs.complexity_factor) ||
    !isFeeLevel(inputs.risk) ||
    !isPositiveNumber(inputs.risk_factor) ||
    !isRecord(calculation) ||
    !isNonNegativeInteger(calculation.monthly_fixed_costs_cents) ||
    !isNullableNonNegativeInteger(calculation.operational_hour_cost_cents) ||
    !isNullableNonNegativeInteger(calculation.minimum_sustainable_cost_cents) ||
    !isNullableNonNegativeInteger(calculation.technical_estimate_cents) ||
    !Array.isArray(warnings)
  ) {
    throw unexpectedResponse(200);
  }

  return {
    service: { id: service.id, areaId: service.area_id, name: service.name },
    oabReference: {
      amountCents: reference.amount_cents,
      percentageMin: reference.percentage_min,
      percentageMax: reference.percentage_max,
    },
    inputs: {
      estimatedHours: inputs.estimated_hours,
      billableHoursPerMonth: inputs.billable_hours_per_month,
      complexity: inputs.complexity,
      complexityFactor: inputs.complexity_factor,
      risk: inputs.risk,
      riskFactor: inputs.risk_factor,
    },
    calculation: {
      monthlyFixedCostsCents: calculation.monthly_fixed_costs_cents,
      operationalHourCostCents: calculation.operational_hour_cost_cents,
      minimumSustainableCostCents: calculation.minimum_sustainable_cost_cents,
      technicalEstimateCents: calculation.technical_estimate_cents,
    },
    warnings: warnings.map(parseWarning),
  };
}

function parseWarning(value: unknown): { code: string; message: string } {
  if (!isRecord(value) || typeof value.code !== "string" || typeof value.message !== "string") {
    throw unexpectedResponse(200);
  }
  return { code: value.code, message: value.message };
}

function isFeeLevel(value: unknown): value is FeeLevel {
  return value === "low" || value === "medium" || value === "high";
}

function isPositiveInteger(value: unknown): value is number {
  return isNonNegativeInteger(value) && value > 0;
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isPositiveNumber(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value > 0;
}

function isNullableNonNegativeInteger(value: unknown): value is number | null {
  return value === null || isNonNegativeInteger(value);
}

function isNullableNumber(value: unknown): value is number | null {
  return value === null || (typeof value === "number" && Number.isFinite(value));
}
