// Every in-app URL in one place, so page structure can change without hunting through actions.
export const routes = {
  home: () => "/",
  login: (callbackUrl?: string) => (callbackUrl ? `/login?callbackUrl=${encodeURIComponent(callbackUrl)}` : "/login"),
  restaurant: (id: string) => `/restaurants/${id}`,
  reserve: (restaurantId: string) => `/restaurants/${restaurantId}/reserve`,
  review: (restaurantId: string) => `/restaurants/${restaurantId}/review`,
  myReservations: (filter?: string) => (filter ? `/me/reservations?filter=${filter}` : "/me/reservations"),
  editReservation: (id: string) => `/me/reservations/${id}/edit`,
  myRestaurants: () => "/me/restaurants",
  newRestaurant: () => "/me/restaurants/new",
  editRestaurant: (id: string) => `/me/restaurants/${id}/edit`,
};
