// User-facing text for API error codes. The API's own message is the fallback, so a new code still
// shows something sensible.
const messages: Record<string, string> = {
  UNAUTHORIZED: "Your session has expired. Please sign in again.",
  FORBIDDEN: "You don't have permission to do that.",
  NOT_FOUND: "We couldn't find what you were looking for.",
  RESTAURANT_NOT_FOUND: "This restaurant no longer exists.",
  RESERVATION_NOT_FOUND: "This reservation no longer exists.",
  REVIEW_NOT_FOUND: "This review no longer exists.",
  IMAGE_NOT_FOUND: "This image no longer exists.",
  VALIDATION_ERROR: "Please check the highlighted fields.",
  INSUFFICIENT_CAPACITY: "Not enough seats are available at that time. Try another time or a smaller party.",
  PARTY_SIZE_EXCEEDS_CAPACITY: "Your party is larger than this restaurant can seat.",
  RESERVATION_IN_PAST: "Please choose a time in the future.",
  OUTSIDE_OPENING_HOURS: "The whole booking must be within the restaurant's opening hours.",
  CANCELLATION_DEADLINE_PASSED: "It's too late to cancel this reservation under the restaurant's cancellation policy.",
  EDIT_DEADLINE_PASSED: "It's too late to change this reservation under the restaurant's cancellation policy.",
  RESERVATION_NOT_ACTIVE: "This reservation has already been cancelled.",
  CANNOT_REVIEW_OWN_RESTAURANT: "You can't review your own restaurant.",
  REVIEW_ALREADY_EXISTS: "You've already reviewed this restaurant. You can edit your review instead.",
  LAST_IMAGE: "A restaurant needs at least one photo.",
  TOO_MANY_IMAGES: "A restaurant can have at most 10 photos.",
  RESTAURANT_HAS_UPCOMING_RESERVATIONS: "This restaurant still has upcoming reservations, so it can't be deleted yet.",
  CONFLICTS_WITH_EXISTING_RESERVATIONS:
    "This change would conflict with upcoming reservations (seats or opening hours). Adjust it or wait until they have passed.",
  API_UNAVAILABLE: "The service is temporarily unavailable. Please try again shortly.",
  INTERNAL_ERROR: "Something went wrong on our side. Please try again.",
};

export function errorMessage(code: string, fallback?: string): string {
  return messages[code] ?? fallback ?? "Something went wrong. Please try again.";
}
