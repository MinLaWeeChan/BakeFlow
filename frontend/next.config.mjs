/** @type {import('next').NextConfig} */
const nextConfig = {
  /* config options here */
  reactStrictMode: true,
  
  // Serve static HTML files from public folder
  async rewrites() {
    return [
      {
        source: '/order-form.html',
        destination: '/order-form.html',
      },
      {
        source: '/saved-orders.html',
        destination: '/saved-orders.html',
      },
      {
        source: '/order-details.html',
        destination: '/order-details.html',
      },
      {
        source: '/css/:path*',
        destination: '/css/:path*',
      },
      {
        source: '/js/:path*',
        destination: '/js/:path*',
      },
    ];
  },
};

export default nextConfig;
