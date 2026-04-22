registerPage('#about', function(app) {
    const page = document.createElement('div');

    const card = document.createElement('div');
    card.className = 'card';
    card.style.maxWidth = '520px';

    const title = document.createElement('h2');
    title.className = 'section-title';
    title.textContent = t('about.title');
    card.appendChild(title);

    const versionEl = document.createElement('div');
    versionEl.style.fontSize = '13px';
    versionEl.style.color = 'var(--color-text-muted)';
    versionEl.style.marginBottom = '20px';
    versionEl.textContent = 'Version ...';
    card.appendChild(versionEl);

    (async () => {
        try {
            const v = await window.go.main.App.GetVersion();
            versionEl.textContent = 'Version ' + v;
        } catch(e) { /* ignore */ }
    })();

    const desc = document.createElement('p');
    desc.style.marginBottom = '20px';
    desc.style.lineHeight = '1.6';
    desc.textContent = t('about.description');
    card.appendChild(desc);

    function linkBlock(labelKey, url) {
        const div = document.createElement('div');
        div.style.marginBottom = '12px';

        const lbl = document.createElement('div');
        lbl.style.fontSize = '12px';
        lbl.style.color = 'var(--color-text-muted)';
        lbl.style.marginBottom = '2px';
        lbl.textContent = t(labelKey);
        div.appendChild(lbl);

        const link = document.createElement('a');
        link.href = url;
        link.textContent = url;
        link.style.color = 'var(--color-primary)';
        link.style.fontSize = '13px';
        link.style.textDecoration = 'none';
        link.onmouseover = function() { link.style.textDecoration = 'underline'; };
        link.onmouseout = function() { link.style.textDecoration = 'none'; };
        link.onclick = function(e) {
            e.preventDefault();
            window.runtime.BrowserOpenURL(url);
        };
        div.appendChild(link);

        return div;
    }

    const editorTitle = document.createElement('h3');
    editorTitle.style.fontSize = '13px';
    editorTitle.style.fontWeight = '600';
    editorTitle.style.marginBottom = '12px';
    editorTitle.style.marginTop = '8px';
    editorTitle.textContent = t('about.editor');
    card.appendChild(editorTitle);

    card.appendChild(linkBlock('about.website', 'https://www.ecommercerie.fr'));
    card.appendChild(linkBlock('about.source', 'https://github.com/ecommercerie/picklight'));

    const creditsTitle = document.createElement('h3');
    creditsTitle.style.fontSize = '13px';
    creditsTitle.style.fontWeight = '600';
    creditsTitle.style.marginBottom = '12px';
    creditsTitle.style.marginTop = '20px';
    creditsTitle.textContent = t('about.credits');
    card.appendChild(creditsTitle);

    const creditsText = document.createElement('p');
    creditsText.style.marginBottom = '8px';
    creditsText.style.lineHeight = '1.6';
    creditsText.style.fontSize = '13px';
    creditsText.textContent = t('about.creditsText');
    card.appendChild(creditsText);

    card.appendChild(linkBlock('about.reference', 'https://github.com/JnyJny/busylight'));

    const coauthorTitle = document.createElement('h3');
    coauthorTitle.style.fontSize = '13px';
    coauthorTitle.style.fontWeight = '600';
    coauthorTitle.style.marginBottom = '12px';
    coauthorTitle.style.marginTop = '20px';
    coauthorTitle.textContent = t('about.coauthored');
    card.appendChild(coauthorTitle);

    card.appendChild(linkBlock('about.coauthored', 'https://claude.ai/claude-code'));

    const licenseNote = document.createElement('div');
    licenseNote.style.marginTop = '20px';
    licenseNote.style.padding = '12px';
    licenseNote.style.background = 'var(--color-bg)';
    licenseNote.style.border = '1px solid var(--color-border)';
    licenseNote.style.borderRadius = '4px';
    licenseNote.style.fontSize = '12px';
    licenseNote.style.color = 'var(--color-text-muted)';
    licenseNote.style.lineHeight = '1.5';
    licenseNote.textContent = t('about.license');
    card.appendChild(licenseNote);

    page.appendChild(card);
    app.appendChild(page);
});
