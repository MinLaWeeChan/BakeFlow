/**
 * BakeFlow - Main Application
 * Initializes all modules and handles app startup
 */

async function init() {
    // Initialize language first
    if (typeof initLanguage === 'function') {
        initLanguage();
    }

    // Load products
    await loadProducts();

    // Load promotions
    if (typeof loadPromotions === 'function') {
        await loadPromotions();
        if (typeof renderPromotionBanner === 'function') {
            renderPromotionBanner();
        }
    }

    // Render UI
    renderTagChips();
    renderProducts();
    updateCart();
    await Promise.resolve(renderRecentOrders());
    await Promise.resolve(renderSavedOrders());
    renderPlacesBar();
    renderRecurringOrders();

    if (typeof initPreorderUI === 'function') {
        initPreorderUI();
    }

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

    // Load custom notes if available
    const notesKey = `pending_notes_${userId}`;
    const notesKeyTok = tok ? `pending_notes_t_${tok}` : '';
    const pendingNotes = localStorage.getItem(notesKey) || (notesKeyTok ? localStorage.getItem(notesKeyTok) : null);
    let customNotes = {};
    if (pendingNotes) {
        try {
            customNotes = JSON.parse(pendingNotes);
        } catch (e) { }
        localStorage.removeItem(notesKey);
        if (notesKeyTok) localStorage.removeItem(notesKeyTok);
    }

    if (pendingCart) {
        try {
            const items = JSON.parse(pendingCart);
            const newCart = {};
            const unavailableItems = [];
            const itemsWithNotes = [];

            items.forEach(it => {
                const p = window.products.find(px => px.id == it.id);
                if (p) {
                    newCart[it.id] = it.qty;
                    // Check for custom notes from item or from notes storage
                    const note = it.note || customNotes[it.id] || '';
                    if (note) {
                        itemsWithNotes.push({ id: it.id, note });
                    }
                } else {
                    unavailableItems.push(it.name || `Item #${it.id}`);
                }
            });
            setCart(newCart);

            // Store custom notes globally for order submission
            if (itemsWithNotes.length > 0) {
                window.cartItemNotes = {};
                itemsWithNotes.forEach(item => {
                    window.cartItemNotes[item.id] = item.note;
                });
            }

            updateCart();
            localStorage.removeItem(pendingCartKey);
            if (pendingCartKeyTok) localStorage.removeItem(pendingCartKeyTok);

            if (unavailableItems.length > 0) {
                const itemNames = unavailableItems.slice(0, 3).join(', ');
                const more = unavailableItems.length > 3 ? ` and ${unavailableItems.length - 3} more` : '';
                showToast(`Some items unavailable: ${itemNames}${more}`, 'warning');
            } else if (itemsWithNotes.length > 0) {
                showToast('✓ Order loaded with notes', 'success');
            } else {
                showToast('✓ Order loaded', 'success');
            }
        } catch (e) {
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
            const unavailableItems = [];
            (list[0].items || []).forEach(it => {
                const p = window.products.find(px => (px.name || '').toLowerCase() === (it.name || '').toLowerCase());
                const id = p ? p.id : null;
                if (id) {
                    newCart[id] = it.qty;
                } else {
                    unavailableItems.push(it.name || 'Unknown item');
                }
            });
            setCart(newCart);
            updateCart();

            if (unavailableItems.length > 0) {
                const itemNames = unavailableItems.slice(0, 3).join(', ');
                const more = unavailableItems.length > 3 ? ` and ${unavailableItems.length - 3} more` : '';
                showToast(`Some items unavailable: ${itemNames}${more}`, 'warning');
            } else {
                showToast('✓ Order loaded', 'success');
            }
        }
    }

    // Handle schedule flag
    if (params.get('schedule') === '1') {
        setTimeout(() => {
            const d = new Date();
            const dateEl = document.getElementById('scheduleDate');
            dateEl.min = '';
            dateEl.max = '';
            dateEl.value = d.toISOString().slice(0, 10);
            document.getElementById('scheduleTime').value = '';
            openSheet('scheduleSheet');
        }, 300);
    }

    // Wire navigation links
    wireNavigationLinks(userId);

    // Setup map tabs
    setupMapTabs();

    // Initialize customer info management
    if (window.initCustomerInfoManagement) {
        window.initCustomerInfoManagement();
    }

    if (window.loadActiveOrder) {
        await Promise.resolve(window.loadActiveOrder());
    }
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

    // Clear inline validation errors on input
    document.addEventListener('input', e => {
        if (e.target.classList.contains('invalid')) {
            e.target.classList.remove('invalid');
            const errEl = document.getElementById(e.target.id + 'Error');
            if (errEl) { errEl.textContent = ''; errEl.style.display = 'none'; }
        }
    });

    // Wire checkout and back buttons
    document.getElementById('barCheckout').addEventListener('click', openDeliveryForm);
    document.getElementById('backToCartBtn').addEventListener('click', backToCart);
    document.getElementById('placeOrderBtn').addEventListener('click', submitOrder);

    const refreshReviewsBtn = document.getElementById('refreshReviewsBtn');
    if (refreshReviewsBtn) {
        refreshReviewsBtn.addEventListener('click', () => {
            const reviewsSection = document.getElementById('reviewsSection');
            const toggleReviewsBtn = document.getElementById('toggleReviewsBtn');
            if (reviewsSection && reviewsSection.classList.contains('reviews-collapsed')) {
                reviewsSection.classList.remove('reviews-collapsed');
                if (toggleReviewsBtn) toggleReviewsBtn.textContent = 'Hide';
                refreshReviewsBtn.style.display = '';
            }
            if (typeof loadLatestReviews === 'function') {
                reviewsSection && (reviewsSection.dataset.reviewsLoaded = '1');
                loadLatestReviews();
            }
        });
    }

    const toggleReviewsBtn = document.getElementById('toggleReviewsBtn');
    const reviewsSection = document.getElementById('reviewsSection');
    if (toggleReviewsBtn && reviewsSection) {
        if (!reviewsSection.classList.contains('reviews-collapsed')) {
            reviewsSection.classList.add('reviews-collapsed');
        }
        toggleReviewsBtn.textContent = 'Show';
        if (refreshReviewsBtn) refreshReviewsBtn.style.display = 'none';

        toggleReviewsBtn.addEventListener('click', () => {
            const isCollapsed = reviewsSection.classList.toggle('reviews-collapsed');
            toggleReviewsBtn.textContent = isCollapsed ? 'Show' : 'Hide';
            if (refreshReviewsBtn) refreshReviewsBtn.style.display = isCollapsed ? 'none' : '';
            if (!isCollapsed && typeof loadLatestReviews === 'function') {
                if (reviewsSection.dataset.reviewsLoaded !== '1') {
                    reviewsSection.dataset.reviewsLoaded = '1';
                    loadLatestReviews();
                }
            }
        });
    }

    // Wire clear cart button
    const clearBtn = document.getElementById('barClear');
    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            const currentCart = window.getCart ? window.getCart() : {};
            if (Object.keys(currentCart).length === 0) {
                showToast('Cart is already empty');
                return;
            }
            if (confirm('Clear all items from cart?')) {
                window.setCart({});
                window.updateCart();
                showToast('Cart cleared', 'success');
            }
        });
    }

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
