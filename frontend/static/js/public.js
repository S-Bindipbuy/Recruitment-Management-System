// ── Public pages (job listing + application form) ──
// The public layout has no app shell and no auth token; these two forms post
// their files to /apply and swap a success/error block in place.

function showPublicSuccess(form, successId) {
    const success = document.getElementById(successId);
    if (success) success.classList.remove('hidden');
    if (form) form.classList.add('hidden');
}

async function submitPublicForm(form, opts) {
    const button = form.querySelector('[type="submit"]');
    if (button) button.disabled = true;

    try {
        const res = await fetch(form.getAttribute('action') || '/apply', {
            method: (form.getAttribute('method') || 'POST').toUpperCase(),
            body: new FormData(form),
        });

        if (res.ok) {
            showPublicSuccess(form, opts.successId);
            return;
        }

        let message = 'Something went wrong.';
        try {
            const data = await res.json();
            if (data && (data.error || data.message)) message = data.error || data.message;
        } catch (e) {
            const text = (await res.text()).trim();
            if (text) message = text;
        }

        if (opts.errorTargetId) {
            const target = document.getElementById(opts.errorTargetId);
            if (target) {
                target.textContent = message;
                target.classList.remove('hidden');
            }
            if (form) form.classList.add('hidden');
        } else if (opts.errorTextId) {
            const target = document.getElementById(opts.errorTextId);
            if (target) target.textContent = message;
        } else {
            alert(message);
        }
    } catch (e) {
        if (opts.errorTargetId) {
            const target = document.getElementById(opts.errorTargetId);
            if (target) {
                target.textContent = 'Connection error. Please try again.';
                target.classList.remove('hidden');
            }
        } else {
            alert('Connection error. Please try again.');
        }
    } finally {
        if (button) button.disabled = false;
    }
}

document.addEventListener('submit', function(e) {
    const form = e.target;

    if (form.id === 'apply-form') {
        e.preventDefault();
        submitPublicForm(form, { successId: 'apply-success', errorTargetId: 'apply-error' });
    } else if (form.id === 'cv-submit-form') {
        e.preventDefault();
        submitPublicForm(form, { successId: 'cv-submit-success', errorTextId: 'cv-submit-error' });
    }
});
