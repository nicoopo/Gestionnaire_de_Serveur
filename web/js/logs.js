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
