import Link from "next/link";

import { ChevronLeft } from "@/components/ui/icons";

export function BackLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link href={href} className="flex h-11 items-center gap-1.5 self-start text-[15px] font-semibold text-ink-2 no-underline hover:text-ink">
      <ChevronLeft size={18} />
      {children}
    </Link>
  );
}
