import Link from "next/link";

import { logout } from "@/features/auth/actions";
import { api, ApiError } from "@/lib/api";
import type { Me } from "@/types/api";

// Temporary Phase 2 page to verify the Next.js → Keycloak → Go API flow; replaced by the restaurant list.
export default async function Home() {
  let me: Me | null = null;
  let error: string | null = null;
  try {
    me = await api<Me>("/api/me");
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 401) error = err instanceof Error ? err.message : String(err);
  }

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-4 p-8">
      <h1 className="text-2xl font-semibold">TableTalk</h1>
      {me ? (
        <>
          <p>
            Signed in as <strong>{me.display_name}</strong> ({me.username})
          </p>
          <pre className="rounded bg-zinc-100 p-4 text-sm dark:bg-zinc-900">{JSON.stringify(me, null, 2)}</pre>
          <form action={logout}>
            <button className="rounded-md border px-4 py-2">Sign out</button>
          </form>
        </>
      ) : (
        <Link href="/login" className="w-fit rounded-md bg-zinc-900 px-4 py-2 text-white">
          Sign in
        </Link>
      )}
      {error && <p className="text-red-600">{error}</p>}
    </main>
  );
}
