/**
 * BakeFlow - Scheduling Module
 * Handles schedule and recurring orders functionality
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
        const d = new Date();
        document.getElementById('scheduleDate').value = d.toISOString().slice(0, 10);
        document.getElementById('scheduleTime').value = '';
        openSheet('scheduleSheet');
    });
    
    document.getElementById('scheduleCancel').addEventListener('click', closeSheets);
    
    document.getElementById('scheduleConfirm').addEventListener('click', () => {
        const date = document.getElementById('scheduleDate').value;
        const time = document.getElementById('scheduleTime').value;
        
        if (!date || !time) { showToast('Pick date and time'); return; }
        if (!isValidBusinessTime(date, time)) { showToast('Outside business hours'); return; }
        
        const userId = getUserId();
        pendingSchedule = { type: deliveryType || 'pickup', date, time };
        localStorage.setItem(`pending_schedule_${userId}`, JSON.stringify(pendingSchedule));
        updateCart();
        closeSheets();
        showToast('Schedule applied', 'success');
    });
}

function initRecurringSheet() {
    document.getElementById('barRecurring').addEventListener('click', () => {
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
    
    document.getElementById('recurringCancel').addEventListener('click', closeSheets);
    
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
        showToast('Recurring scheduled');
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
        const upcoming = computeUpcoming(r).slice(0, 3).join(' • ');
        return `
        <div style="display:flex; align-items:center; justify-content:space-between; gap:8px;">
            <div style="font-size:13px;">
                <div style="font-weight:700;">${escapeHtml(r.name) || 'Recurring Order'}</div>
                <div style="color:#666;">${r.days.join(', ')} · ${r.time} · from ${r.start}${r.end ? ' to ' + r.end : ''}</div>
                <div style="color:#888; font-size:12px;">Next: ${upcoming || '—'}</div>
            </div>
            <div style="display:flex; gap:6px;">
                <button style="padding:8px 10px; border:1px solid #e7eaf3; border-radius:10px; background:#fff; display:flex; align-items:center; gap:6px; cursor:pointer;" onclick="editRecurring(${i})"><i data-lucide="pencil"></i> Edit</button>
                <button style="padding:8px 10px; border:1px solid #e7eaf3; border-radius:10px; background:#fff; display:flex; align-items:center; gap:6px; cursor:pointer;" onclick="pauseRecurring(${i})"><i data-lucide="${r.paused ? 'play' : 'pause'}"></i> ${r.paused ? 'Resume' : 'Pause'}</button>
                <button style="padding:8px 10px; border:1px solid #e7eaf3; border-radius:10px; background:#fff; display:flex; align-items:center; gap:6px; cursor:pointer;" onclick="cancelRecurring(${i})"><i data-lucide="trash-2"></i></button>
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
                out.push(occ.toLocaleString());
            }
        }
        cur.setDate(cur.getDate() + 1);
    }
    return out;
}

// Edit recurring order
let editingRecurringIdx = -1;

window.editRecurring = function(i) {
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    const r = list[i];
    if (!r) return;
    
    editingRecurringIdx = i;
    
    // Pre-fill form with existing values
    document.getElementById('recurringName').value = r.name || '';
    ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'].forEach(d => {
        const el = document.getElementById('day' + d);
        if (el) el.checked = r.days.includes(d);
    });
    document.getElementById('recurringTime').value = r.time || '';
    document.getElementById('recStart').value = r.start || '';
    document.getElementById('recEnd').value = r.end || '';
    
    // Change button to "Update"
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
        list[editingRecurringIdx] = { 
            ...list[editingRecurringIdx], 
            name, days, time, start, end 
        };
        localStorage.setItem(key, JSON.stringify(list));
    }
    
    // Reset button
    const confirmBtn = document.getElementById('recurringConfirm');
    confirmBtn.textContent = 'Add';
    confirmBtn.onclick = null;
    editingRecurringIdx = -1;
    
    renderRecurringOrders();
    closeSheets();
    showToast('Recurring updated');
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

window.addRecurringOrder = function(cfg) {
    const userId = getUserId();
    const key = `recurring_orders_${userId}`;
    const list = JSON.parse(localStorage.getItem(key) || '[]');
    list.push({
        name: cfg.name || 'Recurring Order',
        days: cfg.days || ['Friday'],
        time: cfg.time || '15:00',
        start: cfg.start || new Date().toISOString().slice(0, 10),
        end: cfg.end || '',
        paused: false
    });
    localStorage.setItem(key, JSON.stringify(list));
    renderRecurringOrders();
    showToast('Recurring order added');
};
