// Booking draft helpers shared by the booking panel and Edit Booking. Checks here only give fast
// feedback; the Go API re-validates everything (and alone decides seat availability).

export type BookingDraft = { date: string; start: string; end: string; guests: number };

export type BookingRules = {
  opening_time: string;
  closing_time: string;
  capacity: number;
  timezone: string;
  cancellation_notice_hours: number;
};

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

export function toMinutes(t: string): number | null {
  const m = /^(\d{2}):(\d{2})$/.exec(t);
  return m ? Number(m[1]) * 60 + Number(m[2]) : null;
}

const pad = (n: number) => String(n).padStart(2, "0");
const hhmm = (minutes: number) => `${pad(Math.floor(minutes / 60))}:${pad(minutes % 60)}`;

/** Current date and minute-of-day in the restaurant's timezone. */
export function nowInZone(timeZone: string, at = new Date()): { date: string; minutes: number } {
  const parts = Object.fromEntries(
    new Intl.DateTimeFormat("en-CA", { timeZone, year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", hourCycle: "h23" })
      .formatToParts(at)
      .map((p) => [p.type, p.value]),
  );
  return { date: `${parts.year}-${parts.month}-${parts.day}`, minutes: Number(parts.hour) * 60 + Number(parts.minute) };
}

function addDays(date: string, days: number): string {
  const [y, m, d] = date.split("-").map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d + days));
  return `${dt.getUTCFullYear()}-${pad(dt.getUTCMonth() + 1)}-${pad(dt.getUTCDate())}`;
}

/** "Sat, 10 Oct 2026" */
export function dateLabel(date: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) return "—";
  const [y, m, d] = date.split("-").map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  return `${WEEKDAYS[dt.getUTCDay()]}, ${d} ${MONTHS[m - 1]} ${y}`;
}

/** "18:00 – 20:00 (2 h)" */
export function timeLabel(start: string, end: string, withDuration = true): string {
  const s = toMinutes(start);
  const e = toMinutes(end);
  let dur = "";
  if (withDuration && s !== null && e !== null && e > s) {
    const m = e - s;
    dur = ` (${[Math.floor(m / 60) ? `${Math.floor(m / 60)} h` : "", m % 60 ? `${m % 60} min` : ""].filter(Boolean).join(" ")})`;
  }
  return `${start || "--:--"} – ${end || "--:--"}${dur}`;
}

export const guestsLabel = (n: number) => `${n} ${n === 1 ? "guest" : "guests"}`;

/** Free-cancellation deadline as restaurant-local wall time: "16:00, Sat 10 Oct". */
export function deadlineLabel(date: string, start: string, noticeHours: number): string {
  const s = toMinutes(start);
  if (s === null || !/^\d{4}-\d{2}-\d{2}$/.test(date)) return "";
  const [y, m, d] = date.split("-").map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d, 0, s) - noticeHours * 3600_000);
  return `${hhmm(dt.getUTCHours() * 60 + dt.getUTCMinutes())}, ${WEEKDAYS[dt.getUTCDay()]} ${dt.getUTCDate()} ${MONTHS[dt.getUTCMonth()]}`;
}

export type BookingCheck = { errors: string[]; dateInvalid: boolean; timeInvalid: boolean };

export function checkBooking(draft: BookingDraft, rules: BookingRules, now = nowInZone(rules.timezone)): BookingCheck {
  const errors: string[] = [];
  let dateInvalid = false;
  let timeInvalid = false;
  const s = toMinutes(draft.start);
  const e = toMinutes(draft.end);
  const open = toMinutes(rules.opening_time)!;
  const close = toMinutes(rules.closing_time)!;

  if (!draft.date || s === null || e === null) {
    errors.push("Choose a date, start time and end time.");
  } else {
    if (draft.date < now.date || (draft.date === now.date && s <= now.minutes)) {
      errors.push("Selected time has already passed. Choose a later date or time.");
      dateInvalid = true;
      if (draft.date === now.date) timeInvalid = true;
    }
    if (e <= s) {
      errors.push("End time must be later than start time.");
      timeInvalid = true;
    } else if (s < open || e > close) {
      errors.push(`Selected time is outside opening hours (${rules.opening_time} – ${rules.closing_time}).`);
      timeInvalid = true;
    }
  }
  if (draft.guests < 1 || draft.guests > rules.capacity) {
    errors.push(`Number of guests must be between 1 and ${rules.capacity}.`);
  }
  return { errors, dateInvalid, timeInvalid };
}

/** A sensible first suggestion: 18:00–20:00 when it fits the hours, else the first two open hours; tomorrow if already past. */
export function defaultDraft(rules: BookingRules, now = nowInZone(rules.timezone)): BookingDraft {
  const open = toMinutes(rules.opening_time)!;
  const close = toMinutes(rules.closing_time)!;
  let start = 18 * 60;
  if (start < open || start + 120 > close) start = open;
  const end = Math.min(start + 120, close);
  const date = start <= now.minutes ? addDays(now.date, 1) : now.date;
  return { date, start: hhmm(start), end: hhmm(end), guests: Math.min(2, rules.capacity) };
}

/** Reads a draft from URL params (used to restore the form after login), falling back to defaults. */
export function draftFromParams(params: Record<string, string | string[] | undefined>, rules: BookingRules): BookingDraft {
  const d = defaultDraft(rules);
  const str = (k: string) => (typeof params[k] === "string" ? (params[k] as string) : undefined);
  const guests = Number(str("guests"));
  return {
    date: /^\d{4}-\d{2}-\d{2}$/.test(str("date") ?? "") ? str("date")! : d.date,
    start: toMinutes(str("start") ?? "") !== null ? str("start")! : d.start,
    end: toMinutes(str("end") ?? "") !== null ? str("end")! : d.end,
    guests: Number.isInteger(guests) && guests >= 1 && guests <= rules.capacity ? guests : d.guests,
  };
}
