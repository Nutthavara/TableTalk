"use client";

import Image from "next/image";
import { useState } from "react";

import { Dialog } from "@/components/ui/dialog";
import { Close, Photo } from "@/components/ui/icons";
import { FoodIllustration } from "@/features/restaurants/components/restaurant-image";
import { cn } from "@/lib/cn";
import type { RestaurantImage } from "@/types/api";

/** Cover + two side photos (desktop), single cover (mobile), and a dialog with every photo. */
export function Gallery({ images, name, compact = false }: { images: RestaurantImage[]; name: string; compact?: boolean }) {
  const [open, setOpen] = useState(false);
  const [first, second, third] = images;

  const photo = (img: RestaurantImage | undefined, i: number, sizes: string, priority = false) =>
    img ? <Image src={img.url} alt={`${name} — photo ${i + 1}`} fill unoptimized sizes={sizes} priority={priority} className="object-cover" /> : <FoodIllustration seed={i} />;

  return (
    <section aria-label="Photos" className={cn("relative grid overflow-hidden rounded-3xl", compact ? "h-[220px] grid-cols-1 sm:h-[340px] sm:grid-cols-[2fr_1fr] sm:gap-3" : "h-[240px] grid-cols-1 sm:h-[460px] sm:grid-cols-[2fr_1fr] sm:grid-rows-2 sm:gap-3")}>
      <div className={cn("relative overflow-hidden", !compact && "sm:row-span-2")}>{photo(first, 0, "(min-width: 640px) 66vw, 100vw", true)}</div>
      <div className="relative hidden overflow-hidden sm:block">{photo(second ?? first, 1, "33vw")}</div>
      {!compact && <div className="relative hidden overflow-hidden sm:block">{photo(third ?? second ?? first, 2, "33vw")}</div>}
      {images.length > 1 && (
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="absolute right-3 bottom-3 flex h-11 cursor-pointer items-center gap-2 rounded-xl bg-surface px-4 text-[15px] font-semibold shadow-[0_4px_16px_rgba(34,28,24,.18)] hover:bg-hover sm:right-5 sm:bottom-5"
        >
          <Photo size={18} />
          Show all {images.length} photos
        </button>
      )}
      <Dialog open={open} onClose={() => setOpen(false)} labelledBy="gallery-h" sheetOnMobile={false} className="sm:w-[960px]">
        <div className="flex items-center justify-between gap-4">
          <h2 id="gallery-h" className="m-0 font-display text-2xl font-bold">
            {name} — {images.length} photos
          </h2>
          <button type="button" aria-label="Close photos" onClick={() => setOpen(false)} className="flex size-11 cursor-pointer items-center justify-center rounded-xl hover:bg-hover">
            <Close size={18} />
          </button>
        </div>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {images.map((img, i) => (
            <div key={img.id} className="relative aspect-[4/3] overflow-hidden rounded-2xl">
              <Image src={img.url} alt={`${name} — photo ${i + 1}`} fill unoptimized sizes="(min-width: 640px) 460px, 100vw" className="object-cover" />
            </div>
          ))}
        </div>
      </Dialog>
    </section>
  );
}
