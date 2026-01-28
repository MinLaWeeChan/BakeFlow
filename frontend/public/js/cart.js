/**
 * BakeFlow - Cart Module
 * Cart state management and UI updates
 * Supports multiple instances of same product with different notes
 */

// Legacy cart format for backward compatibility: { productId: qty }
let cart = {};

// New cart format with notes support
// Array of: { id: string, productId: number, qty: number, note: string }
let cartItems = [];

let pendingSchedule = null;

// Generate unique cart item ID
function generateCartItemId() {
    return 'ci_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
}

/**
 * Add item to cart with note (new approach)
 * If note is empty and item already exists without note, increment qty
 * Otherwise create new cart item entry
 */
function addToCartWithNote(productId, qty, note) {
    const product = window.products?.find(p => p.id == productId);
    if (product && product.availability_status === 'sold_out') {
        showToast('This item is sold out');
        return;
    }
    
    // Check if same product with same note already exists
    const existingItem = cartItems.find(item => 
        item.productId == productId && 
        (item.note || '') === (note || '')
    );
    
    if (existingItem) {
        // Increment existing item
        existingItem.qty += qty;
    } else {
        // Add new cart item
        cartItems.push({
            id: generateCartItemId(),
            productId: parseInt(productId),
            qty: qty,
            note: note || ''
        });
    }
    
    // Sync legacy cart format
    syncLegacyCart();
    updateCart();
    markProductCardInCart(productId);
}

/**
 * Update existing cart item
 */
function updateCartItem(cartItemId, newQty, newNote) {
    const item = cartItems.find(i => i.id === cartItemId);
    if (!item) return;
    
    if (newQty <= 0) {
        removeCartItemById(cartItemId);
        return;
    }
    
    item.qty = newQty;
    item.note = newNote || '';
    
    syncLegacyCart();
    updateCart();
}

/**
 * Get cart item by ID
 */
function getCartItem(cartItemId) {
    return cartItems.find(i => i.id === cartItemId);
}

/**
 * Get all cart items
 */
function getCartItems() {
    return [...cartItems];
}

/**
 * Remove cart item by ID
 */
function removeCartItemById(cartItemId) {
    const idx = cartItems.findIndex(i => i.id === cartItemId);
    if (idx !== -1) {
        const item = cartItems[idx];
        cartItems.splice(idx, 1);
        syncLegacyCart();
        updateCart();
        
        // Check if product still in cart
        const stillInCart = cartItems.some(i => i.productId === item.productId);
        if (!stillInCart) {
            unmarkProductCard(item.productId);
        }
    }
}

/**
 * Sync legacy cart format from cartItems
 */
function syncLegacyCart() {
    cart = {};
    cartItems.forEach(item => {
        if (!cart[item.productId]) {
            cart[item.productId] = 0;
        }
        cart[item.productId] += item.qty;
    });
}

/**
 * Mark product card as in-cart
 */
function markProductCardInCart(productId) {
    const card = document.querySelector(`[data-product-id="${productId}"]`);
    if (card) {
        card.classList.add('in-cart');
    }
}

/**
 * Unmark product card
 */
function unmarkProductCard(productId) {
    const card = document.querySelector(`[data-product-id="${productId}"]`);
    if (card) {
        card.classList.remove('in-cart');
    }
}

// Legacy function - now opens product detail for customization
function increaseQty(productId) {
    // Check if product is sold out before adding
    const product = window.products?.find(p => p.id == productId);
    if (product && product.availability_status === 'sold_out') {
        showToast('This item is sold out');
        return;
    }
    
    // Open product detail sheet for customization
    if (window.openProductDetail) {
        window.openProductDetail(productId);
        return;
    }
    
    // Fallback to direct add if product-detail module not loaded
    addToCartWithNote(productId, 1, '');
}

// Legacy function - decrease by finding item without note first
function decreaseQty(productId) {
    // Find items for this product, prefer items without notes
    const itemsForProduct = cartItems.filter(i => i.productId == productId);
    if (itemsForProduct.length === 0) return;
    
    // Sort: items without notes first
    itemsForProduct.sort((a, b) => {
        if (!a.note && b.note) return -1;
        if (a.note && !b.note) return 1;
        return 0;
    });
    
    const item = itemsForProduct[0];
    if (item.qty > 1) {
        item.qty--;
    } else {
        removeCartItemById(item.id);
    }
    
    // Add micro-interaction
    const btn = document.getElementById(`dec-${productId}`);
    if (btn) {
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => btn.style.transform = '', 100);
    }
    
    syncLegacyCart();
    updateCart();
}

