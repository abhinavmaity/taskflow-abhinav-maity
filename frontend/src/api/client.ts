import type { ApiErrorResponse, ApiValidationFields } from "./types";

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || "http://localhost:8080";

type RequestOptions = {
  method?: "GET" | "POST" | "PATCH" | "DELETE";
  body?: unknown;
  token?: string;
  signal?: AbortSignal;
};

export class ApiError extends Error {
  status: number;
  fields?: ApiValidationFields;

  constructor(message: string, status: number, fields?: ApiValidationFields) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = fields;
  }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
    signal: options.signal
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const raw = await response.text();
  const data = raw ? (JSON.parse(raw) as unknown) : undefined;

  if (!response.ok) {
    const errorData = (data ?? {}) as Partial<ApiErrorResponse>;
    throw new ApiError(
      typeof errorData.error === "string" ? errorData.error : "Request failed",
      response.status,
      errorData.fields
    );
  }

  return data as T;
}

export function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Something went wrong";
}
