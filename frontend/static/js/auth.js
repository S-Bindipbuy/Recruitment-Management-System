// ── Auth view toggles ──
function showLogin() {
    const loginSec = document.getElementById('login-section');
    const regSec = document.getElementById('register-section');
    if (loginSec) loginSec.classList.remove('hidden');
    if (regSec) regSec.classList.add('hidden');

    const tabLogin = document.getElementById('tab-login');
    const tabReg = document.getElementById('tab-register');
    if (tabLogin) {
        tabLogin.className = "auth-tab-btn active";
    }
    if (tabReg) {
        tabReg.className = "auth-tab-btn";
    }
}

function showRegister() {
    const loginSec = document.getElementById('login-section');
    const regSec = document.getElementById('register-section');
    if (loginSec) loginSec.classList.add('hidden');
    if (regSec) regSec.classList.remove('hidden');

    const tabLogin = document.getElementById('tab-login');
    const tabReg = document.getElementById('tab-register');
    if (tabLogin) {
        tabLogin.className = "auth-tab-btn";
    }
    if (tabReg) {
        tabReg.className = "auth-tab-btn active";
    }
}

// ── View switching lives in app.js (showAuth/showApp/logout) ──

// ── Event Listeners for Login & Register forms via Delegation ──
document.addEventListener('submit', async function(e) {
    if (!e.target) return;

    if (e.target.id === 'login-form') {
        e.preventDefault();
        const fd = new FormData(e.target);
        try {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: fd.get('email'), password: fd.get('password') })
            });
            const data = await res.json();
            if (!res.ok) {
                const errEl = document.getElementById('login-error');
                const errTxt = document.getElementById('login-error-text');
                if (errTxt) errTxt.textContent = data.error || 'Login failed';
                else if (errEl) errEl.textContent = data.error || 'Login failed';
                if (errEl) errEl.classList.remove('hidden');
                return;
            }
            setAuthToken(data.token);
            openApp(currentPageTarget() || 'dashboard');
        } catch (err) {
            const errEl = document.getElementById('login-error');
            const errTxt = document.getElementById('login-error-text');
            if (errTxt) errTxt.textContent = 'Connection error';
            else if (errEl) errEl.textContent = 'Connection error';
            if (errEl) errEl.classList.remove('hidden');
        }
    } else if (e.target.id === 'register-form') {
        e.preventDefault();
        const fd = new FormData(e.target);
        try {
            const res = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: fd.get('name'), email: fd.get('email'), password: fd.get('password') })
            });
            const data = await res.json();
            if (!res.ok) {
                const errEl = document.getElementById('register-error');
                const errTxt = document.getElementById('register-error-text');
                if (errTxt) errTxt.textContent = data.error || 'Registration failed';
                else if (errEl) errEl.textContent = data.error || 'Registration failed';
                if (errEl) errEl.classList.remove('hidden');
                return;
            }
            showLogin();
        } catch (err) {
            const errEl = document.getElementById('register-error');
            const errTxt = document.getElementById('register-error-text');
            if (errTxt) errTxt.textContent = 'Connection error';
            else if (errEl) errEl.textContent = 'Connection error';
            if (errEl) errEl.classList.remove('hidden');
        }
    }
});

