import Link from "next/link";

import { Clock, Users } from "@/components/ui/icons";
import { RatingPill } from "@/components/ui/stars";
import { RestaurantPhoto } from "@/features/restaurants/components/restaurant-image";
import { formatOpeningHours, formatReviewCount } from "@/lib/format";
import { routes } from "@/lib/routes";
import type { Restaurant } from "@/types/api";

export function RestaurantCard({ restaurant: r, priority = false }: { restaurant: Restaurant; priority?: boolean }) {
  const hours = formatOpeningHours(r.opening_time, r.closing_time);
  const rated = r.average_rating !== null;
  const aria = `${r.name}, ${rated ? `rated ${r.average_rating!.toFixed(1)} out of 5 from ${formatReviewCount(r.review_count)}` : "no reviews yet"}, open ${hours}, ${r.capacity} seats`;

  return (
    <Link
      href={routes.restaurant(r.id)}
      aria-label={aria}
      className="group flex flex-col overflow-hidden rounded-[20px] border border-line-soft bg-surface text-ink no-underline shadow-card transition-[box-shadow,transform] duration-200 hover:-translate-y-[3px] hover:text-ink hover:shadow-card-hover"
    >
      <div className="relative h-[200px] sm:h-[240px]">
        <RestaurantPhoto image={r.images[0]} alt="" sizes="(min-width: 1024px) 400px, (min-width: 640px) 50vw, 100vw" priority={priority} />
        {!rated && (
          <span className="absolute top-4 left-4 flex h-7 items-center rounded-full bg-surface px-3 text-[13px] font-bold text-saffron-ink">New on TableTalk</span>
        )}
      </div>
      <div className="flex flex-col gap-2.5 px-[22px] pt-5 pb-[22px]">
        <div className="flex items-start justify-between gap-3">
          <h2 className="m-0 font-display text-[22px] leading-tight font-semibold tracking-[-0.01em]">{r.name}</h2>
          {rated && <RatingPill average={r.average_rating!} />}
        </div>
        <span className={rated ? "text-sm font-semibold text-ink-2" : "text-sm font-semibold text-ink-3"}>{formatReviewCount(r.review_count)}</span>
        <p className="m-0 line-clamp-2 min-h-[45px] text-[15px] leading-normal text-ink-2">{r.description}</p>
        <div className="flex items-center gap-4 border-t border-sand-2 pt-3 text-sm text-ink-2">
          <span className="flex items-center gap-1.5">
            <Clock size={16} />
            {hours}
          </span>
          <span className="flex items-center gap-1.5">
            <Users size={16} />
            {r.capacity} seats
          </span>
        </div>
      </div>
    </Link>
  );
}

export function RestaurantCardSkeleton() {
  return (
    <div aria-hidden className="flex flex-col overflow-hidden rounded-[20px] border border-line-soft bg-surface">
      <div className="h-[200px] animate-pulse bg-line-soft sm:h-[240px]" />
      <div className="flex flex-col gap-3 p-[22px]">
        <div className="h-6 w-2/3 animate-pulse rounded-md bg-line-soft" />
        <div className="h-4 w-1/3 animate-pulse rounded-md bg-line-soft" />
        <div className="h-10 animate-pulse rounded-md bg-line-soft" />
      </div>
    </div>
  );
}
