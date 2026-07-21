const token = localStorage.getItem('token');
const maxHistoryPoints = 150;
let cpuHistory = [];
let ramHistory = [];
let diskHistory = [];
let gpuHistory = [];
let uploadHistory = [];
let downloadHistory = [];
if (!token) {
    window.location.href = '/login.html';
}

let explorerLoaded = false;
let captureLoaded = false;

document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.style.display = 'none');

        btn.classList.add('active');
        document.getElementById(`tab-${btn.dataset.tab}`).style.display = '';

        if (btn.dataset.tab === 'explorer' && !explorerLoaded) {
            loadExplorerRoots();
            explorerLoaded = true;
        }

        if (btn.dataset.tab === 'capture' && !captureLoaded) {
            loadInterfaces();
            captureLoaded = true;
        }
    });
});

function authFetch(url, options = {}) {
    options.headers = {
        ...(options.headers || {}),
        'Authorization': `Bearer ${token}`
    };
    return fetch(url, options).then(res => {
        if (res.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login.html';
        }
        return res;
    });
}

function segBarHTML(percent, segCount = 24) {
    const lit = Math.round((percent / 100) * segCount);
    let html = '';
    for (let i = 0; i < segCount; i++) {
        let cls = 'seg';
        if (i < lit) {
            const pos = i / segCount;
            cls += pos < 0.6 ? ' lit-ok' : pos < 0.85 ? ' lit-warn' : ' lit-danger';
        }
        html += `<div class="${cls}"></div>`;
    }
    return html;
}

function levelColor(percent) {
    if (percent < 60) return 'var(--ok)';
    if (percent < 85) return 'var(--warn)';
    return 'var(--danger)';
}

function formatUptime(seconds) {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `uptime ${h}h${String(m).padStart(2, '0')}`;
}

function updateClock() {
    document.getElementById('clock').textContent = new Date().toLocaleTimeString('fr-FR');
}

const fileIcons = {
    dir: '📁',
    '.txt': '📄', '.md': '📝', '.log': '📋',
    '.json': '🔧', '.yml': '🔧', '.yaml': '🔧', '.env': '🔧',
    '.go': '🐹', '.js': '📜', '.html': '🌐', '.css': '🎨',
    '.png': '🖼️', '.jpg': '🖼️', '.jpeg': '🖼️', '.gif': '🖼️', '.svg': '🖼️',
    '.zip': '📦', '.tar': '📦', '.gz': '📦',
    '.pdf': '📕',
};

function getFileIcon(entry) {
    if (entry.is_dir) return fileIcons.dir;
    const ext = entry.name.substring(entry.name.lastIndexOf('.'));
    return fileIcons[ext.toLowerCase()] || '📄';
}


function renderSystemInfo(data) {
    document.getElementById('cpu-value').textContent = `${data.cpu_percent.toFixed(1)}%`;
    document.getElementById('cpu-bar').innerHTML = segBarHTML(data.cpu_percent);

    document.getElementById('ram-value').textContent = `${data.ram_percent.toFixed(1)}%`;
    document.getElementById('ram-bar').innerHTML = segBarHTML(data.ram_percent);
    document.getElementById('ram-sub').textContent = `${(data.ram_used_mb/1024).toFixed(1)} / ${(data.ram_total_mb/1024).toFixed(1)} GB`;

    document.getElementById('uptime-tag').textContent = formatUptime(data.uptime_seconds);
    document.getElementById('proc-count').textContent = data.process_count;
    document.getElementById('core-count').textContent = `${data.cpu_core_count} cœurs`;

    const grid = document.getElementById('core-grid');
    grid.innerHTML = '';
    data.cpu_per_core.forEach((pct, i) => {
        const cell = document.createElement('div');
        cell.className = 'core-cell';
        cell.title = `${pct.toFixed(0)}%`;
        cell.innerHTML = `<style>.core-cell:nth-child(${i + 1})::after{height:${Math.max(pct,4)}%;background:${levelColor(pct)};}</style>`;
        grid.appendChild(cell);
    });

    cpuHistory.push(data.cpu_percent);
    if (cpuHistory.length > maxHistoryPoints) cpuHistory.shift();

    ramHistory.push(data.ram_percent);
    if (ramHistory.length > maxHistoryPoints) ramHistory.shift();

    diskHistory.push(data.disk_percent);
    if (diskHistory.length > maxHistoryPoints) diskHistory.shift();


    document.getElementById('net-value').innerHTML =
        `<span class="net-arrow-up">↑ ${formatSpeed(data.upload_kbs)}</span> &nbsp; <span class="net-arrow-down">↓ ${formatSpeed(data.download_kbs)}</span>`;

    uploadHistory.push(data.upload_kbs);
    if (uploadHistory.length > maxHistoryPoints) uploadHistory.shift();

    downloadHistory.push(data.download_kbs);
    if (downloadHistory.length > maxHistoryPoints) downloadHistory.shift();

    renderNetSparkline();

    renderHistoryChart();
}

