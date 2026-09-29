"use client";

import { ErrorState } from "@/components/ui/states";

// Unexpected failures while rendering a page. Messages never include raw backend text.
export default function Error({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <main className="mx-auto flex w-full max-w-[720px] flex-col px-4 py-16 sm:py-24">
      <ErrorState title="Something went wrong" onRetry={reset} />
    </main>
  );
}
