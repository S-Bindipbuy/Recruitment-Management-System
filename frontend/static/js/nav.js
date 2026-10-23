// ── EMS navigation and requests (vanilla JS, replaces htmx) ──
//
// One place owns every request the app makes outside the JSON API:
//   • internal links (<a href="/pages/...">)
//   • forms marked data-ajax
//   • buttons marked data-request
// Server responses are HTML fragments swapped into a target element, or JSON
// errors that are shown as toasts.
//
// Attribute API (the templates use these):
//   data-ajax               form is submitted with fetch instead of a page load
//   data-request="POST"     button/link performs a request
//   data-url="..."          request URL (buttons; forms fall back to action)
//   data-method="PUT"       override the HTTP verb for a form or button
//   data-target="#page"     element whose innerHTML receives the response
//   data-swap="outerHTML"   replace the target element itself
//   data-swap="none"        do not swap (handler only returns a status)
//   data-push-url="/path"   push this URL into history after a form post
//   data-confirm="..."      ask before sending
//   data-vals='{"k":"v"}'   extra request values (JSON)
//   data-include="closest form"  also send the enclosing form's fields
//   data-close-dialog        close the enclosing <dialog> on success
//   data-error-target="id"   show the error body in that element on failure
//   data-form-status="..."   write this text into [data-form-status] on success

// ── Current location helpers ──
function currentPageName() {
    const parts = location.pathname.split('/').filter(Boolean);
    if (parts[0] !== 'pages') return null;
    return parts[1] === 'evaluation-fill' ? 'evaluations' : parts[1];
}

function currentPagePath() {
    const parts = location.pathname.split('/').filter(Boolean);
    if (parts[0] !== 'pages') return null;
    return parts.slice(1).join('/');
}

// Path relative to /pages/, with the current query string preserved.
function currentPageTarget() {
    const path = currentPagePath();
    return path ? path + location.search : null;
}

function updateActiveNav() {
    const active = currentPageName();
    document.querySelectorAll('.nav-link').forEach(el => {
        el.classList.toggle('active', !!active && el.dataset.page === active);
    });
}

// ── Auth-aware request ──
// Returns the Response. Redirects to the login view on 401.
async function emsFetch(url, options = {}) {
    const token = getAuthToken();
    const headers = Object.assign({}, options.headers || {});
    if (token) headers['Authorization'] = 'Bearer ' + token;

    const res = await fetch(url, Object.assign({}, options, { headers }));

    // Server-driven events ride on HX-Trigger ({"showToast": {...}}).
    const trigger = res.headers.get('HX-Trigger');
    if (trigger) dispatchTriggers(trigger);

    if (res.status === 401) {
        clearAuthToken();
        showAuth();
        throw new Error('Unauthorized');
    }
    return res;
}

function dispatchTriggers(raw) {
    let triggers;
    try { triggers = JSON.parse(raw); } catch (e) { return; }
    Object.keys(triggers).forEach(name => {
        const detail = triggers[name];
        if (name === 'showToast') {
            const msg = typeof detail === 'string' ? detail : detail.message;
            const type = (detail && detail.type) || 'error';
            if (msg) showToast(msg, type);
        } else {
            document.dispatchEvent(new CustomEvent(name, { detail }));
        }
    });
}

// ── Response handling ──
async function handleResponse(res, opts = {}) {
    const contentType = res.headers.get('Content-Type') || '';

    if (contentType.includes('application/json')) {
        let data = {};
        try { data = await res.json(); } catch (e) { /* empty body */ }
        if (!res.ok) {
            const msg = data.error || data.message || 'Request failed (' + res.status + ').';
            if (opts.errorTarget) {
                opts.errorTarget.textContent = msg;
                opts.errorTarget.classList.remove('hidden');
            } else {
                showToast(msg, 'error');
            }
            return false;
        }
        if (data.message) showToast(data.message, 'success');
        return true;
    }

    const html = await res.text();

    if (!res.ok) {
        const msg = oneLine(html) || 'Request failed (' + res.status + ').';
        if (opts.errorTarget) {
            opts.errorTarget.textContent = msg;
            opts.errorTarget.classList.remove('hidden');
        } else {
            showToast(msg, 'error');
        }
        return false;
    }

    if (opts.swap !== 'none' && html.trim()) {
        swapContent(opts.target, html, opts.swap);
    }
    return true;
}

