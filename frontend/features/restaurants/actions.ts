"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import { formFiles, formInt, formString } from "@/lib/form";
import { routes } from "@/lib/routes";
import type { Restaurant } from "@/types/api";

const textFields = ["name", "description", "opening_time", "closing_time", "timezone"] as const;

function restaurantFields(fd: FormData) {
  const fieldErrors: Record<string, string> = {};
  const body = {
    name: formString(fd, "name"),
    description: formString(fd, "description"),
    capacity: formInt(fd, "capacity", fieldErrors),
    opening_time: formString(fd, "opening_time"),
    closing_time: formString(fd, "closing_time"),
    timezone: formString(fd, "timezone"),
    cancellation_notice_hours: formInt(fd, "cancellation_notice_hours", fieldErrors),
  };
  return { body, fieldErrors };
}

function invalid(fieldErrors: Record<string, string>): ActionState<never> {
  return { status: "error", code: "VALIDATION_ERROR", message: "Please check the highlighted fields.", fieldErrors };
}

/** Form fields: name, description, capacity, opening_time, closing_time, timezone, cancellation_notice_hours, images (≥1). */
export async function createRestaurant(_prev: ActionState, fd: FormData): Promise<ActionState> {
  const { fieldErrors } = restaurantFields(fd);
  const images = formFiles(fd, "images");
  if (images.length === 0) fieldErrors.images = "Please add at least one photo.";
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);

  // Forward only the known fields, so nothing else the browser sent reaches the API.
  const upload = new FormData();
  for (const name of [...textFields, "capacity", "cancellation_notice_hours"]) upload.set(name, formString(fd, name));
  for (const image of images) upload.append("images", image);

  let created: Restaurant;
  try {
    created = await api<Restaurant>("/api/restaurants", { method: "POST", body: upload });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.restaurant(created.id));
}

export async function updateRestaurant(id: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const { body, fieldErrors } = restaurantFields(fd);
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);
  try {
    await api<Restaurant>(`/api/restaurants/${encodeURIComponent(id)}`, { method: "PUT", body });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.restaurant(id));
}

export async function deleteRestaurant(id: string): Promise<ActionState> {
  try {
    await api(`/api/restaurants/${encodeURIComponent(id)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.myRestaurants());
}

/** Form field: images (one or more files). */
export async function addRestaurantImages(id: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const images = formFiles(fd, "images");
  if (images.length === 0) return invalid({ images: "Please choose at least one photo." });
  const upload = new FormData();
  for (const image of images) upload.append("images", image);
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
