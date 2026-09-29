import { Info, Lock, Warning } from "@/components/ui/icons";
import { cn } from "@/lib/cn";

type Tone = "error" | "info" | "success" | "warning";

const tones: Record<Tone, string> = {
  error: "bg-chili-soft text-chili font-semibold",
  info: "bg-sand text-body",
  success: "bg-herb-soft text-herb-ink",
  warning: "bg-saffron-soft text-[#5E3F0B]",
};

/** Inline message block used for validation summaries, policy notes and API errors. */
export function Alert({ tone = "info", title, children, className, id }: { tone?: Tone; title?: string; children: React.ReactNode; className?: string; id?: string }) {
  const Icon = tone === "error" ? Warning : tone === "warning" ? Lock : Info;
  return (
    <div id={id} role={tone === "error" ? "alert" : undefined} className={cn("flex items-start gap-2.5 rounded-xl px-3.5 py-3 text-sm leading-[1.45]", tones[tone], className)}>
      <Icon size={18} className="mt-px shrink-0" />
      <div className="flex flex-col gap-1">
        {title && <strong className="text-base">{title}</strong>}
        <span>{children}</span>
      </div>
    </div>
  );
}
