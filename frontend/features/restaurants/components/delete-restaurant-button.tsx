"use client";

import { useState, useTransition } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { inputClass } from "@/components/ui/field";
import { Trash } from "@/components/ui/icons";
import { toast } from "@/components/ui/toaster";
import { deleteRestaurant } from "@/features/restaurants/actions";

/** Delete with a type-the-name confirmation (it cannot be undone). */
export function DeleteRestaurantButton({ id, name }: { id: string; name: string }) {
  const [open, setOpen] = useState(false);
  const [typed, setTyped] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();
  const ok = typed.trim() === name.trim();

  const confirm = () =>
    startTransition(async () => {
      const result = await deleteRestaurant(id);
      if (result.status === "error") {
        setError(result.message);
        return;
      }
      setOpen(false);
      toast({ message: `${name} has been deleted.` });
    });

  return (
    <>
      <Button
        variant="dangerOutline"
        className="h-11"
        aria-label={`Delete ${name}`}
        onClick={() => {
          setTyped("");
          setError(null);
          setOpen(true);
        }}
      >
        <Trash size={16} />
        Delete
      </Button>
      <Dialog open={open} onClose={() => setOpen(false)} role="alertdialog" labelledBy={`dr-h-${id}`} describedBy={`dr-d-${id}`} className="border-t-[6px] border-chili sm:w-[500px]">
        <span aria-hidden className="hidden size-[52px] items-center justify-center rounded-full bg-chili-soft text-chili sm:flex">
          <Trash size={26} />
        </span>
        <div className="flex flex-col gap-2">
          <h2 id={`dr-h-${id}`} className="m-0 font-display text-2xl font-bold tracking-[-0.015em] sm:text-[26px]">
            Delete {name}?
          </h2>
          <p id={`dr-d-${id}`} className="m-0 text-base leading-[1.55] text-body">
            This permanently deletes the restaurant page, its photos and reviews. <strong>This can&apos;t be undone.</strong>
          </p>
        </div>
        <div className="flex flex-col gap-1.5">
          <label htmlFor={`dr-confirm-${id}`} className="text-sm font-semibold">
            Type <strong>{name}</strong> to confirm
          </label>
          <input id={`dr-confirm-${id}`} type="text" autoComplete="off" value={typed} onChange={(e) => setTyped(e.target.value)} className={inputClass()} />
        </div>
        {error && <Alert tone="error">{error}</Alert>}
        <div className="flex flex-col-reverse justify-end gap-3 sm:flex-row">
          <Button variant="secondary" onClick={() => setOpen(false)}>
            Keep restaurant
          </Button>
          <Button variant="danger" onClick={confirm} disabled={!ok || pending}>
            {pending ? "Deleting…" : "Delete restaurant"}
          </Button>
        </div>
      </Dialog>
    </>
  );
}
