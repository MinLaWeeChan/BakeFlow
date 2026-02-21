/**
 * BakeFlow - Customer Info Management Module
 * Save and reuse customer information across orders
 */

const CUSTOMER_INFO_STORAGE_KEY = 'bf_saved_customer_info';
const CUSTOMER_INFO_TIMESTAMP_KEY = 'bf_saved_customer_info_timestamp';

/**
 * Save current customer information to localStorage
 */
function saveCustomerInfo() {
    const name = document.getElementById('customerName').value.trim();
    const phone = document.getElementById('customerPhone').value.trim();
    const address = document.getElementById('customerAddress').value.trim();
    const deliveryType = window.getDeliveryType ? window.getDeliveryType() : null;

    if (!name || !phone) {
        window.showToast('Please fill in name and phone before saving', 'info');
        return false;
    }

    const customerInfo = {
        name: name,
        phone: phone,
        address: address,
        deliveryType: deliveryType,
        savedAt: new Date().toISOString()
    };

    try {
        localStorage.setItem(CUSTOMER_INFO_STORAGE_KEY, JSON.stringify(customerInfo));
        localStorage.setItem(CUSTOMER_INFO_TIMESTAMP_KEY, Date.now().toString());
        window.showToast('Customer info saved! ✓', 'success');
        updateLoadCustomerButton();
        return true;
    } catch (e) {
        console.error('Failed to save customer info:', e);
        window.showToast('Failed to save customer info', 'error');
        return false;
    }
}

/**
 * Load saved customer information from localStorage
 */
function loadCustomerInfo() {
    try {
        const saved = localStorage.getItem(CUSTOMER_INFO_STORAGE_KEY);
        if (!saved) {
            window.showToast('No saved customer info found', 'info');
            return false;
        }

        const customerInfo = JSON.parse(saved);
        const nameEl = document.getElementById('customerName');
        const phoneEl = document.getElementById('customerPhone');
        const addressEl = document.getElementById('customerAddress');

        if (nameEl) nameEl.value = customerInfo.name || '';
        if (phoneEl) phoneEl.value = customerInfo.phone || '';
        if (addressEl) addressEl.value = customerInfo.address || '';

        // Set delivery type if saved
        if (customerInfo.deliveryType && window.selectDeliveryType) {
            window.selectDeliveryType(customerInfo.deliveryType);
        }

        // Update phone UI validation
        if (window.updatePhoneUi) {
            window.updatePhoneUi();
        }

        window.showToast('Customer info loaded! ✓', 'success');
        return true;
    } catch (e) {
        console.error('Failed to load customer info:', e);
        window.showToast('Failed to load customer info', 'error');
        return false;
    }
}

/**
 * Clear saved customer information
 */
function clearSavedCustomerInfo() {
    try {
        localStorage.removeItem(CUSTOMER_INFO_STORAGE_KEY);
        localStorage.removeItem(CUSTOMER_INFO_TIMESTAMP_KEY);
        window.showToast('Saved customer info cleared', 'success');
        updateLoadCustomerButton();
        return true;
    } catch (e) {
        console.error('Failed to clear saved customer info:', e);
        return false;
    }
}

/**
 * Check if customer info is saved
 */
function hasSavedCustomerInfo() {
    try {
        const saved = localStorage.getItem(CUSTOMER_INFO_STORAGE_KEY);
        return !!saved;
    } catch (e) {
        return false;
    }
}

/**
 * Get saved customer info (returns null if not found)
 */
function getSavedCustomerInfo() {
    try {
        const saved = localStorage.getItem(CUSTOMER_INFO_STORAGE_KEY);
        if (!saved) return null;
        return JSON.parse(saved);
    } catch (e) {
        return null;
    }
}

/**
 * Update the load button visibility based on saved data
 */
function updateLoadCustomerButton() {
    const loadBtn = document.getElementById('loadCustomerBtn');
    const clearBtn = document.getElementById('clearCustomerBtn');
    const hasSaved = hasSavedCustomerInfo();

    if (loadBtn) {
        loadBtn.style.display = hasSaved ? 'block' : 'none';
    }
    if (clearBtn) {
        clearBtn.style.display = hasSaved ? 'block' : 'none';
    }
}

/**
 * Initialize customer info management
 */
function initCustomerInfoManagement() {
    const saveBtn = document.getElementById('saveCustomerBtn');
    const loadBtn = document.getElementById('loadCustomerBtn');
    const clearBtn = document.getElementById('clearCustomerBtn');

    if (saveBtn) {
        saveBtn.addEventListener('click', saveCustomerInfo);
    }

    if (loadBtn) {
        loadBtn.addEventListener('click', loadCustomerInfo);
    }

    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            if (confirm('Clear saved customer info?')) {
                clearSavedCustomerInfo();
            }
        });
    }

    // Update button visibility on load
    updateLoadCustomerButton();
}

// Export functions
window.saveCustomerInfo = saveCustomerInfo;
window.loadCustomerInfo = loadCustomerInfo;
window.clearSavedCustomerInfo = clearSavedCustomerInfo;
window.hasSavedCustomerInfo = hasSavedCustomerInfo;
window.getSavedCustomerInfo = getSavedCustomerInfo;
window.updateLoadCustomerButton = updateLoadCustomerButton;
window.initCustomerInfoManagement = initCustomerInfoManagement;
