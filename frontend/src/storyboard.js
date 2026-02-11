import { applyLanguage, getLanguage, i18nData } from './i18n.js';

// ============================================================
// State
// ============================================================

let activeTasks = new Set();
let isPolling = false;
let pageData = null; // Holds models, audioSupportedModels
let currentProjectId = null;
let newFirstFramePath = '';
let newLastFramePath = '';

// ============================================================
// Storyboard Page
// ============================================================

export async function renderStoryboardPage(container, projectId) {
    currentProjectId = projectId;
    activeTasks = new Set();
    isPolling = false;
    newFirstFramePath = '';
    newLastFramePath = '';

    try {
        const data = await window.go.main.App.GetProject(projectId);
        pageData = data;
        renderHTML(container, data);
        attachEvents(container, data);
        applyLanguage();

        // Start polling for running tasks
        if (activeTasks.size > 0) {
            startAdaptivePolling();
            isPolling = true;
        }
    } catch (err) {
        container.innerHTML = `<div class="alert alert-error mt-4">Failed to load project: ${escapeHtml(String(err))}</div>`;
    }
}

// ============================================================
// HTML Rendering
// ============================================================

function renderHTML(container, data) {
    const project = data.project;
    const models = data.models || [];
    const audioSupportedModels = data.audio_supported_models || [];

    container.innerHTML = `
        <!-- Back Button -->
        <div class="mb-4">
            <a href="#/" class="btn btn-ghost btn-sm gap-1 text-primary">
                <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                    <path d="m313-440 224 224-57 56-320-320 320-320 57 56-224 224h487v80H313Z" />
                </svg>
                <span data-i18n="sb.back">Back to Projects</span>
            </a>
        </div>

        <!-- Project Header -->
        <div class="mb-6 p-4 bg-base-300 rounded-md">
            <div class="flex justify-between items-start">
                <div>
                    <h1 class="text-2xl font-bold text-base-content mb-1">${escapeHtml(project.name)}</h1>
                    <p class="text-xs text-base-content/50">Campaign ID: ${project.id}</p>
                </div>
                <button id="export-project-btn" class="btn btn-primary btn-sm gap-1" title="Export all succeeded videos as ZIP with FCPXML">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z" />
                    </svg>
                    <span data-i18n="sb.export">Export Bundle</span>
                </button>
            </div>
        </div>

        <!-- Storyboard List -->
        <div class="space-y-8" id="storyboard-list">
            ${(project.storyboards || []).map(sb => renderStoryboard(sb)).join('')}
        </div>

        <!-- Add New Storyboard Form -->
        ${renderNewStoryboardForm(models, audioSupportedModels)}

        <!-- Edit Storyboard Modal -->
        ${renderEditDialog(models, audioSupportedModels)}
    `;
}

function renderStoryboard(sb) {
    const takes = sb.takes || [];
    const activeTake = sb.active_take;
    if (!activeTake) return '';

    // Determine active version number
    let activeVersion = takes.length;
    for (let i = 0; i < takes.length; i++) {
        if (takes[i].id === activeTake.id) {
            activeVersion = i + 1;
            break;
        }
    }

    // Track running tasks
    if (activeTake.status === 'Running' || activeTake.status === 'Queued') {
        activeTasks.add(activeTake.id);
    }

    return `
    <div class="card card-compact bg-base-100 border border-base-content/10" id="sb-container-${sb.id}">
        <!-- Takes Tab Bar -->
        <div class="tabs tabs-boxed bg-base-200 p-1 gap-1 overflow-x-auto rounded-t-md">
            <span class="text-xs font-bold text-base-content/60 uppercase tracking-wider mr-2 self-center px-2">Versions:</span>
            ${takes.map((take, idx) => {
                const version = idx + 1;
                const isActive = take.id === activeTake.id;
                return `<button data-switch-take="${take.id}" data-sb-id="${sb.id}" data-version="${version}"
                    class="take-tab-${sb.id} tab tab-sm ${isActive ? 'tab-active' : ''}">
                    v${version}${take.is_good ? '<span class="ml-1" title="Marked as Good Take">🩷</span>' : ''}
                </button>`;
            }).join('')}
        </div>

        <!-- Active Take Content -->
        <div id="take-content-${sb.id}" class="take-content" data-take-id="${activeTake.id}" data-current-version="${activeVersion}">
            ${renderTakeContent(sb.id, activeTake, activeVersion)}
        </div>
    </div>`;
}

