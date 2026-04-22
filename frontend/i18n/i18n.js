// i18n — lightweight translation system for PickLight
// Usage: t('status.pollBtn') or t('update.available', {current: '1.0', latest: '1.1'})

(function() {
    'use strict';

    const locales = { fr: window.I18N_FR, en: window.I18N_EN };
    let currentLang = 'fr';

    // Validate key parity between all locales at load time
    (function validate() {
        const langs = Object.keys(locales);
        if (langs.length < 2) return;
        const refLang = langs[0];
        const refKeys = Object.keys(locales[refLang]).sort();
        for (let i = 1; i < langs.length; i++) {
            const lang = langs[i];
            const keys = Object.keys(locales[lang]).sort();
            // Keys in ref but missing in lang
            refKeys.forEach(function(k) {
                if (!(k in locales[lang])) {
                    console.warn('[i18n] Missing ' + lang.toUpperCase() + ' key: "' + k + '"');
                }
            });
            // Keys in lang but missing in ref
            keys.forEach(function(k) {
                if (!(k in locales[refLang])) {
                    console.warn('[i18n] Missing ' + refLang.toUpperCase() + ' key: "' + k + '"');
                }
            });
        }
    })();

    // Set the current language
    window.setLang = function(lang) {
        if (locales[lang]) {
            currentLang = lang;
        } else {
            console.warn('[i18n] Unknown locale: ' + lang + ', falling back to fr');
            currentLang = 'fr';
        }
    };

    // Get the current language
    window.getLang = function() {
        return currentLang;
    };

    // Translate a key with optional interpolation: t('key', {var: 'value'})
    window.t = function(key, params) {
        var dict = locales[currentLang] || locales.fr;
        var str = dict[key];
        if (str === undefined) {
            // Fallback to other locale
            for (var lang in locales) {
                if (locales[lang][key] !== undefined) {
                    str = locales[lang][key];
                    console.warn('[i18n] Key "' + key + '" missing in ' + currentLang.toUpperCase() + ', using ' + lang.toUpperCase() + ' fallback');
                    break;
                }
            }
        }
        if (str === undefined) {
            console.warn('[i18n] Unknown key: "' + key + '"');
            return '[' + key + ']';
        }
        if (params) {
            Object.keys(params).forEach(function(k) {
                str = str.replace(new RegExp('\\{' + k + '\\}', 'g'), params[k]);
            });
        }
        return str;
    };
})();
