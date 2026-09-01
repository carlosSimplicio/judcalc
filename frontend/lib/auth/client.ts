export type AuthUser = {
  id: number;
  email: string;
  name: string;
};

export type AuthSession = {
  user: AuthUser;
  access_token: string;
  token_type: string;
  expires_at: string;
};

export type SignInInput = {
  email: string;
  password: string;
};

export type SignUpInput = SignInInput & {
  name: string;
};

type ErrorEnvelope = {
  error: {
    code: string;
    message: string;
  };
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

export function signIn(input: SignInInput): Promise<AuthSession> {
  return requestSession("/auth/sign-in", input);
}

export function signUp(input: SignUpInput): Promise<AuthSession> {
  return requestSession("/auth/sign-up", input);
}

async function requestSession(
  path: string,
  input: SignInInput | SignUpInput,
): Promise<AuthSession> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
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
      throw new ApiError(
        response.status,
        payload.error.code,
        payload.error.message,
      );
    }
    throw unexpectedResponse(response.status);
  }

  if (!isSessionEnvelope(payload)) {
    throw unexpectedResponse(response.status);
  }
  return payload.data;
}

function unexpectedResponse(status: number): ApiError {
  return new ApiError(
    status,
    "unexpected_response",
    "O servidor retornou uma resposta inesperada.",
  );
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  if (!isRecord(value) || !isRecord(value.error)) {
    return false;
  }
  return (
    typeof value.error.code === "string" &&
    typeof value.error.message === "string"
  );
}

function isSessionEnvelope(value: unknown): value is { data: AuthSession } {
  if (!isRecord(value)) {
    return false;
  }
  const data = value.data;
  if (!isRecord(data)) {
    return false;
  }
  const user = data.user;
  if (!isRecord(user)) {
    return false;
  }
  return (
    typeof user.id === "number" &&
    typeof user.email === "string" &&
    typeof user.name === "string" &&
    typeof data.access_token === "string" &&
    data.access_token.length > 0 &&
    typeof data.token_type === "string" &&
    typeof data.expires_at === "string" &&
    !Number.isNaN(Date.parse(data.expires_at))
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
