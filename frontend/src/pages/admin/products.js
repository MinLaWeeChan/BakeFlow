import { useState, useEffect } from 'react';
import Head from 'next/head';
import Link from 'next/link';
import Sidebar from '../../components/Sidebar';
import TopNavbar from '../../components/TopNavbar';
import { useNotifications } from '../../contexts/NotificationContext';
import { useTranslation } from '../../utils/i18n';
import { formatCurrency } from '../../utils/formatCurrency';

export default function ProductsPage() {
  const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [filter, setFilter] = useState({ category: '', status: '', search: '' });
  const [notification, setNotification] = useState({ show: false, message: '', type: '' });
  const { notifications, unreadCount, hasUnread, markAsRead, markAllRead, clearAll } = useNotifications();
  const { t } = useTranslation();

  const fetchProducts = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (filter.category) params.append('category', filter.category);
      if (filter.status) params.append('status', filter.status);
      if (filter.search) params.append('search', filter.search);
      const res = await fetch(`${API_BASE}/api/products?${params.toString()}`);
      if (!res.ok) throw new Error(`API error ${res.status}`);
      const data = await res.json();
      setProducts(data.products || []);
      setError(null);
    } catch (e) {
      console.error(e);
      setError('Failed to load products');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchProducts(); }, [filter]);

  const showNotification = (message, type) => {
    setNotification({ show: true, message, type });
    setTimeout(() => setNotification({ show: false, message: '', type: '' }), 5000);
  };

  const deleteProduct = async (id) => {
    if (!confirm('Are you sure you want to archive this product?')) return;
    try {
      const res = await fetch(`${API_BASE}/api/products/${id}`, { method: 'DELETE' });
      if (!res.ok) throw new Error(`API error ${res.status}`);
      const data = await res.json();
      if (data.success) { showNotification('Product archived successfully', 'success'); fetchProducts(); }
      else showNotification('Failed to archive product', 'danger');
    } catch {
      showNotification('Error archiving product', 'danger');
    }
  };

  const updateStatus = async (id, newStatus) => {
    try {
      const res = await fetch(`${API_BASE}/api/products/${id}/status`, {
        method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ status: newStatus })
      });
      if (!res.ok) throw new Error(`API error ${res.status}`);
      const data = await res.json();
      if (data.success) { showNotification(`Product ${newStatus === 'active' ? 'published' : 'updated'}`, 'success'); fetchProducts(); }
      else showNotification('Failed to update status', 'danger');
    } catch {
      showNotification('Error updating status', 'danger');
    }
  };

  const getStatusBadge = (status) => {
    const m = {
      draft: 'bg-warning-subtle text-warning border border-warning-subtle',
      active: 'bg-success-subtle text-success border border-success-subtle',
      inactive: 'bg-secondary-subtle text-secondary border border-secondary-subtle',
      archived: 'bg-danger-subtle text-danger border border-danger-subtle'
    };
    return m[status] || 'bg-secondary-subtle text-secondary border border-secondary-subtle';
  };

  return (
    <>
      <Head>
        <title>Products - BakeFlow Admin</title>
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet" />
        <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet" />
      </Head>

      <div className="d-flex vh-100 overflow-hidden bg-light-subtle">
        <Sidebar open={sidebarOpen} toggle={() => setSidebarOpen(!sidebarOpen)} />

        <div className="flex-grow-1 d-flex flex-column overflow-hidden">
          <TopNavbar
            toggleSidebar={() => setSidebarOpen(!sidebarOpen)}
            notifications={notifications}
            unreadCount={unreadCount}
            hasUnread={hasUnread}
            onMarkAllRead={markAllRead}
            onClearAll={clearAll}
            onNotificationClick={(id) => markAsRead(id)}
            pageTitle={t('productsTitle')}
            pageSubtitle={t('productInventory')}
          />

          <div className="flex-grow-1 overflow-auto">
            <div className="container-xl py-4">

              {notification.show && (
                <div className={`alert alert-${notification.type} alert-dismissible fade show position-fixed top-0 end-0 m-4`} style={{ zIndex: 9999 }} role="alert">
                  <strong>{notification.message}</strong>
                  <button type="button" className="btn-close" onClick={() => setNotification({ show: false, message: '', type: '' })}></button>
                </div>
              )}

              {/* Page Header */}
              <div className="d-flex align-items-center justify-content-between mb-4">
                <div>
                  <h1 className="h3 mb-1 fw-semibold">Products</h1>
                  <p className="text-muted mb-0 small">Manage your catalog, visibility, and inventory levels</p>
                </div>
                <div className="d-flex align-items-center gap-2">
                  <button className="btn btn-outline-secondary d-none d-md-inline-flex">
                    <i className="bi bi-download me-2"></i>Export CSV
                  </button>
                  <Link href="/admin/products/new">
                    <button className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>Add Product</button>
                  </Link>
                </div>
              </div>

              {/* KPI Summary Cards */}
              <div className="row g-4 mb-4">
                {[
                  { icon: 'bi-box', label: t('totalProducts'), value: products.length },
                  { icon: 'bi-check-circle', label: t('activeProducts'), value: products.filter(p => p.status === 'active').length },
                  { icon: 'bi-exclamation-triangle', label: t('lowStockItems'), value: products.filter(p => p.low_stock || p.out_of_stock).length },
                  { icon: 'bi-eye', label: t('totalViews'), value: products.reduce((s, p) => s + (p.views || 0), 0) }
                ].map((c, idx) => (
                  <div className="col-lg-3 col-md-6" key={idx}>
                    <div className="card shadow-sm border-light rounded-3 h-100">
                      <div className="card-body p-4">
                        <div className="d-flex align-items-center gap-3">
                          <div className="rounded-3 d-inline-flex align-items-center justify-content-center bg-light text-secondary" style={{ width: 48, height: 48 }}>
                            <i className={`bi ${c.icon}`}></i>
                          </div>
                          <div>
                            <div className="fs-2 fw-semibold lh-1">{c.value}</div>
                            <div className="text-muted text-uppercase small mt-1">{c.label}</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Filters & Actions Bar */}
              <div className="card shadow-sm border-light rounded-3 mb-3">
                <div className="card-body p-3 p-md-4">
                  <div className="bg-light p-3 rounded-3 mb-2">
                    <div className="row g-3">
                      {/* LEFT: Search & Filters */}
                      <div className="col-12 col-lg-8">
                        <div className="row g-3">
                          <div className="col-12 col-md-6">
                            <div className="input-group input-group-lg">
                              <span className="input-group-text bg-white border-end-0"><i className="bi bi-search"></i></span>
                              <input
                                type="text"
                                className="form-control border-start-0"
                                placeholder={t('searchProducts')}
                                value={filter.search}
                                onChange={(e) => setFilter({ ...filter, search: e.target.value })}
                              />
                            </div>
                          </div>
                          <div className="col-6 col-md-3">
                            <select
                              className="form-select form-select-lg"
                              value={filter.category}
                              onChange={(e) => setFilter({ ...filter, category: e.target.value })}
                            >
                              <option value="">{t('allCategories')}</option>
                              <option value="Cakes">Cakes</option>
                              <option value="Cupcakes">Cupcakes</option>
                              <option value="Muffins">Muffins</option>
                              <option value="Tarts">Tarts</option>
                              <option value="Cookies">Cookies</option>
                            </select>
                          </div>
                          <div className="col-6 col-md-3">
                            <select
                              className="form-select form-select-lg"
                              value={filter.status}
                              onChange={(e) => setFilter({ ...filter, status: e.target.value })}
                            >
                              <option value="">{t('allStatus')}</option>
                              <option value="draft">Draft</option>
                              <option value="active">{t('activeLabel')}</option>
                              <option value="inactive">Inactive</option>
                              <option value="archived">{t('hiddenLabel')}</option>
                            </select>
                          </div>
                        </div>
                      </div>

                    </div>
                  </div>

                  {/* Products Table */}
                  {loading ? (
                    <div className="text-center py-5">
                      <div className="spinner-border text-primary" role="status">
                        <span className="visually-hidden">{t('loadingOrders')}</span>
                      </div>
                    </div>
                  ) : error ? (
                    <div className="alert alert-danger mb-0">{error}</div>
                  ) : products.length === 0 ? (
                    <div className="text-center py-5">
                      <i className="bi bi-box-seam fs-1 text-muted mb-3 d-block"></i>
                      <h5 className="text-muted">{t('noProductsYet')}</h5>
                      <p className="text-muted">{t('createYourFirstProduct')}</p>
                      <Link href="/admin/products/new">
                        <button className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>{t('createProduct')}</button>
                      </Link>
                    </div>
                  ) : (
                    <div className="table-responsive">
                      <table className="table table-hover align-middle mb-0">
                        <thead className="table-light sticky-top" style={{ top: 0, zIndex: 1 }}>
                          <tr>
                            <th className="text-uppercase text-muted small fw-semibold py-3">{t('productColumn')}</th>
                            <th className="text-uppercase text-muted small fw-semibold py-3">{t('priceColumn')}</th>
                            <th className="text-uppercase text-muted small fw-semibold py-3">{t('stockColumn')}</th>
                            <th className="text-uppercase text-muted small fw-semibold py-3">{t('performanceColumn')}</th>
                            <th className="text-uppercase text-muted small fw-semibold py-3">{t('statusColumn')}</th>
                            <th className="text-uppercase text-muted small fw-semibold text-end py-3">{t('actionsColumn')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {products.map((product) => (
                            <tr key={product.id}>
                              <td>
                                <div className="d-flex align-items-center gap-3">
                                  <div className="rounded-3 overflow-hidden bg-light" style={{ width: 64, height: 64 }}>
                                    {product.image_url ? (
                                      <img src={product.image_url} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                                    ) : (
                                      <div className="d-flex align-items-center justify-content-center h-100">
                                        <i className="bi bi-image text-muted"></i>
                                      </div>
                                    )}
                                  </div>
                                  <div>
                                    <div className="fw-semibold mb-1 d-flex align-items-center gap-2">
                                      {product.name}
                                      <span className="badge rounded-pill bg-secondary-subtle text-secondary border border-secondary-subtle">{product.category}</span>
                                    </div>
                                    <div className="text-muted small d-none d-md-block" style={{ maxWidth: 420 }}>
                                      {product.description?.substring(0, 80)}{product.description && product.description.length > 80 ? '…' : ''}
                                    </div>
                                  </div>
                                </div>
                              </td>
                              <td className="fw-semibold">{formatCurrency(product.price)}</td>
                              <td>
                                <div className="d-flex flex-column">
                                  <span className="fw-semibold">{product.stock}</span>
                                  <span className="mt-2">
                                    {product.out_of_stock ? (
                                      <span className="badge rounded-pill bg-danger-subtle text-danger border border-danger-subtle d-inline-flex align-items-center gap-2"><span className="rounded-circle bg-danger" style={{ width: 8, height: 8, opacity: .2 }}></span> {t('outOfStock')}</span>
                                    ) : product.low_stock ? (
                                      <span className="badge rounded-pill bg-warning-subtle text-warning border border-warning-subtle d-inline-flex align-items-center gap-2"><span className="rounded-circle bg-warning" style={{ width: 8, height: 8, opacity: .2 }}></span> {t('lowStock')}</span>
                                    ) : (
                                      <span className="badge rounded-pill bg-success-subtle text-success border border-success-subtle d-inline-flex align-items-center gap-2"><span className="rounded-circle bg-success" style={{ width: 8, height: 8, opacity: .2 }}></span> {t('goodStock')}</span>
                                    )}
                                  </span>
                                </div>
                              </td>
                              <td>
                                <div className="d-flex align-items-center gap-2">
                                  <i className="bi bi-eye"></i>
                                  <span className="fw-semibold">{product.views || 0}</span>
                                  <span className="text-muted small">0% today</span>
                                </div>
                              </td>
                              <td>
                                <span className={`badge rounded-pill ${getStatusBadge(product.status)}`}>
                                  {product.status === 'active' ? t('activeLabel') : product.status === 'draft' ? 'Draft' : t('hiddenLabel')}
                                </span>
                              </td>
                              <td>
                                <div className="d-flex justify-content-end gap-2">
                                  <Link href={`/admin/products/${product.id}`}>
                                    <button className="btn btn-light btn-sm rounded-circle border" style={{ width: 36, height: 36 }} title={t('editTitle')} data-bs-toggle="tooltip">
                                      <i className="bi bi-pencil"></i>
                                    </button>
                                  </Link>
                                  {product.status === 'active' ? (
                                    <button
                                      className="btn btn-light btn-sm rounded-circle border text-warning"
                                      style={{ width: 36, height: 36 }}
                                      onClick={() => updateStatus(product.id, 'inactive')}
                                      title={t('hideTitle')}
                                      data-bs-toggle="tooltip"
                                    >
                                      <i className="bi bi-eye-slash"></i>
                                    </button>
                                  ) : (
                                    <button
                                      className="btn btn-light btn-sm rounded-circle border text-warning"
                                      style={{ width: 36, height: 36 }}
                                      onClick={() => updateStatus(product.id, 'active')}
                                      title={t('showTitle')}
                                      data-bs-toggle="tooltip"
                                    >
                                      <i className="bi bi-eye"></i>
                                    </button>
                                  )}
                                  <button
                                    className="btn btn-light btn-sm rounded-circle border text-danger"
                                    style={{ width: 36, height: 36 }}
                                    onClick={() => deleteProduct(product.id)}
                                    title={t('deleteTitle')}
                                    data-bs-toggle="tooltip"
                                  >
                                    <i className="bi bi-trash"></i>
                                  </button>
                                </div>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

