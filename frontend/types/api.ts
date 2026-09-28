// Response shapes of the Go API. Times ending in _at are UTC ISO-8601 strings.

export type ApiErrorBody = {
  error: {
    code: string;
    message: string;
    details?: Record<string, string>;
  };
};

export type Page<T> = {
  items: T[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type Me = {
  id: string;
  username: string;
  display_name: string;
  email?: string;
};

export type RestaurantImage = {
  id: string;
  url: string;
  position: number;
};

export type Restaurant = {
  id: string;
  owner_id: string;
  owner_name: string;
  name: string;
  description: string;
  capacity: number;
  /** "HH:MM" in the restaurant's timezone. */
  opening_time: string;
  closing_time: string;
  /** IANA timezone, e.g. "Asia/Bangkok". */
  timezone: string;
  cancellation_notice_hours: number;
  images: RestaurantImage[];
  /** null when the restaurant has no reviews yet (not a 0-star rating). */
  average_rating: number | null;
  review_count: number;
  is_owner: boolean;
  /** Only on the detail endpoint, for a signed-in viewer who has reviewed it. */
  my_review_id?: string;
  created_at: string;
  updated_at: string;
};

export type RestaurantSort = "rating" | "reviews";

export type ReservationStatus = "ACTIVE" | "CANCELLED";
export type ReservationFilter = "upcoming" | "past" | "cancelled";

export type Reservation = {
  id: string;
  restaurant_id: string;
  restaurant_name: string;
  timezone: string;
  party_size: number;
  start_at: string;
  end_at: string;
  /** Restaurant-local date "YYYY-MM-DD" and times "HH:MM" as booked. */
  date: string;
  start_time: string;
  end_time: string;
  status: ReservationStatus;
  created_at: string;
  updated_at: string;
  cancelled_at: string | null;
  cancellation_notice_hours: number;
  /** Last instant the reservation can be edited or cancelled. */
  cancellation_deadline: string;
  /** Informational only; the API enforces the rule on every change. */
  can_modify: boolean;
};

export type Review = {
  id: string;
  restaurant_id: string;
  user_id: string;
  author_name: string;
  rating: number;
  comment: string;
  is_mine: boolean;
  created_at: string;
  updated_at: string;
};
