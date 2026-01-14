/**
 * BakeFlow - Main Application
 * Initializes all modules and handles app startup
 */

async function init() {
    // Load products
    await loadProducts();
    
    // Render UI
    renderProducts();
    updateCart();
    await Promise.resolve(renderRecentOrders());
    await Promise.resolve(renderSavedOrders());
    renderPlacesBar();
    renderRecurringOrders();
    
    // Initialize Lucide icons
    if (window.lucide && window.lucide.createIcons) {
        window.lucide.createIcons();
    }
    
    // Handle safe area padding
    adjustSafePadding();
    window.addEventListener('resize', adjustSafePadding);
    window.addEventListener('orientationchange', adjustSafePadding);

    // Handle pending reorder cart from Saved Orders page
    const params = new URLSearchParams(window.location.search);
    const userId = params.get('user_id') || 'guest';
    const tok = (window.getWebviewToken && window.getWebviewToken()) || params.get('t') || '';
    const pendingCartKey = `pending_reorder_${userId}`;
    const pendingCartKeyTok = tok ? `pending_reorder_t_${tok}` : '';
    const pendingCart = localStorage.getItem(pendingCartKey) || (pendingCartKeyTok ? localStorage.getItem(pendingCartKeyTok) : null);
    
    if (pendingCart) {
        try {
            const items = JSON.parse(pendingCart);
            const newCart = {};
            items.forEach(it => {
                const p = window.products.find(px => px.id == it.id);
                if (p) newCart[it.id] = it.qty;
            });
            setCart(newCart);
            updateCart();
            localStorage.removeItem(pendingCartKey);
            if (pendingCartKeyTok) localStorage.removeItem(pendingCartKeyTok);
            showToast('✓ Order loaded', 'success');
        } catch(e) { 
            console.log('Failed to load pending cart', e); 
        }
    }

    // Restore pending schedule (if any)
    try {
        const scheduleKey = `pending_schedule_${getUserId()}`;
        const raw = localStorage.getItem(scheduleKey);
        if (raw) {
            const sched = JSON.parse(raw);
            if (sched && sched.date && sched.time) {
                const when = new Date(`${sched.date}T${sched.time}:00`);
                if (!Number.isNaN(when.getTime()) && when.getTime() > Date.now()) {
                    window.setPendingSchedule(sched);
                    window.updateCart();
                } else {
                    localStorage.removeItem(scheduleKey);
                }
            }
        }
    } catch (e) {
        // ignore
    }
    
    // Handle reorder flag from order details
    if (params.get('reorder') === '1' && !pendingCart) {
        const key = `saved_orders_${userId}`;
        const list = JSON.parse(localStorage.getItem(key) || '[]');
        if (list.length > 0) {
            const newCart = {};
            (list[0].items || []).forEach(it => {
                const p = window.products.find(px => (px.name || '').toLowerCase() === (it.name || '').toLowerCase());
                const id = p ? p.id : null;
                if (id) newCart[id] = it.qty;
            });
            setCart(newCart);
            updateCart();
            showToast('✓ Order loaded', 'success');
        }
    }
    
    // Handle schedule flag
    if (params.get('schedule') === '1') {
        setTimeout(() => {
            const d = new Date();
            document.getElementById('scheduleDate').value = d.toISOString().slice(0,10);
            document.getElementById('scheduleTime').value = '';
            openSheet('scheduleSheet');
        }, 300);
    }

    // Wire navigation links
    wireNavigationLinks(userId);
    
    // Setup map tabs
    setupMapTabs();
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Initialize all modules
    init();
    initSheetListeners();
    initSaveSheet();
    initPlacesSheet();
    initScheduleSheet();
    initRecurringSheet();
    
    // Wire checkout and back buttons
    document.getElementById('barCheckout').addEventListener('click', openDeliveryForm);
    document.getElementById('backToCartBtn').addEventListener('click', backToCart);
    document.getElementById('placeOrderBtn').addEventListener('click', submitOrder);
    
    // Direct handler for Manage button
    const manageBtn = document.getElementById('manageSavedOrdersBtn');
    if (manageBtn) {
        manageBtn.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            goToSavedOrders();
        });
    }
    
    // Event delegation backup
    document.addEventListener('click', (e) => {
        if (e.target.closest('#manageSavedOrdersBtn')) {
            e.preventDefault();
            goToSavedOrders();
        }
    });
});
