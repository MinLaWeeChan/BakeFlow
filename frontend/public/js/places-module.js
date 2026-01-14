/**
 * BakeFlow - Saved Places Module
 * Handles saved delivery locations functionality
 */

function getPlaces() {
    const key = `saved_places_${getUserId()}`;
    return JSON.parse(localStorage.getItem(key) || '[]');
}

function setPlaces(list) {
    const key = `saved_places_${getUserId()}`;
    localStorage.setItem(key, JSON.stringify(list));
}

function renderPlaces() {
    const list = getPlaces();
    const wrap = document.getElementById('placesList');
    if (!wrap) return;
    
    if (!list.length) {
        wrap.innerHTML = `<div class="p-card" style="padding:12px; box-shadow:none; border:1px dashed #e7eaf3; color:#666; font-weight:700;">No saved places yet.</div>`;
        return;
    }
    
    wrap.innerHTML = list.map((p, i) => {
        const addr = escapeHtml(p.address || '');
        const name = escapeHtml(p.name || 'Saved place');
        return `
            <div class="p-card" style="padding:12px; box-shadow:none; border:1px solid #e7eaf3; align-items:flex-start;">
                <div style="flex:1; min-width:0;">
                    <div style="font-weight:800; color:#111; font-size:14px; display:flex; align-items:center; gap:6px;"><i data-lucide="bookmark"></i> ${name}</div>
                    <div style="color:#666; font-size:12px; margin-top:4px;">${addr || '—'}</div>
                    ${p.directions ? `<div style="color:#888; font-size:12px; margin-top:4px;">Directions: ${escapeHtml(p.directions)}</div>` : ''}
                </div>
                <div style="display:flex; flex-direction:column; gap:8px;">
                    <button class="sheet-btn primary" style="height:40px;" onclick="applyPlace(${i})"><i data-lucide="check"></i> Use</button>
                    <button class="sheet-btn" style="height:40px;" onclick="deletePlace(${i})"><i data-lucide="trash-2"></i> Delete</button>
                </div>
            </div>
        `;
    }).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

function renderPlacesBar() {
    const list = getPlaces();
    const bar = document.getElementById('savedPlacesBar');
    const chips = document.getElementById('savedPlacesChips');
    if (!bar || !chips) return;
    
    if (!list.length) { bar.style.display = 'none'; return; }
    
    bar.style.display = 'block';
    chips.innerHTML = list.slice(0, 6).map((p, i) => `
        <button class="mini-btn" style="white-space:nowrap;" onclick="quickUsePlace(${i})">
            <i data-lucide="map-pin"></i>
            <span>${escapeHtml(p.name || 'Place')}</span>
        </button>
    `).join('');
    
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

window.quickUsePlace = function(i) {
    const list = getPlaces();
    const p = list[i];
    if (!p) return;
    
    openDeliveryForm();
    selectDeliveryType('delivery');
    document.getElementById('addressGroup').style.display = 'block';
    document.getElementById('customerAddress').value = p.address || '';
    
    const dirEl = document.getElementById('deliveryDirections');
    if (dirEl) dirEl.value = p.directions || '';
    
    geo = p.geo || null;
    if (geo && typeof geo.lat === 'number' && typeof geo.lon === 'number') {
        showMap(geo.lat, geo.lon);
        activateTab('map');
    } else {
        activateTab('dir');
    }
    showToast('Place loaded');
};

window.applyPlace = async function(i) {
    const list = getPlaces();
    const p = list[i];
    if (!p) return;
    
    selectDeliveryType('delivery');
    document.getElementById('addressGroup').style.display = 'block';
    document.getElementById('customerAddress').value = p.address || '';
    
    const dirEl = document.getElementById('deliveryDirections');
    if (dirEl) dirEl.value = p.directions || '';
    
    geo = p.geo || null;
    if (geo && typeof geo.lat === 'number' && typeof geo.lon === 'number') {
        showMap(geo.lat, geo.lon);
    }
    closeSheets();
    showToast('Place applied');
};

window.deletePlace = function(i) {
    const list = getPlaces();
    list.splice(i, 1);
    setPlaces(list);
    renderPlaces();
    renderPlacesBar();
    showToast('Deleted');
};

function initPlacesSheet() {
    const openPlacesBtn = document.getElementById('openPlacesBtn');
    if (openPlacesBtn) {
        openPlacesBtn.addEventListener('click', () => {
            const nameEl = document.getElementById('placeName');
            const countEl = document.getElementById('placeCount');
            if (nameEl) nameEl.value = '';
            if (countEl) countEl.textContent = '0/30';
            renderPlaces();
            openSheet('placesSheet');
        });
    }
    
    const managePlacesBtn = document.getElementById('managePlacesBtn');
    if (managePlacesBtn) {
        managePlacesBtn.addEventListener('click', () => {
            const nameEl = document.getElementById('placeName');
            const countEl = document.getElementById('placeCount');
            if (nameEl) nameEl.value = '';
            if (countEl) countEl.textContent = '0/30';
            renderPlaces();
            openSheet('placesSheet');
        });
    }
    
    const placeNameEl = document.getElementById('placeName');
    if (placeNameEl) {
        placeNameEl.addEventListener('input', (e) => {
            const v = e.target.value || '';
            const c = document.getElementById('placeCount');
            if (c) c.textContent = `${v.length}/30`;
        });
    }
    
    const placeCloseBtn = document.getElementById('placeCloseBtn');
    if (placeCloseBtn) placeCloseBtn.addEventListener('click', closeSheets);
    
    const placeSaveBtn = document.getElementById('placeSaveBtn');
    if (placeSaveBtn) {
        placeSaveBtn.addEventListener('click', () => {
            const address = (document.getElementById('customerAddress')?.value || '').trim();
            const directions = (document.getElementById('deliveryDirections')?.value || '').trim();
            const name = (document.getElementById('placeName')?.value || '').trim();
            
            if (!address) { showToast('Enter an address first'); return; }
            if (!name) { showToast('Name your place'); return; }
            
            const list = getPlaces();
            list.unshift({ name, address, directions, geo: geo || null, created_at: new Date().toISOString() });
            setPlaces(list.slice(0, 20));
            renderPlaces();
            renderPlacesBar();
            showToast('Saved place');
        });
    }
}
