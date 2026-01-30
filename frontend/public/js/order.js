/**
 * BakeFlow - Order Submission Module
 */

let isSubmitting = false; // Prevent double submissions
let activeOrder = null;
let activeOrderEditable = false;

function getAuthToken() {
    const params = new URLSearchParams(window.location.search);
    const tok = params.get('t');
    if (tok) return tok;
    if (window.getWebviewToken) return window.getWebviewToken();
    return '';
}

function splitNamePhone(rawName) {
    const nameValue = String(rawName || '').trim();
    const match = nameValue.match(/^(.*)\((.*)\)\s*$/);
    if (!match) return { name: nameValue || '', phone: '' };
    return { name: match[1].trim() || nameValue, phone: match[2].trim() || '' };
}

function setEditLock(isLocked) {
    const nameEl = document.getElementById('customerName');
    const phoneEl = document.getElementById('customerPhone');
    const addressEl = document.getElementById('customerAddress');
    if (nameEl) nameEl.disabled = !!isLocked;
    if (phoneEl) phoneEl.disabled = !!isLocked;
    if (addressEl) addressEl.disabled = !!isLocked;
    document.querySelectorAll('.radio-option').forEach(opt => {
        opt.classList.toggle('disabled', !!isLocked);
    });
}

function applyActiveOrder(order) {
    const banner = document.getElementById('activeOrderBanner');
    if (!banner) return;
    // Hide the banner - we now use the choice dialog instead
    banner.style.display = 'none';

    if (!order || !order.id) {
        setEditLock(false);
        return;
    }

    const { name, phone } = splitNamePhone(order.customer_name);
    const nameEl = document.getElementById('customerName');
    const phoneEl = document.getElementById('customerPhone');
    const addressEl = document.getElementById('customerAddress');
    if (nameEl && name) nameEl.value = name;
    if (phoneEl && phone) phoneEl.value = phone;

    if (order.delivery_type && window.selectDeliveryType) {
        window.selectDeliveryType(order.delivery_type);
        if (order.delivery_type === 'delivery' && addressEl) {
            addressEl.value = order.address || '';
        }
        if (order.delivery_type !== 'delivery' && addressEl) {
            addressEl.value = '';
        }
    }

    setEditLock(!activeOrderEditable);
}

async function loadActiveOrder() {
    const tok = getAuthToken();
    if (!tok) return;
    try {
        const res = await fetch(`/api/me/active-order?t=${encodeURIComponent(tok)}`);
        if (!res.ok) return;
        const data = await res.json().catch(() => ({}));
        if (!data || !data.order) return;
        activeOrder = data.order;
        activeOrderEditable = !!data.editable;
        applyActiveOrder(activeOrder);
    } catch (e) {
        console.log('Failed to load active order', e);
    }
}

function sanitizeMyanmarPhoneInput(raw) {
    const v = (raw || '').trim();
    const hasPlus = v.startsWith('+');
    const digits = v.replace(/\D/g, '');
    if (!digits) return hasPlus ? '+' : '';
    return hasPlus ? `+${digits}` : digits;
}

function normalizeMyanmarPhoneE164(raw) {
    const s = sanitizeMyanmarPhoneInput(raw);
    if (!s) return null;
    if (/^\+959\d{9}$/.test(s)) return s;
    if (/^09\d{9}$/.test(s)) return `+959${s.slice(2)}`;
    return null;
}

function isValidMyanmarPhone(raw) {
    return !!normalizeMyanmarPhoneE164(raw);
}

/**
 * Validate cart stock before checkout
 * Returns { valid: boolean, message: string, items: [...] }
 */
