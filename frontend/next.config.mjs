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
    const backendBase = (process.env.NEXT_PUBLIC_API_BASE || 'https://bakeflow.onrender.com').replace(/\/+$/, '');
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
        destination: `${backendBase}/api/:path*`
      },
      {
        source: '/qr_codes/:path*',
        destination: `${backendBase}/qr_codes/:path*`
      },
      {
        source: '/uploads/:path*',
        destination: `${backendBase}/uploads/:path*`
      },
      {
        source: '/webhook',
        destination: `${backendBase}/webhook`
      },
      {
        source: '/promotions/:path*',
        destination: `${backendBase}/promotions/:path*`
      },
      {
        source: '/checkout',
        destination: `${backendBase}/checkout`
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
