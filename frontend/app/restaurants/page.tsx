import type { Metadata } from "next";
import Link from "next/link";

import { buttonClass } from "@/components/ui/button";
import { Close, Search } from "@/components/ui/icons";
import { EmptyState, ErrorState } from "@/components/ui/states";
import { RestaurantCard } from "@/features/restaurants/components/restaurant-card";
import { listRestaurants } from "@/features/restaurants/queries";
import { cn } from "@/lib/cn";
import { routes } from "@/lib/routes";
import type { Page, Restaurant, RestaurantSort } from "@/types/api";

export const metadata: Metadata = { title: "Restaurants" };

const sorts: { id: RestaurantSort; label: string }[] = [
  { id: "rating", label: "Highest Rating" },
  { id: "reviews", label: "Most Reviewed" },
];

export default async function RestaurantsPage({ searchParams }: PageProps<"/restaurants">) {
  const sp = await searchParams;
  const q = typeof sp.q === "string" ? sp.q.slice(0, 120) : "";
  const sort: RestaurantSort = sp.sort === "reviews" ? "reviews" : "rating";
  const page = Math.max(1, Number(sp.page) || 1);

  let data: Page<Restaurant> | null = null;
  try {
    data = await listRestaurants({ q, sort, page });
  } catch {
    data = null;
  }

  const trimmed = q.trim();
  const countNote = data
    ? trimmed
      ? `${data.total} ${data.total === 1 ? "result" : "results"} for “${trimmed}”`
      : `${data.total} ${data.total === 1 ? "restaurant" : "restaurants"}`
    : "";
  const sortNote = sort === "rating" ? "sorted by highest rating (restaurants without reviews appear last)" : "sorted by number of reviews";

  return (
    <main className="mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-4 pt-5 pb-10 sm:gap-8 sm:px-8 sm:pt-14 sm:pb-20 lg:px-20">
      <div className="flex max-w-[720px] flex-col gap-3">
        <h1 className="m-0 font-display text-[30px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-5xl">Find a table you&apos;ll love</h1>
        <p className="m-0 hidden text-lg leading-[1.55] text-ink-2 sm:block">
          Browse restaurants, see what other diners say, and reserve your table in a few taps.
        </p>
      </div>

      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between lg:gap-6">
        <form role="search" action={routes.restaurants()} className="flex max-w-[560px] grow">
          <label htmlFor="q" className="sr-only">
            Search restaurants by name
          </label>
          {sort !== "rating" && <input type="hidden" name="sort" value={sort} />}
          <div className="flex h-[52px] w-full items-center gap-2.5 rounded-[14px] border-[1.5px] border-line bg-surface pr-2 pl-4 shadow-[0_1px_2px_rgba(34,28,24,.04)] focus-within:border-ink">
            <Search size={20} className="shrink-0 text-ink-3" />
            <input
              id="q"
              name="q"
              type="search"
              defaultValue={q}
              maxLength={120}
              placeholder="Search restaurants by name"
              autoComplete="off"
              className="h-12 min-w-0 grow border-0 bg-transparent text-base [&::-webkit-search-cancel-button]:hidden"
            />
            {trimmed && (
              <Link
                href={routes.restaurants({ sort: sort === "rating" ? undefined : sort })}
                aria-label="Clear search"
                className="flex size-9 items-center justify-center rounded-[10px] text-ink-2 hover:bg-hover hover:text-ink-2"
              >
                <Close size={16} />
              </Link>
            )}
          </div>
        </form>
        <div className="flex items-center gap-4">
          <span id="sort-label" className="text-sm font-semibold text-ink-2">
            Sort by
          </span>
          <div role="group" aria-labelledby="sort-label" className="flex gap-1 rounded-[14px] bg-sand-2 p-1">
            {sorts.map((s) => {
              const on = s.id === sort;
              return (
                <Link
                  key={s.id}
                  href={routes.restaurants({ q: trimmed || undefined, sort: s.id === "rating" ? undefined : s.id })}
                  aria-current={on ? "true" : undefined}
                  className={cn(
                    "flex h-10 items-center rounded-[10px] px-4 text-[15px] font-semibold no-underline",
                    on ? "bg-surface text-ink shadow-[0_1px_3px_rgba(34,28,24,.12)] hover:text-ink" : "text-ink-2 hover:bg-hover hover:text-ink-2",
                  )}
                >
                  {s.label}
                </Link>
              );
            })}
          </div>
        </div>
      </div>

      {!data ? (
        <ErrorState title="We couldn't load restaurants" retryHref={routes.restaurants({ q: trimmed || undefined, sort: sort === "rating" ? undefined : sort, page })} />
      ) : (
        <>
          <p aria-live="polite" className="m-0 text-sm text-ink-3">
            {countNote} · {sortNote}
          </p>
          {data.items.length > 0 ? (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 sm:gap-8 lg:grid-cols-3">
              {data.items.map((r, i) => (
                <RestaurantCard key={r.id} restaurant={r} priority={i < 3} />
              ))}
            </div>
          ) : trimmed ? (
            <EmptyState
              icon={<Search size={44} className="text-faint" />}
              title={`No restaurants match “${trimmed}”`}
              body="Check the spelling or try a shorter name."
              action={
                <Link href={routes.restaurants({ sort: sort === "rating" ? undefined : sort })} className={buttonClass("secondary", "md")}>
                  Clear search
                </Link>
              }
            />
          ) : (
            <EmptyState
              title="No restaurants yet"
              body="Be the first to list a restaurant on TableTalk."
              action={
                <Link href={routes.newRestaurant()} className={buttonClass("primary", "lg")}>
                  Add a restaurant
                </Link>
              }
            />
          )}
          {data.total_pages > 1 && (
            <nav aria-label="Pagination" className="flex items-center justify-center gap-3">
              {page > 1 ? (
                <Link href={routes.restaurants({ q: trimmed || undefined, sort: sort === "rating" ? undefined : sort, page: page - 1 })} className={buttonClass("secondary", "md")}>
                  Previous
                </Link>
              ) : (
                <span className={buttonClass("secondary", "md", "pointer-events-none opacity-50")} aria-disabled="true">
                  Previous
                </span>
              )}
              <span className="text-sm font-semibold text-ink-2">
                Page {page} of {data.total_pages}
              </span>
              {page < data.total_pages ? (
                <Link href={routes.restaurants({ q: trimmed || undefined, sort: sort === "rating" ? undefined : sort, page: page + 1 })} className={buttonClass("secondary", "md")}>
                  Next
                </Link>
              ) : (
                <span className={buttonClass("secondary", "md", "pointer-events-none opacity-50")} aria-disabled="true">
                  Next
                </span>
              )}
            </nav>
          )}
        </>
      )}
    </main>
  );
}
