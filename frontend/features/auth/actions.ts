"use server";

import { redirect } from "next/navigation";

import { signOut } from "@/auth";
import { keycloakIssuer } from "@/lib/keycloak";
import { getIdToken } from "@/lib/session";

/** Clears the local session and ends the Keycloak SSO session, so the next login asks for credentials. */
export async function logout() {
  const idToken = await getIdToken();
  await signOut({ redirect: false });

  const url = new URL(`${keycloakIssuer}/protocol/openid-connect/logout`);
  url.searchParams.set("post_logout_redirect_uri", process.env.AUTH_URL ?? "http://localhost:3000");
  if (idToken) url.searchParams.set("id_token_hint", idToken);
  redirect(url.toString());
}