function oneLine(html) {
    return (html || '').replace(/\s+/g, ' ').trim().slice(0, 300);
}

// Replace the target's content and run any inline scripts it brought with it.
function swapContent(target, html, mode) {
    const el = target || document.getElementById('page');
    if (!el) return;

    if (mode === 'outerHTML') {
        const tpl = document.createElement('template');
        tpl.innerHTML = html;
        const replacement = tpl.content.firstElementChild;
        if (!replacement) return;
        el.replaceWith(replacement);
        runScripts(replacement);
        updateActiveNav();
        return;
    } else {
        el.innerHTML = html;
        runScripts(el);
    }

    updateActiveNav();
    const mc = document.querySelector('.main-content');
    if (mc) mc.scrollTop = 0;
}

// Scripts inserted through innerHTML do not execute; re-create them.
function runScripts(root) {
    if (!root || !root.querySelectorAll) return;
    root.querySelectorAll('script').forEach(oldScript => {
        const fresh = document.createElement('script');
        Array.from(oldScript.attributes).forEach(a => fresh.setAttribute(a.name, a.value));
        fresh.textContent = oldScript.textContent;
        oldScript.parentNode.replaceChild(fresh, oldScript);
    });
}

// ── Requests ──
function formBody(form) {
    const formData = new FormData(form);
    const hasFile = Array.from(formData.values()).some(v => v instanceof File && v.size > 0);
    if (hasFile) {
        return { body: formData, headers: {} };
    }
    return {
        body: new URLSearchParams(formData).toString(),
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    };
}

function sendForm(form) {
    const method = (form.dataset.method || form.getAttribute('method') || 'POST').toUpperCase();
    const url = form.dataset.url || form.getAttribute('action');
    const targetSelector = form.dataset.target;
    const target = targetSelector ? document.querySelector(targetSelector) : document.getElementById('page');

    if (form.dataset.confirm && !confirm(form.dataset.confirm)) return;

    const encoded = formBody(form);
    let fetchUrl = url;
    let init;

    if (method === 'GET') {
        const query = encoded.body.toString ? encoded.body.toString() : '';
        fetchUrl = url + (query ? (url.includes('?') ? '&' : '?') + query : '');
        init = { method: 'GET' };
    } else {
        init = { method: method, body: encoded.body, headers: encoded.headers };
    }

    const errorTarget = form.dataset.errorTarget ? document.getElementById(form.dataset.errorTarget) : null;

    emsFetch(fetchUrl, init)
        .then(res => handleResponse(res, {
            target,
            swap: form.dataset.swap,
            errorTarget,
        }).then(ok => {
            if (!ok) return;
            if (form.dataset.closeDialog) {
                const dialog = form.closest('dialog');
                if (dialog) dialog.close();
            }
            if (form.dataset.formStatus) {
                const status = form.querySelector('[data-form-status]');
                if (status) status.textContent = form.dataset.formStatus;
            }
            // GET forms (filters) put their values in the address bar so the
            // filtered view can be bookmarked and restored on back/forward.
            if (method === 'GET') {
                history.pushState({ ems: true }, '', fetchUrl);
            } else if (form.dataset.pushUrl) {
                history.pushState({ ems: true }, '', form.dataset.pushUrl);
            }
        }))
        .catch(err => {
            if (err.message !== 'Unauthorized') showToast('Request failed. Please try again.', 'error');
        });
}

