import Link from "next/link";

import { buttonClass } from "@/components/ui/button";
import { Retry, Warning } from "@/components/ui/icons";
import { cn } from "@/lib/cn";

/** Dashed empty state that always offers the next step. */
export function EmptyState({
  title,
  body,
  action,
  icon,
  className,
}: {
  title: string;
  body?: string;
  action?: React.ReactNode;
  icon?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col items-center gap-3 rounded-3xl border-[1.5px] border-dashed border-line bg-surface px-6 py-16 text-center", className)}>
      {icon}
      <strong className="font-display text-2xl font-bold">{title}</strong>
      {body && <span className="max-w-md text-[15px] leading-normal text-ink-2">{body}</span>}
      {action}
    </div>
  );
}

export function ErrorState({ title, body, onRetry, retryHref }: { title: string; body?: string; onRetry?: () => void; retryHref?: string }) {
  return (
    <div role="alert" className="flex flex-col items-center gap-3 rounded-3xl border border-chili-line bg-surface px-6 py-16 text-center">
      <span aria-hidden className="flex size-14 items-center justify-center rounded-full bg-chili-soft text-chili">
        <Warning size={26} />
      </span>
      <strong className="font-display text-2xl font-bold">{title}</strong>
      <span className="max-w-md text-[15px] leading-normal text-ink-2">
        {body ?? "Something went wrong on our side or your connection dropped. Please try again in a moment."}
      </span>
      {onRetry && (
        <button type="button" onClick={onRetry} className={buttonClass("secondary", "md")}>
          <Retry size={18} />
          Try again
        </button>
      )}
      {retryHref && (
        <Link href={retryHref} className={buttonClass("secondary", "md")}>
          <Retry size={18} />
          Try again
        </Link>
      )}
    </div>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div aria-hidden className={cn("animate-pulse rounded-lg bg-line-soft", className)} />;
}
