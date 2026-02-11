import './style.css';
import { applyLanguage, getLanguage, setLanguage } from './i18n.js';
import { renderProjectsPage } from './projects.js';
import { renderStoryboardPage } from './storyboard.js';
import { renderStoryboardV2Page } from './storyboard_v2.js';

// ============================================================
// Router
// ============================================================

function getRoute() {
    const hash = window.location.hash || '#/';
    if (hash === '#/' || hash === '#' || hash === '') return { view: 'projects' };
    const match = hash.match(/^#\/projects\/(\d+)$/);
    if (match) return { view: 'storyboard', id: parseInt(match[1]) };
    return { view: 'projects' };
}

async function navigate() {
    const route = getRoute();
    const container = document.getElementById('page-content');
    if (!container) return;

    if (route.view === 'storyboard') {
        // Fetch project to determine version, then route accordingly
        try {
            const data = await window.go.main.App.GetProject(route.id);
            const version = data.project.model_version || 'v1.x';
            if (version === 'v2.0') {
                await renderStoryboardV2Page(container, route.id, data);
            } else {
                await renderStoryboardPage(container, route.id);
            }
        } catch (err) {
            container.innerHTML = `<div class="alert alert-error mt-4">Failed to load project: ${err}</div>`;
        }
    } else {
        await renderProjectsPage(container);
    }

    applyLanguage();
}

// ============================================================
// App Shell
// ============================================================

function renderShell() {
    document.getElementById('app').innerHTML = `
        <!-- Top Navbar -->
        <div class="navbar bg-base-100 border-b border-base-content/10 mb-4 sticky top-0 z-50">
            <div class="navbar-start">
                <a href="#/" class="btn btn-ghost btn-sm text-lg gap-2">
                    <span class="material-symbols-outlined text-xl">movie_creation</span>
                    <span data-i18n="app.title">Seedance Manager</span>
                </a>
            </div>
            <div class="navbar-end">
                <button id="open-settings-btn" class="btn btn-ghost btn-sm gap-1">
                    <svg xmlns="http://www.w3.org/2000/svg" height="20" viewBox="0 -960 960 960" width="20" fill="currentColor">
                        <path d="m370-80-16-128q-13-5-24.5-12T307-235l-119 50L78-375l103-78q-1-7-1-13.5v-27q0-6.5 1-13.5L78-585l110-190 119 50q11-8 23-15t24-12l16-128h220l16 128q13 5 24.5 12t22.5 15l119-50 110 190-103 78q1 7 1 13.5v27q0 6.5-1 13.5l103 78-110 190-119-50q-11 8-23 15t-24 12l-16 128H370Zm112-260q58 0 99-41t41-99q0-58-41-99t-99-41q-59 0-99.5 41T342-480q0 58 40.5 99t99.5 41Zm0-80q-25 0-42.5-17.5T422-480q0-25 17.5-42.5T482-540q25 0 42.5 17.5T542-480q0 25-17.5 42.5T482-420Z" />
                    </svg>
                    <span data-i18n="nav.settings">Settings</span>
                </button>
            </div>
        </div>

        <!-- Settings Modal -->
        <dialog id="settings-dialog" class="modal">
            <div class="modal-box">
                <h3 class="text-lg font-bold mb-4" data-i18n="settings.title">Settings</h3>
                <div class="form-control mb-3">
                    <label class="label py-1">
                        <span class="label-text text-sm font-medium" data-i18n="settings.language">Language</span>
                    </label>
                    <select id="lang-selector" class="select select-bordered select-sm w-full">
                        <option value="en">English (US)</option>
                        <option value="zh">简体中文 (Simplified Chinese)</option>
                    </select>
                </div>
                <div class="form-control mb-3">
                    <label class="label py-1">
                        <span class="label-text text-sm font-medium" data-i18n="settings.apikey">API Key</span>
                    </label>
                    <input type="password" id="apikey-input" data-i18n="settings.apikey.placeholder"
                        placeholder="Enter new API Key" class="input input-bordered input-sm w-full">
                    <label class="label py-1">
                        <span class="label-text-alt text-xs" data-i18n="settings.apikey.hint">
                            Updating this will restart the client connection.
                        </span>
                    </label>
                </div>
                <div class="modal-action">
                    <button id="cancel-settings-btn" class="btn btn-ghost btn-sm" data-i18n="btn.cancel">Cancel</button>
                    <button id="save-apikey-btn" class="btn btn-primary btn-sm" data-i18n="btn.save">Save</button>
                </div>
            </div>
            <form method="dialog" class="modal-backdrop"><button>close</button></form>
        </dialog>

        <!-- Page Content -->
        <div class="container mx-auto px-4 max-w-6xl" id="page-content"></div>

        <!-- Footer -->
        <footer class="footer footer-center p-6 bg-base-100 border-t border-base-content/10 text-base-content/60 mt-8 text-sm">
            <p data-i18n="footer.text">&copy; 2026 Seedance Client</p>
        </footer>
    `;

    // Attach settings events
    document.getElementById('open-settings-btn').addEventListener('click', () => {
        document.getElementById('settings-dialog').showModal();
    });

    document.getElementById('cancel-settings-btn').addEventListener('click', () => {
        document.getElementById('settings-dialog').close();
    });

    const langSelector = document.getElementById('lang-selector');
    langSelector.value = getLanguage();
    langSelector.addEventListener('change', (e) => {
        setLanguage(e.target.value);
        applyLanguage();
    });

    document.getElementById('save-apikey-btn').addEventListener('click', async () => {
        const apiKey = document.getElementById('apikey-input').value;
        if (apiKey) {
            try {
                await window.go.main.App.UpdateAPIKey(apiKey);
            } catch (err) {
                alert('Failed to update API key: ' + err);
                return;
            }
            document.getElementById('settings-dialog').close();
            document.getElementById('apikey-input').value = '';
        }
    });
}

// ============================================================
// Initialize
// ============================================================

window.addEventListener('DOMContentLoaded', () => {
    renderShell();
    applyLanguage();
    navigate();
});

window.addEventListener('hashchange', navigate);
