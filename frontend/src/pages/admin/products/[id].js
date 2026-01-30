import { useState, useEffect, useCallback } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Link from 'next/link';
import Image from 'next/image';
import Sidebar from '../../../components/Sidebar';
import TopNavbar from '../../../components/TopNavbar';
import { useNotifications } from '../../../contexts/NotificationContext';

export default function ProductFormPage() {
  const router = useRouter();
  const { id } = router.query;
  const isEdit = id && id !== 'new';
  const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';

  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [notification, setNotification] = useState({ show: false, message: '', type: '' });
  const { notifications, unreadCount, hasUnread, markAsRead, markAllRead, clearAll } = useNotifications();

  const [form, setForm] = useState({
    name: '',
    description: '',
    category: 'Cakes',
    price: '',
    stock: '',
    image_url: '',
    status: 'draft'
  });

  const [tagsInput, setTagsInput] = useState('');
  const [tagsList, setTagsList] = useState([]);
  const [availableTags, setAvailableTags] = useState([]);
  const [tagsOpen, setTagsOpen] = useState(false);

  const [file, setFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState('');
  const [uploading, setUploading] = useState(false);

  const [errors, setErrors] = useState({});

  const showNotification = useCallback((message, type) => {
    setNotification({ show: true, message, type });
    setTimeout(() => setNotification({ show: false, message: '', type: '' }), 5000);
  }, []);

  const fetchProduct = useCallback(async () => {
    if (!isEdit || !id) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/products/${id}`);
      const data = await res.json();
      if (data.product) {
        setForm({
          name: data.product.name || '',
          description: data.product.description || '',
          category: data.product.category || 'Cakes',
          price: data.product.price || '',
          stock: data.product.stock || '',
          image_url: data.product.image_url || '',
          status: data.product.status || 'draft'
        });
        setTagsList(Array.isArray(data.product.tags) ? data.product.tags : []);
        setTagsInput('');
        if (data.product.image_url) {
          setPreviewUrl(data.product.image_url);
        } else {
          setPreviewUrl('');
        }
      }
    } catch (e) {
      showNotification('Failed to load product', 'danger');
    } finally {
      setLoading(false);
    }
  }, [API_BASE, id, isEdit, showNotification]);

  useEffect(() => {
    fetchProduct();
  }, [fetchProduct]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch(`${API_BASE}/api/products/tags`);
        const data = await res.json();
        if (!cancelled && res.ok) {
          const norm = (value) => {
            const v = String(value || '').trim().toLowerCase();
            if (!v) return '';
            return v.length > 32 ? v.slice(0, 32) : v;
          };
          const list = Array.isArray(data.tags) ? data.tags : [];
          const set = new Set(list.map((t) => norm(t)).filter(Boolean));
          setAvailableTags(Array.from(set).sort((a, b) => a.localeCompare(b)));
        }
      } catch {
        if (!cancelled) setAvailableTags([]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [API_BASE]);

  const normalizeTag = (value) => {
    const v = String(value || '').trim().toLowerCase();
    if (!v) return '';
    return v.length > 32 ? v.slice(0, 32) : v;
  };

  const parseTagsFromText = (text) => {
    return String(text || '')
      .split(',')
      .map((t) => normalizeTag(t))
      .filter(Boolean);
  };

  const addTagsFromText = (text) => {
    const incoming = parseTagsFromText(text);
    if (!incoming.length) return;
    setAvailableTags((prev) => {
      const set = new Set(prev.map((t) => normalizeTag(t)));
      incoming.forEach((t) => set.add(t));
      return Array.from(set).sort((a, b) => a.localeCompare(b));
    });
    setTagsList((prev) => {
      const existing = new Set(prev.map((t) => normalizeTag(t)));
      const next = [...prev];
      for (const t of incoming) {
        if (!existing.has(t)) {
          next.push(t);
          existing.add(t);
        }
        if (next.length >= 20) break;
      }
      return next;
    });
  };

  const removeTag = (tag) => {
    const target = normalizeTag(tag);
    setTagsList((prev) => prev.filter((t) => normalizeTag(t) !== target));
  };

  const validate = () => {
    const newErrors = {};
    
    if (!form.name.trim()) newErrors.name = 'Product name is required';
    if (form.name.length > 255) newErrors.name = 'Name must be less than 255 characters';
    if (!form.category) newErrors.category = 'Category is required';
    if (!form.price || parseFloat(form.price) < 0) newErrors.price = 'Valid price is required';
    if (!form.stock || parseInt(form.stock) < 0) newErrors.stock = 'Valid stock quantity is required';
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validate()) {
      showNotification('Please fix the errors', 'danger');
      return;
    }

    setSaving(true);
    try {
      const url = isEdit 
        ? `${API_BASE}/api/products/${id}`
        : `${API_BASE}/api/products`;
      
      const method = isEdit ? 'PUT' : 'POST';

      // Upload new image if a file was selected
      let imageUrl = form.image_url || '';
      if (file) {
        try {
          setUploading(true);
          const formData = new FormData();
          formData.append('file', file);
          formData.append('folder', 'bakeflow/products');
          const up = await fetch(`${API_BASE}/api/uploads/cloudinary`, { method: 'POST', body: formData });
          const upData = await up.json();
          if (!up.ok) throw new Error(upData?.error || 'Upload failed');
          imageUrl = upData.url;
        } catch (err) {
          showNotification(`Image upload failed: ${err.message}`, 'danger');
          setUploading(false);
          setSaving(false);
          return;
        }
        setUploading(false);
      }
      
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...form,
          image_url: imageUrl,
          tags: Array.from(
            new Set([...tagsList, ...parseTagsFromText(tagsInput)].map((t) => normalizeTag(t)).filter(Boolean))
          ).slice(0, 20),
          price: parseFloat(form.price),
          stock: parseInt(form.stock)
        })
      });
      
      const data = await res.json();
      
      if (data.success) {
        showNotification(
          isEdit ? 'Product updated successfully' : 'Product created successfully',
          'success'
        );
        setTimeout(() => router.push('/admin/products'), 1500);
      } else {
        showNotification(data.error || 'Failed to save product', 'danger');
      }
    } catch (e) {
      showNotification('Error saving product', 'danger');
    } finally {
      setSaving(false);
    }
  };

  const handlePublish = async () => {
    setForm({...form, status: 'active'});
    setTimeout(() => {
      document.getElementById('product-form').requestSubmit();
    }, 100);
  };

  const selectedTagSet = new Set(tagsList.map((t) => normalizeTag(t)));
  const availableTagSet = new Set(availableTags.map((t) => normalizeTag(t)));
  const normalizedTagsInput = normalizeTag(tagsInput);
  const tagSuggestions = normalizedTagsInput
    ? availableTags
      .filter((t) => {
        const v = normalizeTag(t);
        return v && !selectedTagSet.has(v) && v.includes(normalizedTagsInput);
      })
      .slice(0, 8)
    : availableTags
      .filter((t) => {
        const v = normalizeTag(t);
        return v && !selectedTagSet.has(v);
      })
      .slice(0, 12);
  const canCreateTag = normalizedTagsInput
    && !selectedTagSet.has(normalizedTagsInput)
    && !availableTagSet.has(normalizedTagsInput);

  return (
    <>
      <Head>
        <title>{isEdit ? 'Edit Product' : 'Add New Product'} - BakeFlow Admin</title>
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet" />
        <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet" />
      </Head>

      <div className="d-flex vh-100 overflow-hidden bg-light">
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
          />

          <div className="flex-grow-1 overflow-auto">
            <div className="container-fluid px-4 py-4">
              
              {/* Notification Toast */}
              {notification.show && (
                <div className={`alert alert-${notification.type} alert-dismissible fade show position-fixed top-0 end-0 m-4`} style={{zIndex: 9999}} role="alert">
                  <strong>{notification.message}</strong>
                  <button type="button" className="btn-close" onClick={() => setNotification({show: false, message: '', type: ''})}></button>
                </div>
              )}

              {/* Header */}
              <div className="mb-4">
                <div className="d-flex align-items-center gap-2 mb-2">
                  <Link href="/admin/products">
                    <button className="btn btn-sm btn-outline-secondary">
                      <i className="bi bi-arrow-left"></i>
                    </button>
                  </Link>
                  <h1 className="h3 fw-bold mb-0">{isEdit ? 'Edit Product' : 'Add New Product'}</h1>
                </div>
                <p className="text-muted mb-0">Fill in the product details below</p>
              </div>

              {loading ? (
                <div className="text-center py-5">
                  <div className="spinner-border text-primary" role="status">
                    <span className="visually-hidden">Loading...</span>
                  </div>
                </div>
              ) : (
                <div className="row">
                  <div className="col-lg-8">
                    <div className="card shadow-sm">
                      <div className="card-body">
                        <form id="product-form" onSubmit={handleSubmit}>
                          {/* Product Name */}
                          <div className="mb-3">
                            <label className="form-label fw-semibold">Product Name *</label>
                            <input
                              type="text"
                              className={`form-control ${errors.name ? 'is-invalid' : ''}`}
                              value={form.name}
                              onChange={(e) => setForm({...form, name: e.target.value})}
                              placeholder="e.g., Chocolate Fudge Cake"
                            />
                            {errors.name && <div className="invalid-feedback">{errors.name}</div>}
                          </div>

                          {/* Description */}
                          <div className="mb-3">
                            <label className="form-label fw-semibold">Description</label>
                            <textarea
                              className="form-control"
                              rows="4"
                              value={form.description}
                              onChange={(e) => setForm({...form, description: e.target.value})}
                              placeholder="Describe your product..."
                            ></textarea>
                          </div>

                          {/* Category */}
                          <div className="mb-3">
                            <label className="form-label fw-semibold">Category *</label>
                            <select
                              className={`form-select ${errors.category ? 'is-invalid' : ''}`}
                              value={form.category}
                              onChange={(e) => setForm({...form, category: e.target.value})}
                            >
                              <option value="Cakes">Cakes</option>
                              <option value="Cupcakes">Cupcakes</option>
                              <option value="Muffins">Muffins</option>
                              <option value="Tarts">Tarts</option>
                              <option value="Cookies">Cookies</option>
                              <option value="Pastries">Pastries</option>
                              <option value="Breads">Breads</option>
                            </select>
                            {errors.category && <div className="invalid-feedback">{errors.category}</div>}
                          </div>

                          <div className="mb-3">
                            <label className="form-label fw-semibold">Tags</label>
                            {tagsList.length > 0 && (
                              <div className="d-flex flex-wrap gap-2 mb-2">
                                {tagsList.map((t) => (
                                  <span key={t} className="badge text-bg-secondary d-inline-flex align-items-center gap-2 px-3 py-2">
                                    <span>{t}</span>
                                    <button
                                      type="button"
                                      className="btn btn-sm btn-link p-0 text-white text-decoration-none"
                                      aria-label={`Remove tag ${t}`}
                                      onClick={() => removeTag(t)}
                                      style={{ lineHeight: 1 }}
                                    >
                                      <i className="bi bi-x-lg"></i>
                                    </button>
                                  </span>
                                ))}
                              </div>
                            )}
                            <div className="position-relative">
                              <input
                                type="text"
                                className="form-control"
                                value={tagsInput}
                                onChange={(e) => setTagsInput(e.target.value)}
                                onFocus={() => setTagsOpen(true)}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter' || e.key === ',') {
                                    e.preventDefault();
                                    addTagsFromText(tagsInput);
                                    setTagsInput('');
                                  } else if (e.key === 'Backspace' && !tagsInput && tagsList.length > 0) {
                                    removeTag(tagsList[tagsList.length - 1]);
                                  }
                                }}
                                onBlur={() => {
                                  if (tagsInput.includes(',')) {
                                    addTagsFromText(tagsInput);
                                    setTagsInput('');
                                  }
                                  setTimeout(() => setTagsOpen(false), 120);
                                }}
                                placeholder="Type a tag and press Enter (or use commas)"
                                autoComplete="off"
                              />
                              {tagsOpen && (canCreateTag || tagSuggestions.length > 0) && (
                                <div className="position-absolute start-0 end-0 mt-1 bg-white border rounded shadow-sm" style={{ zIndex: 1000 }}>
                                  <div className="list-group list-group-flush" style={{ maxHeight: 240, overflowY: 'auto' }}>
                                    {canCreateTag && (
                                      <button
                                        type="button"
                                        className="list-group-item list-group-item-action"
                                        onMouseDown={(e) => {
                                          e.preventDefault();
                                          addTagsFromText(normalizedTagsInput);
                                          setTagsInput('');
                                        }}
                                      >
                                        Add &quot;{normalizedTagsInput}&quot;
                                      </button>
                                    )}
                                    {tagSuggestions.map((t) => (
                                      <button
                                        key={t}
                                        type="button"
                                        className="list-group-item list-group-item-action"
                                        onMouseDown={(e) => {
                                          e.preventDefault();
                                          addTagsFromText(t);
                                          setTagsInput('');
                                        }}
                                      >
                                        {t}
                                      </button>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </div>
                          </div>

                          {/* Price and Stock */}
                          <div className="row">
                            <div className="col-md-6 mb-3">
                              <label className="form-label fw-semibold">Price ($) *</label>
                              <input
                                type="number"
                                step="0.01"
                                className={`form-control ${errors.price ? 'is-invalid' : ''}`}
                                value={form.price}
                                onChange={(e) => setForm({...form, price: e.target.value})}
                                placeholder="0.00"
                              />
                              {errors.price && <div className="invalid-feedback">{errors.price}</div>}
                            </div>

                            <div className="col-md-6 mb-3">
                              <label className="form-label fw-semibold">Stock Quantity *</label>
                              <input
                                type="number"
                                className={`form-control ${errors.stock ? 'is-invalid' : ''}`}
                                value={form.stock}
                                onChange={(e) => setForm({...form, stock: e.target.value})}
                                placeholder="0"
                              />
                              {errors.stock && <div className="invalid-feedback">{errors.stock}</div>}
                            </div>
                          </div>

                          {/* Image URL */}
                          <div className="mb-3">
                            <label className="form-label fw-semibold">Image</label>
                            <input
                              type="file"
                              accept="image/*"
                              className="form-control"
                              onChange={(e) => {
                                const f = e.target.files?.[0] || null;
                                setFile(f);
                                if (f) setPreviewUrl(URL.createObjectURL(f));
                                else setPreviewUrl(form.image_url || '');
                              }}
                            />
                            {previewUrl && (
                              <div className="mt-2">
                                <Image
                                  src={previewUrl}
                                  alt="Preview"
                                  width={200}
                                  height={200}
                                  style={{ maxWidth: 200, maxHeight: 200, borderRadius: 8, height: 'auto' }}
                                  unoptimized
                                />
                              </div>
                            )}
                          </div>
                        </form>
                      </div>
                    </div>
                  </div>

                  {/* Sidebar */}
                  <div className="col-lg-4">
                    <div className="card shadow-sm mb-3">
                      <div className="card-body">
                        <h6 className="card-title fw-semibold mb-3">Status</h6>
                        <select
                          className="form-select mb-3"
                          value={form.status}
                          onChange={(e) => setForm({...form, status: e.target.value})}
                        >
                          <option value="draft">Draft</option>
                          <option value="active">Active (Published)</option>
                          <option value="inactive">Inactive</option>
                          <option value="archived">Archived</option>
                        </select>

                        <div className="d-grid gap-2">
                          <button 
                            type="button"
                            className="btn btn-primary"
                            onClick={handleSubmit}
                            disabled={saving}
                          >
                            {saving ? (
                              <>
                                <span className="spinner-border spinner-border-sm me-2"></span>
                                Saving...
                              </>
                            ) : (
                              <>
                                <i className="bi bi-save me-2"></i>
                                {isEdit ? 'Update Product' : 'Create Product'}
                              </>
                            )}
                          </button>

                          {form.status === 'draft' && (
                            <button 
                              type="button"
                              className="btn btn-success"
                              onClick={handlePublish}
                              disabled={saving}
                            >
                              <i className="bi bi-check-circle me-2"></i>
                              Save & Publish
                            </button>
                          )}

                          <Link href="/admin/products">
                            <button type="button" className="btn btn-outline-secondary w-100">
                              Cancel
                            </button>
                          </Link>
                        </div>
                      </div>
                    </div>

                    {/* Stock Alert */}
                    {form.stock && parseInt(form.stock) < 10 && (
                      <div className="alert alert-warning">
                        <i className="bi bi-exclamation-triangle me-2"></i>
                        <strong>Low Stock!</strong> Consider restocking soon.
                      </div>
                    )}

                    {form.stock && parseInt(form.stock) === 0 && (
                      <div className="alert alert-danger">
                        <i className="bi bi-x-circle me-2"></i>
                        <strong>Out of Stock!</strong> This product is unavailable.
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
