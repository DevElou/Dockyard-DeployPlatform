import { ApiError } from "@/lib/types/api";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json().catch(() => ({}));

  if (!res.ok) {
    throw new ApiError(
      body?.error ?? "unknown_error",
      body?.message ?? `Request failed with status ${res.status}`,
      res.status,
    );
  }

  return body as T;
}

export function apiGet<T>(path: string): Promise<T> {
  return apiFetch<T>(path, { method: "GET" });
}

export function apiPost<T>(path: string, data?: unknown): Promise<T> {
  return apiFetch<T>(path, {
    method: "POST",
    body: data !== undefined ? JSON.stringify(data) : undefined,
  });
}

export function apiPut<T>(path: string, data?: unknown): Promise<T> {
  return apiFetch<T>(path, {
    method: "PUT",
    body: data !== undefined ? JSON.stringify(data) : undefined,
  });
}

export function apiPatch<T>(path: string, data?: unknown): Promise<T> {
  return apiFetch<T>(path, {
    method: "PATCH",
    body: data !== undefined ? JSON.stringify(data) : undefined,
  });
}

export function apiDelete<T = void>(path: string): Promise<T> {
  return apiFetch<T>(path, { method: "DELETE" });
}
