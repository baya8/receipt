import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  output: 'standalone',
  // 開発PCのIPはDHCPで最後のオクテットが変わるため、サブネット単位でワイルドカード許可する
  allowedDevOrigins: [
    '192.168.50.*',
    ...(process.env.ALLOWED_DEV_ORIGINS ? process.env.ALLOWED_DEV_ORIGINS.split(',') : [])
  ],
};

export default nextConfig;
