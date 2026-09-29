import type { Metadata } from "next";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { buttonClass } from "@/components/ui/button";
import { Calendar, Check, Clock, Lock, Pencil, Users } from "@/components/ui/icons";
import { EmptyState, ErrorState } from "@/components/ui/states";
import { nowInZone } from "@/features/reservations/booking";
import { CancelBookingButton } from "@/features/reservations/components/cancel-button";
import { reservationView, type ReservationKind } from "@/features/reservations/components/reservation-view";
import { countMyReservations, listMyReservations } from "@/features/reservations/queries";
import { restaurantCovers } from "@/features/restaurants/covers";
import { RestaurantPhoto } from "@/features/restaurants/components/restaurant-image";
import { ApiError } from "@/lib/api";
import { requireSignedIn } from "@/lib/auth-guard";
import { cn } from "@/lib/cn";
import { routes } from "@/lib/routes";
import type { Page, Reservation } from "@/types/api";

export const metadata: Metadata = { title: "My Reservations" };

const tabs: { id: ReservationKind; label: string }[] = [
  { id: "upcoming", label: "Upcoming" },
  { id: "past", label: "Past" },
  { id: "cancelled", label: "Cancelled" },
];

const statusBadge: Record<ReservationKind, { tone: "confirmed" | "completed" | "cancelled"; label: string }> = {
  upcoming: { tone: "confirmed", label: "Confirmed" },
  past: { tone: "completed", label: "Completed" },
  cancelled: { tone: "cancelled", label: "Cancelled" },
};

const emptyText: Record<ReservationKind, string> = {
  upcoming: "You don't have any upcoming reservations.",
  past: "No past reservations yet.",
  cancelled: "No cancelled reservations.",
};

