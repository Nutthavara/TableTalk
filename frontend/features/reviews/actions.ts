"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import { formInt, formString } from "@/lib/form";
import { routes } from "@/lib/routes";
import type { Review } from "@/types/api";

/** Form fields: rating (1–5), comment (optional). */
function reviewBody(fd: FormData) {
  const fieldErrors: Record<string, string> = {};
  const body = { rating: formInt(fd, "rating", fieldErrors), comment: formString(fd, "comment") };
  if (!fieldErrors.rating && (body.rating < 1 || body.rating > 5)) fieldErrors.rating = "Please choose a rating from 1 to 5.";
  return { body, fieldErrors };
}

function invalid(fieldErrors: Record<string, string>): ActionState<never> {
  return { status: "error", code: "VALIDATION_ERROR", message: "Please check the highlighted fields.", fieldErrors };
}

export async function createReview(restaurantId: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const { body, fieldErrors } = reviewBody(fd);
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);
  try {
    await api<Review>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reviews`, { method: "POST", body });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.restaurant(restaurantId));
}

export async function updateReview(id: string, restaurantId: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const { body, fieldErrors } = reviewBody(fd);
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);
  try {
    await api<Review>(`/api/reviews/${encodeURIComponent(id)}`, { method: "PUT", body });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.restaurant(restaurantId));
}

export async function deleteReview(id: string): Promise<ActionState> {
  try {
    await api(`/api/reviews/${encodeURIComponent(id)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}
