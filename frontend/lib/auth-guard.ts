import "server-only";

import { notFound, redirect } from "next/navigation";

import { auth } from "@/auth";
import { ApiError } from "@/lib/api";
import { routes } from "@/lib/routes";

/**
 * For pages that need a signed-in user: redirects to login and back to `returnTo` afterwards.
 * This is for UX only; the Go API authorizes every request itself.
 */
export async function requireSignedIn(returnTo: string) {
  const session = await auth();
  if (!session?.user || session.error) redirect(routes.login(returnTo));
  return session;
}

/**
 * Turns API errors on page loads into the matching page behaviour: missing or not-yours resources
 * render the 404 page (without revealing whether the resource exists), an expired session goes to
 * login. Anything else is rethrown to the nearest error boundary.
 */
export function handlePageError(err: unknown, returnTo: string): never {
  if (err instanceof ApiError) {
    if (err.status === 401) redirect(routes.login(returnTo));
    if (err.status === 403 || err.status === 404) notFound();
  }
  throw err;
}
