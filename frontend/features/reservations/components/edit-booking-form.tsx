"use client";

import Link from "next/link";
import { useMemo, useState, useTransition } from "react";

import { Alert } from "@/components/ui/alert";
import { buttonClass } from "@/components/ui/button";
import { Info } from "@/components/ui/icons";
import { updateReservation } from "@/features/reservations/actions";
import {
  type BookingDraft,
  type BookingRules,
  checkBooking,
  dateLabel,
  deadlineLabel,
  guestsLabel,
  nowInZone,
  timeLabel,
} from "@/features/reservations/booking";
import { BookingErrors, BookingFields } from "@/features/reservations/components/booking-fields";
import { RestaurantPhoto } from "@/features/restaurants/components/restaurant-image";
import { routes } from "@/lib/routes";
import type { RestaurantImage } from "@/types/api";

type Props = { reservationId: string; current: BookingDraft; rules: BookingRules; cover?: RestaurantImage; restaurantName: string };

const overline = "text-[13px] font-bold tracking-[.04em] uppercase";

export function EditBookingForm({ reservationId, current, rules, cover, restaurantName }: Props) {
  const [draft, setDraft] = useState(current);
  const [apiError, setApiError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();
  const now = useMemo(() => nowInZone(rules.timezone), [rules.timezone]);
  const check = checkBooking(draft, rules, now);

  const changed = {
    date: draft.date !== current.date,
    start: draft.start !== current.start,
    end: draft.end !== current.end,
    guests: draft.guests !== current.guests,
  };
  const anyChange = Object.values(changed).some(Boolean);
  const timeChanged = changed.start || changed.end;

  const save = (e: React.FormEvent) => {
    e.preventDefault();
    if (check.errors.length || !anyChange) return;
    startTransition(async () => {
      const result = await updateReservation(reservationId, draft);
      if (result?.status === "error") {
        setApiError(result.code === "INSUFFICIENT_CAPACITY" ? `Not enough seats available for ${guestsLabel(draft.guests)} at that time. Try another time or fewer guests.` : result.message);
      }
    });
  };

  return (
    <div className="grid grid-cols-1 items-start gap-8 lg:grid-cols-[minmax(0,1fr)_380px] lg:gap-10">
      <form onSubmit={save} aria-label="Edit booking details" className="flex flex-col gap-[22px] rounded-3xl border border-line-soft bg-surface p-5 sm:p-8">
        <h2 className="m-0 text-xl font-bold">New details</h2>
        <BookingFields
          idPrefix="ed"
          draft={draft}
          onChange={(d) => {
            setDraft(d);
            setApiError(null);
          }}
          check={check}
          capacity={rules.capacity}
          openingTime={rules.opening_time}
          closingTime={rules.closing_time}
          minDate={now.date}
          changed={changed}
          disabled={pending}
          compactStepper
        />
        <BookingErrors id="ed-errors" messages={apiError ? [...check.errors, apiError] : check.errors} />
        <div className="flex gap-2.5 rounded-xl bg-sand px-4 py-3.5 text-sm leading-normal text-body">
          <Info size={18} className="mt-px shrink-0" />
          <span>
            Changes are checked like a new booking: the time must be in the future, within opening hours, and there must be enough free seats. Your current booking
            stays as it is until the change is confirmed.
          </span>
        </div>
        <div className="flex flex-col-reverse justify-end gap-3 pt-1 sm:flex-row">
          <Link href={routes.myReservations()} className={buttonClass("ghost", "lg")}>
            Discard changes
          </Link>
          <button type="submit" disabled={pending || !anyChange || check.errors.length > 0} className={buttonClass("primary", "lg", "px-7")}>
            {pending ? "Saving…" : "Save Changes"}
          </button>
        </div>
      </form>

      <aside aria-label="Booking comparison" className="flex flex-col gap-4">
        <div className="flex flex-col gap-3.5 rounded-3xl border border-line-soft bg-surface p-4">
          <div className="relative h-[150px] overflow-hidden rounded-[14px]">
            <RestaurantPhoto image={cover} alt={restaurantName} sizes="380px" />
          </div>
          <div className="flex flex-col gap-3 px-2 pt-1 pb-2">
            <span className={`${overline} text-ink-3`}>Current booking</span>
            <dl className="m-0 grid grid-cols-[64px_1fr] gap-y-2 text-[15px]">
              <dt className="text-ink-2">Date</dt>
              <dd className="m-0 font-bold">{dateLabel(current.date)}</dd>
              <dt className="text-ink-2">Time</dt>
              <dd className="m-0 font-bold">{timeLabel(current.start, current.end, false)}</dd>
              <dt className="text-ink-2">Guests</dt>
              <dd className="m-0 font-bold">{guestsLabel(current.guests)}</dd>
            </dl>
          </div>
        </div>
        {anyChange ? (
          <div className="flex flex-col gap-3 rounded-[20px] border-[1.5px] border-peach-line bg-[#FFF6F1] px-6 py-5">
            <span className={`${overline} text-peach-ink`}>After saving</span>
            <dl className="m-0 grid grid-cols-[64px_1fr] gap-y-2 text-[15px]">
              <dt className="text-ink-2">Date</dt>
              <dd className="m-0">
                {changed.date ? (
                  <>
                    <s className="text-ink-3">{dateLabel(current.date)}</s> <strong>{dateLabel(draft.date)}</strong>
                  </>
                ) : (
                  <>
                    <strong>{dateLabel(draft.date)}</strong> <span className="font-medium text-ink-3">(no change)</span>
                  </>
                )}
              </dd>
              <dt className="text-ink-2">Time</dt>
              <dd className="m-0">
                {timeChanged ? (
                  <>
                    <s className="text-ink-3">{timeLabel(current.start, current.end, false)}</s> <strong>{timeLabel(draft.start, draft.end, false)}</strong>
                  </>
                ) : (
                  <>
                    <strong>{timeLabel(draft.start, draft.end, false)}</strong> <span className="font-medium text-ink-3">(no change)</span>
                  </>
                )}
              </dd>
              <dt className="text-ink-2">Guests</dt>
              <dd className="m-0">
                {changed.guests ? (
                  <>
                    <s className="text-ink-3">{current.guests}</s> <strong>{guestsLabel(draft.guests)}</strong>
                  </>
                ) : (
                  <>
                    <strong>{guestsLabel(draft.guests)}</strong> <span className="font-medium text-ink-3">(no change)</span>
                  </>
                )}
              </dd>
            </dl>
            {(changed.date || changed.start) && (
              <span className="text-sm leading-[1.45] text-body">Free cancellation would move to {deadlineLabel(draft.date, draft.start, rules.cancellation_notice_hours)}.</span>
            )}
          </div>
        ) : (
          <Alert tone="info">Change the date, time or number of guests, then save.</Alert>
        )}
      </aside>
    </div>
  );
}
