/**
 * BakeFlow - Delivery Form Module
 * Address, location, and delivery management
 */

let deliveryType = null;
let geo = null;
let map = null;
let marker = null;

// ========== Delivery Form ==========
function openDeliveryForm() {
    document.querySelectorAll('.section').forEach(s => {
        if (!s.classList.contains('delivery-form')) {
            s.style.display = 'none';
        }
    });
    document.getElementById('savedPlacesBar').style.display = 'none';
    document.getElementById('recurringList').style.display = 'none';
    
    document.getElementById('deliveryForm').classList.add('active');
    document.getElementById('cartBar').style.display = 'none';
    document.getElementById('orderBar').style.display = 'block';
    adjustSafePadding();
    if (window.lucide && window.lucide.createIcons) window.lucide.createIcons();
}

function backToCart() {
    document.getElementById('deliveryForm').classList.remove('active');
    document.querySelectorAll('.section').forEach(s => s.style.display = '');
    document.getElementById('cartBar').style.display = 'block';
    document.getElementById('orderBar').style.display = 'none';
    renderSavedOrders();
    renderPlacesBar();
    renderRecurringOrders();
    adjustSafePadding();
}

function selectDeliveryType(type) {
    deliveryType = type;
    document.querySelectorAll('.radio-option').forEach(opt => opt.classList.remove('selected'));
    document.querySelector(`[data-type="${type}"]`).classList.add('selected');
    
    const addressGroup = document.getElementById('addressGroup');
    const addressInput = document.getElementById('customerAddress');
    if (type === 'delivery') {
        addressGroup.style.display = 'block';
        addressInput.required = true;
    } else {
        addressGroup.style.display = 'none';
        addressInput.required = false;
    }
}

// ========== Location & Map ==========
async function useMyLocation() {
    if (!navigator.geolocation) { 
        showError('Location is not available on this device'); 
        return; 
    }
    
    try {
        const pos = await new Promise((resolve, reject) => {
            navigator.geolocation.getCurrentPosition(resolve, reject, { 
                enableHighAccuracy: true, 
                timeout: 10000, 
                maximumAge: 0 
            });
        });
        const { latitude: lat, longitude: lon } = pos.coords;
        const address = await reverseGeocode(lat, lon);
        document.getElementById('customerAddress').value = address;
        selectDeliveryType('delivery');
        document.getElementById('addressGroup').style.display = 'block';
        geo = { lat, lon, address };
        showMap(lat, lon);
        activateTab('map');
        showToast('Location found', 'success');
    } catch(e) {
        showError('Unable to get location. Check permissions.');
    }
}

async function reverseGeocode(lat, lon) {
    const url = `https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
    const res = await fetch(url, { headers: { 'Accept': 'application/json' } });
    const j = await res.json();
    return j.display_name || `${lat.toFixed(5)}, ${lon.toFixed(5)}`;
}

async function findOnMap() {
    const q = document.getElementById('customerAddress').value.trim();
    if (!q) { 
        showError('Enter an address to find on map'); 
        return; 
    }
    
    try {
        const url = `https://nominatim.openstreetmap.org/search?format=jsonv2&q=${encodeURIComponent(q)}`;
        const res = await fetch(url, { headers: { 'Accept': 'application/json' } });
        const list = await res.json();
        if (!list || !list.length) { 
            showError('Address not found'); 
            return; 
        }
        const { lat, lon, display_name } = list[0];
        const latNum = parseFloat(lat), lonNum = parseFloat(lon);
        document.getElementById('customerAddress').value = display_name;
        geo = { lat: latNum, lon: lonNum, address: display_name };
        selectDeliveryType('delivery');
        document.getElementById('addressGroup').style.display = 'block';
        showMap(latNum, lonNum);
        activateTab('map');
    } catch(e) { 
        showError('Search failed. Try again.'); 
    }
}

function showMap(lat, lon) {
    const el = document.getElementById('map');
    if (!el || !window.L) return;
    
    if (!map) {
        map = L.map('map', { zoomControl: false }).setView([lat, lon], 16);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { 
            attribution: '© OpenStreetMap' 
        }).addTo(map);
        marker = L.marker([lat, lon], { draggable: true }).addTo(map);
        marker.on('dragend', async () => {
            const p = marker.getLatLng();
            const address = await reverseGeocode(p.lat, p.lng);
            document.getElementById('customerAddress').value = address;
            geo = { lat: p.lat, lon: p.lng, address };
        });
    } else {
        map.setView([lat, lon], 16);
        if (marker) { 
            marker.setLatLng([lat, lon]); 
        } else { 
            marker = L.marker([lat, lon], { draggable: true }).addTo(map); 
        }
    }
}

function activateTab(which) {
    const tabMap = document.getElementById('tabMap');
    const tabDir = document.getElementById('tabDirections');
    const mapEl = document.getElementById('map');
    const dirEl = document.getElementById('deliveryDirections');
    if (!tabMap || !tabDir || !mapEl || !dirEl) return;
    
    if (which === 'map') {
        tabMap.classList.add('active'); 
        tabDir.classList.remove('active');
        mapEl.style.display = 'block'; 
        dirEl.style.display = 'none';
        if (map) setTimeout(() => map.invalidateSize(), 100);
    } else {
        tabDir.classList.add('active'); 
        tabMap.classList.remove('active');
        mapEl.style.display = 'none'; 
        dirEl.style.display = 'block';
    }
}

function setupMapTabs() {
    const hasLeaflet = !!window.L;
    activateTab(hasLeaflet ? 'map' : 'dir');
    
    const tabMap = document.getElementById('tabMap');
    const tabDir = document.getElementById('tabDirections');
    if (tabMap) tabMap.addEventListener('click', () => activateTab('map'));
    if (tabDir) tabDir.addEventListener('click', () => activateTab('dir'));
}

function appendNote(text) {
    const el = document.getElementById('orderNotes');
    el.value = el.value ? (el.value + ' · ' + text) : text;
}

function showError(msg) {
    const b = document.getElementById('errorBanner');
    b.textContent = msg;
    b.style.display = 'block';
    setTimeout(() => { b.style.display = 'none'; }, 3000);
}

function getDeliveryType() { return deliveryType; }
function getGeo() { return geo; }

// Export
window.openDeliveryForm = openDeliveryForm;
window.backToCart = backToCart;
window.selectDeliveryType = selectDeliveryType;
window.useMyLocation = useMyLocation;
window.findOnMap = findOnMap;
window.appendNote = appendNote;
window.showError = showError;
window.setupMapTabs = setupMapTabs;
window.getDeliveryType = getDeliveryType;
window.getGeo = getGeo;
