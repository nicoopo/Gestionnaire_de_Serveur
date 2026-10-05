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

function formatSpeed(kbs) {
    if (kbs >= 1024) return (kbs / 1024).toFixed(1) + ' MB/s';
    return kbs.toFixed(0) + ' KB/s';
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
