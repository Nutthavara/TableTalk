import Link from "next/link";

import { buttonClass } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/states";
import { routes } from "@/lib/routes";

export default function NotFound() {
  return (
    <main className="mx-auto flex w-full max-w-[720px] flex-col px-4 py-16 sm:py-24">
      <EmptyState
        title="We couldn't find that page"
        body="It may have been removed, or it isn't available to your account."
        action={
          <Link href={routes.restaurants()} className={buttonClass("primary", "lg")}>
            Browse restaurants
          </Link>
        }
      />
    </main>
  );
}
