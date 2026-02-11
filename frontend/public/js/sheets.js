/**
 * BakeFlow - Sheet UI Module
 * Bottom sheets and toast notifications
 */

function openSheet(id) {
    const backdrop = document.getElementById('sheetBackdrop');
    backdrop.classList.add('active');
    document.querySelectorAll('.sheet').forEach(s => s.classList.remove('active'));
    const el = document.getElementById(id);
    if (el) el.classList.add('active');
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

function closeSheets() {
    document.getElementById('sheetBackdrop').classList.remove('active');
    document.querySelectorAll('.sheet').forEach(s => s.classList.remove('active'));
    
    // Reset recurring sheet button if needed
    const confirmBtn = document.getElementById('recurringConfirm');
    if (confirmBtn) {
        confirmBtn.textContent = 'Add';
        confirmBtn.onclick = null;
    }
}

function showToast(msg, type) {
    const t = document.getElementById('toast');
    // Reset classes first for re-trigger
    t.classList.remove('show', 'success', 'warning');
    void t.offsetWidth; // force reflow so transition re-triggers
    t.textContent = msg;
    t.classList.add('show');
    if (type === 'success') t.classList.add('success');
    if (type === 'warning') t.classList.add('warning');
    clearTimeout(window._toastTimer);
    window._toastTimer = setTimeout(() => t.classList.remove('show', 'success', 'warning'), 2500);
}

function initSheetListeners() {
    document.getElementById('sheetBackdrop').addEventListener('click', closeSheets);
    window.addEventListener('keydown', (e) => { 
        if (e.key === 'Escape') closeSheets(); 
    });
}

// Export for use in other modules
window.openSheet = openSheet;
window.closeSheets = closeSheets;
window.showToast = showToast;
window.initSheetListeners = initSheetListeners;
