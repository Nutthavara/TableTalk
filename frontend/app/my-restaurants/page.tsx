import type { Metadata } from "next";
import Link from "next/link";

import { OwnerBadge } from "@/components/ui/badge";
import { buttonClass } from "@/components/ui/button";
import { Eye, Pencil, Plus, Star } from "@/components/ui/icons";
import { EmptyState, ErrorState } from "@/components/ui/states";
import { DeleteRestaurantButton } from "@/features/restaurants/components/delete-restaurant-button";
import { RestaurantPhoto } from "@/features/restaurants/components/restaurant-image";
import { listMyRestaurants } from "@/features/restaurants/queries";
import { requireSignedIn } from "@/lib/auth-guard";
import { formatOpeningHours } from "@/lib/format";
import { routes } from "@/lib/routes";
import type { Restaurant } from "@/types/api";

export const metadata: Metadata = { title: "My Restaurants" };

function Stat({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-[13px] font-semibold text-ink-3">{label}</span>
      <strong className="text-base">{children}</strong>
    </div>
  );
}

export default async function MyRestaurantsPage() {
  await requireSignedIn(routes.myRestaurants());
  let items: Restaurant[] | null = null;
  try {
    items = (await listMyRestaurants()).items ?? [];
  } catch {
    items = null;
  }

  return (
    <main className="mx-auto flex w-full max-w-[1120px] flex-col gap-8 px-4 pt-5 pb-12 sm:px-8 sm:pt-14 sm:pb-20 xl:px-0">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <div className="flex flex-col gap-2.5">
          <h1 className="m-0 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-[44px]">My Restaurants</h1>
          <p className="m-0 text-[17px] text-ink-2">Restaurants you own. Only you can edit or delete them.</p>
        </div>
        <Link href={routes.newRestaurant()} className={buttonClass("primary", "lg", "rounded-[14px]")}>
          <Plus size={18} />
          Add Restaurant
        </Link>
      </div>

      {items === null ? (
        <ErrorState title="We couldn't load your restaurants" retryHref={routes.myRestaurants()} />
      ) : items.length === 0 ? (
        <EmptyState
          title="You haven't created a restaurant yet."
          body="List your restaurant so diners can book tables and leave reviews."
          action={
            <Link href={routes.newRestaurant()} className={buttonClass("primary", "lg", "rounded-[14px]")}>
              Create your first restaurant
            </Link>
          }
        />
      ) : (
        <div className="flex flex-col gap-5">
          {items.map((r) => (
            <article key={r.id} aria-label={r.name} className="grid grid-cols-1 items-center gap-5 rounded-[22px] border border-line-soft bg-surface p-4 md:grid-cols-[260px_minmax(0,1fr)_auto] md:gap-7">
              <div className="relative h-44 overflow-hidden rounded-[14px] md:h-[180px]">
                <RestaurantPhoto image={r.images[0]} alt="" sizes="(min-width: 768px) 260px, 100vw" />
              </div>
              <div className="flex flex-col gap-3.5">
                <div className="flex flex-wrap items-center gap-3">
                  <h2 className="m-0 font-display text-[26px] font-bold tracking-[-0.015em]">{r.name}</h2>
                  <OwnerBadge />
                </div>
                <div className="grid grid-cols-2 gap-x-8 gap-y-4 lg:grid-cols-[repeat(4,auto)] lg:justify-start">
                  <Stat label="Opening hours">{formatOpeningHours(r.opening_time, r.closing_time)}</Stat>
                  <Stat label="Capacity">{r.capacity} seats</Stat>
                  <Stat label="Rating">
                    {r.average_rating !== null ? (
                      <span className="flex items-center gap-1">
                        <Star size={16} fill="#D9941A" />
                        {r.average_rating.toFixed(1)} <span className="font-medium text-ink-2">· {r.review_count} {r.review_count === 1 ? "review" : "reviews"}</span>
                      </span>
                    ) : (
                      <span className="font-semibold text-ink-2">No reviews yet</span>
                    )}
                  </Stat>
                  <Stat label="Cancellation notice">
                    {r.cancellation_notice_hours} {r.cancellation_notice_hours === 1 ? "hour" : "hours"}
                  </Stat>
                </div>
              </div>
              <div className="grid grid-cols-3 gap-2.5 md:flex md:min-w-[170px] md:flex-col md:pr-2">
                <Link href={routes.restaurant(r.id)} aria-label={`View ${r.name}`} className={buttonClass("soft", "md", "h-11")}>
                  <Eye size={16} />
                  View
                </Link>
                <Link href={routes.editRestaurant(r.id)} aria-label={`Edit ${r.name}`} className={buttonClass("secondary", "md", "h-11")}>
                  <Pencil size={16} />
                  Edit
                </Link>
                <DeleteRestaurantButton id={r.id} name={r.name} />
              </div>
            </article>
          ))}
        </div>
      )}
    </main>
  );
}
