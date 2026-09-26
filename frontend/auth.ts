import NextAuth from "next-auth";
import Keycloak from "next-auth/providers/keycloak";

import { expiresSoon, keycloakIssuer, refreshTokens } from "@/lib/keycloak";

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    Keycloak({
      clientId: process.env.AUTH_KEYCLOAK_ID,
      clientSecret: process.env.AUTH_KEYCLOAK_SECRET,
      issuer: keycloakIssuer,
      checks: ["pkce", "state"],
    }),
  ],
  // Matches the Keycloak realm SSO max lifespan (10h); refresh keeps the access token fresh within it.
  session: { strategy: "jwt", maxAge: 10 * 60 * 60 },
  pages: { signIn: "/login" },
  callbacks: {
    // Tokens live only in the encrypted session cookie; they are never copied into the session object
    // (which is readable from the browser via /api/auth/session).
    async jwt({ token, account }) {
      if (account) {
        return {
          ...token,
          access_token: account.access_token,
          expires_at: account.expires_at,
          refresh_token: account.refresh_token,
          id_token: account.id_token,
        };
      }
      if (token.error || !expiresSoon(token, 60)) return token;
      return refreshTokens(token);
    },
    async session({ session, token }) {
      session.error = token.error;
      return session;
    },
  },
});
