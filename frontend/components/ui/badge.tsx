import { Store } from "@/components/ui/icons";
import { cn } from "@/lib/cn";

type Tone = "owner" | "today" | "mine" | "neutral" | "confirmed" | "completed" | "cancelled" | "changed";

const tones: Record<Tone, string> = {
  owner: "bg-herb text-white",
  today: "bg-saffron-soft text-saffron-ink",
  mine: "bg-peach text-peach-ink",
  changed: "bg-peach text-peach-ink",
  neutral: "bg-sand text-ink-3",
  confirmed: "bg-herb-soft text-herb-ink",
  completed: "bg-sand-2 text-ink-2",
  cancelled: "bg-chili-soft text-chili",
};

export function Badge({ tone, children, className }: { tone: Tone; children: React.ReactNode; className?: string }) {
  return (
    <span className={cn("inline-flex h-[26px] shrink-0 items-center gap-1.5 rounded-full px-2.5 text-[13px] font-bold whitespace-nowrap", tones[tone], className)}>
      {children}
    </span>
  );
}

export function OwnerBadge({ large = false }: { large?: boolean }) {
  return (
    <Badge tone="owner" className={large ? "h-[30px] px-3 text-sm" : undefined}>
      <Store size={large ? 15 : 14} strokeWidth={2} />
      Owner
    </Badge>
  );
}