async function validateCartStock() {
    const items = window.getCartItemsForOrder ? window.getCartItemsForOrder() : [];
    if (items.length === 0) {
        return { valid: false, message: 'Your cart is empty', items: [] };
    }

    // Build validation request
    const cartForValidation = items.map(item => ({
        product_id: item.product_id,
        quantity: item.qty
    }));

    try {
        const res = await fetch('/api/stock/validate-cart', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ items: cartForValidation })
        });
        const data = await res.json();
        return data;
    } catch (e) {
        console.error('Stock validation error:', e);
        // Allow order to proceed if validation fails (fail-open for better UX)
        return { valid: true, message: '', items: [] };
    }
}

function submitOrder() {
    console.log('🔍 Submit order clicked');
    
    // Prevent double submission
    if (isSubmitting) {
        console.log('⚠️ Order already submitting, ignoring click');
        return;
    }
    
    const name = document.getElementById('customerName').value.trim();
    const phoneRaw = document.getElementById('customerPhone').value.trim();
    const address = document.getElementById('customerAddress').value.trim();
    const notes = document.getElementById('orderNotes').value.trim();
    const deliveryType = window.getDeliveryType();

    if (!name) { showError('Please enter your name'); return; }
    if (!phoneRaw) { showError('Please enter your phone number'); return; }
    const phone = normalizeMyanmarPhoneE164(phoneRaw);
    if (!phone || !isValidMyanmarPhone(phone)) {
        showError('Enter 09xxxxxxxxx or +959xxxxxxxxx');
        return;
    }
    if (!deliveryType) { showError('Please select delivery type'); return; }
    if (deliveryType === 'delivery' && !address) { showError('Please enter delivery address'); return; }

    const phoneEl = document.getElementById('customerPhone');
    if (phoneEl && phoneEl.value.trim() !== phone) {
        phoneEl.value = phone;
    }

    if (activeOrder && activeOrder.id && activeOrderEditable) {
        const normalize = (v) => String(v || '').trim().toLowerCase();
        const activeParsed = splitNamePhone(activeOrder.customer_name);
        const activeName = normalize(activeParsed.name);
        const activePhone = normalizeMyanmarPhoneE164(activeParsed.phone) || normalize(activeParsed.phone);
        const nextPhone = normalizeMyanmarPhoneE164(phone) || normalize(phone);
        const activeType = normalize(activeOrder.delivery_type);
        const nextType = normalize(deliveryType);
        const activeAddress = normalize(activeOrder.delivery_type === 'delivery' ? activeOrder.address : 'pickup');
        const nextAddress = normalize(deliveryType === 'delivery' ? address : 'pickup');

        const nameChanged = normalize(name) !== activeName;
        const phoneChanged = nextPhone !== activePhone;
        const typeChanged = nextType !== activeType;
        const addressChanged = nextAddress !== activeAddress;

        if (nameChanged || phoneChanged || typeChanged || addressChanged) {
            const ok = window.confirm(`You already have an active order (#${activeOrder.id}). Updating your details will update that same order. Continue?`);
            if (!ok) {
                resetSubmitButton();
                return;
            }
        }
    }

    const urlParams = new URLSearchParams(window.location.search);
    const urlUserId = urlParams.get('user_id') || '';
    let storedUserId = '';
    try {
        storedUserId = localStorage.getItem('bf_psid') || localStorage.getItem('bf_user_id') || '';
    } catch (e) {}
    const resolvedUserId = (storedUserId && storedUserId !== 'guest')
        ? storedUserId
        : ((urlUserId && urlUserId !== 'guest')
            ? urlUserId
            : (window.getUserId ? window.getUserId() : 'guest'));
    let userId = resolvedUserId || 'guest';
    const tok = urlParams.get('t') || (window.getWebviewToken ? window.getWebviewToken() : '');
    if (tok && userId === 'guest') {
        userId = '';
    }
    if (!tok && (!userId || userId === 'guest')) {
        showError('Please open this order form from Messenger to receive confirmation.');
        return;
    }

    // Use new cart items with notes if available
    let items;
    if (window.getCartItemsForOrder) {
        items = window.getCartItemsForOrder();
    } else {
        // Fallback to legacy cart format
        const cart = window.getCart();
        items = Object.keys(cart).map(id => {
            const product = window.products.find(p => p.id == id);
            return {
                product_id: parseInt(id),
                name: product.name,
                qty: cart[id],
                price: product.price,
                note: ''
            };
        });
    }

    if (!items.length) { showError('Your cart is empty'); return; }

    // Disable button and show loading state
    isSubmitting = true;
    const submitBtn = document.getElementById('placeOrderBtn');
    if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.innerHTML = '<span class="spinner"></span> Checking availability...';
    }

    // Validate stock availability before submitting
    validateCartStock().then(validation => {
        if (!validation.valid) {
            // Build detailed error message
            const unavailableItems = validation.items?.filter(i => !i.is_available) || [];
            if (unavailableItems.length > 0) {
                const itemMessages = unavailableItems.map(i => {
                    const product = window.products.find(p => p.id == i.product_id);
                    const productName = product ? product.name : `Product #${i.product_id}`;
                    return `${productName}: ${i.message}`;
                });
                showError(`Some items are no longer available:\n${itemMessages.join('\n')}`);
            } else {
                showError(validation.message || 'Unable to validate cart');
            }
            resetSubmitButton();
            return;
        }

        // Stock validated, proceed with order
        if (submitBtn) {
            submitBtn.innerHTML = '<span class="spinner"></span> Placing order...';
        }
        processOrderSubmission(name, phone, address, notes, deliveryType, userId, tok, items);
    }).catch(err => {
        console.error('Stock validation error:', err);
        // Proceed anyway on network failure (fail-open)
        if (submitBtn) {
            submitBtn.innerHTML = '<span class="spinner"></span> Placing order...';
        }
        processOrderSubmission(name, phone, address, notes, deliveryType, userId, tok, items);
    });
}

