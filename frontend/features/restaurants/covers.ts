import "server-only";

import { getRestaurant } from "@/features/restaurants/queries";
import type { RestaurantImage } from "@/types/api";

/**
 * Cover photos for a set of restaurants (reservations don't carry images). Missing or deleted
 * restaurants simply get no cover.
 */
export async function restaurantCovers(ids: string[]): Promise<Map<string, RestaurantImage | undefined>> {
  const unique = [...new Set(ids)];
  const results = await Promise.allSettled(unique.map((id) => getRestaurant(id)));
  return new Map(unique.map((id, i) => [id, results[i].status === "fulfilled" ? results[i].value.images[0] : undefined]));
}
