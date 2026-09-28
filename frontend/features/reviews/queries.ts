import "server-only";

import { api } from "@/lib/api";
import type { Page, Review } from "@/types/api";

/** Newest first; each item's is_mine reflects the signed-in viewer. */
export function listReviews(restaurantId: string, { page = 1, pageSize = 10 }: { page?: number; pageSize?: number } = {}) {
  const q = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  return api<Page<Review>>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reviews?${q}`, { auth: "optional" });
}
