import { useState, useEffect } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';

export default function AdminLoginPage() {
  const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';
  const router = useRouter();

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (typeof window === 'undefined') return;
    let cancelled = false;
    const check = async () => {
      let tok = '';
      try {
        tok = localStorage.getItem('bakeflow_admin_token') || '';
      } catch {
        tok = '';
      }
      if (!tok) return;
      try {
        const res = await fetch(`${API_BASE}/api/admin/me`, {
          headers: { Authorization: `Bearer ${tok}` },
        });
        if (!cancelled && res.ok) {
          const q = router.query || {};
          const raw = typeof q.redirect === 'string' ? q.redirect : '/admin';
          const target = raw && raw.startsWith('/') ? raw : '/admin';
          router.replace(target);
        }
      } catch {
      }
    };
    check();
    return () => {
      cancelled = true;
    };
  }, [API_BASE, router]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!identifier || !password) {
      setError('Please enter both email/username and password.');
      return;
    }
    setError('');
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/admin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ identifier, password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || !data.success || !data.token) {
        const msg =
          data.error ||
          data.message ||
          'Invalid credentials. Please check and try again.';
        setError(msg);
        return;
      }
      if (typeof window !== 'undefined') {
        try {
          localStorage.setItem('bakeflow_admin_token', data.token);
        } catch {
        }
      }
      const q = router.query || {};
      const raw = typeof q.redirect === 'string' ? q.redirect : '/admin';
      const target = raw && raw.startsWith('/') ? raw : '/admin';
      router.replace(target);
    } catch {
      setError('Cannot connect to backend. Make sure Go server is running.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Head>
        <title>BakeFlow Admin - Login</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link
          href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"
          rel="stylesheet"
        />
        <link
          href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css"
          rel="stylesheet"
        />
        <script
          src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"
          defer
        ></script>
      </Head>
      <div className="d-flex align-items-center justify-content-center vh-100 bg-light">
        <div className="card shadow-sm border-0" style={{ maxWidth: 420, width: '100%' }}>
          <div className="card-body p-4 p-md-5">
            <div className="text-center mb-4">
              <div className="mb-2">
                <span className="badge bg-warning text-dark px-3 py-2">
                  BakeFlow Admin
                </span>
              </div>
              <h1 className="h4 mb-1">Sign in</h1>
              <p className="text-muted small mb-0">
                Manage orders, products, and promotions.
              </p>
            </div>
            {error && (
              <div className="alert alert-danger py-2 small" role="alert">
                {error}
              </div>
            )}
            <form onSubmit={handleSubmit} className="mt-3">
              <div className="mb-3">
                <label className="form-label fw-semibold">
                  Email or username
                </label>
                <input
                  type="text"
                  className="form-control"
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  autoComplete="username"
                  autoFocus
                />
              </div>
              <div className="mb-3">
                <label className="form-label fw-semibold">Password</label>
                <input
                  type="password"
                  className="form-control"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                />
              </div>
              <button
                type="submit"
                className="btn btn-primary w-100 d-flex align-items-center justify-content-center"
                disabled={loading}
              >
                {loading && (
                  <span
                    className="spinner-border spinner-border-sm me-2"
                    role="status"
                    aria-hidden="true"
                  ></span>
                )}
                {loading ? 'Signing in…' : 'Sign in'}
              </button>
            </form>
          </div>
        </div>
      </div>
    </>
  );
}

