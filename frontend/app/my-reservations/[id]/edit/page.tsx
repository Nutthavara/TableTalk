import type { Metadata } from "next";
import Link from "next/link";

import { Alert } from "@/components/ui/alert";
import { BackLink } from "@/components/ui/back-link";
import { buttonClass } from "@/components/ui/button";
import { EditBookingForm } from "@/features/reservations/components/edit-booking-form";
import { reservationView } from "@/features/reservations/components/reservation-view";
import { getReservation } from "@/features/reservations/queries";
import { getRestaurant } from "@/features/restaurants/queries";
import { handlePageError, requireSignedIn } from "@/lib/auth-guard";
import { routes } from "@/lib/routes";
import type { Reservation, Restaurant } from "@/types/api";

export const metadata: Metadata = { title: "Edit Booking" };

export default async function EditReservationPage({ params }: PageProps<"/my-reservations/[id]/edit">) {
  const { id } = await params;
  const here = routes.editReservation(id);
  await requireSignedIn(here);

  let reservation: Reservation;
  let restaurant: Restaurant;
  try {
    reservation = await getReservation(id);
    restaurant = await getRestaurant(reservation.restaurant_id);
  } catch (err) {
    handlePageError(err, here);
  }

  const current = { date: reservation.date, start: reservation.start_time, end: reservation.end_time, guests: reservation.party_size };
  const editable = reservation.status === "ACTIVE" && reservation.can_modify;
  const view = reservationView(reservation, "upcoming", "");

  return (
    <main className="mx-auto flex w-full max-w-[1120px] flex-col gap-7 px-4 pt-4 pb-12 sm:px-8 sm:pt-6 sm:pb-20 xl:px-0">
      <BackLink href={routes.myReservations()}>My Reservations</BackLink>
      <div className="flex flex-col gap-2">
        <h1 className="m-0 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-[44px]">Edit Booking</h1>
        <p className="m-0 text-[17px] text-ink-2">
          {restaurant.name} · open {restaurant.opening_time} – {restaurant.closing_time} · {restaurant.capacity} seats
        </p>
      </div>
      {editable ? (
        <EditBookingForm reservationId={reservation.id} current={current} rules={restaurant} cover={restaurant.images[0]} restaurantName={restaurant.name} />
      ) : (
        <div className="flex max-w-2xl flex-col gap-4">
          <Alert tone="warning" title="This booking can no longer be changed">
            {reservation.status === "CANCELLED" ? "This reservation has been cancelled." : view.lockReason}
          </Alert>
          <Link href={routes.myReservations()} className={buttonClass("secondary", "md", "self-start")}>
            Back to My Reservations
          </Link>
        </div>
      )}
    </main>
  );
}
