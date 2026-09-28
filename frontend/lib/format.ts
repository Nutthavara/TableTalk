// Display helpers. Reservation dates/times are shown in the restaurant's timezone.

/** "4.3", or null when there are no reviews (render "No reviews yet", never 0 stars). */
export function formatRating(average: number | null): string | null {
  return average === null ? null : average.toFixed(1);
}

export function formatReviewCount(count: number): string {
  if (count === 0) return "No reviews yet";
  return count === 1 ? "1 review" : `${count.toLocaleString("en")} reviews`;
}

export function formatOpeningHours(opening: string, closing: string): string {
  return `${opening} – ${closing}`;
}

/** e.g. "Sat, 3 Oct 2026" for a restaurant-local "YYYY-MM-DD". */
export function formatLocalDate(date: string): string {
  const [y, m, d] = date.split("-").map(Number);
  return new Intl.DateTimeFormat("en-GB", { weekday: "short", day: "numeric", month: "short", year: "numeric", timeZone: "UTC" }).format(
    new Date(Date.UTC(y, m - 1, d)),
  );
}

/** An instant (e.g. cancellation_deadline) shown in the given timezone: "3 Oct 2026, 16:00". */
export function formatInstant(iso: string, timeZone: string): string {
  return new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short", timeZone }).format(new Date(iso));
}

/** Short label for a timezone, e.g. "GMT+7". */
export function timezoneLabel(timeZone: string, at = new Date()): string {
  const part = new Intl.DateTimeFormat("en-GB", { timeZone, timeZoneName: "short" })
    .formatToParts(at)
    .find((p) => p.type === "timeZoneName");
  return part?.value ?? timeZone;
}

/** Today's date "YYYY-MM-DD" in a timezone, for date input minimums. */
export function todayIn(timeZone: string): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone }).format(new Date());
}
