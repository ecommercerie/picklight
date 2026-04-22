registerPage('#status', function(app) {
    const page = document.createElement('div');

    // Update banner (hidden by default)
    const updateBanner = document.createElement('div');
    updateBanner.style.display = 'none';
    updateBanner.style.background = '#1e3a5f';
    updateBanner.style.border = '1px solid #2563eb';
    updateBanner.style.borderRadius = '8px';
    updateBanner.style.padding = '12px 16px';
    updateBanner.style.marginBottom = '16px';
    updateBanner.style.alignItems = 'center';
    updateBanner.style.justifyContent = 'space-between';
    updateBanner.style.gap = '12px';
    updateBanner.style.fontSize = '13px';

    const updateText = document.createElement('span');
    updateText.style.color = '#93c5fd';
    updateBanner.appendChild(updateText);

    const updateBtn = document.createElement('button');
    updateBtn.className = 'btn btn-primary btn-sm';
    updateBtn.textContent = t('update.btn');
    updateBtn.onclick = async () => {
        updateBtn.disabled = true;
        updateBtn.textContent = t('update.inProgress');
        try {
            await window.go.main.App.ApplyUpdate();
            showToast(t('update.success'), 'success');
        } catch(e) {
            showToast(t('common.error', {msg: e}), 'error');
            updateBtn.textContent = t('update.retry');
            updateBtn.disabled = false;
        }
    };
    updateBanner.appendChild(updateBtn);
    page.appendChild(updateBanner);

    // Check for updates on page load
    (async () => {
        try {
            const status = await window.go.main.App.CheckForUpdate();
            if (status.updateAvailable) {
                updateText.textContent = t('update.available', {current: status.currentVersion, latest: status.latestVersion});
                updateBanner.style.display = 'flex';
            }
        } catch(e) { /* ignore */ }
    })();

    // Big circle with value inside
    const centerCard = document.createElement('div');
    centerCard.className = 'card';
    centerCard.style.textAlign = 'center';

    const circle = document.createElement('div');
    circle.className = 'big-circle';
    circle.style.backgroundColor = '#333';
    circle.style.display = 'flex';
    circle.style.alignItems = 'center';
    circle.style.justifyContent = 'center';
    centerCard.appendChild(circle);

    const bigValue = document.createElement('div');
    bigValue.style.fontSize = '42px';
    bigValue.style.fontWeight = '700';
    bigValue.style.color = '#fff';
    bigValue.style.textShadow = '0 1px 4px rgba(0,0,0,0.5)';
    bigValue.textContent = '-';
    circle.appendChild(bigValue);

    const bigLabel = document.createElement('div');
    bigLabel.className = 'big-label';
    bigLabel.textContent = t('status.waiting');
    centerCard.appendChild(bigLabel);

    // Info row
    const infoRow = document.createElement('div');
    infoRow.style.display = 'flex';
    infoRow.style.justifyContent = 'center';
    infoRow.style.gap = '24px';
    infoRow.style.marginTop = '16px';
    infoRow.style.fontSize = '12px';
    infoRow.style.color = 'var(--color-text-muted)';

    const lastPollSpan = document.createElement('span');
    lastPollSpan.textContent = t('status.lastPoll') + ': -';

    const nextPollSpan = document.createElement('span');
    nextPollSpan.textContent = '';

    const blStatusSpan = document.createElement('span');
    blStatusSpan.innerHTML = '<span class="status-dot" style="background:#ef4444"></span> Status Light';

    infoRow.appendChild(lastPollSpan);
    infoRow.appendChild(nextPollSpan);
    infoRow.appendChild(blStatusSpan);
    centerCard.appendChild(infoRow);

    // Buttons
    const btnRow = document.createElement('div');
    btnRow.style.textAlign = 'center';
    btnRow.style.marginTop = '16px';
    btnRow.style.display = 'flex';
    btnRow.style.justifyContent = 'center';
    btnRow.style.gap = '8px';

    const pollBtn = document.createElement('button');
    pollBtn.className = 'btn btn-primary';
    pollBtn.textContent = t('status.pollBtn');
    pollBtn.onclick = async () => {
        try {
            await window.go.main.App.PollNow();
            showToast(t('status.pollStarted'), 'success');
        } catch(e) { showToast(t('common.error', {msg: e}), 'error'); }
    };
    btnRow.appendChild(pollBtn);

    const detectBtn = document.createElement('button');
    detectBtn.className = 'btn btn-secondary';
    detectBtn.textContent = t('status.detectBtn');
    detectBtn.onclick = async () => {
        const name = await window.go.main.App.DetectDevice();
        if (name) {
            showToast(t('status.deviceDetected', {name: name}), 'success');
        } else {
            showToast(t('status.noDeviceFound'), 'error');
        }
        refresh();
    };
    btnRow.appendChild(detectBtn);

    centerCard.appendChild(btnRow);
    page.appendChild(centerCard);

    // Logs
    const logsCard = document.createElement('div');
    logsCard.className = 'card';

    const logsTitle = document.createElement('h2');
    logsTitle.className = 'section-title';
    logsTitle.textContent = t('status.logs');
    logsCard.appendChild(logsTitle);

    const logsContainer = document.createElement('div');
    logsContainer.className = 'logs-container';
    logsCard.appendChild(logsContainer);

    page.appendChild(logsCard);
    app.appendChild(page);

    function formatCountdown(seconds) {
        if (seconds <= 0) return '';
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        if (m > 0) return t('status.nextPollMinSec', {m: m, s: s > 0 ? String(s).padStart(2, '0') : ''});
        return t('status.nextPollSec', {s: s});
    }

    async function refresh() {
        try {
            const status = await window.go.main.App.GetStatus();
            bigValue.textContent = status.ordersPending;

            if (status.activeThreshold) {
                circle.style.backgroundColor = status.activeThreshold.color;
                bigLabel.textContent = status.activeThreshold.label;
            } else {
                circle.style.backgroundColor = '#333';
                bigLabel.textContent = t('status.noThreshold');
            }

            lastPollSpan.textContent = t('status.lastPoll') + ': ' + (status.lastPollTime || '-');
            if (status.lastPollError) {
                lastPollSpan.textContent += ' ' + t('status.lastPollError');
                lastPollSpan.style.color = 'var(--color-danger)';
            } else {
                lastPollSpan.style.color = 'var(--color-text-muted)';
            }

            nextPollSpan.textContent = status.polling ? formatCountdown(status.nextPollIn) : '';

            const dotColor = status.deviceConnected ? '#22c55e' : '#ef4444';
            const dotLabel = status.deviceConnected ? (status.deviceName || t('status.connected')) : t('status.disconnected');
            blStatusSpan.innerHTML = '<span class="status-dot" style="background:' + dotColor + '"></span> ' + dotLabel;

            // Logs
            const logs = await window.go.main.App.GetLogs();
            logsContainer.innerHTML = '';
            if (logs && logs.length > 0) {
                logs.slice().reverse().forEach(entry => {
                    const div = document.createElement('div');
                    div.className = 'log-entry log-' + entry.level;
                    const ts = new Date(entry.timestamp);
                    div.textContent = ts.toLocaleTimeString() + ' [' + entry.level + '] ' + entry.message;
                    logsContainer.appendChild(div);
                });
            }
        } catch(e) { console.error(e); }
    }

    refresh();

    const unsub = window.runtime.EventsOn('status:updated', refresh);

    // Refresh every second for countdown
    const interval = setInterval(refresh, 1000);
    const observer = new MutationObserver(() => {
        if (!document.body.contains(page)) {
            clearInterval(interval);
            if (unsub) unsub();
            observer.disconnect();
        }
    });
    observer.observe(document.body, { childList: true, subtree: true });
});
