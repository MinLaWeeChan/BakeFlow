import { useEffect, useState } from 'react';
import Head from 'next/head';
import Link from 'next/link';
import Sidebar from '../../../components/Sidebar';
import TopNavbar from '../../../components/TopNavbar';
import { useNotifications } from '../../../contexts/NotificationContext';

export default function OrdersArchivePage() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const { notifications, unreadCount, hasUnread, markAsRead, markAllRead, clearAll } = useNotifications();

  useEffect(() => {
    const fetchOrders = async () => {
      try {
        setError(null);
        const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'https://bakeflow.onrender.com';
        const res = await fetch(`${API_BASE}/api/admin/orders`);
        const data = await res.json();
        if (data.error) {
          setError(data.details || data.error);
          setOrders([]);
        } else {
          const delivered = (data.orders || []).filter(o => o.status === 'delivered');
          setOrders(delivered);
        }
      } catch (e) {
        console.error(e);
        setError('Cannot connect to backend.');
      } finally {
        setLoading(false);
      }
    };
    fetchOrders();
  }, []);

  return (
    <>
      <Head>
        <title>BakeFlow Admin - Orders Archive</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet" />
        <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet" />
      </Head>
      <div className="d-flex vh-100 overflow-hidden bg-light">
        <Sidebar open={sidebarOpen} toggle={() => setSidebarOpen(o => !o)} />
        <div className="flex-grow-1 d-flex flex-column overflow-hidden">
          <TopNavbar 
            toggleSidebar={() => setSidebarOpen(o => !o)}
            notifications={notifications}
            unreadCount={unreadCount}
            hasUnread={hasUnread}
            onMarkAllRead={markAllRead}
            onClearAll={clearAll}
            onNotificationClick={(id) => markAsRead(id)}
          />
          <div className="flex-grow-1 overflow-auto">
            <div className="container-fluid px-4 py-4">
              <div className="mb-4 d-flex align-items-center justify-content-between flex-wrap gap-3">
                <div>
                  <h1 className="h3 fw-bold mb-1">Orders Archive</h1>
                  <p className="text-muted mb-0">Delivered orders</p>
                </div>
                <Link href="/admin/orders" className="btn btn-outline-secondary">
                  <i className="bi bi-arrow-left me-2" />Back to Orders
                </Link>
              </div>

              {error && <div className="alert alert-danger">{error}</div>}
              {loading && <div className="text-center py-5"><div className="spinner-border text-primary" role="status" /><p className="mt-3 text-muted">Loading archived orders...</p></div>}

              {!loading && orders.length === 0 && !error && (
                <div className="card border-0 shadow-sm"><div className="card-body text-center py-5"><i className="bi bi-archive fs-1 text-muted mb-3"/><h4 className="text-muted">No archived orders</h4><p className="text-secondary">Delivered orders will appear here.</p></div></div>
              )}

              <div className="row g-4">
                {orders.map(order => (
                  <div key={order.id} className="col-12 col-xl-6">
                    <div className="card border-0 shadow-sm h-100">
                      <div className="card-header bg-white border-bottom py-3 d-flex justify-content-between align-items-center">
                        <div>
                          <h5 className="mb-1 fw-bold">Order #{order.id}</h5>
                          <small className="text-muted"><i className="bi bi-clock me-1"></i>{new Date(order.created_at).toLocaleString()}</small>
                        </div>
                        <span className="badge bg-success px-3 py-2">DELIVERED</span>
                      </div>
                      <div className="card-body p-4">
                        <div className="mb-2">
                          <strong>Customer:</strong> {order.customer_name}
                        </div>
                        {Array.isArray(order.items) && order.items.map((item, idx) => (
                          <div key={idx} className="d-flex justify-content-between align-items-center py-2 border-bottom">
                            <div className="flex-grow-1">
                              <div className="fw-semibold">{item.product}</div>
                              <small className="text-muted">${item.price.toFixed(2)} × {item.quantity}</small>
                            </div>
                            <div className="fw-bold">${(item.price * item.quantity).toFixed(2)}</div>
                          </div>
                        ))}
                        <div className="d-flex justify-content-between mt-3">
                          <span className="text-muted">Total</span>
                          <span className="fw-bold">${(order.total_amount || 0).toFixed(2)}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
