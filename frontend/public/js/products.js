/**
 * BakeFlow - Products Module
 * Product loading and rendering with GrabFood/Uber Eats style UX
 */

window.products = [];
window.stockStatus = {}; // Real-time stock status cache
window.productFilters = {
    search: '',
    category: 'all'
};
window.productsLoadError = '';

// Category definitions with icons
const CATEGORIES = [
    { id: 'all', name: 'All', icon: 'grid-2x2' },
    { id: 'cakes', name: 'Cakes', icon: 'cake' },
    { id: 'pastries', name: 'Pastries', icon: 'croissant' },
    { id: 'bread', name: 'Bread', icon: 'sandwich' },
    { id: 'drinks', name: 'Drinks', icon: 'coffee' },
    { id: 'other', name: 'Other', icon: 'cookie' }
];

// Map product names to categories (basic heuristic)
function getCategoryForProduct(product) {
    const name = (product.name || '').toLowerCase();
    const desc = (product.description || '').toLowerCase();
    const combined = name + ' ' + desc;
    
    if (combined.includes('cake') || combined.includes('cheesecake') || combined.includes('tiramisu')) {
        return 'cakes';
    }
    if (combined.includes('pastry') || combined.includes('croissant') || combined.includes('danish') || 
        combined.includes('muffin') || combined.includes('donut') || combined.includes('éclair') ||
        combined.includes('tart') || combined.includes('pie')) {
        return 'pastries';
    }
    if (combined.includes('bread') || combined.includes('baguette') || combined.includes('loaf') ||
        combined.includes('roll') || combined.includes('sourdough')) {
        return 'bread';
    }
    if (combined.includes('coffee') || combined.includes('latte') || combined.includes('tea') ||
        combined.includes('drink') || combined.includes('juice') || combined.includes('smoothie')) {
        return 'drinks';
    }
    return 'other';
}

// Get product rating - now uses real data from API
function getProductRating(product) {
    // Real ratings come from API (avg_rating, rating_count)
    const rating = product.avg_rating || 0;
    const reviews = product.rating_count || 0;
    return { rating: Math.min(rating, 5.0), reviews };
}

/**
 * Fetch real-time stock status for products
 */
async function fetchStockStatus(productIds) {
    if (!productIds || productIds.length === 0) return;
    
    try {
        const res = await fetch('/api/stock/bulk', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ product_ids: productIds })
        });
        const data = await res.json();
        
        if (data.products) {
            data.products.forEach(p => {
                window.stockStatus[p.product_id] = p;
            });
        }
    } catch (e) {
        console.log('Failed to fetch stock status', e);
    }
}

async function loadProducts() {
    window.productsLoadError = '';
    try {
        const res = await fetch('/api/products?limit=50&sort_by=created_at&sort_dir=DESC&status=active');
        if (!res.ok) {
            throw new Error(`Failed to load products (${res.status})`);
        }
        const data = await res.json();
        const list = Array.isArray(data.products) ? data.products : [];
        window.products = list.map(p => {
            const product = {
                id: p.id,
                name: p.name,
                emoji: '',
                price: Number(p.price) || 0,
                created_at: p.created_at,
                image_url: p.image_url || '',
                description: p.description || '',
                availability_status: p.availability_status || 'available',
                stock: p.stock || 0,
                has_customization: true,
                avg_rating: p.avg_rating || 0,
                rating_count: p.rating_count || 0
            };
            // Add category and rating
            product.category = getCategoryForProduct(product);
            const ratingData = getProductRating(product);
            product.rating = ratingData.rating;
            product.reviewCount = ratingData.reviews;
            return product;
        });
        window.products.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        
        // Fetch real-time stock status for all products
        const productIds = window.products.map(p => p.id);
        await fetchStockStatus(productIds);
        
    } catch (e) {
        console.log('❌ Failed to load products', e);
        window.products = [];
        window.productsLoadError = 'Unable to load the menu right now. Please try again.';
        if (typeof showError === 'function') {
            showError(window.productsLoadError);
        }
    }
    return window.products;
}

/**
 * Get filtered products based on current search and category
 */
