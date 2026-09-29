import { Warning } from "@/components/ui/icons";
import { cn } from "@/lib/cn";

export const inputClass = (invalid?: boolean, changed?: boolean) =>
  cn(
    "h-12 w-full rounded-xl border-[1.5px] bg-surface px-3.5 text-base focus:border-ink disabled:bg-sand disabled:text-disabled-ink",
    invalid ? "border-chili" : changed ? "border-paprika" : "border-line",
  );

export const textareaClass = (invalid?: boolean) =>
  cn("w-full rounded-xl border-[1.5px] bg-surface px-3.5 py-3 text-base leading-normal focus:border-ink", invalid ? "border-chili" : "border-line");

export function Label({ htmlFor, id, children, className }: { htmlFor?: string; id?: string; children: React.ReactNode; className?: string }) {
  const cls = cn("flex items-center gap-2 text-sm font-semibold", className);
  return htmlFor ? (
    <label htmlFor={htmlFor} id={id} className={cls}>
      {children}
    </label>
  ) : (
    <span id={id} className={cls}>
      {children}
    </span>
  );
}

export function Helper({ id, children, className }: { id?: string; children: React.ReactNode; className?: string }) {
  return (
    <span id={id} className={cn("text-[13px] leading-[1.45] text-ink-3", className)}>
      {children}
    </span>
  );
}

export function FieldError({ id, children }: { id?: string; children?: React.ReactNode }) {
  if (!children) return null;
  return (
    <span id={id} className="flex items-start gap-1.5 text-[13px] font-semibold text-chili">
      <Warning size={16} className="mt-px shrink-0" />
      {children}
    </span>
  );
}

/** Input with a trailing unit (e.g. "seats", "hours"). */
export function SuffixInput({ suffix, invalid, ...props }: React.InputHTMLAttributes<HTMLInputElement> & { suffix: string; invalid?: boolean }) {
  return (
    <div className={cn("flex h-12 items-center overflow-hidden rounded-xl border-[1.5px] bg-surface focus-within:border-ink", invalid ? "border-chili" : "border-line")}>
      <input {...props} className="h-[45px] min-w-0 grow border-0 bg-transparent px-3.5 text-base focus-visible:outline-offset-[-3px]" />
      <span aria-hidden className="px-3.5 text-[15px] font-semibold text-ink-3">
        {suffix}
      </span>
    </div>
  );
}
