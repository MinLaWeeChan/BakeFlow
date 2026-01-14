/**
 * BakeFlow - Bottom Sheet UI Module
 * Handles sheet open/close and toast notifications
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
    t.textContent = msg;
    t.className = 'toast show';
    if (type === 'success') t.classList.add('success');
    setTimeout(() => t.classList.remove('show', 'success'), 1800);
}

function initSheetListeners() {
    document.getElementById('sheetBackdrop').addEventListener('click', closeSheets);
    window.addEventListener('keydown', (e) => { 
        if (e.key === 'Escape') closeSheets(); 
    });
}