function renderTakeContent(sbId, take, version) {
    const statusClass = getStatusBadgeClass(take.status);

    return `
        <!-- Header -->
        <div class="px-4 py-3 border-b border-base-content/10 flex justify-between items-center bg-base-100">
            <div class="flex items-center gap-3">
                <span class="text-sm font-semibold"><span data-i18n="sb.scene">Version</span> ${version}</span>
                <span id="sb-status-${take.id}" class="badge badge-sm ${statusClass}">${take.status}</span>
                <button data-toggle-good="${take.id}" class="btn btn-circle btn-ghost btn-xs ${take.is_good ? 'text-pink-500' : 'text-base-content/30'}" title="Mark as Good Take">
                    <span class="material-symbols-outlined text-sm">favorite</span>
                </button>
            </div>
            <div class="flex items-center gap-1">
                <button data-edit-sb="${sbId}" data-take='${JSON.stringify({
                    prompt: take.prompt, model_id: take.model_id, ratio: take.ratio,
                    duration: take.duration, generate_audio: take.generate_audio,
                    service_tier: take.service_tier, first_frame_path: take.first_frame_path || '',
                    last_frame_path: take.last_frame_path || ''
                }).replace(/'/g, '&#39;')}'
                    class="btn btn-ghost btn-circle btn-xs text-primary" title="Edit & Save as New Take">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M200-200h57l391-391-57-57-391 391v57Zm-80 80v-170l528-527q12-11 26.5-17t30.5-6q16 0 31 6t26 18l55 56q12 11 17.5 26t5.5 30q0 16-5.5 30.5T817-647L290-120H120Zm640-584-56-56 56 56Zm-141 85-28-29 57 57-29-28Z" />
                    </svg>
                </button>
                <button data-delete-take="${take.id}" class="btn btn-ghost btn-circle btn-xs text-error" title="Delete Take">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M280-120q-33 0-56.5-23.5T200-200v-520h-40v-80h200v-40h240v40h200v80h-40v520q0 33-23.5 56.5T680-120H280Zm400-600H280v520h400v-520ZM360-280h80v-360h-80v360Zm160 0h80v-360h-80v360ZM280-720v520-520Z" />
                    </svg>
                </button>
                <button data-delete-storyboard="${sbId}" class="btn btn-ghost btn-circle btn-xs text-base-content/30 hover:text-error" title="Delete Entire Storyboard">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M280-120q-33 0-56.5-23.5T200-200v-520h-40v-80h200v-40h240v40h200v80h-40v520q0 33-23.5 56.5T680-120H280Zm400-600H280v520h400v-520ZM360-280h80v-360h-80v360Zm160 0h80v-360h-80v360ZM280-720v520-520Z" />
                    </svg>
                </button>
            </div>
        </div>

        <div class="flex flex-col md:flex-row">
            <!-- Left: Inputs & Config -->
            <div class="md:w-1/2 p-4 space-y-4">
                <div>
                    <h3 class="text-xs font-bold text-primary mb-1 uppercase tracking-wider" data-i18n="sb.prompt">Prompt</h3>
                    <p class="text-sm text-base-content leading-relaxed take-prompt">${escapeHtml(take.prompt)}</p>
                </div>
                <div class="grid grid-cols-2 gap-3 take-frame-grid">
                    ${take.first_frame_path ? `
                    <div class="relative group">
                        <label class="block text-xs font-bold text-base-content/60 mb-1 uppercase" data-i18n="sb.first_frame">First Frame</label>
                        <div class="rounded-md overflow-hidden border border-base-content/10">
                            <img src="/${take.first_frame_path}" class="w-full h-24 object-cover">
                        </div>
                    </div>` : ''}
                    ${take.last_frame_path ? `
                    <div class="relative group">
                        <label class="block text-xs font-bold text-base-content/60 mb-1 uppercase" data-i18n="sb.last_frame">Last Frame</label>
                        <div class="rounded-md overflow-hidden border border-base-content/10">
                            <img src="/${take.last_frame_path}" class="w-full h-24 object-cover">
                        </div>
                    </div>` : ''}
                </div>
                <div class="flex flex-wrap gap-1 mt-2">
                    <span class="badge badge-secondary badge-sm">${take.model_id}</span>
                    <span class="badge badge-secondary badge-sm">${take.service_tier === 'flex' ? 'Flex' : 'Standard'}</span>
                    <span class="badge badge-secondary badge-sm">${take.ratio}</span>
                    <span class="badge badge-secondary badge-sm">${take.duration}s</span>
                </div>
            </div>

            <!-- Right: Media Output -->
            <div id="sb-right-col-${take.id}" class="md:w-1/2 bg-neutral flex items-center justify-center relative min-h-[280px]">
                ${renderRightColumn(take)}
            </div>
        </div>`;
}

