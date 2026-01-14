/**
 * BakeFlow - Products Module
 * Product loading and rendering
 */

window.products = [];

async function loadProducts() {
    try {
        const res = await fetch('/api/products?limit=50&sort_by=created_at&sort_dir=DESC');
        const data = await res.json();
        const list = Array.isArray(data.products) ? data.products : [];
        window.products = list.map(p => ({
            id: p.id,
            name: p.name,
            emoji: '',
            price: Number(p.price) || 0,
            created_at: p.created_at,
            image_url: p.image_url || '',
            description: p.description || ''
        }));
        window.products.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
    } catch (e) {
        console.log('❌ Failed to load products', e);
        window.products = [];
    }
    return window.products;
}

function renderProducts() {
    const container = document.getElementById('products');
    const countEl = document.getElementById('productsCount');
    
    if (countEl) countEl.textContent = `${window.products.length} items`;
    
    if (!window.products.length) {
        container.innerHTML = `
            <div class="products-empty">
                <div class="products-empty-icon">
                    <i data-lucide="cake"></i>
                </div>
                <div class="products-empty-title">No products yet</div>
                <div class="products-empty-desc">Check back soon for fresh treats!</div>
            </div>
        `;
        return;
    }
    
    container.innerHTML = window.products.map(p => {
        const img = p.image_url && p.image_url.length > 0 
            ? p.image_url 
            : `https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop`;
        const cart = window.getCart ? window.getCart() : {};
        const inCart = cart[p.id] > 0;
        
        return `
        <div class="p-card${inCart ? ' in-cart' : ''}">
            <img class="p-thumb" src="${img}" alt="${escapeHtml(p.name)}" loading="lazy" />
            <div class="p-info">
                <div class="p-name">${escapeHtml(p.name)}</div>
                ${p.description ? `<div class="p-desc">${escapeHtml(p.description)}</div>` : ''}
            </div>
            <div class="p-cta">
                <div class="p-price">$${p.price.toFixed(2)}</div>
                <div class="qty-controls">
                    <button class="qty-btn minus" onclick="decreaseQty(${p.id})" id="dec-${p.id}" ${!inCart ? 'disabled' : ''}>−</button>
                    <div class="qty-display" id="qty-${p.id}">${cart[p.id] || 0}</div>
                    <button class="qty-btn plus" onclick="increaseQty(${p.id})">+</button>
                </div>
            </div>
        </div>`;
    }).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

// Export
window.loadProducts = loadProducts;
window.renderProducts = renderProducts;
