"use client";

import { Badge } from "@/components/ui/badge";
import { Helper, inputClass, Label } from "@/components/ui/field";
import { Minus, Plus } from "@/components/ui/icons";
import { type BookingCheck, type BookingDraft, guestsLabel } from "@/features/reservations/booking";
import { cn } from "@/lib/cn";

type Props = {
  idPrefix: string;
  draft: BookingDraft;
  onChange: (d: BookingDraft) => void;
  check: BookingCheck;
  capacity: number;
  openingTime: string;
  closingTime: string;
  minDate: string;
  disabled?: boolean;
  /** Edit Booking marks fields that differ from the current booking. */
  changed?: Partial<Record<keyof BookingDraft, boolean>>;
  compactStepper?: boolean;
};

function ChangedTag({ show }: { show?: boolean }) {
  return show ? (
    <Badge tone="changed" className="h-5 px-2 text-xs">
      Changed
    </Badge>
  ) : null;
}

/** Date, start/end time and guest stepper: the four fields of every booking (new and edit). */
export function BookingFields({ idPrefix, draft, onChange, check, capacity, openingTime, closingTime, minDate, disabled, changed = {}, compactStepper }: Props) {
  const set = (patch: Partial<BookingDraft>) => onChange({ ...draft, ...patch });
  const errorsId = `${idPrefix}-errors`;

  return (
    <>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor={`${idPrefix}-date`}>
          Date <ChangedTag show={changed.date} />
        </Label>
        <input
          id={`${idPrefix}-date`}
          type="date"
          min={minDate}
          value={draft.date}
          disabled={disabled}
          onChange={(e) => set({ date: e.target.value })}
          aria-invalid={check.dateInvalid}
          aria-describedby={errorsId}
          className={inputClass(check.dateInvalid, changed.date)}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`${idPrefix}-start`}>
            Start Time <ChangedTag show={changed.start} />
          </Label>
          <input
            id={`${idPrefix}-start`}
            type="time"
            value={draft.start}
            disabled={disabled}
            onChange={(e) => set({ start: e.target.value })}
            aria-invalid={check.timeInvalid}
            aria-describedby={`${idPrefix}-hours ${errorsId}`}
            className={inputClass(check.timeInvalid, changed.start)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`${idPrefix}-end`}>
            End Time <ChangedTag show={changed.end} />
          </Label>
          <input
            id={`${idPrefix}-end`}
            type="time"
            value={draft.end}
            disabled={disabled}
            onChange={(e) => set({ end: e.target.value })}
            aria-invalid={check.timeInvalid}
            aria-describedby={`${idPrefix}-hours ${errorsId}`}
            className={inputClass(check.timeInvalid, changed.end)}
          />
        </div>
      </div>
      <Helper id={`${idPrefix}-hours`} className="-mt-3">
        Your booking must start and end within {openingTime} – {closingTime}.
      </Helper>
      <div className="flex flex-col gap-1.5">
        <Label id={`${idPrefix}-guests`}>
          Number of Guests <ChangedTag show={changed.guests} />
        </Label>
        <div
          role="group"
          aria-labelledby={`${idPrefix}-guests`}
          className={cn(
            "flex h-14 items-center justify-between rounded-xl border-[1.5px] px-1.5",
            changed.guests ? "border-paprika" : "border-line",
            compactStepper && "w-full sm:w-[280px]",
          )}
        >
          <button
            type="button"
            aria-label="Remove a guest"
            disabled={disabled || draft.guests <= 1}
            onClick={() => set({ guests: Math.max(1, draft.guests - 1) })}
            className="flex size-11 cursor-pointer items-center justify-center rounded-[10px] hover:bg-hover disabled:cursor-not-allowed disabled:text-[#C9BFB5] disabled:hover:bg-transparent"
          >
            <Minus size={20} />
          </button>
          <output aria-live="polite" className="text-lg font-bold">
            {guestsLabel(draft.guests)}
          </output>
          <button
            type="button"
            aria-label="Add a guest"
            disabled={disabled || draft.guests >= capacity}
            onClick={() => set({ guests: Math.min(capacity, draft.guests + 1) })}
            className="flex size-11 cursor-pointer items-center justify-center rounded-[10px] hover:bg-hover disabled:cursor-not-allowed disabled:text-[#C9BFB5] disabled:hover:bg-transparent"
          >
            <Plus size={20} />
          </button>
        </div>
      </div>
    </>
  );
}

export function BookingErrors({ id, messages }: { id: string; messages: string[] }) {
  return (
    <div id={id} aria-live="polite" className="flex flex-col gap-2 empty:hidden">
      {messages.map((m) => (
        <div key={m} className="flex items-start gap-2.5 rounded-xl bg-chili-soft px-3.5 py-3 text-sm leading-[1.45] font-semibold text-chili">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden className="mt-px shrink-0">
            <path d="M12 3l10 18H2L12 3z" />
            <path d="M12 10v4.5M12 17.5h.01" />
          </svg>
          <span>{m}</span>
        </div>
      ))}
    </div>
  );
}
