export type AuthUser = {
  name: string;
  email: string;
};

export type AuthSession = {
  token: string;
  user: AuthUser;
};

const SESSION_STORAGE_KEY = "taskflow.auth.session";

export function loadSession(): AuthSession | null {
  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<AuthSession>;
    if (
      typeof parsed.token !== "string" ||
      !parsed.user ||
      typeof parsed.user.name !== "string" ||
      typeof parsed.user.email !== "string"
    ) {
      return null;
    }

    return {
      token: parsed.token,
      user: {
        name: parsed.user.name,
        email: parsed.user.email
      }
    };
  } catch {
    return null;
  }
}

export function saveSession(session: AuthSession): void {
  window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

export function clearSession(): void {
  window.localStorage.removeItem(SESSION_STORAGE_KEY);
}
