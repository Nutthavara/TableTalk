import "server-only";

import { ApiError } from "@/lib/api";
import { errorMessage } from "@/lib/error-messages";

export type { ActionState } from "@/lib/action-state-types";
import type { ActionState } from "@/lib/action-state-types";

/** Converts an error thrown while calling the API into state a form can render. */
export function toActionError(err: unknown): ActionState<never> {
  if (err instanceof ApiError) {
    return {
      status: "error",
      code: err.code,
      message: errorMessage(err.code, err.message),
      fieldErrors: err.details,
    };
  }
  console.error("Unexpected error in server action", err);
  return { status: "error", code: "INTERNAL_ERROR", message: errorMessage("INTERNAL_ERROR") };
}
