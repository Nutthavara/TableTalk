import "next-auth";
import "next-auth/jwt";

declare module "next-auth" {
  interface Session {
    error?: "RefreshTokenError";
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    access_token?: string;
    /** Access token expiry, seconds since epoch. */
    expires_at?: number;
    refresh_token?: string;
    id_token?: string;
    error?: "RefreshTokenError";
  }
}