async function loadDisks() {
    const res = await authFetch('/api/disks');
    const data = await res.json();

    const container = document.getElementById('disks-list');
    container.innerHTML = '';

    data.forEach(d => {
        const block = document.createElement('div');
        block.className = 'meter';
        block.innerHTML = `
            <div class="meter-label">
                <span>${d.mountpoint}</span>
                <span class="meter-value">${d.percent.toFixed(1)}%</span>
            </div>
            <div class="segbar">${segBarHTML(d.percent)}</div>
            <div class="meter-sub">${d.used_gb} / ${d.total_gb} GB · ${d.fstype}</div>
        `;
        container.appendChild(block);
    });
}

async function loadGPU() {
    const res = await authFetch('/api/gpu');
    const data = await res.json();

    const container = document.getElementById('gpu-block');
    container.innerHTML = '';

    if (data.length === 0) return;

    data.forEach(gpu => {
        const memPercent = (gpu.mem_used_mb / gpu.mem_total_mb) * 100;

        const block = document.createElement('div');
        block.innerHTML = `
            <div class="meter-sub" style="margin-bottom:0.5rem;">${gpu.name} · ${gpu.temp_c.toFixed(0)}°C</div>
            <div class="meter">
                <div class="meter-label"><span>GPU</span><span class="meter-value">${gpu.usage_percent.toFixed(0)}%</span></div>
                <div class="segbar">${segBarHTML(gpu.usage_percent)}</div>
            </div>
            <div class="meter">
                <div class="meter-label"><span>VRAM</span><span class="meter-value">${memPercent.toFixed(0)}%</span></div>
                <div class="segbar">${segBarHTML(memPercent)}</div>
                <div class="meter-sub">${gpu.mem_used_mb} / ${gpu.mem_total_mb} MB</div>
            </div>
        `;
        container.appendChild(block);
    });

    if (data.length > 0) {
        gpuHistory.push(data[0].usage_percent);
        if (gpuHistory.length > maxHistoryPoints) gpuHistory.shift();
        renderHistoryChart();
    }
}


function formatSpeed(kbs) {
    if (kbs >= 1024) return (kbs / 1024).toFixed(1) + ' MB/s';
    return kbs.toFixed(0) + ' KB/s';
}

function renderNetSparkline() {
    const width = 280, height = 40;
    const maxVal = Math.max(...uploadHistory, ...downloadHistory, 1); // évite division par zéro

    const buildPath = (values) => {
        if (values.length < 2) return '';
        const step = width / (maxHistoryPoints - 1);
        const offset = maxHistoryPoints - values.length;
        return values.map((v, i) => {
            const x = (offset + i) * step;
            const y = height - (v / maxVal) * height;
            return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
        }).join(' ');
    };

    document.getElementById('net-sparkline').innerHTML = `
        <svg viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" style="width:100%;height:40px;">
            <path d="${buildPath(downloadHistory)}" fill="none" stroke="var(--ok)" stroke-width="1.5"/>
            <path d="${buildPath(uploadHistory)}" fill="none" stroke="var(--danger)" stroke-width="1.5"/>
        </svg>
    `;
}

