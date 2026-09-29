import "server-only";

import { api } from "@/lib/api";
import type { Page, Restaurant, RestaurantSort } from "@/types/api";

export function listRestaurants({
  q = "",
  sort = "rating",
  page = 1,
  pageSize = 12,
}: { q?: string; sort?: RestaurantSort; page?: number; pageSize?: number } = {}) {
  const params = new URLSearchParams({ sort, page: String(page), page_size: String(pageSize) });
  if (q.trim()) params.set("q", q.trim());
  return api<Page<Restaurant>>(`/api/restaurants?${params}`, { auth: "optional" });
}

/** Includes is_owner and my_review_id for a signed-in viewer. */
export function getRestaurant(id: string) {
  return api<Restaurant>(`/api/restaurants/${encodeURIComponent(id)}`, { auth: "optional" });
}

export function listMyRestaurants() {
  return api<{ items: Restaurant[] }>("/api/me/restaurants");
}
