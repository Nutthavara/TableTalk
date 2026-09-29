import "server-only";

import { auth } from "@/auth";
import { keycloakIssuer } from "@/lib/keycloak";

export type Viewer = { name: string; email?: string; initial: string; accountUrl: string } | null;

/** The signed-in user for UI purposes (names, menus). Authorization is always done by the API. */
export async function getViewer(): Promise<Viewer> {
  const session = await auth();
  if (!session?.user || session.error) return null;
  const name = session.user.name || session.user.email || "Account";
  return {
    name,
    email: session.user.email ?? undefined,
    initial: name.trim().charAt(0).toUpperCase(),
    accountUrl: `${keycloakIssuer}/account`,
  };
}