async function loadProcesses(filter = '') {
    const url = filter ? `/api/processes?name=${encodeURIComponent(filter)}` : '/api/processes';
    const res = await authFetch(url);
    const data = await res.json();

    const tbody = document.getElementById('processes-body');
    tbody.innerHTML = '';

    data.slice(0, 30).forEach(proc => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${proc.pid}</td>
            <td>${proc.name}</td>
            <td>${proc.cpu_percent.toFixed(1)}%</td>
            <td>${proc.ram_percent.toFixed(1)}%</td>
            <td>${proc.status}</td>
            <td><button class="danger" onclick="killProcess(${proc.pid})">arrêter</button></td>
        `;
        tbody.appendChild(row);
    });
}

async function killProcess(pid) {
    if (!confirm(`Arrêter le processus ${pid} ?`)) return;
    const res = await authFetch(`/api/processes/${pid}/kill`, { method: 'POST' });
    if (res.ok) loadProcesses(document.getElementById('process-filter').value);
    else alert('Erreur lors de l\'arrêt du processus');
}

async function loadServices() {
    const res = await authFetch('/api/services');
    const data = await res.json();
    const running = data.filter(s => s.status === 'running');

    document.getElementById('services-count').textContent = `${running.length} en ligne`;

    const tbody = document.getElementById('services-body');
    tbody.innerHTML = '';

    running.forEach(svc => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><span class="status-dot running"></span></td>
            <td>${svc.display_name}</td>
            <td>${svc.status}</td>
            <td>
                <button class="danger" onclick="stopService('${svc.name}')">stop</button>
                <button onclick="startService('${svc.name}')">start</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

async function startService(name) {
    const res = await authFetch(`/api/services/${name}/start`, { method: 'POST' });
    if (res.ok) loadServices();
    else alert('Erreur lors du démarrage du service');
}

async function stopService(name) {
    if (!confirm(`Arrêter le service ${name} ?`)) return;
    const res = await authFetch(`/api/services/${name}/stop`, { method: 'POST' });
    if (res.ok) loadServices();
    else alert('Erreur lors de l\'arrêt du service');
}

function connectWS() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(`${proto}://${location.host}/ws?token=${encodeURIComponent(token)}`);

    socket.onopen = () => {
        document.getElementById('conn-led').style.background = 'var(--ok)';
    };

    socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);

        if (msg.type === 'system') {
            renderSystemInfo(msg.data);
        } else if (msg.type === 'alerts') {
            showAlerts(msg.data);
        }
    };

    socket.onclose = () => {
        document.getElementById('conn-led').style.background = 'var(--danger)';
        setTimeout(connectWS, 2000);
    };

    socket.onerror = () => socket.close();
}

