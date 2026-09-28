import type { NextConfig } from "next";

// Public base URL of images served by the Go API (see PUBLIC_BASE_URL in the root .env).
const imageBase = new URL(process.env.IMAGE_BASE_URL ?? "http://localhost:8081");

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [new URL("/uploads/**", imageBase)],
    // In development the API runs on localhost, which Next.js 16 refuses to optimize by default.
    // Never enable this in production.
    dangerouslyAllowLocalIP: process.env.NODE_ENV !== "production",
  },
};

export default nextConfig;