function processOrderSubmission(name, phone, address, notes, deliveryType, userId, tok, items) {

    // Combine all item notes into a formatted string for the order notes
    const itemNotesText = items
        .filter(item => item.note && item.note.trim())
        .map(item => `${item.name}: ${item.note}`)
        .join(' | ');
    
    // Combine global notes with item-specific notes
    let combinedNotes = notes;
    if (itemNotesText) {
        combinedNotes = combinedNotes 
            ? `${notes}\n\n📝 Item notes: ${itemNotesText}`
            : `📝 Item notes: ${itemNotesText}`;
    }

    const orderData = {
        user_id: userId,
        items: items,
        channel: 'messenger',
        customer_name: name,
        customer_phone: phone,
        delivery_type: deliveryType,
        address: deliveryType === 'delivery' ? address : 'Pickup at store',
        notes: combinedNotes,
        schedule: window.getPendingSchedule() || null,
        geo: window.getGeo(),
        delivery_directions: document.getElementById('deliveryDirections')?.value.trim() || ''
    };

    // Include promotion data if one was applied
    if (window.currentCheckout) {
        orderData.discount = window.currentCheckout.discount || 0;
        orderData.appliedPromotion = window.currentCheckout.appliedPromotion || null;
        console.log('💰 Applying promotion:', window.currentCheckout.appliedPromotion, 'Discount:', window.currentCheckout.discount);
    }

    console.log('📦 Sending order:', orderData);

    const endpoint = tok ? (`/api/chat/orders?t=${encodeURIComponent(tok)}`) : '/api/chat/orders';

    fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(orderData)
    })
    .then(async res => {
        const text = await res.text();
        let data = null;
        try {
            data = JSON.parse(text);
        } catch (e) {
            data = { success: false, message: text };
        }
        return { res, data };
    })
    .then(({ res, data }) => {
        console.log('📥 Response data:', JSON.stringify(data, null, 2));
        console.log('📥 data.action:', data?.action);
        
        // Check for order choice request FIRST (before success check)
        if (data && data.action === 'ask_user_choice') {
            console.log('✅ Showing choice dialog!');
            showOrderChoiceDialog(data, orderData, name, phone);
            resetSubmitButton();
            return;
        }
        
        if (res.ok && data && data.success) {
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

            try {
                const invoiceKey = `bf_invoice_${data.order_id}`;
                const promotions = Array.isArray(window.currentCheckout?.appliedPromotions)
                    ? window.currentCheckout.appliedPromotions.map(p => ({
                        label: p.label || p.name || p.code || 'Promotion',
                        amount: p.amount ?? p.discount ?? p.value ?? null
                    }))
                    : (window.currentCheckout?.appliedPromotion ? [{ label: window.currentCheckout.appliedPromotion, amount: window.currentCheckout.discount ?? null }] : []);
                const invoiceData = {
                    order_id: data.order_id,
                    created_at: new Date().toISOString(),
                    customer_name: name,
                    customer_phone: phone,
                    address: deliveryType === 'delivery' ? address : 'Pickup at store',
                    notes: combinedNotes,
                    delivery_type: deliveryType === 'pickup' ? 'Pick Up' : 'Delivery',
                    payment_status: 'Pay on delivery',
                    subtotal: window.currentCheckout?.subtotal ?? data.subtotal ?? null,
                    discount: window.currentCheckout?.discount ?? data.discount ?? null,
                    delivery_fee: window.currentCheckout?.delivery_fee ?? data.delivery_fee ?? null,
                    total: window.currentCheckout?.total ?? data.total ?? data.total_amount ?? null,
                    promotions,
                    items: (orderData.items || []).map(it => ({
                        name: it.name,
                        qty: it.qty,
                        price: it.price,
                        line_total: Number(it.price) * Number(it.qty)
                    }))
                };
                localStorage.setItem(invoiceKey, JSON.stringify(invoiceData));
            } catch (e) {
                console.log('Failed to store invoice', e);
            }

            const params = new URLSearchParams();
            if (data.order_id != null) params.set('order_id', data.order_id);
            if (userId) params.set('user_id', userId);
            params.set('delivery_type', deliveryType === 'pickup' ? 'Pick Up' : 'Delivery');
            if (name) params.set('customer_name', name);
            if (phone) params.set('customer_phone', phone);
            if (address) params.set('address', deliveryType === 'delivery' ? address : 'Pickup at store');
            if (combinedNotes) params.set('notes', combinedNotes);

            if (window.currentCheckout) {
                if (window.currentCheckout.subtotal != null) params.set('subtotal', window.currentCheckout.subtotal);
                if (window.currentCheckout.discount != null) params.set('discount', window.currentCheckout.discount);
                if (window.currentCheckout.delivery_fee != null) params.set('delivery_fee', window.currentCheckout.delivery_fee);
                if (window.currentCheckout.total != null) params.set('total', window.currentCheckout.total);
            }
            if (data.total != null) params.set('total', data.total);
            if (data.total_amount != null) params.set('total_amount', data.total_amount);

            window.location.href = `/order-details.html?${params.toString()}`;
        } else {
            // Handle specific errors (like insufficient stock)
            if (data && data.error === 'insufficient_stock') {
                showError(`Sorry, only ${data.available} ${data.product} available. Please reduce quantity.`);
            } else if (data && data.error === 'product_unavailable') {
                showError(data.message || 'Product is no longer available');
            } else {
                showError('Order failed: ' + (data && data.message ? data.message : 'Unknown error'));
            }
            resetSubmitButton();
        }
    })
    .catch(err => {
        console.error('❌ Error:', err);
        showError('Network error. Please try again.');
        resetSubmitButton();
    });
}

