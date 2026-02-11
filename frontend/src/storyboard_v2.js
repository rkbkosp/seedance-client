import { applyLanguage, t } from './i18n.js';

// ============================================================
// Seedance 2.0 Storyboard Page (Placeholder)
//
// This module will contain the new workflow and storyboard UI
// for Seedance 2.0 projects. The implementation is pending the
// release of the Volcengine Seedance 2.0 API documentation.
//
// Key differences expected in v2.0:
// - New API endpoints and request format
// - Enhanced storyboard workflow
// - New model capabilities and parameters
// ============================================================

let currentProjectId = null;

/**
 * Render the Seedance 2.0 storyboard page.
 * Currently shows a placeholder UI indicating the feature is under development.
 *
 * @param {HTMLElement} container - The page content container
 * @param {number} projectId - The project ID
 * @param {object} data - Pre-fetched project data from GetProject API
 */
export async function renderStoryboardV2Page(container, projectId, data) {
    currentProjectId = projectId;
    const project = data.project;

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
                    <div class="flex items-center gap-2 mb-1">
                        <h1 class="text-2xl font-bold text-base-content">${escapeHtml(project.name)}</h1>
                        <span class="badge badge-primary badge-sm">Seedance 2.0</span>
                    </div>
                    <p class="text-xs text-base-content/50">Campaign ID: ${project.id}</p>
                </div>
            </div>
        </div>

        <!-- Placeholder Content -->
        <div class="flex flex-col items-center justify-center py-20">
            <div class="text-center max-w-lg">
                <div class="mb-6">
                    <svg xmlns="http://www.w3.org/2000/svg" height="80" viewBox="0 -960 960 960" width="80" fill="currentColor" class="text-primary/30 mx-auto">
                        <path d="M480-80q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm-40-200h80v-240h-80v240Zm40-320q17 0 28.5-11.5T520-720q0-17-11.5-28.5T480-760q-17 0-28.5 11.5T440-720q0 17 11.5 28.5T480-680Z"/>
                    </svg>
                </div>
                <h2 class="text-xl font-bold text-base-content mb-3" data-i18n="v2.coming_soon">
                    Seedance 2.0 - Coming Soon
                </h2>
                <p class="text-base-content/60 mb-6 leading-relaxed" data-i18n="v2.description">
                    The Seedance 2.0 workflow and storyboard interface is under development. 
                    This page will be updated with new capabilities once the Volcengine Seedance 2.0 API is released.
                </p>
                <div class="space-y-3 text-left bg-base-200 rounded-lg p-4">
                    <h3 class="font-semibold text-sm text-base-content/80" data-i18n="v2.expected_features">Expected Features:</h3>
                    <ul class="text-sm text-base-content/60 space-y-1.5 list-disc list-inside">
                        <li data-i18n="v2.feature_1">New generation model with enhanced quality</li>
                        <li data-i18n="v2.feature_2">Updated workflow and storyboard interface</li>
                        <li data-i18n="v2.feature_3">New API parameters and capabilities</li>
                        <li data-i18n="v2.feature_4">Improved video generation options</li>
                    </ul>
                </div>
                <div class="mt-6">
                    <a href="#/" class="btn btn-ghost btn-sm" data-i18n="sb.back">Back to Projects</a>
                </div>
            </div>
        </div>
    `;

    applyLanguage();
}

// ============================================================
// Helpers
// ============================================================

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
}

// ============================================================
// TODO: Implement v2.0 workflow when API documentation is available
//
// Expected structure:
// - renderV2StoryboardList(container, data)
// - renderV2NewStoryboardForm(models)
// - attachV2Events(container, data)
// - v2 specific polling and status management
// - v2 specific video generation API calls
// ============================================================
