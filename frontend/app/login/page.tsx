import { redirect } from "next/navigation";

import { auth } from "@/auth";
import { login } from "@/features/auth/actions";

export default async function LoginPage({ searchParams }: PageProps<"/login">) {
  const { callbackUrl } = await searchParams;
  const target = typeof callbackUrl === "string" ? callbackUrl : "/";

  const session = await auth();
  if (session?.user && !session.error) redirect(target);

  return (
    <main className="flex flex-1 items-center justify-center p-6">
      <div className="w-full max-w-sm rounded-xl border p-8 text-center shadow-sm">
        <h1 className="text-2xl font-semibold">Welcome to TableTalk</h1>
        <p className="mt-2 text-sm text-zinc-500">Sign in to book tables, write reviews and manage your restaurants.</p>
        {session?.error && <p className="mt-4 text-sm text-red-600">Your session expired. Please sign in again.</p>}
        <form action={login} className="mt-6">
          <input type="hidden" name="callbackUrl" value={target} />
          <button type="submit" className="w-full rounded-md bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-700">
            Sign in with Keycloak
          </button>
        </form>
      </div>
    </main>
  );
}