function getFilteredProducts() {
    let filtered = [...window.products];
    
    // Filter by category
    if (window.productFilters.category && window.productFilters.category !== 'all') {
        filtered = filtered.filter(p => p.category === window.productFilters.category);
    }
    
    // Filter by search
    if (window.productFilters.search) {
        const search = window.productFilters.search.toLowerCase().trim();
        filtered = filtered.filter(p => 
            p.name.toLowerCase().includes(search) ||
            (p.description && p.description.toLowerCase().includes(search))
        );
    }
    
    return filtered;
}

/**
 * Generate star rating HTML
 */
function renderStars(rating) {
    const fullStars = Math.floor(rating);
    const hasHalf = rating % 1 >= 0.5;
    let html = '<div class="p-rating-stars">';
    
    for (let i = 0; i < 5; i++) {
        if (i < fullStars) {
            html += '<svg class="p-rating-star" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>';
        } else if (i === fullStars && hasHalf) {
            html += '<svg class="p-rating-star" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>';
        } else {
            html += '<svg class="p-rating-star empty" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>';
        }
    }
    html += '</div>';
    return html;
}

function normalizeRatingSummary(data) {
    const summary = data && data.summary ? data.summary : data;
    const avg = Number(summary && (summary.avg_rating ?? summary.avgRating) ? (summary.avg_rating ?? summary.avgRating) : 0);
    const count = Number(summary && (summary.rating_count ?? summary.ratingCount) ? (summary.rating_count ?? summary.ratingCount) : 0);
    const safeAvg = Number.isFinite(avg) ? avg : 0;
    const safeCount = Number.isFinite(count) ? count : 0;
    return { avg: Math.max(0, safeAvg), count: Math.max(0, safeCount) };
}

function getReviewCandidates(limit) {
    const list = Array.isArray(window.products) ? window.products : [];
    const withReviews = list
        .map(p => ({
            product: p,
            count: Number(p.reviewCount ?? p.rating_count ?? 0)
        }))
        .filter(p => Number.isFinite(p.count) && p.count > 0)
        .sort((a, b) => b.count - a.count)
        .map(p => p.product);
    return withReviews.slice(0, limit);
}

async function fetchProductReviews(productId) {
    try {
        const res = await fetch(`/api/products/${productId}/ratings`);
        if (!res.ok) return null;
        return await res.json();
    } catch (e) {
        return null;
    }
}

async function loadLatestReviews() {
    const section = document.getElementById('reviewsSection');
    const summaryEl = document.getElementById('reviewsSummary');
    const listEl = document.getElementById('reviewsList');
    if (!section || !summaryEl || !listEl) return;

    listEl.innerHTML = '<div class="reviews-loading">Loading reviews...</div>';
    summaryEl.innerHTML = '';

    const candidates = getReviewCandidates(6);
    if (!candidates.length) {
        listEl.innerHTML = '<div class="reviews-empty">No reviews yet</div>';
        return;
    }

    const results = await Promise.all(candidates.map(async product => ({
        product,
        data: await fetchProductReviews(product.id)
    })));

    const groups = [];

    results.forEach(({ product, data }) => {
        if (!data) return;
        const summary = normalizeRatingSummary(data);
        const ratings = Array.isArray(data.ratings) ? data.ratings : [];
        if (summary.count === 0 && ratings.length === 0) return;
        const sortedRatings = ratings
            .filter(rating => rating && rating.created_at)
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            .slice(0, 3);
        groups.push({ product, summary, ratings: sortedRatings });
    });

    if (!groups.length) {
        listEl.innerHTML = '<div class="reviews-empty">No recent reviews yet</div>';
        return;
    }

    listEl.innerHTML = groups.map(({ product, summary, ratings }) => {
        const img = product.image_url && product.image_url.length > 0
            ? product.image_url
            : 'https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop';
        return `
            <div class="review-group">
                <div class="review-group-header">
                    <img class="review-group-thumb" src="${img}" alt="${escapeHtml(product.name)}" loading="lazy" />
                    <div class="review-group-info">
                        <div class="review-group-name">${escapeHtml(product.name)}</div>
                        <div class="review-group-summary">
                            <div class="review-group-score">${summary.avg.toFixed(1)}</div>
                            <div class="review-group-meta">
                                <div class="review-group-stars">${renderStars(summary.avg)}</div>
                                <div class="review-group-count">${summary.count} review${summary.count === 1 ? '' : 's'}</div>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="review-entry-list">
                    ${ratings.map(rating => {
                        const createdAt = rating.created_at ? new Date(rating.created_at).getTime() : Date.now();
                        const when = window.timeAgo ? window.timeAgo(createdAt) : new Date(createdAt).toLocaleDateString();
                        const comment = rating.comment ? escapeHtml(rating.comment) : 'No written review yet';
                        const emptyClass = rating.comment ? '' : ' review-entry-text--empty';
                        const stars = Number(rating.stars || 0);
                        return `
                            <div class="review-entry">
                                <div class="review-entry-head">
                                    <div class="review-entry-stars">${renderStars(stars)}</div>
                                    <div class="review-entry-time">${when}</div>
                                </div>
                                <div class="review-entry-text${emptyClass}">${comment}</div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    }).join('');
}

