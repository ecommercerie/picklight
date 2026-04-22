registerPage('#settings', function(app) {
    const page = document.createElement('div');

    // General config
    const generalCard = document.createElement('div');
    generalCard.className = 'card';

    const generalTitle = document.createElement('h2');
    generalTitle.className = 'section-title';
    generalTitle.textContent = t('settings.title');
    generalCard.appendChild(generalTitle);

    function field(labelKey, type, id, placeholder) {
        const group = document.createElement('div');
        group.className = 'form-group';
        const lbl = document.createElement('label');
        lbl.textContent = t(labelKey);
        group.appendChild(lbl);
        const input = document.createElement('input');
        input.type = type;
        input.id = id;
        input.placeholder = placeholder || '';
        group.appendChild(input);
        return group;
    }

    function checkbox(labelKey, id) {
        const group = document.createElement('div');
        group.className = 'form-group';
        group.style.display = 'flex';
        group.style.alignItems = 'center';
        group.style.gap = '10px';
        const cb = document.createElement('input');
        cb.type = 'checkbox';
        cb.id = id;
        group.appendChild(cb);
        const lbl = document.createElement('label');
        lbl.htmlFor = id;
        lbl.style.margin = '0';
        lbl.style.cursor = 'pointer';
        lbl.textContent = t(labelKey);
        group.appendChild(lbl);
        return group;
    }

    generalCard.appendChild(field('settings.endpointUrl', 'text', 'endpoint-url', 'https://...'));
    generalCard.appendChild(field('settings.pollInterval', 'number', 'poll-interval', '300'));
    generalCard.appendChild(field('settings.jsonPath', 'text', 'json-path', 'stats.orders_pending'));
    generalCard.appendChild(checkbox('settings.tlsSkip', 'tls-skip'));
    generalCard.appendChild(checkbox('settings.soundEnabled', 'sound-enabled'));
    generalCard.appendChild(checkbox('settings.soundOnChange', 'sound-on-change'));
    generalCard.appendChild(checkbox('settings.startup', 'startup-enabled'));

    // Language selector
    const langGroup = document.createElement('div');
    langGroup.className = 'form-group';
    const langLabel = document.createElement('label');
    langLabel.textContent = t('settings.language');
    langGroup.appendChild(langLabel);
    const langSelect = document.createElement('select');
    langSelect.id = 'language-select';
    [['auto', t('settings.langAuto')], ['fr', 'Fran\u00e7ais'], ['en', 'English']].forEach(([val, label]) => {
        const opt = document.createElement('option');
        opt.value = val;
        opt.textContent = label;
        langSelect.appendChild(opt);
    });
    langGroup.appendChild(langSelect);
    generalCard.appendChild(langGroup);

    const saveBtn = document.createElement('button');
    saveBtn.className = 'btn btn-primary';
    saveBtn.textContent = t('settings.save');
    saveBtn.style.marginTop = '8px';
    generalCard.appendChild(saveBtn);

    page.appendChild(generalCard);

    // Thresholds
    const thresholdCard = document.createElement('div');
    thresholdCard.className = 'card';

    const thresholdTitle = document.createElement('h2');
    thresholdTitle.className = 'section-title';
    thresholdTitle.textContent = t('thresholds.title');
    thresholdCard.appendChild(thresholdTitle);

    const thresholdTable = document.createElement('div');
    thresholdTable.id = 'threshold-table';
    thresholdCard.appendChild(thresholdTable);

    const addBtn = document.createElement('button');
    addBtn.className = 'btn btn-secondary';
    addBtn.textContent = t('thresholds.add');
    addBtn.style.marginTop = '12px';
    addBtn.onclick = () => addThresholdRow({min: 0, max: 0, color: '#FFFFFF', sound: false, label: ''});
    thresholdCard.appendChild(addBtn);

    page.appendChild(thresholdCard);

    // Debug section
    const debugCard = document.createElement('div');
    debugCard.className = 'card';

    const debugTitle = document.createElement('h2');
    debugTitle.className = 'section-title';
    debugTitle.textContent = t('debug.title');
    debugCard.appendChild(debugTitle);

    const debugRow = document.createElement('div');
    debugRow.style.display = 'flex';
    debugRow.style.gap = '8px';
    debugRow.style.alignItems = 'center';
    debugRow.style.flexWrap = 'wrap';

    const colorPicker = document.createElement('input');
    colorPicker.type = 'color';
    colorPicker.value = '#00FF00';
    debugRow.appendChild(colorPicker);

    const testColorBtn = document.createElement('button');
    testColorBtn.className = 'btn btn-secondary btn-sm';
    testColorBtn.textContent = t('debug.testColor');
    testColorBtn.onclick = async () => {
        try {
            await window.go.main.App.TestColor(colorPicker.value);
            showToast(t('debug.colorSent'), 'success');
        } catch(e) { showToast('' + e, 'error'); }
    };
    debugRow.appendChild(testColorBtn);

    const toneSelect = document.createElement('select');
    toneSelect.style.width = '100px';
    for (let i = 1; i <= 15; i++) {
        const opt = document.createElement('option');
        opt.value = i;
        opt.textContent = 'Tone ' + i;
        toneSelect.appendChild(opt);
    }
    debugRow.appendChild(toneSelect);

    const testSoundBtn = document.createElement('button');
    testSoundBtn.className = 'btn btn-secondary btn-sm';
    testSoundBtn.textContent = t('debug.testSound');
    testSoundBtn.onclick = async () => {
        try {
            await window.go.main.App.TestSound(parseInt(toneSelect.value));
            showToast(t('debug.soundSent'), 'success');
        } catch(e) { showToast('' + e, 'error'); }
    };
    debugRow.appendChild(testSoundBtn);

    const testOffBtn = document.createElement('button');
    testOffBtn.className = 'btn btn-danger btn-sm';
    testOffBtn.textContent = t('debug.turnOff');
    testOffBtn.onclick = async () => {
        try {
            await window.go.main.App.TestOff();
            showToast(t('debug.turnedOff'), 'success');
        } catch(e) { showToast('' + e, 'error'); }
    };
    debugRow.appendChild(testOffBtn);

    debugCard.appendChild(debugRow);

    // Diagnose button
    const diagnoseTitle = document.createElement('h3');
    diagnoseTitle.style.fontSize = '13px';
    diagnoseTitle.style.fontWeight = '600';
    diagnoseTitle.style.marginTop = '16px';
    diagnoseTitle.style.marginBottom = '8px';
    diagnoseTitle.textContent = t('debug.diagnoseTitle');
    debugCard.appendChild(diagnoseTitle);

    const diagnoseBtn = document.createElement('button');
    diagnoseBtn.className = 'btn btn-primary btn-sm';
    diagnoseBtn.textContent = t('debug.diagnoseBtn');
    diagnoseBtn.style.marginBottom = '8px';
    debugCard.appendChild(diagnoseBtn);

    const diagnoseResult = document.createElement('div');
    diagnoseResult.style.fontSize = '12px';
    diagnoseResult.style.fontFamily = 'monospace';
    debugCard.appendChild(diagnoseResult);

    diagnoseBtn.onclick = async () => {
        diagnoseBtn.disabled = true;
        diagnoseBtn.textContent = t('debug.diagnosing');
        diagnoseResult.innerHTML = '';
        try {
            const results = await window.go.main.App.DiagnoseDevice();
            if (!results || results.length === 0) {
                diagnoseResult.textContent = t('debug.noInterface');
                return;
            }
            results.forEach((r, idx) => {
                const div = document.createElement('div');
                div.style.marginBottom = '12px';
                div.style.padding = '8px';
                div.style.border = '1px solid var(--color-border)';
                div.style.borderRadius = '4px';
                div.innerHTML = `
                    <b>Interface #${idx + 1}</b> — ${r.product || t('debug.unknown')}<br>
                    Path: ${r.path}<br>
                    VID: ${r.vendorId} PID: ${r.productId}<br>
                    Open: <span style="color:${r.openOk ? '#22c55e' : '#ef4444'}">${r.openOk ? 'OK' : 'FAIL — ' + r.openError}</span><br>
                    Usage: 0x${r.usage.toString(16).toUpperCase()} Page: 0x${r.usagePage.toString(16).toUpperCase()}<br>
                    Output report: ${r.outputReportByteLength} bytes | Feature: ${r.featureReportByteLength} bytes | Input: ${r.inputReportByteLength} bytes<br>
                    <b>WriteFile:</b> <span style="color:${r.writeFileResult.startsWith('OK') ? '#22c55e' : '#ef4444'}">${r.writeFileResult}</span><br>
                    <b>SetOutputReport:</b> <span style="color:${r.setOutputReportResult.startsWith('OK') ? '#22c55e' : '#ef4444'}">${r.setOutputReportResult}</span><br>
                    <b>SetFeature:</b> <span style="color:${r.setFeatureResult.startsWith('OK') ? '#22c55e' : '#ef4444'}">${r.setFeatureResult}</span>
                `;
                diagnoseResult.appendChild(div);
            });
            showToast(t('debug.diagnoseDone'), 'success');
        } catch(e) { showToast(t('common.error', {msg: e}), 'error'); }
        diagnoseBtn.disabled = false;
        diagnoseBtn.textContent = t('debug.diagnoseBtn');
    };

    // USB device list
    const usbTitle = document.createElement('h3');
    usbTitle.style.fontSize = '13px';
    usbTitle.style.fontWeight = '600';
    usbTitle.style.marginTop = '16px';
    usbTitle.style.marginBottom = '8px';
    usbTitle.textContent = t('debug.usbTitle');
    debugCard.appendChild(usbTitle);

    const scanBtn = document.createElement('button');
    scanBtn.className = 'btn btn-secondary btn-sm';
    scanBtn.textContent = t('debug.scanBtn');
    scanBtn.style.marginBottom = '8px';
    debugCard.appendChild(scanBtn);

    const usbTable = document.createElement('div');
    usbTable.style.fontSize = '12px';
    usbTable.style.fontFamily = 'monospace';
    debugCard.appendChild(usbTable);

    scanBtn.onclick = async () => {
        try {
            const devices = await window.go.main.App.ListUSBDevices();
            usbTable.innerHTML = '';
            if (!devices || devices.length === 0) {
                usbTable.textContent = t('debug.noHid');
                return;
            }
            const table = document.createElement('table');
            table.style.width = '100%';
            const thead = document.createElement('thead');
            const headerRow = document.createElement('tr');
            [t('debug.thVendor'), t('debug.thProduct'), t('debug.thManufacturer'), t('debug.thProductName')].forEach(h => {
                const th = document.createElement('th');
                th.textContent = h;
                th.style.textAlign = 'left';
                th.style.padding = '4px 8px';
                headerRow.appendChild(th);
            });
            thead.appendChild(headerRow);
            table.appendChild(thead);

            const tbody = document.createElement('tbody');
            devices.forEach(d => {
                const tr = document.createElement('tr');
                const supportedVendors = ['0x27BB', '0x04D8', '0x27B8', '0x2C0D', '0x0E53'];
                const isSupported = supportedVendors.includes(d.vendorId);
                if (isSupported) tr.style.backgroundColor = 'rgba(34,197,94,0.15)';
                [d.vendorId, d.productId, d.manufacturer, d.product].forEach(val => {
                    const td = document.createElement('td');
                    td.textContent = val || '-';
                    td.style.padding = '4px 8px';
                    tr.appendChild(td);
                });
                tbody.appendChild(tr);
            });
            table.appendChild(tbody);
            usbTable.appendChild(table);
            showToast(t('debug.devicesFound', {n: devices.length}), 'success');
        } catch(e) { showToast(t('common.error', {msg: e}), 'error'); }
    };

    page.appendChild(debugCard);

    app.appendChild(page);

    // Threshold rows
    function addThresholdRow(thr) {
        const row = document.createElement('div');
        row.style.display = 'flex';
        row.style.gap = '8px';
        row.style.alignItems = 'center';
        row.style.marginBottom = '8px';
        row.className = 'threshold-row';

        const minInput = document.createElement('input');
        minInput.type = 'number';
        minInput.value = thr.min;
        minInput.style.width = '60px';
        minInput.placeholder = 'Min';
        minInput.dataset.field = 'min';

        const maxInput = document.createElement('input');
        maxInput.type = 'number';
        maxInput.value = thr.max;
        maxInput.style.width = '60px';
        maxInput.placeholder = 'Max';
        maxInput.dataset.field = 'max';

        const colorInput = document.createElement('input');
        colorInput.type = 'color';
        colorInput.value = thr.color;
        colorInput.dataset.field = 'color';

        const soundCb = document.createElement('input');
        soundCb.type = 'checkbox';
        soundCb.checked = thr.sound;
        soundCb.dataset.field = 'sound';

        const soundLabel = document.createElement('span');
        soundLabel.textContent = t('thresholds.sound');
        soundLabel.style.fontSize = '12px';

        const blinkCb = document.createElement('input');
        blinkCb.type = 'checkbox';
        blinkCb.checked = thr.blink || false;
        blinkCb.dataset.field = 'blink';

        const blinkLabel = document.createElement('span');
        blinkLabel.textContent = t('thresholds.blink');
        blinkLabel.style.fontSize = '12px';

        const labelInput = document.createElement('input');
        labelInput.type = 'text';
        labelInput.value = thr.label;
        labelInput.placeholder = t('thresholds.label');
        labelInput.style.flex = '1';
        labelInput.dataset.field = 'label';

        const testBtn = document.createElement('button');
        testBtn.className = 'btn btn-secondary btn-sm';
        testBtn.textContent = t('thresholds.test');
        testBtn.onclick = async () => {
            try {
                await window.go.main.App.TestThreshold(colorInput.value, soundCb.checked, blinkCb.checked);
                showToast(t('thresholds.testSent'), 'success');
            } catch(e) { showToast('' + e, 'error'); }
        };

        const removeBtn = document.createElement('button');
        removeBtn.className = 'btn btn-danger btn-sm';
        removeBtn.textContent = '\u00d7';
        removeBtn.onclick = () => row.remove();

        row.appendChild(minInput);
        row.appendChild(maxInput);
        row.appendChild(colorInput);
        row.appendChild(soundCb);
        row.appendChild(soundLabel);
        row.appendChild(blinkCb);
        row.appendChild(blinkLabel);
        row.appendChild(labelInput);
        row.appendChild(testBtn);
        row.appendChild(removeBtn);
        thresholdTable.appendChild(row);
    }

    function getThresholds() {
        const rows = thresholdTable.querySelectorAll('.threshold-row');
        return Array.from(rows).map(row => ({
            min: parseInt(row.querySelector('[data-field="min"]').value) || 0,
            max: parseInt(row.querySelector('[data-field="max"]').value) || 0,
            color: row.querySelector('[data-field="color"]').value,
            sound: row.querySelector('[data-field="sound"]').checked,
            blink: row.querySelector('[data-field="blink"]').checked,
            label: row.querySelector('[data-field="label"]').value,
        }));
    }

    // Save handler
    saveBtn.onclick = async () => {
        const cfg = {
            endpointUrl: document.getElementById('endpoint-url').value,
            pollIntervalSeconds: parseInt(document.getElementById('poll-interval').value) || 300,
            jsonPath: document.getElementById('json-path').value || 'stats.orders_pending',
            tlsSkipVerify: document.getElementById('tls-skip').checked,
            soundEnabled: document.getElementById('sound-enabled').checked,
            soundOnChangeOnly: document.getElementById('sound-on-change').checked,
            language: document.getElementById('language-select').value,
            thresholds: getThresholds(),
        };

        try {
            await window.go.main.App.SaveConfig(cfg);
            showToast(t('settings.saved'), 'success');
            // Apply language change immediately
            const newLang = cfg.language === 'auto' ? await window.go.main.App.GetLanguage() : cfg.language;
            if (newLang !== getLang()) {
                setLang(newLang);
                translateStaticElements();
                navigate(); // Re-render current page
            }
        } catch(e) { showToast(t('common.error', {msg: e}), 'error'); }

        // Startup
        const startupEnabled = document.getElementById('startup-enabled').checked;
        try {
            await window.go.main.App.SetStartupEnabled(startupEnabled);
        } catch(e) { showToast(t('common.error', {msg: e}), 'error'); }
    };

    // Load initial config
    (async () => {
        try {
            const cfg = await window.go.main.App.GetConfig();
            document.getElementById('endpoint-url').value = cfg.endpointUrl || '';
            document.getElementById('poll-interval').value = cfg.pollIntervalSeconds || 300;
            document.getElementById('json-path').value = cfg.jsonPath || 'stats.orders_pending';
            document.getElementById('tls-skip').checked = cfg.tlsSkipVerify || false;
            document.getElementById('sound-enabled').checked = cfg.soundEnabled || false;
            document.getElementById('sound-on-change').checked = cfg.soundOnChangeOnly || false;
            document.getElementById('language-select').value = cfg.language || 'auto';

            const startup = await window.go.main.App.IsStartupEnabled();
            document.getElementById('startup-enabled').checked = startup;

            thresholdTable.innerHTML = '';
            if (cfg.thresholds && cfg.thresholds.length > 0) {
                cfg.thresholds.forEach(thr => addThresholdRow(thr));
            }
        } catch(e) { console.error(e); }
    })();
});
