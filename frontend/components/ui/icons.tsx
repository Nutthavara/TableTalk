// Stroke icons from the TableTalk design (24×24 grid, currentColor).
import type { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement> & { size?: number; strokeWidth?: number };

function stroke({ size = 18, strokeWidth = 1.8, ...rest }: IconProps) {
  return {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true,
    ...rest,
  };
}

export const ChevronLeft = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M15 6l-6 6 6 6" /></svg>;
export const ChevronDown = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M6 9l6 6 6-6" /></svg>;
export const ArrowRight = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M5 12h14M13 6l6 6-6 6" /></svg>;
export const Search = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></svg>;
export const Close = (p: IconProps) => <svg {...stroke({ strokeWidth: 2.2, ...p })}><path d="M6 6l12 12M18 6L6 18" /></svg>;
export const Clock = (p: IconProps) => <svg {...stroke(p)}><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>;
export const Users = (p: IconProps) => (
  <svg {...stroke(p)}>
    <circle cx="9" cy="8" r="3.5" />
    <path d="M2.5 20v-1a4.5 4.5 0 0 1 4.5-4.5h4a4.5 4.5 0 0 1 4.5 4.5v1" />
    <path d="M16 4.7a3.5 3.5 0 0 1 0 6.6" />
    <path d="M21.5 20v-1a4.5 4.5 0 0 0-3-4.2" />
  </svg>
);
export const Calendar = (p: IconProps) => <svg {...stroke(p)}><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M3 10h18M8 3v4M16 3v4" /></svg>;
export const CalendarX = (p: IconProps) => (
  <svg {...stroke({ strokeWidth: 2, ...p })}><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M3 10h18M8 3v4M16 3v4M10 13.5l4 4M14 13.5l-4 4" /></svg>
);
export const Pencil = (p: IconProps) => <svg {...stroke(p)}><path d="M4 20h4L19 9l-4-4L4 16v4z" /></svg>;
export const Trash = (p: IconProps) => <svg {...stroke(p)}><path d="M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3" /></svg>;
export const Eye = (p: IconProps) => <svg {...stroke(p)}><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></svg>;
export const Store = (p: IconProps) => <svg {...stroke(p)}><path d="M4 10v10h16V10" /><path d="M3 4h18l-1 6H4L3 4z" /><path d="M10 20v-5h4v5" /></svg>;
export const Utensils = (p: IconProps) => <svg {...stroke({ strokeWidth: 1.9, ...p })}><path d="M5 3v6a2 2 0 0 0 4 0V3M7 11v10M17 21V3c-2 0-3.5 2.5-3.5 6s1.5 4 3.5 4" /></svg>;
export const User = (p: IconProps) => <svg {...stroke({ strokeWidth: 1.9, ...p })}><circle cx="12" cy="8" r="4" /><path d="M4 21v-1a7 7 0 0 1 16 0v1" /></svg>;
export const Plus = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M12 5v14M5 12h14" /></svg>;
export const Minus = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M5 12h14" /></svg>;
export const Check = (p: IconProps) => <svg {...stroke({ strokeWidth: 2.4, ...p })}><path d="M5 12.5l4.5 4.5L19 7.5" /></svg>;
export const Info = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><circle cx="12" cy="12" r="9" /><path d="M12 11v6M12 7.5h.01" /></svg>;
export const Warning = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M12 3l10 18H2L12 3z" /><path d="M12 10v4.5M12 17.5h.01" /></svg>;
export const Lock = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><rect x="5" y="11" width="14" height="10" rx="2" /><path d="M8 11V8a4 4 0 0 1 8 0v3" /></svg>;
export const Photo = (p: IconProps) => <svg {...stroke(p)}><rect x="3" y="4" width="18" height="16" rx="2" /><circle cx="9" cy="10" r="2" /><path d="M21 16l-5-5-9 9" /></svg>;
export const Upload = (p: IconProps) => <svg {...stroke(p)}><path d="M12 16V4M7 9l5-5 5 5" /><path d="M4 16v3a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-3" /></svg>;
export const Retry = (p: IconProps) => <svg {...stroke({ strokeWidth: 2, ...p })}><path d="M20 11a8 8 0 1 0-2.3 5.7" /><path d="M20 4v7h-7" /></svg>;
export const LogOut = (p: IconProps) => <svg {...stroke(p)}><path d="M15 4h4v16h-4M10 8l-4 4 4 4M6 12h10" /></svg>;

const STAR_PATH = "M12 2.5l2.95 6.1 6.55.95-4.75 4.6 1.12 6.55L12 17.6l-5.87 3.1 1.12-6.55L2.5 9.55l6.55-.95L12 2.5z";

export function Star({ size = 18, fill = "#E3A327", ...rest }: SVGProps<SVGSVGElement> & { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" aria-hidden {...rest}>
      <path d={STAR_PATH} fill={fill} />
    </svg>
  );
}

export function Logo({ size = 36, inverted = false }: { size?: number; inverted?: boolean }) {
  const bg = inverted ? "#FBF7F1" : "#BD4719";
  const fg = inverted ? "#BD4719" : "#FFFFFF";
  return (
    <svg width={size} height={size} viewBox="0 0 36 36" aria-hidden>
      <rect width="36" height="36" rx="11" fill={bg} />
      <circle cx="18" cy="17" r="8.5" fill="none" stroke={fg} strokeWidth="2.4" />
      <circle cx="18" cy="17" r="3.2" fill={fg} />
      <path d="M11.5 23.5L9 28.5l5.5-2.4" fill="none" stroke={fg} strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
