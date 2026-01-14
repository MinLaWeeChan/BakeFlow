/**
 * BakeFlow - Order Submission Module
 */

function submitOrder() {
    console.log('🔍 Submit order clicked');
    
    const name = document.getElementById('customerName').value.trim();
    const phone = document.getElementById('customerPhone').value.trim();
    const address = document.getElementById('customerAddress').value.trim();
    const notes = document.getElementById('orderNotes').value.trim();
    const deliveryType = window.getDeliveryType();

    if (!name) { showError('Please enter your name'); return; }
    if (!phone) { showError('Please enter your phone number'); return; }
    if (!deliveryType) { showError('Please select delivery type'); return; }
    if (deliveryType === 'delivery' && !address) { showError('Please enter delivery address'); return; }

    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('user_id');
    const tok = urlParams.get('t');
    const cart = window.getCart();

    const items = Object.keys(cart).map(id => {
        const product = window.products.find(p => p.id == id);
        return {
            product_id: parseInt(id),
            name: product.name,
            qty: cart[id],
            price: product.price
        };
    });

    if (!items.length) { showError('Your cart is empty'); return; }

    const orderData = {
        user_id: userId,
        items: items,
        channel: 'messenger',
        customer_name: name,
        customer_phone: phone,
        delivery_type: deliveryType,
        address: deliveryType === 'delivery' ? address : 'Pickup at store',
        notes: notes,
        schedule: window.getPendingSchedule() || null,
        geo: window.getGeo(),
        delivery_directions: document.getElementById('deliveryDirections')?.value.trim() || ''
    };

    console.log('📦 Sending order:', orderData);

    const endpoint = tok ? (`/api/chat/orders?t=${encodeURIComponent(tok)}`) : '/api/chat/orders';

    fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(orderData)
    })
    .then(res => res.json())
    .then(data => {
        if (data.success) {
            try {
                const userKey = `recent_orders_${getUserId()}`;
                const existing = JSON.parse(localStorage.getItem(userKey) || '[]');

                const itemsWithImages = (orderData.items || []).map(it => {
                    const prod = window.products.find(p => p.id == it.product_id);
                    return {
                        id: it.product_id,
                        qty: it.qty,
                        name: it.name,
                        price: it.price,
                        image: prod ? prod.image_url : ''
                    };
                });

                const entry = {
                    order_id: data.order_id,
                    timestamp: Date.now(),
                    items: itemsWithImages
                };

                const next = Array.isArray(existing) ? existing : [];
                next.unshift(entry);
                localStorage.setItem(userKey, JSON.stringify(next.slice(0, 10)));

                if (window.renderRecentOrders) window.renderRecentOrders();
            } catch (e) {
                console.log('Failed to persist recent order', e);
            }

            showToast(`Order #${data.order_id} placed!`, 'success');
            
            if (window.MessengerExtensions) {
                setTimeout(() => {
                    window.MessengerExtensions.requestCloseBrowser(
                        () => console.log("✅ Webview closed"),
                        (err) => {
                            console.log("❌ Error closing webview:", err);
                            resetOrder();
                        }
                    );
                }, 1500);
            } else {
                setTimeout(resetOrder, 1500);
            }
        } else {
            showError('Order failed: ' + (data.message || 'Unknown error'));
        }
    })
    .catch(err => {
        console.error('❌ Error:', err);
        showError('Network error. Please try again.');
    });
}

function resetOrder() {
    window.setCart({});
    try {
        window.setPendingSchedule(null);
        localStorage.removeItem(`pending_schedule_${getUserId()}`);
    } catch (e) {
        // ignore
    }
    document.getElementById('customerName').value = '';
    document.getElementById('customerPhone').value = '';
    document.getElementById('customerAddress').value = '';
    document.getElementById('orderNotes').value = '';
    backToCart();
    window.updateCart();
}

// Export
window.submitOrder = submitOrder;