function buildLinePath(values, width, height) {
    if (values.length < 2) return '';
    const step = width / (maxHistoryPoints - 1);
    const offset = maxHistoryPoints - values.length;
    return values.map((v, i) => {
        const x = (offset + i) * step;
        const y = height - (v / 100) * height;
        return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(' ');
}

function renderHistoryChart() {
    const width = 600, height = 120;
    const cpuPath = buildLinePath(cpuHistory, width, height);
    const ramPath = buildLinePath(ramHistory, width, height);
    const diskPath = buildLinePath(diskHistory, width, height);
    const gpuPath = buildLinePath(gpuHistory, width, height);

    document.getElementById('history-chart').innerHTML = `
        <svg viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" style="width:100%;height:120px;">
            <line x1="0" y1="${height*0.25}" x2="${width}" y2="${height*0.25}" stroke="var(--border)" stroke-width="1"/>
            <line x1="0" y1="${height*0.5}" x2="${width}" y2="${height*0.5}" stroke="var(--border)" stroke-width="1"/>
            <line x1="0" y1="${height*0.75}" x2="${width}" y2="${height*0.75}" stroke="var(--border)" stroke-width="1"/>
            <path d="${diskPath}" fill="none" stroke="#7aa2f7" stroke-width="1.5" opacity="0.8"/>
            <path d="${gpuPath}" fill="none" stroke="#c678dd" stroke-width="1.5" opacity="0.8"/>
            <path d="${ramPath}" fill="none" stroke="var(--ok)" stroke-width="2"/>
            <path d="${cpuPath}" fill="none" stroke="var(--accent)" stroke-width="2"/>
        </svg>
    `;
}

async function loadHistory() {
    const res = await authFetch('/api/history');
    const data = await res.json();
    cpuHistory = data.map(s => s.cpu_percent);
    ramHistory = data.map(s => s.ram_percent);
    diskHistory = data.map(s => s.disk_percent);
    gpuHistory = data.map(s => s.gpu_percent);
    renderHistoryChart();
}

function showAlerts(alerts) {
    const container = document.getElementById('alerts-bar');

    container.innerHTML = alerts.map(a => `
        <div class="alert alert-${a.level}">
            <span class="alert-dot"></span>
            ${a.message}
        </div>
    `).join('');

    container.style.display = 'flex';

    clearTimeout(showAlerts._timeout);
    showAlerts._timeout = setTimeout(() => {
        container.style.display = 'none';
    }, 6000);
}

// ==================== FICHIERS (sandbox data/) ====================

let currentPath = '';
let currentSort = 'name';
let currentOrder = 'asc';
let currentEntries = [];

function renderBreadcrumb() {
    const parts = currentPath.split('/').filter(Boolean);
    let html = '<a onclick="loadFiles(\'\')">racine</a>';
    let accPath = '';

    parts.forEach(part => {
        accPath += (accPath ? '/' : '') + part;
        html += ` / <a onclick="loadFiles('${accPath}')">${part}</a>`;
    });

    document.getElementById('breadcrumb').innerHTML = html;
}

function renderFilesTable() {
    const searchTerm = document.getElementById('file-search').value.toLowerCase();
    const filtered = currentEntries.filter(e => e.name.toLowerCase().includes(searchTerm));

    const tbody = document.getElementById('files-body');
    tbody.innerHTML = '';

    if (currentPath !== '') {
        const row = document.createElement('tr');
        row.innerHTML = `<td>📁</td><td colspan="4"><a href="#" onclick="goUp(); return false;">..</a></td>`;
        tbody.appendChild(row);
    }

    filtered.forEach(entry => {
        const icon = getFileIcon(entry);
        const row = document.createElement('tr');

        if (entry.is_dir) {
            row.innerHTML = `
                <td>${icon}</td>
                <td><a href="#" onclick="loadFiles('${entry.path}'); return false;">${entry.name}</a></td>
                <td>—</td>
                <td>${entry.mod_time}</td>
                <td></td>
            `;
        } else {
            row.innerHTML = `
                <td>${icon}</td>
                <td><a href="#" onclick="viewFile('${entry.path}'); return false;">${entry.name}</a></td>
                <td>${entry.size_kb} Ko</td>
                <td>${entry.mod_time}</td>
                <td>
                    <button onclick="downloadFile('${entry.path}')">télécharger</button>
                    <button class="danger" onclick="deleteFile('${entry.path}')">supprimer</button>
                </td>
            `;
        }
        tbody.appendChild(row);
    });

    updateSortArrows();
}

function updateSortArrows() {
    document.querySelectorAll('#tab-dashboard th.sortable').forEach(th => {
        const arrow = th.querySelector('.sort-arrow');
        if (th.dataset.sort === currentSort) {
            arrow.textContent = currentOrder === 'asc' ? '▲' : '▼';
        } else {
            arrow.textContent = '';
        }
    });
}

async function loadFiles(path = '') {
    currentPath = path;
    renderBreadcrumb();

    const res = await authFetch(`/api/files?path=${encodeURIComponent(path)}&sort=${currentSort}&order=${currentOrder}`);
    if (!res.ok) {
        alert('Erreur lors du chargement du dossier');
        return;
    }
    currentEntries = await res.json();
    renderFilesTable();
}

function goUp() {
    const parts = currentPath.split('/').filter(Boolean);
    parts.pop();
    loadFiles(parts.join('/'));
}

function setSortBy(field) {
    if (currentSort === field) {
        currentOrder = currentOrder === 'asc' ? 'desc' : 'asc';
    } else {
        currentSort = field;
        currentOrder = 'asc';
    }
    loadFiles(currentPath);
}

async function viewFile(path) {
    const res = await authFetch(`/api/files/read?path=${encodeURIComponent(path)}`);
    if (!res.ok) {
        const err = await res.text();
        alert('Erreur: ' + err);
        return;
    }
    const data = await res.json();

    document.getElementById('viewer-filename').textContent = path;
    document.getElementById('file-content').textContent = data.content;
    document.getElementById('file-viewer').style.display = 'block';
}

function closeFileViewer() {
    document.getElementById('file-viewer').style.display = 'none';
}

function downloadFile(path) {
    const url = `/api/files/download?path=${encodeURIComponent(path)}&token=${encodeURIComponent(token)}`;
    window.open(url, '_blank');
}

document.querySelectorAll('#tab-dashboard th.sortable').forEach(th => {
    th.addEventListener('click', () => setSortBy(th.dataset.sort));
});

document.getElementById('file-search').addEventListener('input', renderFilesTable);

// ==================== EXPLORATEUR MACHINE COMPLÈTE ====================

let explorerRoot = '';
let explorerPath = '';
let explorerSort = 'name';
let explorerOrder = 'asc';
let explorerEntries = [];

async function loadExplorerRoots() {
    const res = await authFetch('/api/explorer/roots');
    const roots = await res.json();

    const select = document.getElementById('explorer-root-selector');
    select.innerHTML = roots.map(r => `<option value="${r}">${r}</option>`).join('');

    select.addEventListener('change', () => {
        explorerRoot = select.value;
        loadExplorerFiles('');
    });

    if (roots.length > 0) {
        explorerRoot = roots[0];
        loadExplorerFiles('');
    }
}

function renderExplorerBreadcrumb() {
    const parts = explorerPath.split('/').filter(Boolean);
    let html = `<a onclick="loadExplorerFiles('')">${explorerRoot}</a>`;
    let accPath = '';

    parts.forEach(part => {
        accPath += (accPath ? '/' : '') + part;
        html += ` / <a onclick="loadExplorerFiles('${accPath}')">${part}</a>`;
    });

    document.getElementById('explorer-breadcrumb').innerHTML = html;
}

function renderExplorerTable() {
    const searchTerm = document.getElementById('explorer-search').value.toLowerCase();
    const filtered = explorerEntries.filter(e => e.name.toLowerCase().includes(searchTerm));

    const tbody = document.getElementById('explorer-files-body');
    tbody.innerHTML = '';

    if (explorerPath !== '') {
        const row = document.createElement('tr');
        row.innerHTML = `<td>📁</td><td colspan="4"><a href="#" onclick="explorerGoUp(); return false;">..</a></td>`;
        tbody.appendChild(row);
    }

    filtered.forEach(entry => {
        const icon = getFileIcon(entry);
        const row = document.createElement('tr');

        if (entry.is_dir) {
            row.innerHTML = `
                <td>${icon}</td>
                <td><a href="#" onclick="loadExplorerFiles('${entry.path}'); return false;">${entry.name}</a></td>
                <td>—</td>
                <td>${entry.mod_time}</td>
                <td><button class="danger" onclick="deleteExplorerDir('${entry.path}', '${entry.name}')">supprimer</button></td>
            `;
        } else {
            row.innerHTML = `
                <td>${icon}</td>
                <td><a href="#" onclick="viewExplorerFile('${entry.path}'); return false;">${entry.name}</a></td>
                <td>${entry.size_kb} Ko</td>
                <td>${entry.mod_time}</td>
                <td>
                    <button onclick="downloadExplorerFile('${entry.path}')">télécharger</button>
                    <button class="danger" onclick="deleteExplorerFile('${entry.path}')">supprimer</button>
                </td>
            `;
        }
        tbody.appendChild(row);
    });

    updateExplorerSortArrows();
}

function updateExplorerSortArrows() {
    document.querySelectorAll('#tab-explorer th.sortable').forEach(th => {
        const arrow = th.querySelector('.sort-arrow');
        if (th.dataset.sort === explorerSort) {
            arrow.textContent = explorerOrder === 'asc' ? '▲' : '▼';
        } else {
            arrow.textContent = '';
        }
    });
}

async function deleteExplorerFile(path) {
    if (!confirm(`Supprimer définitivement "${path}" sur ${explorerRoot} ? Cette action est irréversible.`)) return;

    const res = await authFetch(`/api/explorer/delete?root=${encodeURIComponent(explorerRoot)}&path=${encodeURIComponent(path)}`, { method: 'DELETE' });
    if (res.ok) {
        loadExplorerFiles(explorerPath);
    } else {
        const err = await res.text();
        alert('Erreur: ' + err);
    }
}

async function loadExplorerFiles(path = '') {
    explorerPath = path;
    renderExplorerBreadcrumb();

    const res = await authFetch(`/api/explorer/list?root=${encodeURIComponent(explorerRoot)}&path=${encodeURIComponent(path)}&sort=${explorerSort}&order=${explorerOrder}`);
    if (!res.ok) {
        alert('Erreur lors du chargement du dossier');
        return;
    }
    explorerEntries = await res.json();
    renderExplorerTable();
}

function explorerGoUp() {
    const parts = explorerPath.split('/').filter(Boolean);
    parts.pop();
    loadExplorerFiles(parts.join('/'));
}

function setExplorerSortBy(field) {
    if (explorerSort === field) {
        explorerOrder = explorerOrder === 'asc' ? 'desc' : 'asc';
    } else {
        explorerSort = field;
        explorerOrder = 'asc';
    }
    loadExplorerFiles(explorerPath);
}

async function viewExplorerFile(path) {
    const res = await authFetch(`/api/explorer/read?root=${encodeURIComponent(explorerRoot)}&path=${encodeURIComponent(path)}`);
    if (!res.ok) {
        const err = await res.text();
        alert('Erreur: ' + err);
        return;
    }
    const data = await res.json();

    document.getElementById('explorer-viewer-filename').textContent = path;
    document.getElementById('explorer-file-content').textContent = data.content;
    document.getElementById('explorer-file-viewer').style.display = 'block';
}

function closeExplorerViewer() {
    document.getElementById('explorer-file-viewer').style.display = 'none';
}

function downloadExplorerFile(path) {
    const url = `/api/explorer/download?root=${encodeURIComponent(explorerRoot)}&path=${encodeURIComponent(path)}&token=${encodeURIComponent(token)}`;
    window.open(url, '_blank');
}

document.querySelectorAll('#tab-explorer th.sortable').forEach(th => {
    th.addEventListener('click', () => setExplorerSortBy(th.dataset.sort));
});

document.getElementById('explorer-search').addEventListener('input', renderExplorerTable);

// ==================== LOGS ====================

function connectLogsWS() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(`${proto}://${location.host}/ws/logs?token=${encodeURIComponent(token)}`);

    socket.onopen = () => {
        document.getElementById('logs-led').style.background = 'var(--ok)';
    };

    socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.type === 'log_line') {
            const output = document.getElementById('logs-output');
            output.textContent += msg.line;
            output.scrollTop = output.scrollHeight;
        }
    };

    socket.onclose = () => {
        document.getElementById('logs-led').style.background = 'var(--danger)';
        setTimeout(connectLogsWS, 3000);
    };

    socket.onerror = () => socket.close();
}