function renderRightColumn(take) {
    if (take.status === 'Succeeded') {
        const videoUrl = take.video_url || '';
        const lastFrameUrl = take.last_frame_url || '';
        return `
            <video controls class="w-full h-full max-h-[400px] object-contain" loop>
                <source src="${videoUrl}" type="video/mp4">
            </video>
            <div class="absolute bottom-3 right-3 flex gap-1">
                ${lastFrameUrl ? `
                <button data-use-first-frame="${lastFrameUrl}" class="btn btn-accent btn-circle btn-sm" title="Use last frame as first frame">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M440-280H280q-83 0-141.5-58.5T80-480q0-83 58.5-141.5T280-680h160v80H280q-50 0-85 35t-35 85q0 50 35 85t85 35h160v80ZM320-440v-80h320v80H320Zm200 160v-80h160q50 0 85-35t35-85q0-50-35-85t-85-35H520v-80h160q83 0 141.5 58.5T880-480q0 83-58.5 141.5T680-280H520Z" />
                    </svg>
                </button>
                <a href="${lastFrameUrl}" download class="btn btn-secondary btn-circle btn-sm" title="Download last frame">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M200-120q-33 0-56.5-23.5T120-200v-560q0-33 23.5-56.5T200-840h560q33 0 56.5 23.5T840-760v560q0 33-23.5 56.5T760-120H200Zm0-80h560v-560H200v560Zm40-80h480L570-480 450-320l-90-120-120 160Zm-40 80v-560 560Z" />
                    </svg>
                </a>` : ''}
                <a href="${videoUrl}" download class="btn btn-primary btn-circle btn-sm" title="Download video">
                    <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                        <path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z" />
                    </svg>
                </a>
            </div>`;
    } else if (take.status === 'Running' || take.status === 'Queued') {
        return `
            <div class="text-base-content flex flex-col items-center">
                <span class="loading loading-spinner loading-md text-primary mb-3"></span>
                <p class="text-sm font-medium" data-i18n="sb.generating">GENERATING...</p>
            </div>`;
    } else if (take.status === 'Draft') {
        return `
            <button data-generate="${take.id}" class="btn btn-primary btn-sm gap-1 generate-btn">
                <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 -960 960 960" width="16" fill="currentColor">
                    <path d="M320-200v-560l440 280-440 280Zm80-280Zm0 134 210-134-210-134v268Z" />
                </svg>
                <span data-i18n="sb.start_gen">Generate Video</span>
            </button>`;
    } else {
        // Failed or other status
        return `
            <div class="text-error flex flex-col items-center">
                <svg xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 -960 960 960" width="32" fill="currentColor">
                    <path d="M480-280q17 0 28.5-11.5T520-320q0-17-11.5-28.5T480-360q-17 0-28.5 11.5T440-320q0 17 11.5 28.5T480-280Zm-40-160h80v-240h-80v240Zm40 360q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z" />
                </svg>
                <p class="mt-1 text-sm font-bold" data-i18n="sb.failed">Failed</p>
                <button data-generate="${take.id}" class="btn btn-error btn-sm gap-1 mt-2 generate-btn">
                    <svg xmlns="http://www.w3.org/2000/svg" height="14" viewBox="0 -960 960 960" width="14" fill="currentColor">
                        <path d="M480-160q-134 0-227-93t-93-227q0-134 93-227t227-93q69 0 132 28.5T720-690v-110h80v280H520v-80h168q-32-56-87.5-88T480-720q-100 0-170 70t-70 170q0 100 70 170t170 70q77 0 139-44t87-116h84q-28 106-114 173t-196 67Z" />
                    </svg>
                    <span data-i18n="sb.retry">Retry</span>
                </button>
            </div>`;
    }
}

function renderNewStoryboardForm(models, audioSupportedModels) {
    const modelOptions = models.map(m =>
        `<option value="${m.id}" ${m.default ? 'selected' : ''}>${escapeHtml(m.name)}</option>`
    ).join('');

    return `
    <div class="mt-10 card card-compact bg-base-100 border border-base-content/10">
        <div class="card-body">
            <div class="flex items-center gap-2 mb-4">
                <div class="w-8 h-8 rounded-md bg-primary flex items-center justify-center text-primary-content">
                    <span class="material-symbols-outlined text-lg">add</span>
                </div>
                <h3 class="card-title text-lg" data-i18n="sb.new">New Storyboard</h3>
            </div>

            <div id="new-sb-form" class="space-y-4">
                <div class="form-control">
                    <label class="label py-1">
                        <span class="label-text text-sm font-medium" data-i18n="sb.prompt">Prompt</span>
                    </label>
                    <textarea id="new-prompt" rows="2" required class="textarea textarea-bordered textarea-sm w-full"
                        data-i18n="sb.desc_placeholder" placeholder="Describe your scene in detail..."></textarea>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-sm" data-i18n="sb.first_frame">First Frame</span></label>
                        <div id="drop-first-frame" class="drag-drop-zone relative border border-dashed border-base-content/20 rounded-md p-4 text-center hover:bg-base-200 transition cursor-pointer" data-pick-image="first">
                            <div class="upload-placeholder text-primary text-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 -960 960 960" width="24" fill="currentColor" class="mx-auto mb-1">
                                    <path d="M440-320v-326L336-542l-56-58 200-200 200 200-56 58-104-104v326h-80ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z" />
                                </svg>
                                <span class="font-medium" data-i18n="sb.upload_img">Select Image</span>
                            </div>
                            <div class="preview-container hidden relative">
                                <img class="preview-img w-full h-24 object-cover rounded-md">
                                <button type="button" class="clear-btn btn btn-error btn-circle btn-xs absolute top-1 right-1 z-10" data-clear-image="first">✕</button>
                            </div>
                        </div>
                    </div>
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-sm" data-i18n="sb.last_frame">Last Frame</span></label>
                        <div id="drop-last-frame" class="drag-drop-zone relative border border-dashed border-base-content/20 rounded-md p-4 text-center hover:bg-base-200 transition cursor-pointer" data-pick-image="last">
                            <div class="upload-placeholder text-primary text-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 -960 960 960" width="24" fill="currentColor" class="mx-auto mb-1">
                                    <path d="M440-320v-326L336-542l-56-58 200-200 200 200-56 58-104-104v326h-80ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z" />
                                </svg>
                                <span class="font-medium" data-i18n="sb.upload_img">Select Image</span>
                            </div>
                            <div class="preview-container hidden relative">
                                <img class="preview-img w-full h-24 object-cover rounded-md">
                                <button type="button" class="clear-btn btn btn-error btn-circle btn-xs absolute top-1 right-1 z-10" data-clear-image="last">✕</button>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-xs" data-i18n="sb.model">Model</span></label>
                        <select id="new-model" class="select select-bordered select-sm w-full">${modelOptions}</select>
                    </div>
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-xs" data-i18n="sb.ratio">Ratio</span></label>
                        <select id="new-ratio" class="select select-bordered select-sm w-full">
                            <option value="adaptive">Adaptive</option>
                            <option value="16:9">16:9</option>
                            <option value="9:16">9:16</option>
                        </select>
                    </div>
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-xs" data-i18n="sb.duration">Duration</span></label>
                        <select id="new-duration" class="select select-bordered select-sm w-full">
                            <option value="5" selected>5s</option>
                            <option value="10">10s</option>
                        </select>
                    </div>
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-xs" data-i18n="sb.mode">Mode</span></label>
                        <select id="new-service-tier" class="select select-bordered select-sm w-full">
                            <option value="standard" selected>Standard</option>
                            <option value="flex">Flex</option>
                        </select>
                    </div>
                    <div class="form-control">
                        <label class="label py-1"><span class="label-text text-xs" data-i18n="sb.audio">Audio</span></label>
                        <label class="label cursor-pointer justify-start gap-2 bg-base-200 rounded-md px-2 h-8">
                            <input type="checkbox" id="new-generate-audio" value="true" class="toggle toggle-primary toggle-sm">
                            <span class="label-text text-xs">Audio</span>
                        </label>
                    </div>
                </div>

                <div class="pt-2 flex justify-end">
                    <button id="add-storyboard-btn" class="btn btn-primary btn-sm" data-i18n="sb.add_btn">Add Storyboard</button>
                </div>
            </div>
        </div>
    </div>`;
}

