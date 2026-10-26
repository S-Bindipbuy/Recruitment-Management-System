// ── EMS application shell ──
// Boot, auth view toggles, toasts, permission gating and the sidebar scroll
// indicator. Page loading and form requests live in nav.js.

function showAuth() {
    const authView = document.getElementById('auth-view');
    const appView = document.getElementById('app-view');
    if (authView) authView.classList.remove('hidden');
    if (appView) appView.classList.add('hidden');
}

function showApp() {
    const authView = document.getElementById('auth-view');
    const appView = document.getElementById('app-view');
    if (authView) authView.classList.add('hidden');
    if (appView) appView.classList.remove('hidden');
}

function logout() {
    clearAuthToken();
    const sidebarContainer = document.getElementById('sidebar-container');
    if (sidebarContainer) sidebarContainer.innerHTML = '';
    const headerContainer = document.getElementById('header-container');
    if (headerContainer) headerContainer.innerHTML = '';
    history.replaceState({}, '', '/');
    showAuth();
}

function showToast(message, type = 'error') {
    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        container.className = 'toast-container';
        document.body.appendChild(container);
    }

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    const icon = type === 'error' ? 'fa-circle-exclamation' : (type === 'success' ? 'fa-circle-check' : 'fa-circle-info');
    // The message comes from the server and can carry user-supplied text
    // (a candidate name, for instance). Build the toast with the DOM API so
    // that text is inserted as text and never parsed as markup.
    const iconEl = document.createElement('i');
    iconEl.className = `fa-solid ${icon}`;
    const textEl = document.createElement('span');
    textEl.textContent = message;
    toast.appendChild(iconEl);
    toast.appendChild(textEl);
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(-10px)';
        setTimeout(() => toast.remove(), 300);
    }, 4500);
}

async function downloadCV(candidateID) {
    const response = await emsFetch('/api/candidates/' + candidateID + '/cv');
    if (!response.ok) {
        alert('Unable to download this CV. Please try again.');
        return;
    }
    // The server sets Content-Disposition with a safe filename (cv-{id}.ext).
    // Extract it so the browser saves with the correct name.
    const disposition = response.headers.get('Content-Disposition') || '';
    const match = disposition.match(/filename[^;=\n]*=(['"]?)([^'";\n]+)\1/);
    const filename = match ? match[2] : ('cv-' + candidateID);
    const url = URL.createObjectURL(await response.blob());
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
}

async function initApp() {
    // Import separated components via API
    await Promise.all([
        loadComponent('auth-container', 'auth'),
        loadComponent('header-container', 'header')
    ]);

    const me = await fetchMe();

    if (getAuthToken()) {
        await loadComponent('sidebar-container', 'sidebar');
        initScrollIndicator();
        applyPermissionGate(me);
        showApp();
        loadPage(currentPageTarget() || 'dashboard');
    } else {
        showAuth();
    }
}

async function fetchMe() {
    const token = getAuthToken();
    if (!token) return null;
    try {
        const res = await fetch('/api/auth/me', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!res.ok) return null;
        return await res.json();
    } catch (err) {
        return null;
    }
}

// Hide navigation entries the current user's permissions don't cover.
// data-permission may list several comma-separated permissions; the entry is
// kept when the user holds at least one of them, matching the server's CanAny.
function applyPermissionGate(me) {
    if (!me) return;
    const perms = new Set(Array.isArray(me.permissions) ? me.permissions : []);
    const isAdmin = !!me.is_admin;
    document.querySelectorAll('.nav-link[data-permission]').forEach(el => {
        if (isAdmin) {
            return;
        }
        const wanted = el.dataset.permission.split(',').map(p => p.trim()).filter(Boolean);
        if (wanted.length === 0) return;
        if (!wanted.some(p => perms.has(p))) {
            el.remove();
        }
    });
    document.querySelectorAll('.sidebar-nav-section').forEach(section => {
        const nav = section.querySelector('nav.sidebar-nav');
        if (!nav || nav.querySelectorAll('a.nav-link').length === 0) {
            section.remove();
        }
    });
}

function bootApp() {
    console.log('[EMS] app.js booted, path =', location.pathname);
    initApp().catch(err => {
        console.error('[EMS] init failed:', err);
    });
}

// ── Sidebar Scroll Indicator ──
function initScrollIndicator() {
    const navContent = document.querySelector('.sidebar-nav-content');
    const progress = document.querySelector('.sidebar-scroll-progress');
    const line = document.getElementById('scrollIndicator');
    if (!navContent || !progress || !line) return;

    function checkScroll() {
        const { scrollTop, scrollHeight, clientHeight } = navContent;
        const isScrollable = scrollHeight > clientHeight;
        if (!isScrollable) {
            line.style.display = 'none';
            return;
        }
        line.style.display = '';
        const pct = (scrollTop / (scrollHeight - clientHeight)) * 100;
        progress.style.width = pct + '%';
    }

    navContent.addEventListener('scroll', checkScroll);
    window.addEventListener('resize', checkScroll);
    checkScroll();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', bootApp);
} else {
    bootApp();
}