function updateCart() {
    let total = 0;
    let itemCount = 0;

    // Update product card displays using legacy cart
    if (window.products) {
        window.products.forEach(p => {
            const qty = cart[p.id] || 0;
            const qtyEl = document.getElementById(`qty-${p.id}`);
            const decEl = document.getElementById(`dec-${p.id}`);
            if (qtyEl) qtyEl.textContent = qty;
            if (decEl) decEl.disabled = qty === 0;
            
            // Update in-cart status
            const card = document.querySelector(`[data-product-id="${p.id}"]`);
            if (card) {
                if (qty > 0) {
                    card.classList.add('in-cart');
                } else {
                    card.classList.remove('in-cart');
                }
            }
        });
    }

    // Calculate totals from cartItems (more accurate with notes)
    cartItems.forEach(item => {
        const product = window.products?.find(p => p.id == item.productId);
        if (product) {
            total += (product.price * item.qty);
            itemCount += item.qty;
        }
    });

    const checkoutItems = cartItems
        .map(item => {
            const product = window.products?.find(p => p.id == item.productId);
            if (!product) return null;
            return {
                clientLineId: String(item.id || ''),
                productId: Number(item.productId),
                qty: Number(item.qty),
                unitPrice: Number(product.price)
            };
        })
        .filter(Boolean);

    const barItemsEl = document.getElementById('barItems');
    const barSubtotalEl = document.getElementById('barSubtotal');
    const barDiscountEl = document.getElementById('barDiscount');
    const barDiscountRowEl = document.getElementById('barDiscountRow');
    const barTotalEl = document.getElementById('barTotal');
    const barCheckoutEl = document.getElementById('barCheckout');
    const barSavingsEl = document.getElementById('barSavings');
    
    if (barItemsEl) barItemsEl.textContent = itemCount;
    if (barSubtotalEl) barSubtotalEl.textContent = `$${total.toFixed(2)}`;
    if (barDiscountEl) barDiscountEl.textContent = `-$${(0).toFixed(2)}`;
    if (barDiscountRowEl) barDiscountRowEl.style.display = 'none';
    if (barTotalEl) barTotalEl.textContent = `$${total.toFixed(2)}`;
    if (barCheckoutEl) barCheckoutEl.textContent = `Checkout • $${total.toFixed(2)}`;
    if (barCheckoutEl) barCheckoutEl.disabled = itemCount === 0;

    if (itemCount > 0 && typeof window.calculateCheckoutWithPromotions === 'function') {
        window.calculateCheckoutWithPromotions(checkoutItems)
            .then(checkout => {
                window.currentCheckout = checkout;
                const subtotal = Number(checkout.subtotal || total);
                const finalTotal = Number(checkout.total || total);
                const discount = Number(checkout.discount || 0);
                const appliedPromos = Array.isArray(checkout.appliedPromotions) && checkout.appliedPromotions.length > 0
                    ? checkout.appliedPromotions
                    : (checkout.appliedPromotion ? [checkout.appliedPromotion] : []);
                const hasPromo = appliedPromos.length > 0 && discount > 0;

                if (barSubtotalEl) barSubtotalEl.textContent = `$${subtotal.toFixed(2)}`;
                if (barDiscountEl) barDiscountEl.textContent = `-$${discount.toFixed(2)}`;
                if (barDiscountRowEl) barDiscountRowEl.style.display = hasPromo ? 'flex' : 'none';
                if (barTotalEl) barTotalEl.textContent = `$${finalTotal.toFixed(2)}`;
                if (barCheckoutEl) barCheckoutEl.textContent = `Checkout • $${finalTotal.toFixed(2)}`;
                if (barSavingsEl) {
                    barSavingsEl.style.display = 'none';
                    barSavingsEl.textContent = '';
                }

                if (pendingSchedule && barTotalEl) {
                    const when = `${pendingSchedule.date} ${pendingSchedule.time}`;
                    barTotalEl.textContent = `$${finalTotal.toFixed(2)}`;
                }
            })
            .catch(() => {
                if (barSubtotalEl) barSubtotalEl.textContent = `$${total.toFixed(2)}`;
                if (barDiscountEl) barDiscountEl.textContent = `-$${(0).toFixed(2)}`;
                if (barDiscountRowEl) barDiscountRowEl.style.display = 'none';
                if (barTotalEl) barTotalEl.textContent = `$${total.toFixed(2)}`;
                if (barCheckoutEl) barCheckoutEl.textContent = `Checkout • $${total.toFixed(2)}`;
                if (barSavingsEl) {
                    barSavingsEl.style.display = 'none';
                    barSavingsEl.innerHTML = '';
                }
                if (pendingSchedule && barTotalEl) {
                    const when = `${pendingSchedule.date} ${pendingSchedule.time}`;
                    barTotalEl.textContent = `$${total.toFixed(2)}`;
                }
            });
    } else {
        window.currentCheckout = null;
        if (barSubtotalEl) barSubtotalEl.textContent = `$${total.toFixed(2)}`;
        if (barDiscountEl) barDiscountEl.textContent = `-$${(0).toFixed(2)}`;
        if (barDiscountRowEl) barDiscountRowEl.style.display = 'none';
        if (barTotalEl) barTotalEl.textContent = `$${total.toFixed(2)}`;
        if (barCheckoutEl) barCheckoutEl.textContent = `Checkout • $${total.toFixed(2)}`;
        if (barSavingsEl) {
            barSavingsEl.style.display = 'none';
            barSavingsEl.innerHTML = '';
        }
        if (pendingSchedule && barTotalEl) {
            const when = `${pendingSchedule.date} ${pendingSchedule.time}`;
            barTotalEl.textContent = `$${total.toFixed(2)}`;
        }
    }
    
    adjustSafePadding();
    if (typeof window.renderPromotionBanner === 'function') {
        window.renderPromotionBanner();
    }
    
    // Update cart items sheet if open
    if (window.renderCartItemsList && document.querySelector('#cartItemsSheet.active')) {
        window.renderCartItemsList();
    }
}

