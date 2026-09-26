import type { JWT } from "next-auth/jwt";

export const keycloakIssuer = process.env.AUTH_KEYCLOAK_ISSUER!;

type TokenResponse = {
  access_token: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
};

/**
 * Exchanges the refresh token for new tokens. On failure the token is marked with an error
 * so the UI can ask the user to sign in again.
 */
export async function refreshTokens(token: JWT): Promise<JWT> {
  if (!token.refresh_token) return { ...token, error: "RefreshTokenError" };

  try {
    const res = await fetch(`${keycloakIssuer}/protocol/openid-connect/token`, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: token.refresh_token,
        client_id: process.env.AUTH_KEYCLOAK_ID!,
        client_secret: process.env.AUTH_KEYCLOAK_SECRET!,
      }),
      cache: "no-store",
    });
    if (!res.ok) throw new Error(`refresh failed with ${res.status}`);
    const t = (await res.json()) as TokenResponse;

    return {
      ...token,
      access_token: t.access_token,
      expires_at: Math.floor(Date.now() / 1000) + t.expires_in,
      refresh_token: t.refresh_token ?? token.refresh_token,
      id_token: t.id_token ?? token.id_token,
      error: undefined,
    };
  } catch (err) {
    console.error("Keycloak token refresh failed", err);
    return { ...token, error: "RefreshTokenError" };
  }
}

/** True when the access token expires within `marginSeconds`. */
export function expiresSoon(token: JWT, marginSeconds: number): boolean {
  return !token.expires_at || Date.now() / 1000 + marginSeconds >= token.expires_at;
}
