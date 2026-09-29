"use client";

import Link from "next/link";
import { useMemo, useState, useTransition } from "react";

import { buttonClass } from "@/components/ui/button";
import { Check, Info, Warning } from "@/components/ui/icons";
import { toast } from "@/components/ui/toaster";
import { createReservation } from "@/features/reservations/actions";
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
import { routes } from "@/lib/routes";

type Props = {
  restaurant: BookingRules & { id: string; name: string };
  signedIn: boolean;
  initialDraft: BookingDraft;
  idPrefix?: string;
};

type Step = "form" | "confirm" | "seats" | "success";

const h2 = "m-0 font-display text-[26px] font-bold tracking-[-0.015em]";
const overline = "text-[13px] font-bold tracking-[.04em] text-ink-3 uppercase";

/**
 * Reserve a Table: validate in the browser for quick feedback, show a summary, then book through the
 * API, which is the only judge of availability.
 */
export function BookingPanel({ restaurant, signedIn, initialDraft, idPrefix = "bk" }: Props) {
  const [draft, setDraft] = useState(initialDraft);
  const [step, setStep] = useState<Step>("form");
  const [apiError, setApiError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const now = useMemo(() => nowInZone(restaurant.timezone), [restaurant.timezone]);
  const check = checkBooking(draft, restaurant, now);
  const blocked = check.errors.length > 0;
  const summary = { date: dateLabel(draft.date), time: timeLabel(draft.start, draft.end), guests: guestsLabel(draft.guests) };
  const returnTo = routes.restaurant(restaurant.id, { date: draft.date, start: draft.start, end: draft.end, guests: draft.guests });

  const change = (d: BookingDraft) => {
    setDraft(d);
    setApiError(null);
  };

  const confirm = () =>
    startTransition(async () => {
      const result = await createReservation(restaurant.id, draft);
      if (result.status === "success") {
        setStep("success");
        toast({ message: "Reservation confirmed — see it in My Reservations.", action: { label: "View", href: routes.myReservations() } });
      } else if (result.status === "error" && result.code === "INSUFFICIENT_CAPACITY") {
        setStep("seats");
      } else if (result.status === "error") {
        setApiError(result.message);
        setStep("form");
      }
    });

  if (step === "confirm") {
    return (
      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-1">
          <span className={overline}>Step 2 of 2</span>
          <h2 className={h2}>Confirm your reservation</h2>
        </div>
        <dl className="m-0 grid grid-cols-[90px_1fr] gap-y-2.5 rounded-2xl bg-sand p-[18px] text-base">
          <dt className="text-ink-2">Restaurant</dt>
          <dd className="m-0 font-bold">{restaurant.name}</dd>
          <dt className="text-ink-2">Date</dt>
          <dd className="m-0 font-bold">{summary.date}</dd>
          <dt className="text-ink-2">Time</dt>
          <dd className="m-0 font-bold">{summary.time}</dd>
          <dt className="text-ink-2">Guests</dt>
          <dd className="m-0 font-bold">{summary.guests}</dd>
        </dl>
        <div className="flex gap-2.5 rounded-xl bg-herb-soft p-3.5 text-sm leading-[1.45] text-herb-ink">
          <Info size={18} className="mt-px shrink-0" />
          <span>
            Free cancellation until <strong>{deadlineLabel(draft.date, draft.start, restaurant.cancellation_notice_hours)}</strong> ({restaurant.cancellation_notice_hours}{" "}
            hours before your reservation).
          </span>
        </div>
        <button type="button" onClick={confirm} disabled={pending} className={buttonClass("primary", "xl")}>
          {pending ? "Checking availability…" : "Confirm Reservation"}
        </button>
        <button type="button" onClick={() => setStep("form")} disabled={pending} className={buttonClass("ghost", "md")}>
          Back to edit details
        </button>
      </div>
    );
  }

  if (step === "seats") {
    return (
      <div className="flex flex-col gap-5">
        <h2 className={h2}>Reserve a Table</h2>
        <div role="alert" className="flex gap-3 rounded-[14px] bg-chili-soft p-4 text-chili">
          <Warning size={22} className="shrink-0" />
          <div className="flex flex-col gap-1">
            <strong className="text-base">Not enough seats available</strong>
            <span className="text-sm leading-normal">
              There aren&apos;t enough free seats for {summary.guests} on {summary.date}, {timeLabel(draft.start, draft.end, false)}. Try another time or fewer guests.
            </span>
          </div>
        </div>
        <button type="button" onClick={() => setStep("form")} className={buttonClass("primary", "xl")}>
          Change time or guests
        </button>
      </div>
    );
  }

  if (step === "success") {
    return (
      <div role="status" className="flex flex-col items-center gap-4 py-2 text-center">
        <span aria-hidden className="flex size-16 items-center justify-center rounded-full bg-herb-soft text-herb">
          <Check size={32} />
        </span>
        <h2 className="m-0 font-display text-[28px] font-bold">You&apos;re booked!</h2>
        <p className="m-0 text-base leading-[1.55] text-body">
          {restaurant.name}
          <br />
          {summary.date} · {timeLabel(draft.start, draft.end, false)} · {summary.guests}
        </p>
        <Link href={routes.myReservations()} className={buttonClass("primary", "lg", "w-full")}>
          View My Reservations
        </Link>
        <button type="button" onClick={() => setStep("form")} className={buttonClass("ghost", "md", "w-full")}>
          Make another booking
        </button>
      </div>
    );
  }

  const errors = apiError ? [...check.errors, apiError] : check.errors;
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!blocked && signedIn) setStep("confirm");
      }}
      className="flex flex-col gap-5"
    >
      <div className="flex flex-col gap-1">
        <h2 className={h2}>Reserve a Table</h2>
        <span className="text-sm text-ink-2">
          {signedIn
            ? `Open ${restaurant.opening_time} – ${restaurant.closing_time} · up to ${restaurant.capacity} seats`
            : "Pick a date and time, then log in to confirm. We'll keep what you entered."}
        </span>
      </div>
      <BookingFields
        idPrefix={idPrefix}
        draft={draft}
        onChange={change}
        check={check}
        capacity={restaurant.capacity}
        openingTime={restaurant.opening_time}
        closingTime={restaurant.closing_time}
        minDate={now.date}
      />
      <BookingErrors id={`${idPrefix}-errors`} messages={errors} />
      <div className="flex flex-col gap-2.5 rounded-2xl bg-sand p-[18px]">
        <span className={overline}>Booking summary</span>
        <dl className="m-0 grid grid-cols-[72px_1fr] gap-y-1.5 text-[15px]">
          <dt className="text-ink-2">Date</dt>
          <dd className="m-0 font-bold">{summary.date}</dd>
          <dt className="text-ink-2">Time</dt>
          <dd className="m-0 font-bold">{summary.time}</dd>
          <dt className="text-ink-2">Guests</dt>
          <dd className="m-0 font-bold">{summary.guests}</dd>
        </dl>
      </div>
      {signedIn ? (
        <>
          <button type="submit" disabled={blocked} className={buttonClass("primary", "xl")}>
            Check Availability / Reserve
          </button>
          <span className="text-center text-[13px] leading-[1.45] text-ink-3">Seats are only held once you confirm on the next step.</span>
        </>
      ) : (
        <a href={routes.login(returnTo)} className={buttonClass("primary", "xl")}>
          Log in to reserve
        </a>
      )}
    </form>
  );
}
