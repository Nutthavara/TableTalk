import "server-only";

import { cookies } from "next/headers";

import { FLASH_COOKIE, type Flash } from "@/lib/flash-types";

/**
 * Leaves a one-shot message for the next page (read and deleted by <Toaster />). Use before
 * redirect() in a Server Action. Not httpOnly: it only carries UI text. base64url keeps the value
 * free of characters that cookie serialization would escape.
 */
export async function setFlash(flash: Flash) {
  const value = Buffer.from(JSON.stringify(flash), "utf8").toString("base64url");
  (await cookies()).set(FLASH_COOKIE, value, { path: "/", maxAge: 60, sameSite: "lax" });
}
