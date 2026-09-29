import { Star } from "@/components/ui/icons";

const EMPTY = "#E6DCCD";
const FULL = "#E3A327";

/** Whole-star display for an individual review (1–5). */
export function StarRow({ rating, size = 18 }: { rating: number; size?: number }) {
  return (
    <div role="img" aria-label={`Rated ${rating} out of 5`} className="flex gap-0.5">
      {[1, 2, 3, 4, 5].map((i) => (
        <Star key={i} size={size} fill={i <= rating ? FULL : EMPTY} />
      ))}
    </div>
  );
}

/** Fractional star bar for an average rating (e.g. 4.7 fills 94%). */
export function AverageStars({ average, size = 26 }: { average: number; size?: number }) {
  const gap = 4;
  const width = size * 5 + gap * 4;
  const filled = Math.round((average / 5) * width);
  return (
    <div role="img" aria-label={`Average rating ${average.toFixed(1)} out of 5`} className="relative" style={{ width, height: size }}>
      <div className="absolute top-0 left-0 flex" style={{ gap }}>
        {[0, 1, 2, 3, 4].map((i) => (
          <Star key={i} size={size} fill={EMPTY} />
        ))}
      </div>
      <div className="absolute top-0 left-0 flex overflow-hidden" style={{ gap, width: filled }}>
        {[0, 1, 2, 3, 4].map((i) => (
          <Star key={i} size={size} fill={FULL} className="shrink-0" />
        ))}
      </div>
    </div>
  );
}

/** Pill with a star and the 1-decimal average, e.g. on cards. */
export function RatingPill({ average, large = false }: { average: number; large?: boolean }) {
  return (
    <span
      className={
        large
          ? "flex h-[34px] shrink-0 items-center gap-1.5 rounded-full bg-saffron-soft px-3 text-base font-bold text-ink"
          : "flex h-[30px] shrink-0 items-center gap-[5px] rounded-full bg-saffron-soft px-2.5 text-[15px] font-bold text-ink"
      }
    >
      <Star size={large ? 18 : 16} fill="#D9941A" />
      {average.toFixed(1)}
    </span>
  );
}
