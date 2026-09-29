"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import type { BookingDraft } from "@/features/reservations/booking";
import type { ActionState } from "@/lib/action-state";
import { toActionError } from "@/lib/action-state";
import { api } from "@/lib/api";
import { setFlash } from "@/lib/flash";
import { routes } from "@/lib/routes";
import type { Reservation } from "@/types/api";

function body(draft: BookingDraft) {
  return { party_size: draft.guests, date: draft.date, start_time: draft.start, end_time: draft.end };
}

/** Books a table. Returns the reservation id; the booking panel shows the success state in place. */
export async function createReservation(restaurantId: string, draft: BookingDraft): Promise<ActionState<{ id: string }>> {
  try {
    const r = await api<Reservation>(`/api/restaurants/${encodeURIComponent(restaurantId)}/reservations`, { method: "POST", body: body(draft) });
    return { status: "success", data: { id: r.id } };
  } catch (err) {
    return toActionError(err);
  }
}

/** Same rules as a new booking, enforced by the API. On success returns to My Reservations. */
export async function updateReservation(id: string, draft: BookingDraft): Promise<ActionState> {
  let updated: Reservation;
  try {
    updated = await api<Reservation>(`/api/reservations/${encodeURIComponent(id)}`, { method: "PUT", body: body(draft) });
  } catch (err) {
    return toActionError(err);
  }
  await setFlash({
    message: `Booking updated — ${updated.start_time} – ${updated.end_time}, ${updated.party_size} ${updated.party_size === 1 ? "guest" : "guests"}.`,
  });
  redirect(routes.myReservations());
}

/** Logical cancellation; the reservation moves to the Cancelled tab. */
export async function cancelReservation(id: string): Promise<ActionState> {
  try {
    await api<Reservation>(`/api/reservations/${encodeURIComponent(id)}`, { method: "DELETE" });
  } catch (err) {
    return toActionError(err);
  }
  refresh();
  return { status: "success" };
}