function renderEditDialog(models, audioSupportedModels) {
    const modelOptions = models.map(m =>
        `<option value="${m.id}">${escapeHtml(m.name)}</option>`
    ).join('');

    return `
    <dialog id="edit-dialog" class="modal">
        <div class="modal-box max-w-md">
            <h3 class="text-lg font-bold mb-3" data-i18n="sb.edit_title">Edit Storyboard</h3>
            <div id="edit-form" class="space-y-3">
                <div class="form-control">
                    <label class="label py-1"><span class="label-text text-sm font-medium" data-i18n="sb.prompt">Prompt</span></label>
                    <textarea id="edit-prompt" rows="2" required class="textarea textarea-bordered textarea-sm w-full"></textarea>
                </div>
                <div class="grid grid-cols-3 gap-2">
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.model">Model</span></label>
                        <select id="edit-model" class="select select-bordered select-sm w-full">${modelOptions}</select>
                    </div>
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.ratio">Ratio</span></label>
                        <select id="edit-ratio" class="select select-bordered select-sm w-full">
                            <option value="adaptive">Adaptive</option>
                            <option value="16:9">16:9</option>
                            <option value="9:16">9:16</option>
                        </select>
                    </div>
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.duration">Duration</span></label>
                        <select id="edit-duration" class="select select-bordered select-sm w-full">
                            <option value="5">5s</option>
                            <option value="10">10s</option>
                        </select>
                    </div>
                </div>
                <div class="grid grid-cols-2 gap-2">
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.mode">Mode</span></label>
                        <select id="edit-service-tier" class="select select-bordered select-sm w-full">
                            <option value="standard">Standard</option>
                            <option value="flex">Flex</option>
                        </select>
                    </div>
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.audio">Audio</span></label>
                        <label class="label cursor-pointer justify-start gap-2 bg-base-200 rounded-md px-2 h-8">
                            <input type="checkbox" id="edit-generate-audio" value="true" class="toggle toggle-primary toggle-sm">
                            <span class="label-text text-xs">Audio</span>
                        </label>
                    </div>
                </div>
                <div class="grid grid-cols-2 gap-2">
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.first_frame">First Frame</span></label>
                        <button type="button" id="edit-first-frame-btn" class="btn btn-outline btn-sm w-full gap-1">
                            <svg xmlns="http://www.w3.org/2000/svg" height="14" viewBox="0 -960 960 960" width="14" fill="currentColor"><path d="M440-320v-326L336-542l-56-58 200-200 200 200-56 58-104-104v326h-80ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
                            <span data-i18n="sb.upload_img">Select Image</span>
                        </button>
                        <span id="edit-first-frame-name" class="text-xs text-success mt-1 truncate hidden"></span>
                        <label class="label cursor-pointer justify-start gap-2 mt-1 hidden" id="first-frame-delete-wrapper">
                            <input type="checkbox" id="edit-delete-first-frame" class="checkbox checkbox-xs checkbox-error">
                            <span class="label-text text-xs text-error">Remove frame</span>
                        </label>
                    </div>
                    <div class="form-control">
                        <label class="label py-0.5"><span class="label-text text-xs" data-i18n="sb.last_frame">Last Frame</span></label>
                        <button type="button" id="edit-last-frame-btn" class="btn btn-outline btn-sm w-full gap-1">
                            <svg xmlns="http://www.w3.org/2000/svg" height="14" viewBox="0 -960 960 960" width="14" fill="currentColor"><path d="M440-320v-326L336-542l-56-58 200-200 200 200-56 58-104-104v326h-80ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
                            <span data-i18n="sb.upload_img">Select Image</span>
                        </button>
                        <span id="edit-last-frame-name" class="text-xs text-success mt-1 truncate hidden"></span>
                        <label class="label cursor-pointer justify-start gap-2 mt-1 hidden" id="last-frame-delete-wrapper">
                            <input type="checkbox" id="edit-delete-last-frame" class="checkbox checkbox-xs checkbox-error">
                            <span class="label-text text-xs text-error">Remove frame</span>
                        </label>
                    </div>
                </div>
                <div class="modal-action">
                    <button id="edit-cancel-btn" class="btn btn-ghost btn-sm" data-i18n="btn.cancel">Cancel</button>
                    <button id="edit-submit-btn" class="btn btn-primary btn-sm" data-i18n="sb.update_btn">Update</button>
                </div>
            </div>
        </div>
        <form method="dialog" class="modal-backdrop"><button>close</button></form>
    </dialog>`;
}

