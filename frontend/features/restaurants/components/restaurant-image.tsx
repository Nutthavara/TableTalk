import Image from "next/image";

import { cn } from "@/lib/cn";
import type { RestaurantImage as Img } from "@/types/api";

// Warm plate illustration from the design, shown when a restaurant has no photo to display.
export function FoodIllustration({ className, seed = 0 }: { className?: string; seed?: number }) {
  const palettes = [
    ["#2F5D46", "#D98A2B", "#F4D35E", "#7A9E3B", "#C2410C"],
    ["#8FA7B3", "#E9C46A", "#C0392B", "#6A994E", "#F4F1DE"],
    ["#5B2A3A", "#E0A03A", "#C2572B", "#F3E3C3", "#3F7D3A"],
    ["#9DB7A0", "#EFE6D4", "#D9A441", "#7A5230", "#4E7D3A"],
  ];
  const [bg, food, a1, a2, a3] = palettes[seed % palettes.length];
  return (
    <svg viewBox="0 0 400 260" preserveAspectRatio="xMidYMid slice" className={cn("block size-full", className)} aria-hidden>
      <rect width="400" height="260" fill={bg} />
      <circle cx="252" cy="146" r="118" fill="#000" opacity="0.14" />
      <circle cx="244" cy="136" r="114" fill="#FBF7F1" />
      <circle cx="244" cy="136" r="90" fill="#EFE6D8" />
      <circle cx="244" cy="136" r="72" fill={food} />
      <circle cx="220" cy="114" r="24" fill={a1} />
      <circle cx="270" cy="158" r="19" fill={a2} />
      <circle cx="266" cy="106" r="10" fill={a3} />
      <circle cx="74" cy="62" r="46" fill="#FBF7F1" />
      <circle cx="74" cy="62" r="34" fill={a3} />
    </svg>
  );
}

/**
 * A restaurant photo filling its (relatively positioned) parent. Photos are loaded straight from the
 * API (unoptimized): the Next.js optimizer would have to fetch from the API host server-side.
 */
export function RestaurantPhoto({ image, alt, sizes, priority = false, className }: { image?: Img; alt: string; sizes: string; priority?: boolean; className?: string }) {
  if (!image) return <FoodIllustration className={className} />;
  return <Image src={image.url} alt={alt} fill unoptimized sizes={sizes} priority={priority} className={cn("object-cover", className)} />;
}
