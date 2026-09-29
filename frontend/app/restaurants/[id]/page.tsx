import type { Metadata } from "next";
import Link from "next/link";

import { BackLink } from "@/components/ui/back-link";
import { OwnerBadge } from "@/components/ui/badge";
import { buttonClass } from "@/components/ui/button";
import { Calendar, Clock, Eye, Pencil, Users } from "@/components/ui/icons";
import { RatingPill } from "@/components/ui/stars";
import { draftFromParams } from "@/features/reservations/booking";
import { BookingPanel } from "@/features/reservations/components/booking-panel";
import { MobileBooking } from "@/features/reservations/components/mobile-booking";
import { Gallery } from "@/features/restaurants/components/gallery";
import { getRestaurant } from "@/features/restaurants/queries";
import { ReviewsSection } from "@/features/reviews/components/reviews-section";
import { findMyReview, listReviews } from "@/features/reviews/queries";
import { handlePageError } from "@/lib/auth-guard";
import { formatOpeningHours, formatReviewCount } from "@/lib/format";
import { routes } from "@/lib/routes";
import { getViewer } from "@/lib/viewer";
import type { Restaurant } from "@/types/api";

export async function generateMetadata({ params }: PageProps<"/restaurants/[id]">): Promise<Metadata> {
  const { id } = await params;
  try {
    return { title: (await getRestaurant(id)).name };
  } catch {
    return { title: "Restaurant" };
  }
}

function InfoTile({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl border border-line-soft bg-surface px-5 py-[18px]">
      <span className="flex items-center gap-2 text-[13px] font-bold tracking-[.04em] text-ink-3 uppercase">
        {icon}
        {label}
      </span>
      <span className="text-lg font-bold">{value}</span>
    </div>
  );
}

