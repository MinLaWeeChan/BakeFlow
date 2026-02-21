/**
 * BakeFlow - Scheduling Module
 * Schedule and recurring orders
 */

function isValidBusinessTime(dateStr, timeStr) {
    const [hh, mm] = timeStr.split(':').map(Number);
    if (Number.isNaN(hh) || Number.isNaN(mm)) return false;
    if (hh < 8 || hh > 20 || (hh === 20 && mm > 0)) return false;
    const when = new Date(`${dateStr}T${timeStr}:00`);
    return when.getTime() > Date.now();
}

function initScheduleSheet() {
    document.getElementById('barSchedule').addEventListener('click', () => {
        const dateEl = document.getElementById('scheduleDate');
        const timeEl = document.getElementById('scheduleTime');
        // Schedule must be for today or later
        const today = new Date().toISOString().slice(0, 10);
        dateEl.min = today;
        dateEl.max = '';
        // Pre-fill with existing schedule if set, otherwise default to today
        const existing = window.getPendingSchedule ? window.getPendingSchedule() : null;
        if (existing && existing.date) {
            dateEl.value = existing.date;
            timeEl.value = existing.time || '';
        } else {
            dateEl.value = new Date().toISOString().slice(0, 10);
            timeEl.value = '';
        }
        // Show/hide the Remove button
        const removeBtn = document.getElementById('scheduleRemove');
        if (removeBtn) removeBtn.style.display = (existing && existing.date) ? '' : 'none';
        openSheet('scheduleSheet');
    });
    
    document.getElementById('scheduleCancel').addEventListener('click', closeSheets);

    // Remove schedule button
    document.getElementById('scheduleRemove').addEventListener('click', () => {
        if (window.setPendingSchedule) window.setPendingSchedule(null);
        const userId = typeof getUserId === 'function' ? getUserId() : '';
        if (userId) localStorage.removeItem(`pending_schedule_${userId}`);
        window.updateCart();
        closeSheets();
        showToast('Schedule removed', 'success');
    });
    
    document.getElementById('scheduleConfirm').addEventListener('click', () => {
        const date = document.getElementById('scheduleDate').value;
        const time = document.getElementById('scheduleTime').value;
        
        if (!date || !time) { showToast('Pick date and time'); return; }
        if (!isValidBusinessTime(date, time)) { showToast('Outside business hours (8am-8pm)'); return; }
        
        const userId = getUserId();
        window.setPendingSchedule({ type: window.getDeliveryType() || 'pickup', date, time });
        localStorage.setItem(`pending_schedule_${userId}`, JSON.stringify(window.getPendingSchedule()));
        window.updateCart();
        closeSheets();
        showToast('Schedule applied', 'success');
    });
}

// ========== Clear Cart ==========
function initClearCartButton() {
    const clearBtn = document.getElementById('barClear');
    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            if (Object.keys(window.getCart ? window.getCart() : {}).length === 0) {
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
}

// ========== Recurring Orders ==========
let editingRecurringIdx = -1;

function initRecurringSheet() {
    // Initialize clear cart button
    initClearCartButton();
    
    // Keep recurring functionality for saved recurring orders
    const barRecurring = document.getElementById('barRecurring');
    if (barRecurring) {
        barRecurring.addEventListener('click', () => {
            document.getElementById('recurringName').value = '';
            ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'].forEach(d => {
                const el = document.getElementById('day' + d);
                if (el) el.checked = false;
            });
            document.getElementById('recurringTime').value = '';
            document.getElementById('recStart').value = new Date().toISOString().slice(0, 10);
            document.getElementById('recEnd').value = '';
            openSheet('recurringSheet');
        });
    }
    
    document.getElementById('recurringCancel')?.addEventListener('click', closeSheets);
    
    document.getElementById('recurringConfirm').addEventListener('click', () => {
        const name = document.getElementById('recurringName').value.trim();
        const days = ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'].filter(d => 
            document.getElementById('day' + d)?.checked
        );
        
        if (!days.length) { showToast('Pick at least one day'); return; }
        
        const time = document.getElementById('recurringTime').value;
        const start = document.getElementById('recStart').value || new Date().toISOString().slice(0, 10);
        const end = document.getElementById('recEnd').value || '';
        
        if (!time) { showToast('Pick a time'); return; }
        if (!isValidBusinessTime(start, time)) { showToast('Outside business hours'); return; }
        
        const userId = getUserId();
        const key = `recurring_orders_${userId}`;
        const list = JSON.parse(localStorage.getItem(key) || '[]');
        list.push({ name, days, time, start, end, paused: false });
        localStorage.setItem(key, JSON.stringify(list));
        
        renderRecurringOrders();
        closeSheets();
        showToast('Recurring scheduled', 'success');
    });
}

