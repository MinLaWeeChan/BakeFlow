/**
 * BakeFlow - Utility Functions
 */

function getUserId() {
    const params = new URLSearchParams(window.location.search);
    const userId = params.get('user_id');
    if (userId) return userId;

    const tok = getWebviewToken();
    if (tok) return `t_${tok.slice(0, 16)}`;

    return 'guest';
}

function getWebviewToken() {
    return new URLSearchParams(window.location.search).get('t') || '';
}

function goToSavedOrders(userId) {
    const uid = userId || new URLSearchParams(window.location.search).get('user_id') || '';
    const tok = getWebviewToken();
    const origin = window.location.origin;

    const qs = new URLSearchParams();
    if (uid) qs.set('user_id', uid);
    if (tok) qs.set('t', tok);

    const targetUrl = origin + '/saved-orders.html' + (qs.toString() ? ('?' + qs.toString()) : '');
    console.log('[BakeFlow] Navigating to:', targetUrl);
    window.location.href = targetUrl;
}

function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function timeAgo(timestamp) {
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return 'Just now';
    const mins = Math.floor(seconds / 60);
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    if (days === 1) return 'Yesterday';
    if (days < 7) return `${days}d ago`;
    const weeks = Math.floor(days / 7);
    if (weeks < 4) return `${weeks}w ago`;
    return new Date(timestamp).toLocaleDateString();
}

function adjustSafePadding() {
    const cartBar = document.getElementById('cartBar');
    const orderBar = document.getElementById('orderBar');
    const h1 = cartBar && cartBar.style.display !== 'none' ? cartBar.offsetHeight : 0;
    const h2 = orderBar && orderBar.style.display !== 'none' ? orderBar.offsetHeight : 0;
    const h = Math.max(h1, h2);
    const pad = Math.max(h + 20, 140);
    document.documentElement.style.setProperty('--content-bottom-pad', pad + 'px');
}

function wireNavigationLinks(userId) {
    const headerLink = document.getElementById('headerSavedOrdersLink');
    if (headerLink) {
        headerLink.addEventListener('click', (e) => {
            e.preventDefault();
            goToSavedOrders(userId);
        });
    }
    
    document.querySelectorAll('.viewAllSavedLink').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            goToSavedOrders(userId);
        });
    });
}

// Export
window.getUserId = getUserId;
window.getWebviewToken = getWebviewToken;
window.goToSavedOrders = goToSavedOrders;
window.escapeHtml = escapeHtml;
window.timeAgo = timeAgo;
window.adjustSafePadding = adjustSafePadding;
window.wireNavigationLinks = wireNavigationLinks;
