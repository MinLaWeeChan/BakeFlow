/** @type {import('next').NextConfig} */
const nextConfig = {
  /* config options here */
  reactStrictMode: true,
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'images.unsplash.com' },
      { protocol: 'https', hostname: 'res.cloudinary.com' },
      { protocol: 'http', hostname: 'localhost' },
      { protocol: 'https', hostname: 'localhost' },
    ],
  },

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
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*'
      },
      {
        source: '/qr_codes/:path*',
        destination: 'http://localhost:8080/qr_codes/:path*'
      },
      {
        source: '/uploads/:path*',
        destination: 'http://localhost:8080/uploads/:path*'
      },
      {
        source: '/webhook',
        destination: 'http://localhost:8080/webhook'
      },
      {
        source: '/promotions/:path*',
        destination: 'http://localhost:8080/promotions/:path*'
      },
      {
        source: '/checkout',
        destination: 'http://localhost:8080/checkout'
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
