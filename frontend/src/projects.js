import { applyLanguage } from './i18n.js';

// ============================================================
// Projects Page
// ============================================================

export async function renderProjectsPage(container) {
    try {
        const data = await window.go.main.App.GetProjects();
        renderHTML(container, data);
        attachEvents(container);
        applyLanguage();
    } catch (err) {
        container.innerHTML = `<div class="alert alert-error mt-4">Failed to load projects: ${escapeHtml(String(err))}</div>`;
    }
}

function renderHTML(container, data) {
    const projects = data.projects || [];
    const stats = data.stats || {};

    const projectCards = projects.map(p => {
        const versionBadge = p.model_version === 'v2.0'
            ? '<span class="badge badge-primary badge-xs">2.0</span>'
            : '<span class="badge badge-ghost badge-xs">1.x</span>';

        return `
        <div class="card card-compact bg-base-100 border border-base-content/10">
            <div class="card-body">
                <div class="flex justify-between items-start">
                    <div class="flex items-center gap-2 min-w-0">
                        <h2 class="card-title text-base truncate">${escapeHtml(p.name)}</h2>
                        ${versionBadge}
                    </div>
                    <button data-delete-project="${p.id}" class="btn btn-ghost btn-circle btn-xs text-error flex-shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                            <path d="M280-120q-33 0-56.5-23.5T200-200v-520h-40v-80h200v-40h240v40h200v80h-40v520q0 33-23.5 56.5T680-120H280Zm400-600H280v520h400v-520ZM360-280h80v-360h-80v360Zm160 0h80v-360h-80v360ZM280-720v520-520Z" />
                        </svg>
                    </button>
                </div>
                <p class="text-xs text-base-content/50">
                    ID: ${p.id} &bull; ${formatDate(p.created_at)}
                </p>
                <div class="card-actions justify-end mt-2">
                    <a href="#/projects/${p.id}" class="btn btn-secondary btn-sm" data-i18n="projects.open">Open</a>
                </div>
            </div>
        </div>
    `}).join('');

    container.innerHTML = `
        <div class="mb-6 flex flex-col md:flex-row justify-between items-start md:items-center gap-3">
            <div>
                <h1 class="text-2xl font-bold text-base-content" data-i18n="projects.title">Projects</h1>
                <p class="text-sm text-base-content/60" data-i18n="projects.subtitle">Manage your video generation campaigns</p>
            </div>
            <div class="join">
                <input type="text" id="new-project-name" data-i18n="projects.create.placeholder"
                    placeholder="New Project Name" required class="input input-bordered input-sm join-item w-48">
                <select id="new-project-version" class="select select-bordered select-sm join-item">
                    <option value="v2.0">Seedance 2.0</option>
                    <option value="v1.x">Seedance 1.5 &amp; earlier</option>
                </select>
                <button id="create-project-btn" class="btn btn-primary btn-sm join-item gap-1">
                    <span class="text-lg font-bold">+</span> <span data-i18n="projects.create.btn">Create</span>
                </button>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            ${projectCards}
        </div>

        <!-- Data Monitoring Stats -->
        <div class="mt-8 card card-compact bg-base-100 border border-base-content/10">
            <div class="card-body">
                <h2 class="card-title text-lg mb-4">Data Monitoring</h2>
                <div class="stats stats-vertical lg:stats-horizontal border border-base-content/10 w-full">
                    <div class="stat py-3">
                        <div class="stat-title text-xs">Total Videos</div>
                        <div class="stat-value text-xl text-primary">${stats.total_videos || 0}</div>
                    </div>
                    <div class="stat py-3">
                        <div class="stat-title text-xs">Total Tokens</div>
                        <div class="stat-value text-xl text-secondary">${stats.total_token_usage || 0}</div>
                        <div class="stat-desc text-xs">Output Tokens</div>
                    </div>
                    <div class="stat py-3">
                        <div class="stat-title text-xs">Total Cost</div>
                        <div class="stat-value text-xl text-accent">¥${(stats.total_cost || 0).toFixed(4)}</div>
                        <div class="stat-desc text-xs">Estimated</div>
                    </div>
                    <div class="stat py-3">
                        <div class="stat-title text-xs">Total Savings</div>
                        <div class="stat-value text-xl text-success">¥${(stats.total_savings || 0).toFixed(4)}</div>
                        <div class="stat-desc text-xs">vs Platform Price</div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

function attachEvents(container) {
    // Create project
    const createBtn = document.getElementById('create-project-btn');
    const nameInput = document.getElementById('new-project-name');

    createBtn.addEventListener('click', async () => {
        const name = nameInput.value.trim();
        if (!name) return;
        const version = document.getElementById('new-project-version').value;
        try {
            await window.go.main.App.CreateProject({ name, model_version: version });
            nameInput.value = '';
            await renderProjectsPage(container);
        } catch (err) {
            alert('Failed to create project: ' + err);
        }
    });

    nameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') createBtn.click();
    });

    // Delete project buttons
    container.querySelectorAll('[data-delete-project]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = parseInt(btn.dataset.deleteProject);
            if (!confirm('Are you sure?')) return;
            try {
                await window.go.main.App.DeleteProject(id);
                await renderProjectsPage(container);
            } catch (err) {
                alert('Failed to delete project: ' + err);
            }
        });
    });
}

// ============================================================
// Helpers
// ============================================================

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateStr) {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    return d.toLocaleDateString('en', { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}
