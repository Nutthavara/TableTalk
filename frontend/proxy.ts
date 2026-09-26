// Runs the Auth.js session check on navigations so tokens close to expiry are refreshed and the
// updated cookie is persisted. This is not an authorization boundary: the Go API enforces access.
export { auth as proxy } from "@/auth";

export const config = {
  matcher: ["/((?!api/auth|_next/static|_next/image|favicon.ico|.*\\.(?:png|jpg|jpeg|svg|webp|ico)$).*)"],
};
