"use client";

import { usePathname, useSearchParams } from "next/navigation";
import { useEffect, useSyncExternalStore } from "react";

import { Check, Close, Warning } from "@/components/ui/icons";
import { FLASH_COOKIE, type Flash } from "@/lib/flash-types";

// A tiny global store so any client component can call toast(); server actions that redirect leave
// a one-shot cookie (see lib/flash.ts) that is turned into a toast after navigation.
let current: (Flash & { id: number }) | null = null;
const listeners = new Set<() => void>();
let nextId = 1;

export function toast(flash: Flash) {
  current = { ...flash, id: nextId++ };
  listeners.forEach((l) => l());
}

function dismiss() {
  current = null;
  listeners.forEach((l) => l());
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

function readFlashCookie(): Flash | null {
  const match = document.cookie.split("; ").find((c) => c.startsWith(FLASH_COOKIE + "="));
  if (!match) return null;
  document.cookie = `${FLASH_COOKIE}=; Max-Age=0; path=/`;
  try {
    const b64 = match.slice(FLASH_COOKIE.length + 1).replace(/-/g, "+").replace(/_/g, "/");
    const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
    return JSON.parse(new TextDecoder().decode(bytes)) as Flash;
  } catch {
    return null;
  }
}

export function Toaster() {
  const item = useSyncExternalStore(subscribe, () => current, () => null);
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // Pick up a flash left by a server action after every navigation.
  useEffect(() => {
    const flash = readFlashCookie();
    if (flash) toast(flash);
  }, [pathname, searchParams]);

  useEffect(() => {
    if (!item) return;
    const timer = setTimeout(dismiss, item.kind === "error" ? 8000 : 5000);
    return () => clearTimeout(timer);
  }, [item]);

  return (
    <div aria-live="polite" className="pointer-events-none fixed inset-x-0 top-4 z-50 flex justify-center px-4 max-sm:top-auto max-sm:bottom-24 sm:top-24">
      {item && (
        <div
          key={item.id}
          role={item.kind === "error" ? "alert" : "status"}
          className={
            item.kind === "error"
              ? "pointer-events-auto flex w-full max-w-md items-center gap-3 rounded-[14px] border-[1.5px] border-chili-line bg-surface py-3.5 pr-4 pl-[18px] text-[15px] font-semibold text-chili shadow-[0_16px_40px_rgba(34,28,24,.18)]"
              : "pointer-events-auto flex w-full max-w-md items-center gap-3 rounded-[14px] bg-ink py-3.5 pr-4 pl-[18px] text-[15px] font-semibold text-white shadow-[0_16px_40px_rgba(34,28,24,.3)]"
          }
        >
          {item.kind === "error" ? (
            <Warning size={20} className="shrink-0" />
          ) : (
            <span aria-hidden className="flex size-[26px] shrink-0 items-center justify-center rounded-full bg-[#3C8A5F]">
              <Check size={16} />
            </span>
          )}
          <span className="grow">{item.message}</span>
          {item.action && (
            <a href={item.action.href} className="font-bold text-[#FFC9AE] hover:text-white">
              {item.action.label}
            </a>
          )}
          <button
            type="button"
            aria-label="Dismiss notification"
            onClick={dismiss}
            className="flex size-8 shrink-0 cursor-pointer items-center justify-center rounded-lg"
          >
            <Close size={16} />
          </button>
        </div>
      )}
    </div>
  );
}
