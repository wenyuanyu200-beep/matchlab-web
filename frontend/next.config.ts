import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";
    if (!/^https?:\/\//.test(apiBase)) return [];
    return [{ source: "/api-proxy/:path*", destination: `${apiBase.replace(/\/$/, "")}/:path*` }];
  },
};

export default nextConfig;