async function deleteFile(path) {
    if (!confirm(`Supprimer définitivement "${path}" ? Cette action est irréversible.`)) return;

    const res = await authFetch(`/api/files/delete?path=${encodeURIComponent(path)}`, { method: 'DELETE' });
    if (res.ok) {
        loadFiles(currentPath);
    } else {
        const err = await res.text();
        alert('Erreur: ' + err);
    }
}

async function deleteExplorerDir(path, name) {
    const confirmation = prompt(`Pour supprimer définitivement le dossier "${name}" et TOUT son contenu, tape son nom exact :`);
    if (confirmation !== name) {
        if (confirmation !== null) alert('Nom incorrect, suppression annulée.');
        return;
    }

    const res = await authFetch(`/api/explorer/delete-dir?root=${encodeURIComponent(explorerRoot)}&path=${encodeURIComponent(path)}`, { method: 'DELETE' });
    if (res.ok) {
        loadExplorerFiles(explorerPath);
    } else {
        const err = await res.text();
        alert('Erreur: ' + err);
    }
}

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
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    captureSocket = new WebSocket(`${proto}://${location.host}/ws/capture?iface=${encodeURIComponent(iface)}&token=${encodeURIComponent(token)}`);

    captureSocket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.type === 'packet') addPacketRow(msg.data);
    };

    captureSocket.onopen = () => {
        capturing = true;
        document.getElementById('capture-toggle').textContent = 'arrêter';
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

document.querySelectorAll('.proto-filter').forEach(cb => {
    cb.addEventListener('change', () => {
        document.getElementById('capture-body').innerHTML = '';
    });
});
document.getElementById('capture-search')?.addEventListener('input', () => {
    document.getElementById('capture-body').innerHTML = '';
});

// ==================== DÉMARRAGE ====================

document.getElementById('process-filter').addEventListener('input', (e) => {
    loadProcesses(e.target.value);
});

updateClock();
setInterval(updateClock, 1000);

connectWS();
connectLogsWS();
loadProcesses();
loadServices();
loadDisks();
loadGPU();
loadHistory();
loadFiles();

setInterval(() => {
    loadProcesses(document.getElementById('process-filter').value);
    loadServices();
    loadDisks();
    loadGPU();
}, 5000);