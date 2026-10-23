// ── Auth token helpers ──
function getAuthToken() {
    return sessionStorage.getItem('token');
}

function setAuthToken(token) {
    sessionStorage.setItem('token', token);
}

function clearAuthToken() {
    sessionStorage.removeItem('token');
}

function authHeaders() {
    const token = getAuthToken();
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    return headers;
}

// ── JSON API helpers ──
async function apiFetch(url, opts = {}) {
    const res = await fetch(url, {
        ...opts,
        headers: {
            ...authHeaders(),
            ...(opts.headers || {})
        }
    });

    if (res.status === 401) {
        if (typeof showAuth === 'function') showAuth();
        throw new Error('Unauthorized');
    }

    if (!res.ok) {
        throw new Error('Request failed: ' + res.status);
    }

    return res.json();
}

// ── Server-rendered component fragments (auth form, sidebar, header) ──
async function fetchComponent(name) {
    const res = await fetch(`/api/components/${encodeURIComponent(name)}`);
    if (!res.ok) {
        throw new Error(`Failed to fetch component: ${name}`);
    }
    return res.text();
}

async function loadComponent(targetId, componentName) {
    const target = document.getElementById(targetId);
    if (!target) return;
    try {
        const html = await fetchComponent(componentName);
        target.innerHTML = html;
        if (typeof runScripts === 'function') {
            runScripts(target);
        }
    } catch (err) {
        console.error(`Error loading component ${componentName}:`, err);
    }
}
