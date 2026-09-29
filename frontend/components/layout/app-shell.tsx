import Link from "next/link";
import { Suspense } from "react";

import { DesktopNavLinks, MobileTabLinks } from "@/components/layout/nav-links";
import { MobileProfileTab, ProfileMenu } from "@/components/layout/profile-menu";
import { buttonClass } from "@/components/ui/button";
import { Logo } from "@/components/ui/icons";
import { routes } from "@/lib/routes";
import { getViewer } from "@/lib/viewer";

function Brand({ small = false }: { small?: boolean }) {
  return (
    <Link href={routes.restaurants()} aria-label="TableTalk — go to restaurants" className="flex items-center gap-2.5 text-ink no-underline hover:text-ink">
      <Logo size={small ? 30 : 36} />
      <span className={small ? "font-display text-xl font-bold" : "font-display text-[22px] font-bold tracking-[-0.02em]"}>TableTalk</span>
    </Link>
  );
}

/** Top navigation (desktop), compact header + bottom tab bar (mobile). */
export async function AppShell({ children }: { children: React.ReactNode }) {
  const viewer = await getViewer();
  const signedIn = viewer !== null;

  return (
    <>
      <header className="sticky top-0 z-40 hidden h-[72px] items-center justify-between border-b border-line-soft bg-surface px-6 sm:flex lg:px-20">
        <div className="flex h-[72px] items-center gap-6 lg:gap-12">
          <Brand />
          <Suspense>
            <DesktopNavLinks signedIn={signedIn} />
          </Suspense>
        </div>
        {viewer ? (
          <ProfileMenu viewer={viewer} />
        ) : (
          <div className="flex items-center gap-2">
            <a href={routes.login()} className={buttonClass("ghost", "sm", "h-11 px-4 text-[15px] text-ink")}>
              Log in
            </a>
            <a href={routes.register()} className={buttonClass("primary", "sm", "h-11 px-[18px] text-[15px] font-semibold")}>
              Sign up
            </a>
          </div>
        )}
      </header>

      <header className="sticky top-0 z-40 flex h-[60px] items-center justify-between border-b border-line-soft bg-surface px-4 sm:hidden">
        <Brand small />
        {!viewer && (
          <a href={routes.login()} className={buttonClass("secondary", "sm", "h-11 px-4 text-[15px]")}>
            Log in
          </a>
        )}
      </header>

      <div className="flex flex-1 flex-col pb-24 sm:pb-0">{children}</div>

      <Suspense>
        <MobileTabLinks signedIn={signedIn} last={<MobileProfileTab viewer={viewer} loginHref={routes.login()} />} />
      </Suspense>
    </>
  );
}