function resetSubmitButton() {
    isSubmitting = false;
    const submitBtn = document.getElementById('placeOrderBtn');
    if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.innerHTML = 'Place Order';
    }
}

function resetOrder() {
    window.setCart({});
    // Reset cart items array too
    if (window.cartItems) {
        window.cartItems.length = 0;
    }
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
    resetSubmitButton(); // Reset button for next order
}

// Export
window.submitOrder = submitOrder;
window.validateCartStock = validateCartStock;
window.loadActiveOrder = loadActiveOrder;
// ========== Order Choice Dialog ==========
function showOrderChoiceDialog(choiceData, orderData, name, phone) {
    console.log('🎨 Creating choice dialog');
    
    // Remove any existing dialog first
    const existing = document.getElementById('orderChoiceDialog');
    if (existing) {
        existing.remove();
    }
    
    // Create a modal dialog
    const dialog = document.createElement('div');
    dialog.id = 'orderChoiceDialog';
    dialog.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0,0,0,0.6);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10000;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        animation: fadeIn 0.2s ease-out;
    `;
    
    // Add animations if not already in document
    if (!document.getElementById('orderChoiceAnimations')) {
        const style = document.createElement('style');
        style.id = 'orderChoiceAnimations';
        style.textContent = `
            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }
            @keyframes slideUp {
                from {
                    opacity: 0;
                    transform: translateY(20px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }
        `;
        document.head.appendChild(style);
    }
    
    const content = document.createElement('div');
    content.style.cssText = `
        background: white;
        border-radius: 12px;
        padding: 40px;
        max-width: 480px;
        width: 90%;
        box-shadow: 0 10px 50px rgba(0,0,0,0.25);
        text-align: center;
        animation: slideUp 0.3s ease-out;
    `;
    
    const title = document.createElement('h2');
    title.textContent = 'Active Orders';
    title.style.cssText = `
        margin: 0 0 12px 0;
        font-size: 22px;
        font-weight: 700;
        color: #1a1a1a;
        letter-spacing: -0.5px;
    `;
    
    const msg = document.createElement('p');
    const orderCount = choiceData.orders ? choiceData.orders.length : 1;
    msg.textContent = `You have ${orderCount} active ${orderCount === 1 ? 'order' : 'orders'}. Which would you like to add items to?`;
    msg.style.cssText = `
        margin: 0 0 24px 0;
        font-size: 15px;
        color: #555;
        line-height: 1.6;
    `;
    
    // Create order selection list
    const orderList = document.createElement('div');
    orderList.style.cssText = `
        background: #f9f9f9;
        border-radius: 8px;
        margin-bottom: 24px;
        max-height: 300px;
        overflow-y: auto;
    `;
    
    if (choiceData.orders && choiceData.orders.length > 0) {
        choiceData.orders.forEach(order => {
            const orderItem = document.createElement('div');
            orderItem.style.cssText = `
                padding: 12px 16px;
                border-bottom: 1px solid #eee;
                display: flex;
                justify-content: space-between;
                align-items: center;
                cursor: pointer;
                transition: background 0.2s;
            `;
            orderItem.onmouseover = () => { orderItem.style.background = '#f0f0f0'; };
            orderItem.onmouseout = () => { orderItem.style.background = 'transparent'; };
            
            const orderInfo = document.createElement('div');
            orderInfo.textContent = order.summary;
            orderInfo.style.cssText = `
                flex: 1;
                font-size: 14px;
                color: #333;
                font-weight: 500;
            `;
            
            const selectBtn = document.createElement('button');
            selectBtn.textContent = 'Select';
            selectBtn.style.cssText = `
                padding: 6px 12px;
                background: #4CAF50;
                color: white;
                border: none;
                border-radius: 4px;
                font-size: 12px;
                cursor: pointer;
                transition: background 0.2s;
            `;
            selectBtn.onmouseover = () => { selectBtn.style.background = '#45a049'; };
            selectBtn.onmouseout = () => { selectBtn.style.background = '#4CAF50'; };
            selectBtn.onclick = () => {
                console.log('User chose order:', order.id);
                dialog.remove();
                sendOrderChoice('add_to_existing', order.id, orderData, name, phone);
            };
            
            orderItem.appendChild(orderInfo);
            orderItem.appendChild(selectBtn);
            orderList.appendChild(orderItem);
        });
    }
    
    const buttonGroup = document.createElement('div');
    buttonGroup.style.cssText = `
        display: flex;
        gap: 12px;
        flex-wrap: wrap;
    `;
    
    const addBtn = document.createElement('button');
    addBtn.textContent = '✨ Create New Order Instead';
    addBtn.style.cssText = `
        flex: 1;
        min-width: 200px;
        padding: 14px 24px;
        background: linear-gradient(135deg, #2196F3 0%, #1976D2 100%);
        color: white;
        border: none;
        border-radius: 8px;
        font-size: 15px;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.3s ease;
        box-shadow: 0 4px 12px rgba(33, 150, 243, 0.3);
    `;
    addBtn.onmouseover = () => { 
        addBtn.style.boxShadow = '0 6px 20px rgba(33, 150, 243, 0.4)';
        addBtn.style.transform = 'translateY(-2px)';
    };
    addBtn.onmouseout = () => { 
        addBtn.style.boxShadow = '0 4px 12px rgba(33, 150, 243, 0.3)';
        addBtn.style.transform = 'translateY(0)';
    };
    addBtn.onclick = () => {
        console.log('User chose: new order');
        dialog.remove();
        sendOrderChoice('new_order', null, orderData, name, phone);
    };
    
    buttonGroup.appendChild(addBtn);
    
    content.appendChild(title);
    content.appendChild(msg);
    content.appendChild(orderList);
    content.appendChild(buttonGroup);
    dialog.appendChild(content);
    
    document.body.appendChild(dialog);
    console.log('✅ Dialog appended to body');
}

function sendOrderChoice(choice, existingOrderID, orderData, name, phone) {
    console.log('📤 Sending order choice:', choice);
    
    const tok = getAuthToken(); // Get the token using the same method
    
    const choiceRequest = {
        choice: choice,
        order_id: existingOrderID || 0,
        items: orderData.items,
        customer_name: name,
        delivery_type: orderData.delivery_type,
        address: orderData.address
    };
    
    // Use relative URL with token parameter for ngrok compatibility
    const tokenParam = tok ? `?t=${encodeURIComponent(tok)}` : '';
    const apiUrl = `/api/chat/orders/choice${tokenParam}`;
    
    console.log('📤 API URL:', apiUrl);
    console.log('📤 Token present:', !!tok);
    
    fetch(apiUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(choiceRequest)
    })
    .then(async res => {
        console.log('📥 Response status:', res.status);
        const text = await res.text();
        console.log('📥 Response text:', text);
        try {
            return JSON.parse(text);
        } catch (e) {
            console.error('Failed to parse response:', e);
            return { success: false, message: text };
        }
    })
    .then(data => {
        console.log('📥 Choice response:', data);
        if (data.success) {
            completeOrderSubmission(data, name, phone, orderData);
        } else {
            alert('❌ Error: ' + (data.message || 'Failed to process order'));
            resetSubmitButton();
        }
    })
    .catch(err => {
        console.error('❌ Network error:', err);
        alert('❌ Network error: ' + err.message);
        resetSubmitButton();
    });
}

function completeOrderSubmission(data, name, phone, orderData) {
    const orderMsg = data.action === 'items_merged' 
        ? `Items added to order #${data.orderID}!`
        : `Order #${data.orderID} placed successfully!`;
    
    showToast(orderMsg, 'success');
    
    // Close form after a short delay
    setTimeout(() => {
        if (window.MessengerExtensions) {
            window.MessengerExtensions.requestCloseBrowser(
                () => console.log("✅ Webview closed"),
                (err) => {
                    console.log("❌ Error closing webview:", err);
                    window.location.reload();
                }
            );
        } else {
            resetOrder();
        }
    }, 1500);
}