function sendRequest(el) {
    const method = (el.dataset.request || 'GET').toUpperCase();
    const url = el.dataset.url;
    if (!url) return;
    if (el.dataset.confirm && !confirm(el.dataset.confirm)) return;

    // Values come from data-vals and, for buttons inside a form, the form itself.
    let values = {};
    if (el.dataset.vals) {
        try { values = JSON.parse(el.dataset.vals); } catch (e) { values = {}; }
    }
    const include = el.dataset.include === 'closest form' ? el.closest('form') : null;
    const includeForm = include || (el.closest('form') || null);

    let fetchUrl = url;
    let init = { method };

    if (method === 'GET') {
        const params = new URLSearchParams(values);
        if (includeForm) new URLSearchParams(new FormData(includeForm)).forEach((v, k) => params.set(k, v));
        const query = params.toString();
        fetchUrl = url + (query ? (url.includes('?') ? '&' : '?') + query : '');
    } else {
        const params = new URLSearchParams(values);
        if (includeForm) new URLSearchParams(new FormData(includeForm)).forEach((v, k) => params.set(k, v));
        init.body = params.toString();
        init.headers = { 'Content-Type': 'application/x-www-form-urlencoded' };
    }

    const targetSelector = el.dataset.target;
    const target = targetSelector ? document.querySelector(targetSelector) : document.getElementById('page');

    emsFetch(fetchUrl, init)
        .then(res => handleResponse(res, { target, swap: el.dataset.swap }))
        .catch(err => {
            if (err.message !== 'Unauthorized') showToast('Request failed. Please try again.', 'error');
        });
}

// Load a page fragment into #page (used for navigation and history).
function loadPage(pathWithQuery, options = {}) {
    if (!getAuthToken()) {
        showAuth();
        return;
    }
    const targetSelector = options.target || '#page';
    const target = document.querySelector(targetSelector) || document.getElementById('page');
    const url = '/pages/' + pathWithQuery;

    emsFetch(url, { method: 'GET' })
        .then(res => handleResponse(res, { target, swap: options.swap }))
        .catch(err => {
            if (err.message !== 'Unauthorized') {
                if (target) {
                    target.innerHTML = '<div class="badge badge-danger" style="padding:1rem;display:block;width:100%;">Failed to load this page.</div>';
                }
            }
        });
}

// ── Navigation ──
function navigateTo(page, push = true) {
    const url = '/pages/' + page;
    if (push) {
        history.pushState({ ems: true }, '', url);
    } else {
        history.replaceState({ ems: true }, '', url);
    }
    showApp();
    loadPage(page);
}

function openApp(page) {
    navigateTo(page, false);
}

// ── Event delegation ──
document.addEventListener('click', function(e) {
    // Let the browser handle new-tab / download / external links.
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

    const requestEl = e.target.closest('[data-request]');
    if (requestEl) {
        e.preventDefault();
        sendRequest(requestEl);
        return;
    }

    const navLink = e.target.closest('a[href]');
    if (!navLink) return;
    if (navLink.target === '_blank' || navLink.hasAttribute('download')) return;

    const href = navLink.getAttribute('href');
    if (!href || !href.startsWith('/pages/')) return;

    e.preventDefault();
    if (href === location.pathname + location.search) return;
    history.pushState({ ems: true }, '', href);
    showApp();
    loadPage(href.slice('/pages/'.length));
});

document.addEventListener('submit', function(e) {
    const form = e.target;
    if (!(form instanceof HTMLFormElement) || !form.hasAttribute('data-ajax')) return;
    e.preventDefault();
    sendForm(form);
});

// Browser back/forward: re-query the server for the current URL.
window.addEventListener('popstate', function() {
    if (!getAuthToken()) {
        showAuth();
        return;
    }
    showApp();
    updateActiveNav();
    loadPage(currentPageTarget() || 'dashboard');
});

// Called from inline handlers that used to use htmx.ajax(...).
function emsRequest(method, url, target, body) {
    const opts = { method: method };
    if (body !== undefined) {
        opts.body = new URLSearchParams(body).toString();
        opts.headers = { 'Content-Type': 'application/x-www-form-urlencoded' };
    }
    const targetEl = typeof target === 'string' ? document.querySelector(target) : target;
    emsFetch(url, opts)
        .then(res => handleResponse(res, { target: targetEl || document.getElementById('page') }))
        .catch(err => {
            if (err.message !== 'Unauthorized') showToast('Request failed. Please try again.', 'error');
        });
}
