import { redirect } from "next/navigation";
import type { NextRequest } from "next/server";

import { auth, signIn } from "@/auth";
import { safeCallback } from "@/lib/routes";

// GET /login?callbackUrl=/path — sends the user to Keycloak and back to callbackUrl.
export async function GET(request: NextRequest) {
  const target = safeCallback(request.nextUrl.searchParams.get("callbackUrl"));
  const session = await auth();
  if (session?.user && !session.error) redirect(target);
  await signIn("keycloak", { redirectTo: target });
}
