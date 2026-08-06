import type { NextConfig } from "next";

const apiInternalURL = process.env.API_INTERNAL_URL || "http://localhost:8080";
const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiInternalURL}/api/:path*` }];
  },
};

export default nextConfig;
