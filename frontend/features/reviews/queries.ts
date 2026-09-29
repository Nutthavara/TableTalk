import "server-only";

import { api } from "@/lib/api";
import type { Page, Review } from "@/types/api";

/** Newest first; each item's is_mine reflects the signed-in viewer. */
export function listReviews(restaurantId: string, { page = 1, pageSize = 10 }: { page?: number; pageSize?: number } = {}) {
  const q = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  return api<Page<Review>>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reviews?${q}`, { auth: "optional" });
}

/**
 * Finds the viewer's review (the API has no single-review endpoint). Checks the already loaded page
 * first, then pages through the list.
 */
export async function findMyReview(restaurantId: string, reviewId: string, loaded: Review[]): Promise<Review | null> {
  const hit = loaded.find((r) => r.id === reviewId);
  if (hit) return hit;
  for (let page = 1; page <= 20; page++) {
    const p = await listReviews(restaurantId, { page, pageSize: 50 });
    const found = p.items.find((r) => r.id === reviewId);
    if (found) return found;
    if (page >= p.total_pages) break;
  }
  return null;
}
