import { apiRequest, isRecord, unexpectedResponse } from "@/lib/api/client";

export { ApiError } from "@/lib/api/client";

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
  const payload = await apiRequest(path, { method: "POST", body: input });
  if (!isSessionEnvelope(payload)) {
    throw unexpectedResponse(200);
  }
  return payload.data;
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
