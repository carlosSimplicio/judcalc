export type ErrorEnvelope = {
  error: { code: string; message: string };
};

export type ApiRequestInit = {
  method?: "GET" | "POST" | "PATCH";
  token?: string;
  body?: unknown;
};

const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL
).replace(/\/$/, "");

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export async function apiRequest(
  path: string,
  init: ApiRequestInit = {},
): Promise<unknown> {
  const headers = new Headers();
  if (init.body !== undefined) headers.set("Content-Type", "application/json");
  if (init.token) headers.set("Authorization", `Bearer ${init.token}`);

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: init.method ?? "GET",
      headers,
      body: init.body === undefined ? undefined : JSON.stringify(init.body),
    });
  } catch {
    throw new ApiError(
      0,
      "network_error",
      "Não foi possível conectar ao servidor. Tente novamente.",
    );
  }

  const payload: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    if (isErrorEnvelope(payload)) {
      throw new ApiError(response.status, payload.error.code, payload.error.message);
    }
    throw unexpectedResponse(response.status);
  }
  return payload;
}

export function unexpectedResponse(status: number): ApiError {
  return new ApiError(
    status,
    "unexpected_response",
    "O servidor retornou uma resposta inesperada.",
  );
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    isRecord(value) &&
    isRecord(value.error) &&
    typeof value.error.code === "string" &&
    typeof value.error.message === "string"
  );
}
