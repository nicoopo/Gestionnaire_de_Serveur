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

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('#tab-explorer th.sortable').forEach(th => {
        th.addEventListener('click', () => setExplorerSortBy(th.dataset.sort));
    });

    const searchInput = document.getElementById('explorer-search');
    if (searchInput) searchInput.addEventListener('input', renderExplorerTable);
});
