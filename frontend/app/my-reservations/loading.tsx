import { Skeleton } from "@/components/ui/states";

export default function Loading() {
  return (
    <main aria-busy="true" aria-label="Loading reservations" className="mx-auto flex w-full max-w-[1120px] flex-col gap-7 px-4 pt-14 sm:px-8 xl:px-0">
      <Skeleton className="h-12 w-72" />
      <Skeleton className="h-[52px]" />
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="h-[200px] rounded-[22px]" />
      ))}
    </main>
  );
}
