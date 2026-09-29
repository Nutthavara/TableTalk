"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import { setFlash } from "@/lib/flash";
import { routes } from "@/lib/routes";
import type { Restaurant } from "@/types/api";

export type RestaurantFields = {
  name: string;
  description: string;
  capacity: number;
  opening_time: string;
  closing_time: string;
  timezone: string;
  cancellation_notice_hours: number;
};

const fieldNames = ["name", "description", "capacity", "opening_time", "closing_time", "timezone", "cancellation_notice_hours"] as const;

/**
 * Creates the restaurant with its photos in one request (a restaurant always has at least one photo).
 * Only the known fields and the "images" files are forwarded to the API.
 */
export async function createRestaurant(fd: FormData): Promise<ActionState> {
  const upload = new FormData();
  for (const name of fieldNames) upload.set(name, String(fd.get(name) ?? "").trim());
  for (const file of fd.getAll("images")) if (file instanceof File && file.size > 0) upload.append("images", file);

  let created: Restaurant;
  try {
    created = await api<Restaurant>("/api/restaurants", { method: "POST", body: upload });
  } catch (err) {
    return toActionError(err);
  }
  await setFlash({ message: `${created.name} is now live on TableTalk.` });
  redirect(routes.restaurant(created.id));
}

export async function updateRestaurant(id: string, fields: RestaurantFields): Promise<ActionState> {
  let updated: Restaurant;
  try {
    updated = await api<Restaurant>(`/api/restaurants/${encodeURIComponent(id)}`, { method: "PUT", body: fields });
  } catch (err) {
    return toActionError(err);
  }
  await setFlash({ message: `${updated.name} — changes saved.`, action: { label: "View page", href: routes.restaurant(id) } });
  redirect(routes.myRestaurants());
}

/** Called from My Restaurants itself, so the page refreshes in place (the caller shows the toast). */
export async function deleteRestaurant(id: string): Promise<ActionState> {
  try {
    await api(`/api/restaurants/${encodeURIComponent(id)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}

/** Form field: images (one or more files). Takes effect immediately. */
export async function addRestaurantImages(id: string, fd: FormData): Promise<ActionState> {
  const upload = new FormData();
  for (const file of fd.getAll("images")) if (file instanceof File && file.size > 0) upload.append("images", file);
  try {
    await api(`/api/restaurants/${encodeURIComponent(id)}/images`, { method: "POST", body: upload });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}

export async function deleteRestaurantImage(id: string, imageId: string): Promise<ActionState> {
  try {
    await api(`/api/restaurants/${encodeURIComponent(id)}/images/${encodeURIComponent(imageId)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}