export default async function RestaurantPage({ params, searchParams }: PageProps<"/restaurants/[id]">) {
  const { id } = await params;
  const sp = await searchParams;
  const here = routes.restaurant(id);

  let restaurant: Restaurant;
  try {
    restaurant = await getRestaurant(id);
  } catch (err) {
    handlePageError(err, here);
  }
  const r = restaurant;
  const viewer = await getViewer();
  const mode = r.is_owner ? "owner" : viewer ? "customer" : "guest";

  // "Show more reviews" grows the first page 10 at a time (up to the API's page-size limit).
  const shown = Math.min(50, 10 * Math.max(1, Number(sp.reviews) || 1));
  const reviewsPage = await listReviews(id, { pageSize: shown });
  const mine = r.my_review_id ? await findMyReview(id, r.my_review_id, reviewsPage.items) : null;
  const others = reviewsPage.items.filter((rv) => rv.id !== r.my_review_id);
  const moreHref = reviewsPage.total > shown && shown < 50 ? routes.restaurant(id, { reviews: shown / 10 + 1 }) + "#reviews" : null;

  const rules = {
    id: r.id,
    name: r.name,
    opening_time: r.opening_time,
    closing_time: r.closing_time,
    capacity: r.capacity,
    timezone: r.timezone,
    cancellation_notice_hours: r.cancellation_notice_hours,
  };
  const initialDraft = draftFromParams(sp, rules);
  const hours = formatOpeningHours(r.opening_time, r.closing_time);
  const policy = `Free up to ${r.cancellation_notice_hours} h before`;

  return (
    <>
      {mode === "owner" && (
        <div role="note" className="flex min-h-12 items-center justify-center gap-2.5 bg-herb px-4 py-2 text-center text-[15px] font-semibold text-white">
          <Eye size={18} className="shrink-0" />
          You&apos;re viewing your own restaurant — this is the page diners see.
        </div>
      )}
      <main className="mx-auto flex w-full max-w-[1440px] flex-col gap-7 px-4 pt-4 pb-40 sm:px-8 sm:pt-6 sm:pb-24 lg:px-20">
        {mode === "owner" ? <BackLink href={routes.myRestaurants()}>My Restaurants</BackLink> : <BackLink href={routes.restaurants()}>All restaurants</BackLink>}

        <Gallery images={r.images} name={r.name} compact={mode === "owner"} />

        <div className="grid grid-cols-1 items-start gap-10 lg:grid-cols-[minmax(0,1fr)_400px] lg:gap-16 xl:grid-cols-[minmax(0,1fr)_420px]">
          <div className="flex min-w-0 flex-col gap-10">
            <div className="flex flex-col gap-4">
              <div className="flex flex-wrap items-center gap-3.5">
                <h1 className="m-0 font-display text-[32px] leading-[1.08] font-bold tracking-[-0.025em] sm:text-[44px]">{r.name}</h1>
                {mode === "owner" && <OwnerBadge large />}
              </div>
              <div className="flex flex-wrap items-center gap-3 text-base">
                {r.average_rating !== null ? (
                  <>
                    <RatingPill average={r.average_rating} large />
                    <a href="#reviews" className="font-semibold text-ink-2">
                      {formatReviewCount(r.review_count)}
                    </a>
                  </>
                ) : (
                  <span className="font-semibold text-ink-3">No reviews yet</span>
                )}
                {mode === "owner" && (
                  <>
                    <span className="text-faint">·</span>
                    <span className="text-body">{hours}</span>
                    <span className="text-faint">·</span>
                    <span className="text-body">{r.capacity} seats</span>
                    <span className="text-faint">·</span>
                    <span className="text-body">Free cancellation up to {r.cancellation_notice_hours} h before</span>
                  </>
                )}
              </div>
              {r.description && <p className="m-0 max-w-[680px] text-[17px] leading-[1.65] text-pretty whitespace-pre-line text-body">{r.description}</p>}
            </div>

            {mode !== "owner" && (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <InfoTile icon={<Clock size={16} />} label="Opening hours" value={hours} />
                <InfoTile icon={<Users size={16} />} label="Capacity" value={`${r.capacity} seats`} />
                <InfoTile icon={<Calendar size={16} />} label="Cancellation" value={policy} />
              </div>
            )}

            <ReviewsSection
              key={`${r.my_review_id ?? "none"}-${reviewsPage.total}`}
              restaurantId={r.id}
              restaurantName={r.name}
              average={r.average_rating}
              count={r.review_count}
              mine={mine}
              others={others}
              mode={mode}
              loginHref={routes.login(`${here}?review=1#reviews`)}
              moreHref={moreHref}
              openInitially={sp.review === "1"}
            />
          </div>

          {mode === "owner" ? (
            <aside className="flex flex-col gap-4 lg:sticky lg:top-24">
              <div className="flex flex-col gap-4 rounded-3xl border-[1.5px] border-herb-line bg-surface p-7">
                <div className="flex flex-col gap-1">
                  <h2 className="m-0 text-xl font-bold">Manage this restaurant</h2>
                  <span className="text-sm leading-normal text-ink-2">Update photos, hours, capacity or your cancellation policy.</span>
                </div>
                <Link href={routes.editRestaurant(r.id)} className={buttonClass("dark", "lg", "font-bold")}>
                  <Pencil size={18} />
                  Edit Restaurant
                </Link>
                <Link href={routes.myRestaurants()} className="flex h-11 items-center justify-center rounded-xl text-[15px] font-semibold">
                  Go to My Restaurants
                </Link>
              </div>
              <div role="note" className="flex flex-col gap-2.5 rounded-3xl border border-dashed border-line bg-sand px-7 py-6">
                <div className="flex items-center gap-2.5">
                  <span aria-hidden className="flex size-9 items-center justify-center rounded-full bg-surface text-ink-2">
                    <Calendar size={18} />
                  </span>
                  <h2 className="m-0 text-lg font-bold">Booking is off for owners</h2>
                </div>
                <span className="text-sm leading-normal text-body">You can&apos;t reserve a table at your own restaurant. Diners see the “Reserve a Table” panel here.</span>
              </div>
            </aside>
          ) : (
            <aside id="reserve" aria-label="Reserve a table" className="hidden scroll-mt-24 flex-col gap-5 rounded-3xl border border-line-soft bg-surface p-7 shadow-panel sm:flex lg:sticky lg:top-24">
              <BookingPanel restaurant={rules} signedIn={viewer !== null} initialDraft={initialDraft} />
            </aside>
          )}
        </div>
      </main>
      {mode !== "owner" && <MobileBooking restaurant={rules} signedIn={viewer !== null} initialDraft={initialDraft} />}
    </>
  );
}
