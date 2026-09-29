"use server";

import { refresh } from "next/cache";

import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import type { Review } from "@/types/api";

export type ReviewDraft = { rating: number; comment: string };

export async function createReview(restaurantId: string, draft: ReviewDraft): Promise<ActionState> {
  try {
    await api<Review>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reviews`, { method: "POST", body: draft });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}

export async function updateReview(id: string, draft: ReviewDraft): Promise<ActionState> {
  try {
    await api<Review>(`/api/reviews/${encodeURIComponent(id)}`, { method: "PUT", body: draft });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
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
