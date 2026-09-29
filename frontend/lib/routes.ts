// Every in-app URL in one place (routes follow the design blueprint).
const withQuery = (path: string, params: Record<string, string | number | undefined>) => {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) if (v !== undefined && v !== "") q.set(k, String(v));
  const s = q.toString();
  return s ? `${path}?${s}` : path;
};

export const routes = {
  home: () => "/",
  restaurants: (params: { q?: string; sort?: string; page?: number } = {}) => withQuery("/restaurants", params),
  restaurant: (id: string, params: Record<string, string | number | undefined> = {}) => withQuery(`/restaurants/${id}`, params),
  myReservations: (tab?: string) => withQuery("/my-reservations", { tab }),
  editReservation: (id: string) => `/my-reservations/${id}/edit`,
  myRestaurants: () => "/my-restaurants",
  newRestaurant: () => "/my-restaurants/new",
  editRestaurant: (id: string) => `/my-restaurants/${id}/edit`,
  /** Starts the Keycloak login and comes back to callbackUrl afterwards. */
  login: (callbackUrl?: string) => withQuery("/login", { callbackUrl }),
  register: (callbackUrl?: string) => withQuery("/register", { callbackUrl }),
};

/** Only same-origin relative paths are accepted as post-login destinations (prevents open redirects). */
export function safeCallback(target: string | null | undefined): string {
  return target && target.startsWith("/") && !target.startsWith("//") && !target.startsWith("/\\") ? target : "/restaurants";
}
