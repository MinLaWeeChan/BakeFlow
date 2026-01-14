/**
 * BakeFlow - Cart Module
 * Cart state management and UI updates
 */

let cart = {};
let pendingSchedule = null;

function increaseQty(productId) {
    if (!cart[productId]) cart[productId] = 0;
    cart[productId]++;
    
    // Add micro-interaction
    const btn = document.querySelector(`#qty-${productId}`)?.closest('.qty-controls')?.querySelector('.plus');
    if (btn) {
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => btn.style.transform = '', 100);
    }
    
    updateCart();
    
    // Mark card as "in cart"
    const card = document.querySelector(`#qty-${productId}`)?.closest('.p-card');
    if (card && cart[productId] > 0) {
        card.classList.add('in-cart');
    }
}

function decreaseQty(productId) {
    if (cart[productId] > 0) {
        cart[productId]--;
        
        // Add micro-interaction
        const btn = document.getElementById(`dec-${productId}`);
        if (btn) {
            btn.style.transform = 'scale(0.9)';
            setTimeout(() => btn.style.transform = '', 100);
        }
        
        if (cart[productId] === 0) {
            delete cart[productId];
            // Remove "in cart" class
            const card = document.getElementById(`dec-${productId}`)?.closest('.p-card');
            if (card) card.classList.remove('in-cart');
        }
        updateCart();
    }
}

function updateCart() {
    let total = 0;
    let itemCount = 0;

    window.products.forEach(p => {
        const qty = cart[p.id] || 0;
        const qtyEl = document.getElementById(`qty-${p.id}`);
        const decEl = document.getElementById(`dec-${p.id}`);
        if (qtyEl) qtyEl.textContent = qty;
        if (decEl) decEl.disabled = qty === 0;
    });

    Object.keys(cart).forEach(id => {
        const product = window.products.find(p => p.id == id);
        if (product) {
            const qty = cart[id];
            total += (product.price * qty);
            itemCount += qty;
        }
    });

    const barItemsEl = document.getElementById('barItems');
    const barTotalEl = document.getElementById('barTotal');
    const barTotalBigEl = document.getElementById('barTotalBig');
    const barCheckoutEl = document.getElementById('barCheckout');
    
    if (barItemsEl) barItemsEl.textContent = itemCount;
    if (barTotalEl) barTotalEl.textContent = `$${total.toFixed(2)}`;
    if (barTotalBigEl) barTotalBigEl.textContent = `$${total.toFixed(2)}`;
    if (barCheckoutEl) barCheckoutEl.disabled = itemCount === 0;

    if (pendingSchedule && barTotalEl) {
        const when = `${pendingSchedule.date} ${pendingSchedule.time}`;
        barTotalEl.innerHTML = `$${total.toFixed(2)} · <span style="color:var(--primary);font-size:11px;">📅 ${when}</span>`;
    }
    
    adjustSafePadding();
}

function getCart() {
    return cart;
}

function setCart(newCart) {
    cart = newCart;
}

function getPendingSchedule() {
    return pendingSchedule;
}

function setPendingSchedule(schedule) {
    pendingSchedule = schedule;
}

// Export
window.increaseQty = increaseQty;
window.decreaseQty = decreaseQty;
window.updateCart = updateCart;
window.getCart = getCart;
window.setCart = setCart;
window.getPendingSchedule = getPendingSchedule;
window.setPendingSchedule = setPendingSchedule;
