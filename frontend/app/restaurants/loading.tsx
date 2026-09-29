import { RestaurantCardSkeleton } from "@/features/restaurants/components/restaurant-card";

export default function Loading() {
  return (
    <main aria-busy="true" aria-label="Loading restaurants" className="mx-auto flex w-full max-w-[1440px] flex-col gap-8 px-4 pt-5 pb-10 sm:px-8 sm:pt-14 lg:px-20">
      <div className="h-12 w-2/3 max-w-[560px] animate-pulse rounded-xl bg-line-soft" />
      <div className="h-[52px] max-w-[560px] animate-pulse rounded-[14px] bg-line-soft" />
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 sm:gap-8 lg:grid-cols-3">
        {Array.from({ length: 6 }, (_, i) => (
          <RestaurantCardSkeleton key={i} />
        ))}
      </div>
    </main>
  );
}
