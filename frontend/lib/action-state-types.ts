// Shared between Server Actions and the Client Components that render their results
// (useActionState). Kept free of server-only imports.
export type ActionState<T = undefined> =
  | { status: "idle" }
  | { status: "success"; data?: T }
  | { status: "error"; code: string; message: string; fieldErrors?: Record<string, string> };

export const initialActionState: ActionState<never> = { status: "idle" };
