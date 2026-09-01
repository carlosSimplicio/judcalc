import type { AuthSession } from "./client";

const SESSION_STORAGE_KEY = "judcalc.auth.session.v1";

export function saveSession(session: AuthSession): void {
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

export function getSession(): AuthSession | null {
  const serialized = localStorage.getItem(SESSION_STORAGE_KEY);
  if (!serialized) {
    return null;
  }

  try {
    const session: unknown = JSON.parse(serialized);
    if (!isStoredSession(session) || Date.parse(session.expires_at) <= Date.now()) {
      clearSession();
      return null;
    }
    return session;
  } catch {
    clearSession();
    return null;
  }
}

export function clearSession(): void {
  localStorage.removeItem(SESSION_STORAGE_KEY);
}

function isStoredSession(value: unknown): value is AuthSession {
  if (!isRecord(value) || !isRecord(value.user)) {
    return false;
  }
  return (
    typeof value.user.id === "number" &&
    typeof value.user.email === "string" &&
    typeof value.user.name === "string" &&
    typeof value.access_token === "string" &&
    value.access_token.length > 0 &&
    typeof value.token_type === "string" &&
    typeof value.expires_at === "string" &&
    !Number.isNaN(Date.parse(value.expires_at))
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
