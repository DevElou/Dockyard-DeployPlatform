import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Génère un serveur autonome (node server.js) — nécessaire pour Docker
  output: "standalone",
};

export default nextConfig;
