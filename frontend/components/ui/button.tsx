import { cn } from "@/lib/cn";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "dangerOutline" | "dark" | "soft";
export type ButtonSize = "sm" | "md" | "lg" | "xl";

const base =
  "inline-flex items-center justify-center gap-2 rounded-xl font-semibold no-underline transition-colors cursor-pointer disabled:cursor-not-allowed";

const variants: Record<ButtonVariant, string> = {
  primary: "bg-paprika text-white hover:bg-paprika-dark hover:text-white disabled:bg-disabled disabled:text-disabled-ink",
  secondary:
    "border-[1.5px] border-line bg-surface text-ink hover:bg-hover hover:text-ink disabled:border-line-soft disabled:bg-sand disabled:text-disabled-ink",
  ghost: "bg-transparent text-ink-2 hover:bg-hover hover:text-ink-2 disabled:text-[#B3A89F]",
  danger: "bg-chili text-white hover:bg-chili-dark hover:text-white disabled:bg-disabled disabled:text-disabled-ink",
  dangerOutline:
    "border-[1.5px] border-chili-line bg-surface text-chili hover:bg-chili-soft hover:text-chili disabled:border-line-soft disabled:bg-sand disabled:text-disabled-ink",
  dark: "bg-ink text-white hover:bg-black hover:text-white",
  soft: "bg-sand text-ink hover:bg-hover hover:text-ink",
};

const sizes: Record<ButtonSize, string> = {
  sm: "h-10 px-3.5 text-sm rounded-[10px]",
  md: "h-12 px-5 text-[15px]",
  lg: "h-[52px] px-[22px] text-base",
  xl: "h-14 px-6 text-[17px] font-bold rounded-[14px]",
};

/** Class names for a button-styled element (also used on links). */
export function buttonClass(variant: ButtonVariant = "primary", size: ButtonSize = "md", extra?: string) {
  return cn(base, variants[variant], sizes[size], variant === "primary" && size !== "sm" && "font-bold", extra);
}

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant; size?: ButtonSize };

export function Button({ variant = "primary", size = "md", className, type = "button", ...rest }: ButtonProps) {
  return <button type={type} className={buttonClass(variant, size, className)} {...rest} />;
}