// ============================================================
// Event Handlers
// ============================================================

function attachEvents(container, data) {
    const audioSupportedModels = data.audio_supported_models || [];

    // Export button
    document.getElementById('export-project-btn')?.addEventListener('click', async () => {
        try {
            await window.go.main.App.ExportProject(currentProjectId);
        } catch (err) {
            if (err) alert('Export failed: ' + err);
        }
    });

    // Tab switching
    container.querySelectorAll('[data-switch-take]').forEach(btn => {
        btn.addEventListener('click', () => switchTake(
            parseInt(btn.dataset.switchTake),
            parseInt(btn.dataset.sbId),
            parseInt(btn.dataset.version)
        ));
    });

    // Generate video buttons
    attachGenerateEvents(container);

    // Good take toggle
    container.querySelectorAll('[data-toggle-good]').forEach(btn => {
        btn.addEventListener('click', () => toggleGoodTake(parseInt(btn.dataset.toggleGood), btn));
    });

    // Delete take
    container.querySelectorAll('[data-delete-take]').forEach(btn => {
        btn.addEventListener('click', () => deleteTake(parseInt(btn.dataset.deleteTake)));
    });

    // Delete storyboard
    container.querySelectorAll('[data-delete-storyboard]').forEach(btn => {
        btn.addEventListener('click', () => deleteStoryboard(parseInt(btn.dataset.deleteStoryboard)));
    });

    // Edit storyboard
    container.querySelectorAll('[data-edit-sb]').forEach(btn => {
        btn.addEventListener('click', () => {
            const takeData = JSON.parse(btn.dataset.take);
            openEditDialog(parseInt(btn.dataset.editSb), takeData, audioSupportedModels);
        });
    });

    // Use as first frame
    container.querySelectorAll('[data-use-first-frame]').forEach(btn => {
        btn.addEventListener('click', () => useAsFirstFrame(btn.dataset.useFirstFrame));
    });

    // New storyboard form
    setupNewStoryboardForm(audioSupportedModels);

    // Image picker zones (click to open native file dialog)
    setupImagePickers(container);
}

function attachGenerateEvents(container) {
    container.querySelectorAll('[data-generate]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const takeId = parseInt(btn.dataset.generate);
            const originalContent = btn.innerHTML;
            btn.disabled = true;
            btn.innerHTML = `<span class="loading loading-spinner loading-xs mr-1"></span>Starting...`;

            try {
                await window.go.main.App.GenerateTakeVideo(takeId);
                // Update UI to show running state
                updateStatusBadge(takeId, 'Running');
                const col = document.getElementById(`sb-right-col-${takeId}`);
                if (col) {
                    col.innerHTML = `
                        <div class="text-base-content flex flex-col items-center">
                            <span class="loading loading-spinner loading-md text-primary mb-3"></span>
                            <p class="text-sm font-medium">GENERATING...</p>
                        </div>`;
                }
                activeTasks.add(takeId);
                if (!isPolling) {
                    startAdaptivePolling();
                    isPolling = true;
                }
            } catch (err) {
                alert('Failed to start generation: ' + err);
                btn.disabled = false;
                btn.innerHTML = originalContent;
            }
        });
    });
}

// ============================================================
// Take Operations
// ============================================================

async function switchTake(takeId, sbId, version) {
    // Update tab appearance
    document.querySelectorAll(`.take-tab-${sbId}`).forEach(btn => {
        btn.classList.toggle('tab-active', parseInt(btn.dataset.switchTake) === takeId);
    });

    try {
        const take = await window.go.main.App.GetTake(takeId);
        const contentDiv = document.getElementById(`take-content-${sbId}`);
        if (contentDiv) {
            contentDiv.dataset.takeId = take.id;
            contentDiv.dataset.currentVersion = version;
            contentDiv.innerHTML = renderTakeContent(sbId, take, version);
            applyLanguage();

            // Re-attach events for new content
            reattachTakeEvents(contentDiv);

            if (take.status === 'Running' || take.status === 'Queued') {
                activeTasks.add(take.id);
                if (!isPolling) { startAdaptivePolling(); isPolling = true; }
            }
        }
    } catch (err) {
        console.error('Failed to switch take:', err);
        alert('Failed to switch take');
    }
}

