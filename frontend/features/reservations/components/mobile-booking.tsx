"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Close } from "@/components/ui/icons";
import type { BookingDraft, BookingRules } from "@/features/reservations/booking";
import { BookingPanel } from "@/features/reservations/components/booking-panel";

/** Mobile: sticky "Reserve a Table" bar that opens the booking panel as a bottom sheet. */
export function MobileBooking({ restaurant, signedIn, initialDraft }: { restaurant: BookingRules & { id: string; name: string }; signedIn: boolean; initialDraft: BookingDraft }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <div className="fixed inset-x-0 bottom-20 z-30 flex items-center justify-between gap-3 border-t border-line-soft bg-surface px-4 py-3 shadow-[0_-8px_24px_rgba(34,28,24,.08)] sm:hidden">
        <div className="flex min-w-0 flex-col">
          <strong className="text-[15px]">
            Open {restaurant.opening_time} – {restaurant.closing_time}
          </strong>
          <span className="truncate text-[13px] text-ink-3">Free to cancel up to {restaurant.cancellation_notice_hours} h before</span>
        </div>
        <Button size="lg" className="shrink-0 font-bold" onClick={() => setOpen(true)}>
          Reserve a Table
        </Button>
      </div>
      <Dialog open={open} onClose={() => setOpen(false)} labelledBy="mb-h">
        <div className="flex items-center justify-between">
          <span id="mb-h" className="text-sm font-semibold text-ink-2">
            {restaurant.name}
          </span>
          <button type="button" aria-label="Close booking" onClick={() => setOpen(false)} className="flex size-11 cursor-pointer items-center justify-center rounded-xl hover:bg-hover">
            <Close size={18} />
          </button>
        </div>
        <BookingPanel restaurant={restaurant} signedIn={signedIn} initialDraft={initialDraft} idPrefix="mbk" />
      </Dialog>
    </>
  );
}
