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

async function loadSystemInfo() {
    const res = await fetch('/api/system');
    const data = await res.json();

    document.getElementById('cpu-value').textContent = `${data.cpu_percent.toFixed(1)}%`;
    document.getElementById('cpu-bar').innerHTML = segBarHTML(data.cpu_percent);

    document.getElementById('ram-value').textContent = `${data.ram_percent.toFixed(1)}%`;
    document.getElementById('ram-bar').innerHTML = segBarHTML(data.ram_percent);
    document.getElementById('ram-sub').textContent = `${(data.ram_used_mb/1024).toFixed(1)} / ${(data.ram_total_mb/1024).toFixed(1)} GB`;

    document.getElementById('disk-value').textContent = `${data.disk_percent.toFixed(1)}%`;
    document.getElementById('disk-bar').innerHTML = segBarHTML(data.disk_percent);
    document.getElementById('disk-sub').textContent = `${data.disk_used_gb} / ${data.disk_total_gb} GB`;

    document.getElementById('uptime-tag').textContent = formatUptime(data.uptime_seconds);
    document.getElementById('proc-count').textContent = data.process_count;
    document.getElementById('core-count').textContent = `${data.cpu_core_count} cœurs`;

    const grid = document.getElementById('core-grid');
    grid.innerHTML = '';
    data.cpu_per_core.forEach(pct => {
        const cell = document.createElement('div');
        cell.className = 'core-cell';
        cell.title = `${pct.toFixed(0)}%`;
        cell.style.setProperty('--h', `${Math.max(pct, 4)}%`);
        cell.innerHTML = `<style>.core-cell:nth-child(${grid.children.length + 1})::after{height:${Math.max(pct,4)}%;background:${levelColor(pct)};}</style>`;
        grid.appendChild(cell);
    });
}

async function loadProcesses(filter = '') {
    const url = filter ? `/api/processes?name=${encodeURIComponent(filter)}` : '/api/processes';
    const res = await fetch(url);
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
    const res = await fetch(`/api/processes/${pid}/kill`, { method: 'POST' });
    if (res.ok) loadProcesses(document.getElementById('process-filter').value);
    else alert('Erreur lors de l\'arrêt du processus');
}

async function loadServices() {
    const res = await fetch('/api/services');
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
    const res = await fetch(`/api/services/${name}/start`, { method: 'POST' });
    if (res.ok) loadServices();
    else alert('Erreur lors du démarrage du service');
}

async function stopService(name) {
    if (!confirm(`Arrêter le service ${name} ?`)) return;
    const res = await fetch(`/api/services/${name}/stop`, { method: 'POST' });
    if (res.ok) loadServices();
    else alert('Erreur lors de l\'arrêt du service');
}

document.getElementById('process-filter').addEventListener('input', (e) => {
    loadProcesses(e.target.value);
});

updateClock();
setInterval(updateClock, 1000);

loadSystemInfo();
loadProcesses();
loadServices();

setInterval(() => {
    loadSystemInfo();
    loadProcesses(document.getElementById('process-filter').value);
}, 5000);