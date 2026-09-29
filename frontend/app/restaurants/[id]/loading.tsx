import { Skeleton } from "@/components/ui/states";

export default function Loading() {
  return (
    <main aria-busy="true" aria-label="Loading restaurant" className="mx-auto flex w-full max-w-[1440px] flex-col gap-7 px-4 pt-6 sm:px-8 lg:px-20">
      <Skeleton className="h-11 w-40" />
      <Skeleton className="h-[240px] rounded-3xl sm:h-[460px]" />
      <div className="grid grid-cols-1 gap-10 lg:grid-cols-[minmax(0,1fr)_420px] lg:gap-16">
        <div className="flex flex-col gap-4">
          <Skeleton className="h-12 w-2/3" />
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-24" />
        </div>
        <Skeleton className="h-[520px] rounded-3xl" />
      </div>
    </main>
  );
}