function getCart() {
    return cart;
}

function setCart(newCart) {
    cart = newCart;
    // Convert legacy cart to cartItems format
    cartItems = [];
    Object.keys(cart).forEach(productId => {
        const qty = cart[productId];
        if (qty > 0) {
            cartItems.push({
                id: generateCartItemId(),
                productId: parseInt(productId),
                qty: qty,
                note: ''
            });
        }
    });
}

function getPendingSchedule() {
    return pendingSchedule;
}

function setPendingSchedule(schedule) {
    pendingSchedule = schedule;
}

function clearCart() {
    console.log('clearCart called, cart:', cart);
    if (cartItems.length === 0 && Object.keys(cart).length === 0) {
        showToast('Cart is already empty');
        return;
    }
    cart = {};
    cartItems = [];
    updateCart();
    showToast('Cart cleared', 'success');
}

/**
 * Get cart items formatted for order submission
 * Returns array with product info, notes, and images
 */
function getCartItemsForOrder() {
    return cartItems.map(item => {
        const product = window.products?.find(p => p.id == item.productId);
        return {
            product_id: item.productId,
            name: product?.name || '',
            qty: item.qty,
            price: product?.price || 0,
            note: item.note || '',
            image_url: product?.image_url || ''
        };
    });
}

/**
 * Check if any cart item has a note
 */
function cartHasNotes() {
    return cartItems.some(item => item.note && item.note.trim());
}

// Export
window.increaseQty = increaseQty;
window.decreaseQty = decreaseQty;
window.updateCart = updateCart;
window.getCart = getCart;
window.setCart = setCart;
window.getPendingSchedule = getPendingSchedule;
window.setPendingSchedule = setPendingSchedule;
window.clearCart = clearCart;
window.addToCartWithNote = addToCartWithNote;
window.updateCartItem = updateCartItem;
window.getCartItem = getCartItem;
window.getCartItems = getCartItems;
window.removeCartItem = removeCartItemById;
window.getCartItemsForOrder = getCartItemsForOrder;
window.cartHasNotes = cartHasNotes;