function reattachTakeEvents(contentDiv) {
    const audioSupportedModels = pageData?.audio_supported_models || [];

    contentDiv.querySelectorAll('[data-generate]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const takeId = parseInt(btn.dataset.generate);
            const originalContent = btn.innerHTML;
            btn.disabled = true;
            btn.innerHTML = `<span class="loading loading-spinner loading-xs mr-1"></span>Starting...`;
            try {
                await window.go.main.App.GenerateTakeVideo(takeId);
                updateStatusBadge(takeId, 'Running');
                const col = document.getElementById(`sb-right-col-${takeId}`);
                if (col) col.innerHTML = `<div class="text-base-content flex flex-col items-center"><span class="loading loading-spinner loading-md text-primary mb-3"></span><p class="text-sm font-medium">GENERATING...</p></div>`;
                activeTasks.add(takeId);
                if (!isPolling) { startAdaptivePolling(); isPolling = true; }
            } catch (err) {
                alert('Failed to start generation: ' + err);
                btn.disabled = false;
                btn.innerHTML = originalContent;
            }
        });
    });

    contentDiv.querySelectorAll('[data-toggle-good]').forEach(btn => {
        btn.addEventListener('click', () => toggleGoodTake(parseInt(btn.dataset.toggleGood), btn));
    });

    contentDiv.querySelectorAll('[data-delete-take]').forEach(btn => {
        btn.addEventListener('click', () => deleteTake(parseInt(btn.dataset.deleteTake)));
    });

    contentDiv.querySelectorAll('[data-delete-storyboard]').forEach(btn => {
        btn.addEventListener('click', () => deleteStoryboard(parseInt(btn.dataset.deleteStoryboard)));
    });

    contentDiv.querySelectorAll('[data-edit-sb]').forEach(btn => {
        btn.addEventListener('click', () => {
            const takeData = JSON.parse(btn.dataset.take);
            openEditDialog(parseInt(btn.dataset.editSb), takeData, audioSupportedModels);
        });
    });

    contentDiv.querySelectorAll('[data-use-first-frame]').forEach(btn => {
        btn.addEventListener('click', () => useAsFirstFrame(btn.dataset.useFirstFrame));
    });
}

async function toggleGoodTake(takeId, btn) {
    try {
        const isGood = await window.go.main.App.ToggleGoodTake(takeId);
        if (isGood) {
            btn.classList.add('text-pink-500');
            btn.classList.remove('text-base-content/30');
        } else {
            btn.classList.remove('text-pink-500');
            btn.classList.add('text-base-content/30');
        }

        // Update tab heart icons
        const contentDiv = btn.closest('.take-content');
        if (contentDiv) {
            const sbId = contentDiv.id.replace('take-content-', '');
            document.querySelectorAll(`.take-tab-${sbId}`).forEach(tab => {
                const heart = tab.querySelector('span[title="Marked as Good Take"]');
                if (heart) heart.remove();
            });
            if (isGood) {
                const currentTab = document.querySelector(`.take-tab-${sbId}[data-switch-take="${takeId}"]`);
                if (currentTab) {
                    const span = document.createElement('span');
                    span.className = 'ml-1';
                    span.title = 'Marked as Good Take';
                    span.textContent = '🩷';
                    currentTab.appendChild(span);
                }
            }
        }
    } catch (err) {
        console.error('Toggle good take failed:', err);
    }
}

async function deleteTake(takeId) {
    if (!confirm('Delete this take version?')) return;
    try {
        await window.go.main.App.DeleteTake(takeId);
        // Reload the page
        const container = document.getElementById('page-content');
        await renderStoryboardPage(container, currentProjectId);
    } catch (err) {
        alert('Error deleting take: ' + err);
    }
}

async function deleteStoryboard(sbId) {
    const lang = getLanguage();
    const confirmMsg = (i18nData[lang] && i18nData[lang]['sb.confirm_delete']) || 'Delete entire storyboard container?';
    if (!confirm(confirmMsg)) return;
    try {
        await window.go.main.App.DeleteStoryboard(sbId);
        const container = document.getElementById('page-content');
        await renderStoryboardPage(container, currentProjectId);
    } catch (err) {
        alert('Error deleting storyboard: ' + err);
    }
}

// ============================================================
// Polling
// ============================================================

async function startAdaptivePolling() {
    if (activeTasks.size === 0) {
        isPolling = false;
        return;
    }

    let nextPollDelay = 3000;
    let minReportedInterval = 999999;

    const promises = Array.from(activeTasks).map(async (id) => {
        try {
            const data = await window.go.main.App.GetTakeStatus(id);
            if (data.status) {
                updateStatusBadge(id, data.status);
                updateRightColumn(id, data);
                const s = data.status.toLowerCase();
                if (s === 'succeeded' || s === 'failed') {
                    activeTasks.delete(id);
                }
                if (data.poll_interval) {
                    minReportedInterval = Math.min(minReportedInterval, data.poll_interval);
                }
            }
        } catch (err) {
            console.error('Poll failed for', id, err);
        }
    });

    await Promise.all(promises);

    if (minReportedInterval < 999999) nextPollDelay = minReportedInterval;
    if (nextPollDelay < 1000) nextPollDelay = 1000;

    if (activeTasks.size > 0) {
        setTimeout(startAdaptivePolling, nextPollDelay);
    } else {
        isPolling = false;
    }
}

function updateStatusBadge(id, status) {
    const badge = document.getElementById(`sb-status-${id}`);
    if (!badge) return;
    badge.textContent = status;
    badge.className = 'badge badge-sm ' + getStatusBadgeClass(status);
}

