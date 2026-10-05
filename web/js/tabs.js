let explorerLoaded = false;
let captureLoaded = false;
let devicesLoaded = false;

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
        if (btn.dataset.tab === 'devices' && !devicesLoaded) {
            loadDevices();
            devicesLoaded = true;
        }
    });
});
