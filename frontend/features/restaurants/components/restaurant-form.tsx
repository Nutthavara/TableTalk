"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState, useTransition } from "react";

import { buttonClass } from "@/components/ui/button";
import { FieldError, Helper, inputClass, Label, SuffixInput, textareaClass } from "@/components/ui/field";
import { Check, Close, Upload, Warning } from "@/components/ui/icons";
import { toast } from "@/components/ui/toaster";
import { addRestaurantImages, createRestaurant, deleteRestaurantImage, updateRestaurant } from "@/features/restaurants/actions";
import { cn } from "@/lib/cn";
import { routes } from "@/lib/routes";
import type { Restaurant, RestaurantImage } from "@/types/api";

const MAX_IMAGES = 10;
const MAX_BYTES = 5 * 1024 * 1024;
const IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp"];

type Values = { name: string; description: string; capacity: string; opening_time: string; closing_time: string; cancellation_notice_hours: string };
type Errors = Partial<Record<keyof Values | "images", string>>;

const LABELS: Record<keyof Errors, string> = {
  name: "Restaurant Name",
  description: "Description",
  images: "Restaurant Images",
  capacity: "Capacity",
  opening_time: "Opening Time",
  closing_time: "Closing Time",
  cancellation_notice_hours: "Cancellation Notice",
};

const SECTIONS = [
  { id: "s-basics", label: "Basics", fields: ["name", "description"] },
  { id: "s-photos", label: "Photos", fields: ["images"] },
  { id: "s-seating", label: "Seating & hours", fields: ["capacity", "opening_time", "closing_time"] },
  { id: "s-cancel", label: "Cancellation policy", fields: ["cancellation_notice_hours"] },
] as const;

const isWhole = (s: string) => /^\d+$/.test(s.trim());

/** Mirrors the API's rules for quick feedback; the API validates again on save. */
function validate(v: Values, photoCount: number): Errors {
  const e: Errors = {};
  if (!v.name.trim()) e.name = "Enter the restaurant name.";
  else if (v.name.trim().length > 120) e.name = "Name must be at most 120 characters.";
  if (v.description.length > 2000) e.description = "Description must be at most 2000 characters.";
  if (photoCount < 1) e.images = "Add at least one photo.";
  if (!isWhole(v.capacity) || Number(v.capacity) < 1) e.capacity = "Capacity must be a whole number of at least 1.";
  else if (Number(v.capacity) > 10000) e.capacity = "Capacity must be at most 10,000.";
  if (!v.opening_time) e.opening_time = "Choose an opening time.";
  if (!v.closing_time) e.closing_time = "Choose a closing time.";
  else if (v.opening_time && v.closing_time <= v.opening_time) e.closing_time = "Closing time must be later than opening time.";
  if (!isWhole(v.cancellation_notice_hours) || Number(v.cancellation_notice_hours) < 2) e.cancellation_notice_hours = "Cancellation notice must be at least 2 hours.";
  return e;
}

function checkFiles(files: File[], existing: number): string | null {
  if (existing + files.length > MAX_IMAGES) return `A restaurant can have at most ${MAX_IMAGES} photos.`;
  if (files.some((f) => !IMAGE_TYPES.includes(f.type))) return "Photos must be JPEG, PNG or WebP.";
  if (files.some((f) => f.size > MAX_BYTES)) return "Each photo must be 5 MB or smaller.";
  return null;
}

type NewPhoto = { file: File; url: string };

