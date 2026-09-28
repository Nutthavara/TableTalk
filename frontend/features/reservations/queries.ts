import "server-only";

import { api } from "@/lib/api";
import type { Page, Reservation, ReservationFilter } from "@/types/api";

export function listMyReservations({ filter = "upcoming", page = 1, pageSize = 12 }: { filter?: ReservationFilter; page?: number; pageSize?: number } = {}) {
  const q = new URLSearchParams({ filter, page: String(page), page_size: String(pageSize) });
  return api<Page<Reservation>>(`/api/me/reservations?${q}`);
}

/** Only the user who made the reservation can read it (others get 403). */
export function getReservation(id: string) {
  return api<Reservation>(`/api/reservations/${encodeURIComponent(id)}`);
}