function renderRecurringOrders() {
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    const wrap = document.getElementById('recurringItems');
    const box = document.getElementById('recurringList');
    
    if (!list.length) { 
        if (box) box.style.display = 'none'; 
        return; 
    }
    
    box.style.display = 'block';
    wrap.innerHTML = list.map((r, i) => {
        const upcoming = computeUpcoming(r).slice(0, 2).join(' • ');
        return `
        <div class="recurring-item">
            <div class="recurring-info">
                <div class="recurring-name">${escapeHtml(r.name) || 'Recurring Order'}</div>
                <div class="recurring-schedule">${r.days.join(', ')} · ${r.time}</div>
                <div class="recurring-next">${r.paused ? '⏸️ Paused' : `Next: ${upcoming || '—'}`}</div>
            </div>
            <div class="recurring-actions">
                <button class="recurring-action-btn" onclick="editRecurring(${i})">
                    <i data-lucide="pencil"></i>
                </button>
                <button class="recurring-action-btn" onclick="pauseRecurring(${i})">
                    <i data-lucide="${r.paused ? 'play' : 'pause'}"></i>
                </button>
                <button class="recurring-action-btn" onclick="cancelRecurring(${i})">
                    <i data-lucide="trash-2"></i>
                </button>
            </div>
        </div>`;
    }).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

function computeUpcoming(r) {
    const map = { 'Sun':0, 'Mon':1, 'Tue':2, 'Wed':3, 'Thu':4, 'Fri':5, 'Sat':6 };
    const daysIdx = r.days.map(d => map[d] ?? -1).filter(d => d >= 0);
    const startDate = new Date(r.start + 'T' + (r.time || '09:00') + ':00');
    const endDate = r.end ? new Date(r.end + 'T23:59:00') : null;
    let cur = new Date();
    const out = [];
    
    for (let i = 0; i < 30 && out.length < 5; i++) {
        const d = new Date(cur);
        if (daysIdx.includes(d.getDay())) {
            const occ = new Date(d.toISOString().slice(0, 10) + 'T' + r.time + ':00');
            if (occ >= startDate && (!endDate || occ <= endDate) && !r.paused) {
                out.push(occ.toLocaleDateString());
            }
        }
        cur.setDate(cur.getDate() + 1);
    }
    return out;
}

window.editRecurring = function(i) {
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    const r = list[i];
    if (!r) return;
    
    editingRecurringIdx = i;
    
    document.getElementById('recurringName').value = r.name || '';
    ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'].forEach(d => {
        const el = document.getElementById('day' + d);
        if (el) el.checked = r.days.includes(d);
    });
    document.getElementById('recurringTime').value = r.time || '';
    document.getElementById('recStart').value = r.start || '';
    document.getElementById('recEnd').value = r.end || '';
    
    const confirmBtn = document.getElementById('recurringConfirm');
    confirmBtn.textContent = 'Update';
    confirmBtn.onclick = updateRecurring;
    
    openSheet('recurringSheet');
};

function updateRecurring() {
    if (editingRecurringIdx < 0) return;
    
    const name = document.getElementById('recurringName').value.trim();
    const days = ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'].filter(d => 
        document.getElementById('day' + d)?.checked
    );
    
    if (!days.length) { showToast('Pick at least one day'); return; }
    
    const time = document.getElementById('recurringTime').value;
    const start = document.getElementById('recStart').value || new Date().toISOString().slice(0, 10);
    const end = document.getElementById('recEnd').value || '';
    
    if (!time) { showToast('Pick a time'); return; }
    if (!isValidBusinessTime(start, time)) { showToast('Outside business hours'); return; }
    
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    
    if (list[editingRecurringIdx]) {
        list[editingRecurringIdx] = { ...list[editingRecurringIdx], name, days, time, start, end };
        localStorage.setItem(key, JSON.stringify(list));
    }
    
    const confirmBtn = document.getElementById('recurringConfirm');
    confirmBtn.textContent = 'Add';
    confirmBtn.onclick = null;
    editingRecurringIdx = -1;
    
    renderRecurringOrders();
    closeSheets();
    showToast('Recurring updated', 'success');
}

window.pauseRecurring = function(i) {
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    if (!list[i]) return;
    
    list[i].paused = !list[i].paused;
    localStorage.setItem(key, JSON.stringify(list));
    renderRecurringOrders();
    showToast(list[i].paused ? 'Recurring paused' : 'Recurring resumed');
};

window.cancelRecurring = function(i) {
    if (!confirm('Cancel this recurring order?')) return;
    
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    list.splice(i, 1);
    localStorage.setItem(key, JSON.stringify(list));
    renderRecurringOrders();
    showToast('Recurring cancelled');
};

// Export
window.initScheduleSheet = initScheduleSheet;
window.initRecurringSheet = initRecurringSheet;
window.renderRecurringOrders = renderRecurringOrders;