export function RestaurantForm({ restaurant }: { restaurant?: Restaurant }) {
  const edit = Boolean(restaurant);
  const [values, setValues] = useState<Values>({
    name: restaurant?.name ?? "",
    description: restaurant?.description ?? "",
    capacity: restaurant ? String(restaurant.capacity) : "",
    opening_time: restaurant?.opening_time ?? "",
    closing_time: restaurant?.closing_time ?? "",
    cancellation_notice_hours: String(restaurant?.cancellation_notice_hours ?? 2),
  });
  const [newPhotos, setNewPhotos] = useState<NewPhoto[]>([]);
  const [photoError, setPhotoError] = useState<string | null>(null);
  const [uploading, setUploading] = useState<string | null>(null);
  const [tried, setTried] = useState(false);
  const [apiErrors, setApiErrors] = useState<Errors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  // Release preview URLs when the form unmounts (removed photos are released in removeNew).
  const previews = useRef<NewPhoto[]>([]);
  useEffect(() => {
    previews.current = newPhotos;
  }, [newPhotos]);
  useEffect(() => () => previews.current.forEach((p) => URL.revokeObjectURL(p.url)), []);
  const removeNew = (photo: NewPhoto) => {
    URL.revokeObjectURL(photo.url);
    setNewPhotos((all) => all.filter((x) => x !== photo));
  };

  const existing: RestaurantImage[] = restaurant?.images ?? [];
  const photoCount = edit ? existing.length : newPhotos.length;
  const localErrors = useMemo(() => validate(values, photoCount), [values, photoCount]);
  // API errors win over local ones, but a cleared API error must not hide a local one.
  const errors: Errors = { ...(tried ? localErrors : {}) };
  for (const [k, v] of Object.entries(apiErrors)) if (v) errors[k as keyof Errors] = v;
  if (!tried && localErrors.cancellation_notice_hours && values.cancellation_notice_hours !== "") errors.cancellation_notice_hours = localErrors.cancellation_notice_hours;
  const errorFields = (Object.keys(errors) as (keyof Errors)[]).filter((k) => errors[k]);

  const set = (key: keyof Values) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setValues((v) => ({ ...v, [key]: e.target.value }));
    setApiErrors((a) => ({ ...a, [key]: undefined }));
    setFormError(null);
  };

  const onFiles = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);
    e.target.value = "";
    if (!files.length) return;
    const problem = checkFiles(files, photoCount);
    setPhotoError(problem);
    if (problem) return;
    setApiErrors((a) => ({ ...a, images: undefined }));

    if (!edit) {
      setNewPhotos((p) => [...p, ...files.map((file) => ({ file, url: URL.createObjectURL(file) }))]);
      return;
    }
    // Edit mode: photos are saved straight away.
    const fd = new FormData();
    files.forEach((f) => fd.append("images", f));
    setUploading(files.length === 1 ? files[0].name : `${files.length} photos`);
    startTransition(async () => {
      const result = await addRestaurantImages(restaurant!.id, fd);
      setUploading(null);
      if (result.status === "error") setPhotoError(result.message);
      else toast({ message: files.length === 1 ? "Photo added." : `${files.length} photos added.` });
    });
  };

  const removeExisting = (img: RestaurantImage) =>
    startTransition(async () => {
      const result = await deleteRestaurantImage(restaurant!.id, img.id);
      if (result.status === "error") setPhotoError(result.message);
      else toast({ message: "Photo removed." });
    });

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setTried(true);
    if (Object.keys(localErrors).length) {
      setFormError(null);
      return;
    }
    const fields = {
      name: values.name.trim(),
      description: values.description.trim(),
      capacity: Number(values.capacity),
      opening_time: values.opening_time,
      closing_time: values.closing_time,
      timezone: restaurant?.timezone ?? "Asia/Bangkok",
      cancellation_notice_hours: Number(values.cancellation_notice_hours),
    };
    startTransition(async () => {
      let result;
      if (edit) {
        result = await updateRestaurant(restaurant!.id, fields);
      } else {
        const fd = new FormData();
        Object.entries(fields).forEach(([k, v]) => fd.set(k, String(v)));
        newPhotos.forEach((p) => fd.append("images", p.file));
        result = await createRestaurant(fd);
      }
      if (result?.status === "error") {
        setApiErrors((result.fieldErrors ?? {}) as Errors);
        setFormError(result.message);
      }
    });
  };

  const sectionState = (fields: readonly string[]) => (fields.some((f) => errors[f as keyof Errors]) ? "error" : fields.every((f) => !localErrors[f as keyof Errors]) ? "done" : "todo");
  const noticeOk = !localErrors.cancellation_notice_hours;
  const card = "flex scroll-mt-24 flex-col gap-5 rounded-3xl border border-line-soft bg-surface p-5 sm:p-8";

  return (
    <div className="grid grid-cols-1 items-start gap-8 lg:grid-cols-[240px_minmax(0,1fr)] lg:gap-12">
      <nav aria-label="Form sections" className="hidden flex-col gap-1 pt-2 lg:sticky lg:top-24 lg:flex">
        {SECTIONS.map((s, i) => {
          const state = sectionState(s.fields);
          return (
            <a key={s.id} href={`#${s.id}`} className="flex h-12 items-center gap-3 rounded-xl px-3.5 text-[15px] font-semibold text-body no-underline hover:bg-surface hover:text-ink">
              <span
                className={cn(
                  "flex size-[26px] items-center justify-center rounded-full text-[13px] font-bold",
                  state === "done" && "bg-herb text-white",
                  state === "todo" && "border-2 border-ink-3 text-ink-3",
                  state === "error" && "border-2 border-chili text-chili",
                )}
              >
                {state === "done" ? <Check size={14} /> : i + 1}
              </span>
              {s.label}
            </a>
          );
        })}
      </nav>

      <form onSubmit={submit} aria-label={edit ? "Edit restaurant" : "Add a restaurant"} noValidate className="flex flex-col gap-6">
        {(formError || (tried && errorFields.length > 0)) && (
          <div role="alert" className="flex items-start gap-3 rounded-[14px] border-[1.5px] border-chili-line bg-surface px-[18px] py-3.5 text-[15px] font-semibold text-chili">
            <Warning size={20} className="mt-0.5 shrink-0" />
            <span>
              {errorFields.length > 0
                ? `Please fix ${errorFields.length} ${errorFields.length === 1 ? "field" : "fields"} before saving — ${errorFields.map((f) => LABELS[f]).join(", ")}.`
                : formError}
            </span>
          </div>
        )}

        <section id="s-basics" aria-labelledby="h-basics" className={card}>
          <div className="flex flex-col gap-1">
            <h2 id="h-basics" className="m-0 text-[22px] font-bold">
              1. Basics
            </h2>
            <span className="text-[15px] text-ink-2">What diners see first on the restaurant list.</span>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="f-name">Restaurant Name</Label>
            <input id="f-name" type="text" maxLength={120} value={values.name} onChange={set("name")} aria-invalid={!!errors.name} aria-describedby="f-name-err" className={inputClass(!!errors.name)} />
            <FieldError id="f-name-err">{errors.name}</FieldError>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="f-desc">Description</Label>
            <textarea id="f-desc" rows={4} maxLength={2000} value={values.description} onChange={set("description")} aria-describedby="f-desc-help f-desc-err" className={textareaClass(!!errors.description)} />
            <Helper id="f-desc-help">Shown on your restaurant page and in a shortened form on restaurant cards.</Helper>
            <FieldError id="f-desc-err">{errors.description}</FieldError>
          </div>
        </section>

        <section id="s-photos" aria-labelledby="h-photos" className={card}>
          <div className="flex flex-col gap-1">
            <h2 id="h-photos" className="m-0 text-[22px] font-bold">
              2. Restaurant Images
            </h2>
            <span className="text-[15px] text-ink-2">The first photo is used as the cover on cards and at the top of your page.</span>
          </div>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {edit
              ? existing.map((img, i) => (
                  <PhotoTile
                    key={img.id}
                    src={img.url}
                    index={i}
                    onRemove={() => removeExisting(img)}
                    removeDisabled={pending || existing.length <= 1}
                    removeHint={existing.length <= 1 ? "A restaurant needs at least one photo" : undefined}
                  />
                ))
              : newPhotos.map((p, i) => (
                  <PhotoTile
                    key={p.url}
                    src={p.url}
                    index={i}
                    local
                    onRemove={() => removeNew(p)}
                    removeDisabled={pending}
                  />
                ))}
            {uploading && (
              <div className="flex h-[130px] flex-col justify-center gap-2.5 rounded-[14px] bg-sand p-4">
                <span className="truncate text-[13px] font-semibold text-body">Uploading {uploading}</span>
                <div role="progressbar" aria-label={`Uploading ${uploading}`} className="h-1.5 overflow-hidden rounded-full bg-line-mid">
                  <div className="h-1.5 w-1/2 animate-pulse bg-paprika" />
                </div>
              </div>
            )}
            {photoCount < MAX_IMAGES && (
              <label
                htmlFor="f-photos"
                className="flex h-[130px] cursor-pointer flex-col items-center justify-center gap-1.5 rounded-[14px] border-[1.5px] border-dashed border-[#C9BBA8] bg-surface text-center text-sm font-semibold text-body hover:bg-sand"
              >
                <Upload size={22} />
                Add photos
                <span className="font-medium text-ink-3">JPEG, PNG or WebP · max 5 MB</span>
              </label>
            )}
            <input id="f-photos" type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={onFiles} disabled={pending} className="sr-only" />
          </div>
          <FieldError>{photoError ?? errors.images}</FieldError>
          {edit && <Helper>Photo changes are saved immediately.</Helper>}
        </section>

        <section id="s-seating" aria-labelledby="h-seating" className={card}>
          <div className="flex flex-col gap-1">
            <h2 id="h-seating" className="m-0 text-[22px] font-bold">
              3. Seating &amp; opening hours
            </h2>
            <span className="text-[15px] text-ink-2">Used to check every reservation.</span>
          </div>
          <div className="flex max-w-[320px] flex-col gap-1.5">
            <Label htmlFor="f-cap">Capacity</Label>
            <SuffixInput id="f-cap" suffix="seats" type="number" min={1} step={1} inputMode="numeric" value={values.capacity} onChange={set("capacity")} invalid={!!errors.capacity} aria-invalid={!!errors.capacity} aria-describedby="f-cap-help f-cap-err" />
          </div>
          <Helper id="f-cap-help" className="-mt-3">
            Total number of guests (people, not tables) you can seat at the same time. Overlapping bookings can&apos;t go over this number.
          </Helper>
          <FieldError id="f-cap-err">{errors.capacity}</FieldError>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-[repeat(2,minmax(0,320px))]">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-open">Opening Time</Label>
              <input id="f-open" type="time" value={values.opening_time} onChange={set("opening_time")} aria-invalid={!!errors.opening_time} className={inputClass(!!errors.opening_time)} />
              <FieldError>{errors.opening_time}</FieldError>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-close">Closing Time</Label>
              <input id="f-close" type="time" value={values.closing_time} onChange={set("closing_time")} aria-invalid={!!errors.closing_time} className={inputClass(!!errors.closing_time)} />
              <FieldError>{errors.closing_time}</FieldError>
            </div>
          </div>
          <Helper className="-mt-2">Reservations must start and end within these hours. Overnight opening hours are not supported.</Helper>
        </section>

        <section id="s-cancel" aria-labelledby="h-cancel" className={card}>
          <div className="flex flex-col gap-1">
            <h2 id="h-cancel" className="m-0 text-[22px] font-bold">
              4. Cancellation policy
            </h2>
            <span className="text-[15px] text-ink-2">Shown to diners before they book and when they cancel.</span>
          </div>
          <div className="flex max-w-[320px] flex-col gap-1.5">
            <Label htmlFor="f-notice">Cancellation Notice</Label>
            <SuffixInput
              id="f-notice"
              suffix="hours"
              type="number"
              min={2}
              step={1}
              inputMode="numeric"
              value={values.cancellation_notice_hours}
              onChange={set("cancellation_notice_hours")}
              invalid={!!errors.cancellation_notice_hours}
              aria-invalid={!!errors.cancellation_notice_hours}
              aria-describedby="f-notice-help f-notice-err"
            />
          </div>
          <Helper id="f-notice-help" className="-mt-3">
            Customers must cancel at least this many hours before their reservation. Minimum 2 hours.
          </Helper>
          <div id="f-notice-err" aria-live="polite">
            <FieldError>{errors.cancellation_notice_hours}</FieldError>
          </div>
          <div className="rounded-xl bg-sand px-4 py-3.5 text-sm leading-normal text-body">
            Diners will see: <strong>{noticeOk ? `Free cancellation until ${Number(values.cancellation_notice_hours)} hours before your reservation.` : "—"}</strong>
          </div>
        </section>

        <div className="sticky bottom-24 z-10 flex flex-col gap-4 rounded-[20px] border border-line-soft bg-surface px-6 py-5 shadow-[0_-4px_24px_rgba(34,28,24,.06)] sm:bottom-4 sm:flex-row sm:items-center sm:justify-between">
          <span className="text-sm text-ink-2">
            {edit ? "Existing reservations keep their booked time." : "Your restaurant goes live on TableTalk as soon as it is created."}
          </span>
          <div className="flex gap-3">
            <Link href={routes.myRestaurants()} className={buttonClass("ghost", "lg")}>
              Cancel
            </Link>
            <button type="submit" disabled={pending} className={buttonClass("primary", "lg", "grow px-7 sm:grow-0")}>
              {pending && !uploading ? "Saving…" : edit ? "Save Changes" : "Create Restaurant"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

function PhotoTile({ src, index, onRemove, removeDisabled, removeHint, local }: { src: string; index: number; onRemove: () => void; removeDisabled?: boolean; removeHint?: string; local?: boolean }) {
  return (
    <div className="relative h-[130px] overflow-hidden rounded-[14px] bg-sand">
      {local ? (
        // Local previews are blob: URLs, which next/image cannot optimise.
        // eslint-disable-next-line @next/next/no-img-element
        <img src={src} alt={`Photo ${index + 1}`} className="size-full object-cover" />
      ) : (
        <Image src={src} alt={`Photo ${index + 1}`} fill unoptimized sizes="200px" className="object-cover" />
      )}
      {index === 0 && <span className="absolute top-2 left-2 h-6 rounded-full bg-surface px-2.5 text-xs leading-6 font-bold">Cover</span>}
      <button
        type="button"
        aria-label={`Remove photo ${index + 1}`}
        title={removeHint}
        onClick={onRemove}
        disabled={removeDisabled}
        className="absolute top-1.5 right-1.5 flex size-8 cursor-pointer items-center justify-center rounded-full bg-[rgba(34,28,24,.72)] text-white disabled:cursor-not-allowed disabled:opacity-40"
      >
        <Close size={16} />
      </button>
    </div>
  );
}
