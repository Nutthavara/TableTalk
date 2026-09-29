import type { Metadata } from "next";

import { BackLink } from "@/components/ui/back-link";
import { RestaurantForm } from "@/features/restaurants/components/restaurant-form";
import { requireSignedIn } from "@/lib/auth-guard";
import { routes } from "@/lib/routes";

export const metadata: Metadata = { title: "Add a restaurant" };

export default async function NewRestaurantPage() {
  await requireSignedIn(routes.newRestaurant());
  return (
    <main className="mx-auto flex w-full max-w-[1120px] flex-col gap-7 px-4 pt-4 pb-12 sm:px-8 sm:pt-6 sm:pb-16 xl:px-0">
      <BackLink href={routes.myRestaurants()}>My Restaurants</BackLink>
      <div className="flex flex-col gap-2">
        <h1 className="m-0 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-[44px]">Add a restaurant</h1>
        <p className="m-0 text-[17px] text-ink-2">Tell diners what makes your place special. You can change these details any time.</p>
      </div>
      <RestaurantForm />
    </main>
  );
}
