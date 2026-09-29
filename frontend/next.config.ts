import type { NextConfig } from "next";

// Creating a restaurant sends up to 10 photos of 5 MB each through a Server Action.
const uploadLimit = "52mb";

const nextConfig: NextConfig = {
  experimental: {
    serverActions: { bodySizeLimit: uploadLimit },
    // proxy.ts buffers request bodies; keep uploads intact.
    proxyClientMaxBodySize: uploadLimit,
  },
};

export default nextConfig;
