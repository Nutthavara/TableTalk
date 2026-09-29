import type { NextRequest } from "next/server";

import { signIn } from "@/auth";
import { safeCallback } from "@/lib/routes";

// GET /register — opens Keycloak's registration form (OIDC prompt=create), then returns to callbackUrl.
export async function GET(request: NextRequest) {
  const target = safeCallback(request.nextUrl.searchParams.get("callbackUrl"));
  await signIn("keycloak", { redirectTo: target }, { prompt: "create" });
}