function updateRightColumn(id, data) {
    const col = document.getElementById(`sb-right-col-${id}`);
    if (!col) return;
    const status = data.status.toLowerCase();

    if (status === 'succeeded') {
        col.innerHTML = renderRightColumn({
            status: 'Succeeded',
            video_url: data.video_url,
            last_frame_url: data.last_frame_url,
            id: id
        });
        // Re-attach events
        col.querySelectorAll('[data-use-first-frame]').forEach(btn => {
            btn.addEventListener('click', () => useAsFirstFrame(btn.dataset.useFirstFrame));
        });
    } else if (status === 'failed') {
        col.innerHTML = renderRightColumn({ status: 'Failed', id: id });
        // Re-attach generate button
        col.querySelectorAll('[data-generate]').forEach(btn => {
            btn.addEventListener('click', async () => {
                const takeId = parseInt(btn.dataset.generate);
                const originalContent = btn.innerHTML;
                btn.disabled = true;
                btn.innerHTML = `<span class="loading loading-spinner loading-xs mr-1"></span>Starting...`;
                try {
                    await window.go.main.App.GenerateTakeVideo(takeId);
                    updateStatusBadge(takeId, 'Running');
                    col.innerHTML = `<div class="text-base-content flex flex-col items-center"><span class="loading loading-spinner loading-md text-primary mb-3"></span><p class="text-sm font-medium">GENERATING...</p></div>`;
                    activeTasks.add(takeId);
                    if (!isPolling) { startAdaptivePolling(); isPolling = true; }
                } catch (err) {
                    alert('Failed: ' + err);
                    btn.disabled = false;
                    btn.innerHTML = originalContent;
                }
            });
        });
    } else if (status === 'running' || status === 'queued') {
        if (!col.innerHTML.includes('loading')) {
            col.innerHTML = `<div class="text-base-content flex flex-col items-center"><span class="loading loading-spinner loading-md text-primary mb-3"></span><p class="text-sm font-medium">GENERATING...</p></div>`;
        }
    }
    applyLanguage();
}

// ============================================================
// Edit Dialog
// ============================================================

let editStoryboardId = null;
let editFirstFramePath = '';
let editLastFramePath = '';

function openEditDialog(sbId, takeData, audioSupportedModels) {
    editStoryboardId = sbId;
    editFirstFramePath = '';
    editLastFramePath = '';
    const dialog = document.getElementById('edit-dialog');

    document.getElementById('edit-prompt').value = takeData.prompt || '';
    document.getElementById('edit-model').value = takeData.model_id || '';
    document.getElementById('edit-ratio').value = takeData.ratio || 'adaptive';
    document.getElementById('edit-duration').value = String(takeData.duration || 5);
    document.getElementById('edit-service-tier').value = takeData.service_tier || 'standard';

    const audioCheckbox = document.getElementById('edit-generate-audio');
    audioCheckbox.checked = takeData.generate_audio === true;
    toggleAudioCheckbox(document.getElementById('edit-model'), audioCheckbox, audioSupportedModels);

    // First frame delete option
    const firstDelWrapper = document.getElementById('first-frame-delete-wrapper');
    const firstDelInput = document.getElementById('edit-delete-first-frame');
    if (firstDelInput) firstDelInput.checked = false;
    if (takeData.first_frame_path) {
        firstDelWrapper.classList.remove('hidden');
    } else {
        firstDelWrapper.classList.add('hidden');
    }

    // Last frame delete option
    const lastDelWrapper = document.getElementById('last-frame-delete-wrapper');
    const lastDelInput = document.getElementById('edit-delete-last-frame');
    if (lastDelInput) lastDelInput.checked = false;
    if (takeData.last_frame_path) {
        lastDelWrapper.classList.remove('hidden');
    } else {
        lastDelWrapper.classList.add('hidden');
    }

    // Reset file picker labels
    document.getElementById('edit-first-frame-name')?.classList.add('hidden');
    document.getElementById('edit-last-frame-name')?.classList.add('hidden');

    // Edit model change -> audio toggle
    document.getElementById('edit-model').onchange = () => {
        toggleAudioCheckbox(document.getElementById('edit-model'), audioCheckbox, audioSupportedModels);
    };

    // File picker buttons
    document.getElementById('edit-first-frame-btn').onclick = async () => {
        try {
            const path = await window.go.main.App.SelectImageFile();
            if (path) {
                editFirstFramePath = path;
                const nameEl = document.getElementById('edit-first-frame-name');
                nameEl.textContent = '\u2713 ' + path.split('/').pop();
                nameEl.classList.remove('hidden');
            }
        } catch (err) {
            console.error('File selection failed:', err);
        }
    };

    document.getElementById('edit-last-frame-btn').onclick = async () => {
        try {
            const path = await window.go.main.App.SelectImageFile();
            if (path) {
                editLastFramePath = path;
                const nameEl = document.getElementById('edit-last-frame-name');
                nameEl.textContent = '\u2713 ' + path.split('/').pop();
                nameEl.classList.remove('hidden');
            }
        } catch (err) {
            console.error('File selection failed:', err);
        }
    };

    // Cancel button
    document.getElementById('edit-cancel-btn').onclick = () => dialog.close();

    // Submit button
    document.getElementById('edit-submit-btn').onclick = () => submitEditForm(audioSupportedModels);

    dialog.showModal();
    applyLanguage();
}