/**
 * Render category tabs
 */
function renderCategoryTabs() {
    const container = document.getElementById('categoryTabs');
    if (!container) return;
    
    // Count products per category
    const counts = { all: window.products.length };
    CATEGORIES.forEach(cat => {
        if (cat.id !== 'all') {
            counts[cat.id] = window.products.filter(p => p.category === cat.id).length;
        }
    });
    
    // Only show categories that have products
    const visibleCategories = CATEGORIES.filter(cat => counts[cat.id] > 0);
    
    container.innerHTML = visibleCategories.map(cat => `
        <button class="category-tab${window.productFilters.category === cat.id ? ' active' : ''}" 
                data-category="${cat.id}" onclick="handleCategoryClick('${cat.id}')">
            <i data-lucide="${cat.icon}"></i>
            <span>${cat.name}</span>
            ${cat.id !== 'all' ? `<span class="category-tab-count">${counts[cat.id]}</span>` : ''}
        </button>
    `).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

/**
 * Handle category tab click
 */
function handleCategoryClick(categoryId) {
    window.productFilters.category = categoryId;
    renderCategoryTabs();
    renderProducts();
}

/**
 * Handle search input
 */
function handleSearchInput(value) {
    window.productFilters.search = value;
    renderProducts();
    
    // Toggle clear button visibility
    const clearBtn = document.querySelector('.products-search-clear');
    if (clearBtn) {
        clearBtn.style.display = value ? 'flex' : 'none';
    }
}

/**
 * Clear all filters
 */
function clearFilters() {
    window.productFilters.search = '';
    window.productFilters.category = 'all';
    
    const searchInput = document.getElementById('productSearch');
    if (searchInput) searchInput.value = '';
    
    renderCategoryTabs();
    renderProducts();
}

function getBogoPromotionForProduct(productId) {
    const promos = typeof window.getActivePromotions === 'function' ? window.getActivePromotions() : [];
    const applies = typeof window.promotionAppliesToProduct === 'function'
        ? window.promotionAppliesToProduct
        : () => true;

    const applicable = promos
        .filter(p => p && p.type === 'BUY_X_GET_Y' && applies(p, productId))
        .sort((a, b) => (Number(b.priority) || 0) - (Number(a.priority) || 0));

    const pid = Number(productId);
    const promo = applicable.find(p => {
        const buyIds = Array.isArray(p.buyProductIds) ? p.buyProductIds.map(Number).filter(Number.isFinite) : [];
        const getIds = Array.isArray(p.getProductIds) ? p.getProductIds.map(Number).filter(Number.isFinite) : [];
        return buyIds.includes(pid) && getIds.includes(pid);
    });
    if (!promo) return null;

    let rules = {};
    if (typeof window.parsePromotionRules === 'function') {
        rules = window.parsePromotionRules(promo.rules);
    } else if (promo.rules) {
        try {
            rules = typeof promo.rules === 'string' ? (JSON.parse(promo.rules) || {}) : promo.rules;
        } catch {
            rules = {};
        }
    }

    const buyQty = Number(promo.buyQty ?? rules.buyQty) || 0;
    const getQty = Number(promo.getQty ?? rules.getQty) || 0;
    if (!buyQty || !getQty) return null;

    return { id: promo.id, buyQty, getQty };
}

function getCheckoutLineSummaryForProduct(productId) {
    const lines = Array.isArray(window.currentCheckout?.lineItems) ? window.currentCheckout.lineItems : [];
    const pid = Number(productId);
    if (!Number.isFinite(pid) || lines.length === 0) return null;
    let paidQty = 0;
    let freeQty = 0;
    let lineTotal = 0;
    let lineSubtotal = 0;
    for (const li of lines) {
        if (!li || Number(li.productId) !== pid) continue;
        const qty = Number(li.qty || 0);
        const free = Number(li.freeQty || 0);
        let paid = Number(li.paidQty ?? (qty - free));
        if (!Number.isFinite(paid)) paid = 0;
        paidQty += paid;
        freeQty += free;
        const rawSubtotal = li.lineSubtotal ?? (qty * Number(li.unitPrice || 0));
        let subtotal = Number(rawSubtotal);
        if (!Number.isFinite(subtotal)) subtotal = 0;
        const rawTotal = li.lineTotal ?? subtotal;
        let total = Number(rawTotal);
        if (!Number.isFinite(total)) total = 0;
        lineSubtotal += subtotal;
        lineTotal += total;
    }
    if (paidQty === 0 && freeQty === 0) return null;
    return { paidQty, freeQty, lineTotal, lineSubtotal };
}

function renderProducts() {
    const container = document.getElementById('products');
    const countEl = document.getElementById('productsCount');
    const filterInfoEl = document.getElementById('productsFilterInfo');
    
    const filtered = getFilteredProducts();
    const isFiltering = window.productFilters.search || window.productFilters.category !== 'all';
    
    // Update count
    if (countEl) {
        countEl.textContent = `${filtered.length} item${filtered.length !== 1 ? 's' : ''}`;
    }
    
    // Show/hide filter info
    if (filterInfoEl) {
        if (isFiltering && filtered.length > 0) {
            filterInfoEl.style.display = 'flex';
            filterInfoEl.innerHTML = `
                <span>Showing <strong>${filtered.length}</strong> of ${window.products.length} items</span>
                <button class="products-filter-clear" onclick="clearFilters()">Clear filters</button>
            `;
        } else {
            filterInfoEl.style.display = 'none';
        }
    }
    
    // Empty state - no products at all
    if (!window.products.length) {
        const emptyTitle = window.productsLoadError ? 'Menu unavailable' : 'No products yet';
        const emptyDesc = window.productsLoadError ? 'Please check your connection and try again.' : 'Check back soon for fresh treats!';
        container.innerHTML = `
            <div class="products-empty">
                <div class="products-empty-icon">
                    <i data-lucide="cake"></i>
                </div>
                <div class="products-empty-title">${emptyTitle}</div>
                <div class="products-empty-desc">${emptyDesc}</div>
            </div>
        `;
        if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
        return;
    }
    
    // No results state - filtered to nothing
    if (filtered.length === 0) {
        const searchTerm = window.productFilters.search;
        container.innerHTML = `
            <div class="products-no-results">
                <div class="products-no-results-icon">
                    <i data-lucide="search-x"></i>
                </div>
                <div class="products-no-results-title">No items found</div>
                <div class="products-no-results-desc">
                    ${searchTerm ? `We couldn't find anything matching "${escapeHtml(searchTerm)}"` : 'Try a different category'}
                </div>
                <button class="products-no-results-btn" onclick="clearFilters()">
                    <i data-lucide="rotate-ccw"></i>
                    Show all items
                </button>
            </div>
        `;
        if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
        return;
    }
    
    // Render filtered products
    container.innerHTML = filtered.map(p => {
        const img = p.image_url && p.image_url.length > 0 
            ? p.image_url 
            : `https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop`;
        const cart = window.getCart ? window.getCart() : {};
        const inCart = cart[p.id] > 0;
        
        // Check real-time stock status
        const stockInfo = window.stockStatus[p.id];
        const availableStock = stockInfo ? stockInfo.available_stock : p.stock;
        const isOutOfStock = stockInfo ? stockInfo.status === 'out_of_stock' : false;
        const isLowStock = stockInfo ? stockInfo.status === 'low_stock' : false;
        
        // Determine product status (real-time stock takes priority)
        const status = isOutOfStock ? 'sold_out' : (p.availability_status || 'available');
        const isSoldOut = status === 'sold_out' || isOutOfStock;
        const isLimited = status === 'limited' || isLowStock;
        
        // Status badge HTML - prioritize real-time stock info
        let statusBadge = '';
        if (isSoldOut) {
            statusBadge = '<div class="p-status p-status--sold-out"><i data-lucide="x-circle"></i>Sold out</div>';
        } else if (isLowStock && stockInfo) {
            statusBadge = `<div class="p-status p-status--limited"><i data-lucide="alert-circle"></i>Only ${availableStock} left</div>`;
        } else if (isLimited) {
            statusBadge = '<div class="p-status p-status--limited"><i data-lucide="clock"></i>Limited availability</div>';
        }
        
        // Rating HTML
        const ratingHtml = p.rating ? `
            <div class="p-rating">
                ${renderStars(p.rating)}
                <span class="p-rating-value">${p.rating.toFixed(1)}</span>
                <span class="p-rating-count">(${p.reviewCount})</span>
            </div>
        ` : '';

        const promo = !isSoldOut && typeof window.getBestPromotionForProduct === 'function'
            ? window.getBestPromotionForProduct(p.id)
            : null;
        const promoBadge = promo ? promo.badgeText : null;
        const promoBadgeVariant = promo ? promo.badgeVariant : null;
        const promoText = promo ? promo.promoText : null;
        const hasPercentPromo = promo && promo.type === 'PERCENT_OFF' && promo.percent > 0 && promo.percent <= 100;
        const discountedPrice = hasPercentPromo ? (p.price * (1 - (promo.percent / 100))) : null;
        const checkoutSummary = getCheckoutLineSummaryForProduct(p.id);
        const isFreeApplied = promo && promo.type === 'BUY_X_GET_Y'
            && checkoutSummary
            && checkoutSummary.freeQty > 0
            && (checkoutSummary.paidQty === 0 || checkoutSummary.lineTotal === 0);

        const qtyInCart = cart[p.id] || 0;
        const bogoPromo = !isSoldOut ? getBogoPromotionForProduct(p.id) : null;
        let bogoHint = '';
        if (bogoPromo && qtyInCart > 0) {
            const groupSize = bogoPromo.buyQty + bogoPromo.getQty;
            const freeItems = Math.floor(qtyInCart / groupSize) * bogoPromo.getQty;
            const remainder = qtyInCart % groupSize;
            if (qtyInCart < groupSize) {
                bogoHint = `Add ${groupSize - qtyInCart} more to get ${bogoPromo.getQty} free`;
            } else if (freeItems > 0 && remainder === 0) {
                bogoHint = `Buy ${bogoPromo.buyQty} Get ${bogoPromo.getQty} applied: ${freeItems} FREE`;
            } else if (freeItems > 0) {
                bogoHint = `Buy ${bogoPromo.buyQty} Get ${bogoPromo.getQty} applied: ${freeItems} FREE · Add ${groupSize - remainder} more for ${bogoPromo.getQty} more`;
            }
        }
        
        // Card classes
        const cardClasses = ['p-card'];
        if (inCart) cardClasses.push('in-cart');
        if (isSoldOut) cardClasses.push('p-card--sold-out');
        if (isLimited || isLowStock) cardClasses.push('p-card--limited');
        if (promo) cardClasses.push('p-card--promo');
        
        return `
        <div class="${cardClasses.join(' ')}" data-product-id="${p.id}" onclick="handleProductCardClick(event, ${p.id})">
            <div class="p-thumb-wrap">
                <img class="p-thumb" src="${img}" alt="${escapeHtml(p.name)}" loading="lazy" />
                ${isSoldOut ? '<div class="p-thumb-overlay"></div>' : ''}
                ${promoBadge ? `<div class="product-promotion-badge${promoBadgeVariant ? ` product-promotion-badge--${escapeHtml(promoBadgeVariant)}` : ''}">${escapeHtml(promoBadge)}</div>` : ''}
            </div>
            <div class="p-info">
                <div class="p-name">${escapeHtml(p.name)}</div>
                ${statusBadge}
                ${!statusBadge ? ratingHtml : ''}
                ${p.description ? `<div class="p-desc">${escapeHtml(p.description)}</div>` : ''}
                ${promoText ? `<div class="p-promo-sub">${escapeHtml(promoText)}</div>` : ''}
            </div>
            <div class="p-cta">
                ${isSoldOut ? `<div class="p-price">—</div>` : (
                    isFreeApplied
                        ? `<div class="p-price-wrap">
                             <div class="p-price p-price--promo">$0.00</div>
                             <div class="p-price-old">$${p.price.toFixed(2)}</div>
                           </div>`
                        : hasPercentPromo
                        ? `<div class="p-price-wrap">
                             <div class="p-price p-price--promo">$${discountedPrice.toFixed(2)}</div>
                             <div class="p-price-old">$${p.price.toFixed(2)}</div>
                           </div>`
                        : `<div class="p-price">$${p.price.toFixed(2)}</div>`
                )}
                ${bogoHint ? `<div class="p-promo-hint">${escapeHtml(bogoHint)}</div>` : ''}
                <div class="qty-controls${isSoldOut ? ' qty-controls--disabled' : ''}">
                    <button class="qty-btn minus" onclick="handleDecreaseClick(event, ${p.id})" id="dec-${p.id}" ${!inCart || isSoldOut ? 'disabled' : ''}>−</button>
                    <div class="qty-display" id="qty-${p.id}">${cart[p.id] || 0}</div>
                    <button class="qty-btn plus" onclick="handleAddClick(event, ${p.id})" ${isSoldOut ? 'disabled' : ''}>+</button>
                </div>
            </div>
        </div>`;
    }).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

/**
 * Handle click on product card (not on buttons)
 * Opens product detail sheet for editing/viewing
 */
function handleProductCardClick(event, productId) {
    // Don't trigger if clicking on buttons or qty controls
    if (event.target.closest('.qty-btn') || event.target.closest('.qty-controls')) {
        return;
    }
    
    event.preventDefault();
    event.stopPropagation();
    
    const product = window.products?.find(p => p.id == productId);
    if (!product) return;
    
    if (product.availability_status === 'sold_out') {
        showToast('This item is sold out');
        return;
    }
    
    // Open product detail sheet
    if (window.openProductDetail) {
        window.openProductDetail(productId);
    }
}

/**
 * Handle add button click
 * Opens product detail sheet for customization
 */
function handleAddClick(event, productId) {
    event.preventDefault();
    event.stopPropagation(); // Prevent card click
    
    const product = window.products?.find(p => p.id == productId);
    if (!product) return;
    
    if (product.availability_status === 'sold_out') {
        showToast('This item is sold out');
        return;
    }
    
    // Add micro-interaction
    const btn = event.currentTarget;
    btn.style.transform = 'scale(0.9)';
    setTimeout(() => btn.style.transform = '', 100);
    
    // Open product detail sheet for customization
    if (window.openProductDetail) {
        window.openProductDetail(productId);
    } else {
        // Fallback to direct add
        window.addToCartWithNote ? window.addToCartWithNote(productId, 1, '') : window.increaseQty(productId);
    }
}

/**
 * Handle decrease button click
 */
function handleDecreaseClick(event, productId) {
    event.preventDefault();
    event.stopPropagation(); // Prevent card click
    window.decreaseQty(productId);
}

// Export
window.loadProducts = loadProducts;
window.renderProducts = renderProducts;
window.renderCategoryTabs = renderCategoryTabs;
window.handleProductCardClick = handleProductCardClick;
window.handleAddClick = handleAddClick;
window.handleDecreaseClick = handleDecreaseClick;
window.handleCategoryClick = handleCategoryClick;
window.handleSearchInput = handleSearchInput;
window.clearFilters = clearFilters;
window.fetchStockStatus = fetchStockStatus;
window.CATEGORIES = CATEGORIES;
window.loadLatestReviews = loadLatestReviews;
