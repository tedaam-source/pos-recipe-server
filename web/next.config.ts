import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === 'production';

const nextConfig: NextConfig = {
  output: isProd ? 'export' : undefined,
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: '/admin/:path*',
        destination: 'http://localhost:8081/admin/:path*',
      },
      {
        source: '/health',
        destination: 'http://localhost:8081/health',
      },
    ];
  },
};

export default nextConfig;
