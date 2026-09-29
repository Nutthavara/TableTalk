"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { ChevronDown, LogOut, User } from "@/components/ui/icons";
import { logout } from "@/features/auth/actions";
import { cn } from "@/lib/cn";
import { routes } from "@/lib/routes";
import type { Viewer } from "@/lib/viewer";

function Avatar({ initial, size = 36 }: { initial: string; size?: number }) {
  return (
    <span aria-hidden className="flex shrink-0 items-center justify-center rounded-full bg-herb font-bold text-white" style={{ width: size, height: size, fontSize: size * 0.42 }}>
      {initial}
    </span>
  );
}

const itemClass = "flex h-11 w-full cursor-pointer items-center gap-3 rounded-[10px] px-3 text-left text-[15px] font-semibold text-ink no-underline hover:bg-hover hover:text-ink";

function MenuItems({ viewer, onNavigate }: { viewer: NonNullable<Viewer>; onNavigate: () => void }) {
  return (
    <>
      <div className="flex items-center gap-3 border-b border-line-soft px-3 pt-1 pb-3">
        <Avatar initial={viewer.initial} size={40} />
        <div className="flex min-w-0 flex-col">
          <strong className="truncate text-base">{viewer.name}</strong>
          {viewer.email && <span className="truncate text-sm text-ink-3">{viewer.email}</span>}
        </div>
      </div>
      <a role="menuitem" href={viewer.accountUrl} className={itemClass}>
        Profile
      </a>
      <Link role="menuitem" href={routes.myReservations()} className={itemClass} onClick={onNavigate}>
        My Reservations
      </Link>
      <Link role="menuitem" href={routes.myRestaurants()} className={itemClass} onClick={onNavigate}>
        My Restaurants
      </Link>
      <form action={logout} className="border-t border-line-soft pt-1">
        <button role="menuitem" type="submit" className={cn(itemClass, "text-chili hover:text-chili")}>
          <LogOut size={18} />
          Log out
        </button>
      </form>
    </>
  );
}

function useDismiss(open: boolean, close: () => void, ref: React.RefObject<HTMLElement | null>) {
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) close();
    };
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && close();
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open, close, ref]);
}

/** Account button + dropdown in the desktop top nav. */
export function ProfileMenu({ viewer }: { viewer: NonNullable<Viewer> }) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useDismiss(open, () => setOpen(false), ref);

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        aria-label={`Account menu for ${viewer.name}`}
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={() => setOpen((o) => !o)}
        className="flex h-11 cursor-pointer items-center gap-2.5 rounded-full border border-line-mid bg-surface pr-3 pl-1 text-[15px] font-semibold"
      >
        <Avatar initial={viewer.initial} />
        <span className="max-w-40 truncate">{viewer.name}</span>
        <ChevronDown size={16} />
      </button>
      {open && (
        <div role="menu" className="absolute top-[52px] right-0 z-50 flex w-72 flex-col gap-1 rounded-2xl border border-line-soft bg-surface p-2 shadow-dialog">
          <MenuItems viewer={viewer} onNavigate={() => setOpen(false)} />
        </div>
      )}
    </div>
  );
}

/** Fourth tab of the mobile tab bar: Profile (menu sheet) or Log in. */
export function MobileProfileTab({ viewer, loginHref }: { viewer: Viewer; loginHref: string }) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useDismiss(open, () => setOpen(false), ref);

  if (!viewer) {
    return (
      <a href={loginHref} className="flex flex-col items-center justify-center gap-1 text-xs font-semibold text-ink-2 no-underline">
        <User size={24} />
        Log in
      </a>
    );
  }
  return (
    <div ref={ref} className="relative flex">
      <button
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={() => setOpen((o) => !o)}
        className="flex w-full cursor-pointer flex-col items-center justify-center gap-1 text-xs font-semibold text-ink-2"
      >
        <User size={24} />
        Profile
      </button>
      {open && (
        <div role="menu" className="fixed inset-x-3 bottom-24 z-50 flex flex-col gap-1 rounded-2xl border border-line-soft bg-surface p-2 shadow-dialog">
          <MenuItems viewer={viewer} onNavigate={() => setOpen(false)} />
        </div>
      )}
    </div>
  );
}
