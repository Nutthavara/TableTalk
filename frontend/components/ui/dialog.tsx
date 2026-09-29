"use client";

import { useEffect, useRef } from "react";

import { cn } from "@/lib/cn";

type DialogProps = {
  open: boolean;
  onClose: () => void;
  labelledBy: string;
  describedBy?: string;
  /** alertdialog for destructive confirmations. */
  role?: "dialog" | "alertdialog";
  /** Bottom sheet on small screens, centred card from `sm` up (the design's mobile pattern). */
  sheetOnMobile?: boolean;
  className?: string;
  children: React.ReactNode;
};

/**
 * Native <dialog> shown with showModal(): focus is trapped and restored to the trigger, Esc closes,
 * and the page behind is inert.
 */
export function Dialog({ open, onClose, labelledBy, describedBy, role = "dialog", sheetOnMobile = true, className, children }: DialogProps) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (open && !el.open) el.showModal();
    if (!open && el.open) el.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      role={role}
      aria-modal="true"
      aria-labelledby={labelledBy}
      aria-describedby={describedBy}
      onClose={onClose}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose(); // click on the backdrop
      }}
      className={cn(
        "m-auto max-h-[92vh] overflow-y-auto bg-surface p-0 text-ink shadow-dialog",
        sheetOnMobile
          ? "mb-0 w-full max-w-full rounded-t-3xl rounded-b-none sm:mb-auto sm:w-[480px] sm:max-w-[calc(100vw-32px)] sm:rounded-3xl"
          : "w-[480px] max-w-[calc(100vw-32px)] rounded-3xl",
        className,
      )}
    >
      {open && (
        <div className="flex flex-col gap-5 px-4 pt-2.5 pb-7 sm:p-8">
          {sheetOnMobile && <span aria-hidden className="h-[5px] w-10 self-center rounded-full bg-line sm:hidden" />}
          {children}
        </div>
      )}
    </dialog>
  );
}
