const pages = {};

function registerPage(hash, render) {
    pages[hash] = render;
}

function navigate() {
    const hash = window.location.hash || '#status';
    const app = document.getElementById('app');
    app.innerHTML = '';

    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.toggle('active', link.getAttribute('href') === hash);
    });

    const render = pages[hash];
    if (render) {
        render(app);
    }
}

function showToast(message, type) {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = 'toast show ' + (type || 'success');
    setTimeout(() => { toast.className = 'toast'; }, 3000);
}

// Translate all elements with data-i18n attribute
function translateStaticElements() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        el.textContent = t(el.getAttribute('data-i18n'));
    });
}

window.addEventListener('hashchange', navigate);
window.addEventListener('DOMContentLoaded', async function() {
    // Init language from backend before first render
    try {
        const lang = await window.go.main.App.GetLanguage();
        setLang(lang);
    } catch(e) { /* default fr */ }

    translateStaticElements();
    navigate();

    // Display version in sidebar
    try {
        const version = await window.go.main.App.GetVersion();
        const el = document.getElementById('app-version');
        if (el && version) el.textContent = 'v' + version;
    } catch(e) { /* ignore */ }
});
