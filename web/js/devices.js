const deviceCategoryLabels = {
    keyboards: 'Claviers',
    mice: 'Souris',
    hid_devices: 'Périphériques HID (détail)',
    cameras: 'Caméras',
    monitors: 'Écrans',
    graphics_cards: 'Cartes graphiques',
    audio_devices: 'Périphériques audio',
    usb_devices: 'Périphériques USB',
    storage_drives: 'Disques',
    network_adapters: 'Cartes réseau',
    printers: 'Imprimantes',
    memory_modules: 'Barrettes mémoire',
};

async function loadDevices() {
    const res = await authFetch('/api/devices');
    if (!res.ok) {
        document.getElementById('devices-categories').innerHTML = '<p class="meter-sub">Erreur lors du chargement des périphériques.</p>';
        return;
    }
    const data = await res.json();
    renderDevicesSummary(data);
    renderDevicesCategories(data);
}

function renderDevicesSummary(data) {
    const cards = [
        { label: 'Processeur', value: data.processor?.name, sub: data.processor?.extra },
        { label: 'Carte mère', value: data.motherboard?.name, sub: data.motherboard?.manufacturer },
        { label: 'BIOS', value: 'Version ' + (data.bios?.name || '—'), sub: data.bios?.manufacturer },
    ];

    document.getElementById('devices-summary').innerHTML = cards.map(c => `
        <div class="device-summary-card">
            <div class="label">${c.label}</div>
            <div class="value">${c.value || '—'}</div>
            ${c.sub ? `<div class="sub">${c.sub}</div>` : ''}
        </div>
    `).join('');
}

function renderDevicesCategories(data) {
    const container = document.getElementById('devices-categories');
    container.innerHTML = '';

    Object.entries(deviceCategoryLabels).forEach(([key, label]) => {
        const items = data[key] || [];
        if (items.length === 0) return;

        const block = document.createElement('div');
        block.className = 'device-category';
        block.innerHTML = `
            <div class="device-category-title">
                ${label} <span class="device-count">${items.length}</span>
            </div>
            <div class="device-list">
                ${items.map(deviceItemHTML).join('')}
            </div>
        `;
        container.appendChild(block);
    });

    if (container.innerHTML === '') {
        container.innerHTML = '<p class="meter-sub">Aucun périphérique détecté.</p>';
    }
}

function deviceItemHTML(d) {
    const statusClass = d.status === 'OK' || d.status === 'connecté' ? 'status-ok' : 'status-other';
    return `
        <div class="device-item">
            <div class="device-name">${d.name || 'Sans nom'}</div>
            ${d.manufacturer ? `<div class="device-manufacturer">${d.manufacturer}</div>` : ''}
            ${d.extra ? `<div class="device-manufacturer">${d.extra}</div>` : ''}
            ${d.status ? `<span class="device-status ${statusClass}">${d.status}</span>` : ''}
        </div>
    `;
}