export default async function MyReservationsPage({ searchParams }: PageProps<"/my-reservations">) {
  const sp = await searchParams;
  const tab: ReservationKind = sp.tab === "past" || sp.tab === "cancelled" ? sp.tab : "upcoming";
  const page = Math.max(1, Number(sp.page) || 1);
  await requireSignedIn(routes.myReservations(tab === "upcoming" ? undefined : tab));

  let list: Page<Reservation> | null = null;
  let counts: Record<ReservationKind, number> | null = null;
  try {
    [list, counts] = await Promise.all([listMyReservations({ filter: tab, page }), countMyReservations()]);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) await requireSignedIn(routes.myReservations());
  }
  const covers = list ? await restaurantCovers(list.items.map((r) => r.restaurant_id)) : new Map();
  const noneAtAll = counts && counts.upcoming + counts.past + counts.cancelled === 0;

  return (
    <main className="mx-auto flex w-full max-w-[1120px] flex-col gap-7 px-4 pt-5 pb-12 sm:px-8 sm:pt-14 sm:pb-20 xl:px-0">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <div className="flex flex-col gap-2.5">
          <h1 className="m-0 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-[44px]">My Reservations</h1>
          <p className="m-0 text-[17px] text-ink-2">Manage upcoming bookings and look back on past visits.</p>
        </div>
        <Link href={routes.restaurants()} className={buttonClass("secondary", "md", "hidden sm:inline-flex")}>
          Find a restaurant
        </Link>
      </div>

      <nav aria-label="Reservation status" className="flex gap-2 overflow-x-auto border-b border-line-mid">
        {tabs.map((t) => {
          const on = t.id === tab;
          return (
            <Link
              key={t.id}
              href={routes.myReservations(t.id === "upcoming" ? undefined : t.id)}
              aria-current={on ? "page" : undefined}
              className={cn(
                "flex h-[52px] shrink-0 items-center gap-2 px-4 text-base no-underline",
                on ? "font-bold text-ink shadow-[inset_0_-3px_0_var(--color-paprika)] hover:text-ink" : "rounded-t-[10px] font-semibold text-ink-2 hover:bg-hover hover:text-ink-2",
              )}
            >
              {t.label}
              {counts && (
                <span className={cn("flex h-6 min-w-6 items-center justify-center rounded-full px-[7px] text-[13px]", on ? "bg-ink text-white" : "bg-line-soft text-ink-2")}>
                  {counts[t.id]}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      <div className="flex flex-col gap-5">
        {!list ? (
          <ErrorState title="We couldn't load your reservations" retryHref={routes.myReservations(tab === "upcoming" ? undefined : tab)} />
        ) : list.items.length === 0 ? (
          noneAtAll ? (
            <EmptyState
              title="You don't have any reservations yet."
              body="Find a restaurant you like and book a table in a few taps."
              action={
                <Link href={routes.restaurants()} className={buttonClass("primary", "lg")}>
                  Browse restaurants
                </Link>
              }
            />
          ) : (
            <div className="flex flex-col items-center gap-3 rounded-[20px] border border-dashed border-line bg-surface px-6 py-16 text-center">
              <strong className="text-lg">{emptyText[tab]}</strong>
              <Link href={routes.restaurants()} className="text-[15px] font-semibold">
                Browse restaurants
              </Link>
            </div>
          )
        ) : (
          list.items.map((r) => {
            const v = reservationView(r, tab, nowInZone(r.timezone).date);
            const locked = tab === "upcoming" && !r.can_modify;
            const lockId = `lock-${r.id}`;
            return (
              <article
                key={r.id}
                aria-label={`${r.restaurant_name}, ${v.date}, ${v.time}, ${v.guests}`}
                className="grid grid-cols-1 items-stretch gap-4 rounded-[22px] border border-line-soft bg-surface p-4 shadow-[0_1px_2px_rgba(34,28,24,.04)] md:grid-cols-[220px_minmax(0,1fr)_auto] md:gap-7"
              >
                <div className={cn("relative h-40 overflow-hidden rounded-[14px] md:h-[168px]", tab === "cancelled" && "opacity-55")}>
                  <RestaurantPhoto image={covers.get(r.restaurant_id)} alt="" sizes="(min-width: 768px) 220px, 100vw" />
                </div>
                <div className="flex flex-col gap-3 md:py-2">
                  <div className="flex flex-wrap items-center gap-3">
                    <Link href={routes.restaurant(r.restaurant_id)} className="font-display text-2xl font-bold tracking-[-0.01em] text-ink no-underline hover:text-paprika-dark">
                      {r.restaurant_name}
                    </Link>
                    <Badge tone={statusBadge[tab].tone}>{statusBadge[tab].label}</Badge>
                    {v.isToday && <Badge tone="today">Today</Badge>}
                  </div>
                  <div className="flex flex-wrap gap-x-6 gap-y-2 text-base">
                    <span className="flex items-center gap-2">
                      <Calendar size={18} className="text-ink-3" />
                      <strong>{v.date}</strong>
                    </span>
                    <span className="flex items-center gap-2">
                      <Clock size={18} className="text-ink-3" />
                      <strong>{v.time}</strong>
                    </span>
                    <span className="flex items-center gap-2">
                      <Users size={18} className="text-ink-3" />
                      <strong>{v.guests}</strong>
                    </span>
                  </div>
                  {tab === "upcoming" && !locked && (
                    <span className="flex items-center gap-2 text-sm text-herb-ink">
                      <Check size={16} strokeWidth={2} />
                      {v.policy}
                    </span>
                  )}
                  {locked && (
                    <span id={lockId} className="flex max-w-[560px] items-start gap-2 rounded-[10px] bg-saffron-soft px-3 py-2.5 text-sm leading-[1.45] text-[#5E3F0B]">
                      <Lock size={16} className="mt-0.5 shrink-0" />
                      <span>{v.lockReason}</span>
                    </span>
                  )}
                  {tab === "cancelled" && v.cancelledOn && <span className="text-sm text-ink-3">Cancelled on {v.cancelledOn}</span>}
                </div>
                <div className="flex flex-col justify-center gap-2.5 md:min-w-[190px] md:pr-2">
                  {tab === "upcoming" &&
                    (locked ? (
                      <>
                        <button type="button" disabled aria-describedby={lockId} className={buttonClass("secondary", "md", "h-[46px]")}>
                          <Pencil size={16} />
                          Edit Booking
                        </button>
                        <button type="button" disabled aria-describedby={lockId} className={buttonClass("dangerOutline", "md", "h-[46px]")}>
                          Cancel Booking
                        </button>
                      </>
                    ) : (
                      <>
                        <Link href={routes.editReservation(r.id)} className={buttonClass("secondary", "md", "h-[46px]")}>
                          <Pencil size={16} />
                          Edit Booking
                        </Link>
                        <CancelBookingButton reservation={{ id: r.id, name: r.restaurant_name, date: v.date, time: v.time, guests: v.guests, policy: v.policy }} />
                      </>
                    ))}
                  {tab === "past" && (
                    <>
                      <Link href={routes.restaurant(r.restaurant_id, { review: 1 }) + "#reviews"} className={buttonClass("secondary", "md", "h-[46px]")}>
                        Write a Review
                      </Link>
                      <Link href={routes.restaurant(r.restaurant_id) + "#reserve"} className="flex h-[46px] items-center justify-center rounded-xl text-[15px] font-semibold">
                        Book again
                      </Link>
                    </>
                  )}
                  {tab === "cancelled" && (
                    <Link href={routes.restaurant(r.restaurant_id) + "#reserve"} className={buttonClass("secondary", "md", "h-[46px]")}>
                      Book again
                    </Link>
                  )}
                </div>
              </article>
            );
          })
        )}
        {list && list.total_pages > 1 && (
          <nav aria-label="Pagination" className="flex items-center justify-center gap-3">
            {page > 1 && (
              <Link href={`${routes.myReservations(tab)}&page=${page - 1}`} className={buttonClass("secondary", "md")}>
                Previous
              </Link>
            )}
            <span className="text-sm font-semibold text-ink-2">
              Page {page} of {list.total_pages}
            </span>
            {page < list.total_pages && (
              <Link href={`${routes.myReservations(tab)}&page=${page + 1}`} className={buttonClass("secondary", "md")}>
                Next
              </Link>
            )}
          </nav>
        )}
      </div>
    </main>
  );
}
