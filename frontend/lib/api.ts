import "server-only";

import { getAccessToken } from "@/lib/session";
import type { ApiErrorBody } from "@/types/api";

const baseUrl = process.env.API_BASE_URL ?? "http://localhost:8081";

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
    readonly details?: Record<string, string>,
  ) {
    super(message);
  }
}

type ApiOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  /** JSON-serialisable body, or FormData for uploads. */
  body?: unknown;
  /**
   * "required": fail with 401 when signed out; "optional": send a token when there is one
   * (public endpoints that personalise results); "none": never send a token.
   */
  auth?: "required" | "optional" | "none";
};

/** Calls the Go API from the Next.js server. Throws ApiError with the API's error code on failure. */
export async function api<T>(path: string, { method = "GET", body, auth = "required" }: ApiOptions = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };

  if (auth !== "none") {
    const token = await getAccessToken();
    if (token) headers.Authorization = `Bearer ${token}`;
    else if (auth === "required") throw new ApiError(401, "UNAUTHORIZED", "Please sign in to continue.");
  }

  let payload: BodyInit | undefined;
  if (body instanceof FormData) {
    payload = body;
  } else if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }

  let res: Response;
  try {
    res = await fetch(`${baseUrl}${path}`, { method, headers, body: payload, cache: "no-store" });
  } catch {
    throw new ApiError(503, "API_UNAVAILABLE", "The service is unavailable. Please try again shortly.");
  }

  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const err = (data as ApiErrorBody | null)?.error;
    throw new ApiError(res.status, err?.code ?? "UNKNOWN_ERROR", err?.message ?? "Something went wrong.", err?.details);
  }
  return data as T;
}
