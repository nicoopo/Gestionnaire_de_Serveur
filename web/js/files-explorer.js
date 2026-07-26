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

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('#tab-dashboard th.sortable').forEach(th => {
        th.addEventListener('click', () => setSortBy(th.dataset.sort));
    });

    const searchInput = document.getElementById('file-search');
    if (searchInput) searchInput.addEventListener('input', renderFilesTable);
});