async function submitEditForm() {
    const dialog = document.getElementById('edit-dialog');

    const params = {
        storyboard_id: editStoryboardId,
        prompt: document.getElementById('edit-prompt').value,
        model_id: document.getElementById('edit-model').value,
        ratio: document.getElementById('edit-ratio').value,
        duration: parseInt(document.getElementById('edit-duration').value),
        generate_audio: document.getElementById('edit-generate-audio').checked,
        service_tier: document.getElementById('edit-service-tier').value,
        first_frame_path: editFirstFramePath,
        last_frame_path: editLastFramePath,
        delete_first_frame: document.getElementById('edit-delete-first-frame')?.checked || false,
        delete_last_frame: document.getElementById('edit-delete-last-frame')?.checked || false,
    };

    try {
        await window.go.main.App.UpdateStoryboard(params);
        dialog.close();
        const container = document.getElementById('page-content');
        await renderStoryboardPage(container, currentProjectId);
    } catch (err) {
        alert('Failed to update: ' + err);
    }
}

// ============================================================
// New Storyboard Form
// ============================================================

function setupNewStoryboardForm(audioSupportedModels) {
    const modelSelect = document.getElementById('new-model');
    const audioCheckbox = document.getElementById('new-generate-audio');

    if (modelSelect && audioCheckbox) {
        modelSelect.addEventListener('change', () => toggleAudioCheckbox(modelSelect, audioCheckbox, audioSupportedModels));
        toggleAudioCheckbox(modelSelect, audioCheckbox, audioSupportedModels);
    }

    document.getElementById('add-storyboard-btn')?.addEventListener('click', async () => {
        const prompt = document.getElementById('new-prompt').value.trim();
        if (!prompt) {
            alert('Please enter a prompt');
            return;
        }

        const params = {
            project_id: currentProjectId,
            prompt: prompt,
            model_id: document.getElementById('new-model').value,
            ratio: document.getElementById('new-ratio').value,
            duration: parseInt(document.getElementById('new-duration').value),
            generate_audio: document.getElementById('new-generate-audio').checked,
            service_tier: document.getElementById('new-service-tier').value,
            first_frame_path: newFirstFramePath,
            last_frame_path: newLastFramePath,
        };

        try {
            await window.go.main.App.CreateStoryboard(params);
            const container = document.getElementById('page-content');
            await renderStoryboardPage(container, currentProjectId);
        } catch (err) {
            alert('Failed to create storyboard: ' + err);
        }
    });
}

// ============================================================
// Use As First Frame
// ============================================================

async function useAsFirstFrame(imagePath) {
    try {
        // Strip leading slash to get local path for Go
        const localPath = imagePath.startsWith('/') ? imagePath.substring(1) : imagePath;
        const uploadPath = await window.go.main.App.CopyToUploads(localPath);
        if (!uploadPath) return;

        newFirstFramePath = uploadPath;

        const dropZone = document.getElementById('drop-first-frame');
        const placeholder = dropZone.querySelector('.upload-placeholder');
        const previewContainer = dropZone.querySelector('.preview-container');
        const preview = dropZone.querySelector('.preview-img');

        preview.src = '/' + uploadPath;
        previewContainer.classList.remove('hidden');
        placeholder.classList.add('hidden');

        // Scroll to form
        const formSection = document.querySelector('.mt-10');
        if (formSection) {
            formSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
            dropZone.classList.add('ring-2', 'ring-primary');
            setTimeout(() => dropZone.classList.remove('ring-2', 'ring-primary'), 1500);
        }
    } catch (err) {
        console.error('Failed to set first frame:', err);
        alert('Failed to load the image');
    }
}

// ============================================================
// Image Picker (native file dialog)
// ============================================================

function setupImagePickers(container) {
    // Click on picker zones to open native file dialog
    container.querySelectorAll('[data-pick-image]').forEach(zone => {
        zone.addEventListener('click', async (e) => {
            if (e.target.closest('.clear-btn')) return;
            const type = zone.dataset.pickImage;
            try {
                const path = await window.go.main.App.SelectImageFile();
                if (!path) return;

                if (type === 'first') {
                    newFirstFramePath = path;
                } else {
                    newLastFramePath = path;
                }

                const preview = zone.querySelector('.preview-img');
                const placeholder = zone.querySelector('.upload-placeholder');
                const previewContainer = zone.querySelector('.preview-container');
                preview.src = '/' + path;
                previewContainer.classList.remove('hidden');
                placeholder.classList.add('hidden');
            } catch (err) {
                console.error('File selection failed:', err);
            }
        });
    });

    // Clear image buttons
    container.querySelectorAll('[data-clear-image]').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const type = btn.dataset.clearImage;
            if (type === 'first') {
                newFirstFramePath = '';
            } else {
                newLastFramePath = '';
            }
            const zone = btn.closest('.drag-drop-zone');
            zone.querySelector('.preview-container').classList.add('hidden');
            zone.querySelector('.upload-placeholder').classList.remove('hidden');
        });
    });
}

// ============================================================
// Utilities
// ============================================================

function toggleAudioCheckbox(modelSelect, audioCheckbox, audioSupportedModels) {
    const supportsAudio = audioSupportedModels.includes(modelSelect.value);
    audioCheckbox.disabled = !supportsAudio;
    if (!supportsAudio) {
        audioCheckbox.checked = false;
        audioCheckbox.parentElement.classList.add('opacity-50', 'cursor-not-allowed');
    } else {
        audioCheckbox.parentElement.classList.remove('opacity-50', 'cursor-not-allowed');
    }
}

function getStatusBadgeClass(status) {
    switch ((status || '').toLowerCase()) {
        case 'succeeded': return 'badge-success';
        case 'failed': return 'badge-error';
        case 'running': case 'queued': return 'badge-info';
        default: return 'badge-ghost';
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
}
