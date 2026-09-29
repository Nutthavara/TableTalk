// Display helpers shared by server and client components.

/** "128 reviews"; zero reviews reads "No reviews yet" (never shown as a 0-star rating). */
export function formatReviewCount(count: number): string {
  if (count === 0) return "No reviews yet";
  return count === 1 ? "1 review" : `${count.toLocaleString("en")} reviews`;
}

export function formatOpeningHours(opening: string, closing: string): string {
  return `${opening} – ${closing}`;
}
