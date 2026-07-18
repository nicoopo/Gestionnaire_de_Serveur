const token = localStorage.getItem('token');
if (!token) {
    window.location.href = '/login.html';
}

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
        renderSystemInfo(JSON.parse(event.data));
    };

    socket.onclose = () => {
        document.getElementById('conn-led').style.background = 'var(--danger)';
        setTimeout(connectWS, 2000);
    };

    socket.onerror = () => socket.close();
}

document.getElementById('process-filter').addEventListener('input', (e) => {
    loadProcesses(e.target.value);
});

updateClock();
setInterval(updateClock, 1000);

connectWS();
loadProcesses();
loadServices();
loadDisks();
loadGPU();

setInterval(() => {
    loadProcesses(document.getElementById('process-filter').value);
    loadServices();
    loadDisks();
    loadGPU();
}, 5000);