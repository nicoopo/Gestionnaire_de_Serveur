document.addEventListener('DOMContentLoaded', () => {
    const filterInput = document.getElementById('process-filter');
    if (filterInput) {
        filterInput.addEventListener('input', (e) => {
            loadProcesses(e.target.value);
        });
    }

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
});
