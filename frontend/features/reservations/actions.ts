"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import { formInt, formString } from "@/lib/form";
import { routes } from "@/lib/routes";
import type { Reservation } from "@/types/api";

/** Form fields: party_size, date (YYYY-MM-DD), start_time, end_time (HH:MM, restaurant-local). */
function reservationBody(fd: FormData) {
  const fieldErrors: Record<string, string> = {};
  const body = {
    party_size: formInt(fd, "party_size", fieldErrors),
    date: formString(fd, "date"),
    start_time: formString(fd, "start_time"),
    end_time: formString(fd, "end_time"),
  };
  return { body, fieldErrors };
}

function invalid(fieldErrors: Record<string, string>): ActionState<never> {
  return { status: "error", code: "VALIDATION_ERROR", message: "Please check the highlighted fields.", fieldErrors };
}

export async function createReservation(restaurantId: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const { body, fieldErrors } = reservationBody(fd);
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);
  try {
    await api<Reservation>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reservations`, { method: "POST", body });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.myReservations());
}

export async function updateReservation(id: string, _prev: ActionState, fd: FormData): Promise<ActionState> {
  const { body, fieldErrors } = reservationBody(fd);
  if (Object.keys(fieldErrors).length > 0) return invalid(fieldErrors);
  try {
    await api<Reservation>(`/api/reservations/${encodeURIComponent(id)}`, { method: "PUT", body });
  } catch (err) {
    return toActionError(err);
  }
  redirect(routes.myReservations());
}

/** Logical cancellation; the reservation stays in history with status CANCELLED. */
export async function cancelReservation(id: string): Promise<ActionState> {
  try {
    await api<Reservation>(`/api/reservations/${encodeURIComponent(id)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}
