import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { BackLink } from "@/components/ui/back-link";
import { RestaurantForm } from "@/features/restaurants/components/restaurant-form";
import { getRestaurant } from "@/features/restaurants/queries";
import { handlePageError, requireSignedIn } from "@/lib/auth-guard";
import { routes } from "@/lib/routes";
import type { Restaurant } from "@/types/api";

export const metadata: Metadata = { title: "Edit restaurant" };

export default async function EditRestaurantPage({ params }: PageProps<"/my-restaurants/[id]/edit">) {
  const { id } = await params;
  const here = routes.editRestaurant(id);
  await requireSignedIn(here);

  let restaurant: Restaurant;
  try {
    restaurant = await getRestaurant(id);
  } catch (err) {
    handlePageError(err, here);
  }
  // Not the owner: behave as if it doesn't exist here (the API would reject any change anyway).
  if (!restaurant.is_owner) notFound();

  return (
    <main className="mx-auto flex w-full max-w-[1120px] flex-col gap-7 px-4 pt-4 pb-12 sm:px-8 sm:pt-6 sm:pb-16 xl:px-0">
      <BackLink href={routes.myRestaurants()}>My Restaurants</BackLink>
      <div className="flex flex-col gap-2">
        <h1 className="m-0 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.025em] sm:text-[44px]">Edit restaurant</h1>
        <p className="m-0 text-[17px] text-ink-2">Changes appear on your restaurant page as soon as you save.</p>
      </div>
      <RestaurantForm restaurant={restaurant} />
    </main>
  );
}
