"use client";

import { useState, useTransition } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { CalendarX, Info } from "@/components/ui/icons";
import { toast } from "@/components/ui/toaster";
import { cancelReservation } from "@/features/reservations/actions";

type Summary = { id: string; name: string; date: string; time: string; guests: string; policy: string };

/** "Cancel Booking" with the confirmation dialog from the design. */
export function CancelBookingButton({ reservation }: { reservation: Summary }) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();
  const titleId = `cx-h-${reservation.id}`;
  const descId = `cx-d-${reservation.id}`;

  const confirm = () =>
    startTransition(async () => {
      const result = await cancelReservation(reservation.id);
      if (result.status === "success") {
        setOpen(false);
        toast({ message: `Reservation cancelled — ${reservation.name}, ${reservation.date}.` });
      } else if (result.status === "error") {
        setError(result.message);
      }
    });

  return (
    <>
      <Button
        variant="dangerOutline"
        className="h-[46px]"
        onClick={() => {
          setError(null);
          setOpen(true);
        }}
      >
        Cancel Booking
      </Button>
      <Dialog open={open} onClose={() => setOpen(false)} role="alertdialog" labelledBy={titleId} describedBy={descId}>
        <span aria-hidden className="hidden size-[52px] items-center justify-center rounded-full bg-chili-soft text-chili sm:flex">
          <CalendarX size={26} />
        </span>
        <div className="flex flex-col gap-2">
          <h2 id={titleId} className="m-0 font-display text-2xl font-bold tracking-[-0.015em] sm:text-[26px]">
            Cancel reservation?
          </h2>
          <p id={descId} className="m-0 text-base leading-[1.55] text-body">
            Are you sure you want to cancel this reservation?
          </p>
        </div>
        <dl className="m-0 grid grid-cols-[80px_1fr] gap-y-2 rounded-[14px] bg-sand px-[18px] py-4 text-[15px]">
          <dt className="text-ink-2">Restaurant</dt>
          <dd className="m-0 font-bold">{reservation.name}</dd>
          <dt className="text-ink-2">Date</dt>
          <dd className="m-0 font-bold">{reservation.date}</dd>
          <dt className="text-ink-2">Time</dt>
          <dd className="m-0 font-bold">{reservation.time}</dd>
          <dt className="text-ink-2">Guests</dt>
          <dd className="m-0 font-bold">{reservation.guests}</dd>
        </dl>
        <div className="flex gap-2.5 rounded-xl bg-herb-soft px-3.5 py-3 text-sm leading-[1.45] text-herb-ink">
          <Info size={18} className="mt-px shrink-0" />
          <span>{reservation.policy}. Cancelling now is free.</span>
        </div>
        {error && <Alert tone="error">{error}</Alert>}
        <div className="flex flex-col-reverse justify-end gap-3 sm:flex-row">
          <Button variant="secondary" onClick={() => setOpen(false)} autoFocus>
            Keep reservation
          </Button>
          <Button variant="danger" onClick={confirm} disabled={pending}>
            {pending ? "Cancelling…" : "Yes, cancel booking"}
          </Button>
        </div>
      </Dialog>
    </>
  );
}
