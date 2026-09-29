"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { Calendar, Store, Utensils } from "@/components/ui/icons";
import { cn } from "@/lib/cn";
import { routes } from "@/lib/routes";

export const mainLinks = [
  { href: routes.restaurants(), label: "Restaurants", short: "Restaurants", prefix: "/restaurants", Icon: Utensils, protected: false },
  { href: routes.myReservations(), label: "My Reservations", short: "Reservations", prefix: "/my-reservations", Icon: Calendar, protected: true },
  { href: routes.myRestaurants(), label: "My Restaurants", short: "My Restaurants", prefix: "/my-restaurants", Icon: Store, protected: true },
];

/**
 * Signed-out users go to login first for protected sections, then land on them. Login is a route
 * handler that redirects to Keycloak, so it needs a plain <a> rather than client-side navigation.
 */
function NavAnchor({ href, loginFirst, ...rest }: React.ComponentProps<"a"> & { href: string; loginFirst: boolean }) {
  return loginFirst ? <a href={routes.login(href)} {...rest} /> : <Link href={href} {...rest} />;
}

export function DesktopNavLinks({ signedIn }: { signedIn: boolean }) {
  const pathname = usePathname();
  return (
    <nav aria-label="Main" className="flex h-[72px] items-center gap-1">
      {mainLinks.map((link) => {
        const active = pathname.startsWith(link.prefix);
        return (
          <NavAnchor
            key={link.href}
            href={link.href}
            loginFirst={link.protected && !signedIn}
            aria-current={active ? "page" : undefined}
            className={cn(
              "flex h-[72px] items-center px-3.5 text-[15px] font-semibold no-underline",
              active ? "text-ink shadow-[inset_0_-3px_0_var(--color-paprika)] hover:text-ink" : "text-ink-2 hover:text-ink",
            )}
          >
            {link.label}
          </NavAnchor>
        );
      })}
    </nav>
  );
}

export function MobileTabLinks({ signedIn, last }: { signedIn: boolean; last: React.ReactNode }) {
  const pathname = usePathname();
  return (
    <nav
      aria-label="Main"
      className="fixed inset-x-0 bottom-0 z-40 grid h-20 grid-cols-4 border-t border-line-soft bg-surface pb-4 sm:hidden"
    >
      {mainLinks.map(({ Icon, ...link }) => {
        const active = pathname.startsWith(link.prefix);
        return (
          <NavAnchor
            key={link.href}
            href={link.href}
            loginFirst={link.protected && !signedIn}
            aria-current={active ? "page" : undefined}
            className={cn(
              "flex flex-col items-center justify-center gap-1 text-xs no-underline",
              active ? "font-bold text-paprika hover:text-paprika" : "font-semibold text-ink-2 hover:text-ink",
            )}
          >
            <Icon size={24} />
            {link.short}
          </NavAnchor>
        );
      })}
      {last}
    </nav>
  );
}
