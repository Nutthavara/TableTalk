# TableTalk frontend

Next.js 16 (App Router) with Auth.js v5 (Keycloak) and Tailwind CSS v4.

See the [root README](../README.md) for setup, environment variables and how the frontend talks to the Go API.

```bash
cp .env.example .env.local   # then set AUTH_SECRET
npm install
npm run dev                  # http://localhost:3000
npm run lint
npm run build
```

The Go API is called only from the server (Server Components and Server Actions, via `lib/api.ts`). Access tokens never reach browser JavaScript.
