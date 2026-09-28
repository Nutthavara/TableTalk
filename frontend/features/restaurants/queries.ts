import "server-only";

import { api } from "@/lib/api";
import type { Page, Restaurant, RestaurantSort } from "@/types/api";

export function listRestaurants({ sort = "rating", page = 1, pageSize = 12 }: { sort?: RestaurantSort; page?: number; pageSize?: number } = {}) {
  const q = new URLSearchParams({ sort, page: String(page), page_size: String(pageSize) });
  return api<Page<Restaurant>>(`/api/restaurants?${q}`, { auth: "optional" });
}

/** Includes is_owner and my_review_id for a signed-in viewer. */
export function getRestaurant(id: string) {
  return api<Restaurant>(`/api/restaurants/${encodeURIComponent(id)}`, { auth: "optional" });
}

export function listMyRestaurants() {
  return api<{ items: Restaurant[] }>("/api/me/restaurants");
}
