// Display-ready view of a reservation for cards and dialogs (restaurant-local wall times).
import { dateLabel, guestsLabel, timeLabel } from "@/features/reservations/booking";
import type { Reservation } from "@/types/api";

export type ReservationKind = "upcoming" | "past" | "cancelled";

const localDeadline = (r: Reservation) => {
  const at = new Date(r.cancellation_deadline);
  const time = new Intl.DateTimeFormat("en-GB", { timeZone: r.timezone, hour: "2-digit", minute: "2-digit", hourCycle: "h23" }).format(at);
  // "Sat 3 Oct" from the restaurant-local date (avoids Intl's "Sept").
  const day = dateLabel(new Intl.DateTimeFormat("en-CA", { timeZone: r.timezone }).format(at)).replace(/,/, "").replace(/ \d{4}$/, "");
  return { time, day };
};

export function reservationView(r: Reservation, kind: ReservationKind, today: string) {
  const deadline = localDeadline(r);
  const notice = `${r.cancellation_notice_hours} ${r.cancellation_notice_hours === 1 ? "hour" : "hours"}`;
  const isToday = r.date === today;
  return {
    date: isToday ? `Today, ${dateLabel(r.date).replace(",", "")}` : dateLabel(r.date),
    time: timeLabel(r.start_time, r.end_time, false),
    guests: guestsLabel(r.party_size),
    isToday: isToday && kind === "upcoming",
    policy: `Free cancellation until ${deadline.time}, ${deadline.day} (${notice} before)`,
    // The API allows neither edits nor cancellation after the deadline (same rule for both).
    lockReason: `Changes and cancellation are no longer available. ${r.restaurant_name} requires ${notice}' notice — the free cancellation window closed at ${deadline.time}, ${deadline.day}.`,
    cancelledOn: r.cancelled_at ? dateLabel(new Intl.DateTimeFormat("en-CA", { timeZone: r.timezone }).format(new Date(r.cancelled_at))).replace(/^\w+, /, "") : null,
  };
}
