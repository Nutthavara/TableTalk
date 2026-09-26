import "server-only";

import { headers } from "next/headers";
import { getToken } from "next-auth/jwt";

import { expiresSoon, refreshTokens } from "@/lib/keycloak";

async function readToken() {
  return getToken({
    req: { headers: await headers() },
    secret: process.env.AUTH_SECRET!,
    secureCookie: process.env.AUTH_URL?.startsWith("https://") ?? false,
  });
}

/**
 * Returns a usable access token for calling the Go API, or null when signed out.
 * The proxy normally refreshes tokens ahead of time; this only refreshes (in memory, for this
 * request) when the cookie still holds an expired token.
 */
export async function getAccessToken(): Promise<string | null> {
  let token = await readToken();
  if (!token?.access_token || token.error) return null;
  if (expiresSoon(token, 10)) {
    token = await refreshTokens(token);
    if (token.error) return null;
  }
  return token.access_token ?? null;
}

export async function getIdToken(): Promise<string | undefined> {
  return (await readToken())?.id_token;
}
