let captureSocket = null;
let capturing = false;
const maxPacketRows = 200;
let packetStats = { total: 0, byProto: {}, byIP: {} };

async function loadInterfaces() {
    const res = await authFetch('/api/network/interfaces');
    const ifaces = await res.json();
    const select = document.getElementById('iface-selector');
    select.innerHTML = ifaces.map(i => `<option value="${i.name}">${i.description}</option>`).join('');
}

function toggleCapture() {
    if (capturing) stopCapture();
    else startCapture();
}

function startCapture() {
    resetStats();
    document.getElementById('capture-body').innerHTML = '';

    const iface = document.getElementById('iface-selector').value;
    const protocols = Array.from(document.querySelectorAll('.proto-filter:checked')).map(c => c.value);
    const ipFilter = document.getElementById('capture-ip-filter').value.trim();
    const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';

    const params = new URLSearchParams({ iface, token });
    if (protocols.length) params.set('proto', protocols.join(','));
    if (ipFilter) params.set('ip', ipFilter);

    captureSocket = new WebSocket(`${wsProto}://${location.host}/ws/capture?${params.toString()}`);

    captureSocket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.type === 'packet') addPacketRow(msg.data);
    };

    captureSocket.onopen = () => {
        capturing = true;
        document.getElementById('capture-toggle').textContent = 'arrêter';
    };

    captureSocket.onerror = () => {
        document.getElementById('capture-stats').innerHTML =
            '<div style="color:var(--danger, #d33);">Échec de connexion — vérifie le filtre IP/CIDR saisi.</div>';
    };

    captureSocket.onclose = () => {
        capturing = false;
        document.getElementById('capture-toggle').textContent = 'démarrer';
    };
}

function stopCapture() {
    if (captureSocket) captureSocket.close();
}

function resetStats() {
    packetStats = { total: 0, byProto: {}, byIP: {} };
    updateStatsDisplay();
}

function passesFilters(pkt) {
    const activeProtos = Array.from(document.querySelectorAll('.proto-filter:checked')).map(c => c.value);
    if (!activeProtos.includes(pkt.protocol)) return false;

    const search = document.getElementById('capture-search').value.toLowerCase();
    if (search) {
        const haystack = `${pkt.src_ip} ${pkt.dst_ip} ${pkt.src_host} ${pkt.dst_host}`.toLowerCase();
        if (!haystack.includes(search)) return false;
    }
    return true;
}

function updateStats(pkt) {
    packetStats.total++;
    packetStats.byProto[pkt.protocol] = (packetStats.byProto[pkt.protocol] || 0) + 1;
    packetStats.byIP[pkt.dst_ip] = (packetStats.byIP[pkt.dst_ip] || 0) + pkt.length;
    updateStatsDisplay();
}

function updateStatsDisplay() {
    const protoParts = Object.entries(packetStats.byProto)
        .sort((a, b) => b[1] - a[1])
        .map(([p, c]) => `${p}: ${c}`).join(' · ');

    const topIPs = Object.entries(packetStats.byIP)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 3)
        .map(([ip, bytes]) => `${ip} (${(bytes / 1024).toFixed(1)} Ko)`).join(' · ');

    document.getElementById('capture-stats').innerHTML = `
        <div><strong>${packetStats.total}</strong> paquets — ${protoParts || 'aucun'}</div>
        <div class="meter-sub" style="margin-top:0.3rem;">Top destinations : ${topIPs || '—'}</div>
    `;
}

function addPacketRow(pkt) {
    updateStats(pkt);
    if (!passesFilters(pkt)) return;

    const tbody = document.getElementById('capture-body');
    const row = document.createElement('tr');
    const srcLabel = pkt.src_host || pkt.src_ip;
    const dstLabel = pkt.dst_host || pkt.dst_ip;

    row.innerHTML = `
        <td>${pkt.timestamp}</td>
        <td title="${pkt.src_ip}">${srcLabel}${pkt.src_port ? ':' + pkt.src_port : ''}</td>
        <td title="${pkt.dst_ip}">${dstLabel}${pkt.dst_port ? ':' + pkt.dst_port : ''}</td>
        <td>${pkt.protocol}</td>
        <td>${pkt.tcp_flags || ''}</td>
        <td>${pkt.process_name || '—'}</td>
        <td>${pkt.length} o</td>
    `;
    tbody.insertBefore(row, tbody.firstChild);

    while (tbody.children.length > maxPacketRows) {
        tbody.removeChild(tbody.lastChild);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.proto-filter').forEach(cb => {
        cb.addEventListener('change', () => {
            document.getElementById('capture-body').innerHTML = '';
        });
    });

    const searchInput = document.getElementById('capture-search');
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            document.getElementById('capture-body').innerHTML = '';
        });
    }
});
