import { applyLanguage } from './i18n.js';
import { formatError, reportError, reportErrorOnce } from './errors.js';

const PANELS = {
    breakdown: 'breakdown',
    assets: 'assets',
    workbench: 'workbench'
};

const SCROLL_PRESERVE_SELECTORS = ['.shot-scroll', '.asset-body', '.wb-scroll'];

let state = {
    projectId: null,
    workspace: null,
    activePanel: PANELS.breakdown,
    assetTab: 'character',
    selectedShotId: null,
    selectedTakeByShot: {},
    scrollTopBySelector: {},
    decomposeText: '',
    llmModel: '',
    llmProvider: 'ark_default',
    llmBaseURL: 'https://ark.cn-beijing.volces.com/api/v3',
    llmApiKey: '',
    replaceExisting: true,
};

let rootContainer = null;
let pollTimer = null;
let isLoadingWorkspace = false;
let loadWorkspaceInFlight = null;

async function ensureHasGlobalAPIKey() {
    try {
        const ok = await window.go.main.App.HasAPIKey();
        if (ok) return true;
    } catch (e) {
        // If check fails, fall back to letting the backend return an explicit error.
        return true;
    }

    reportError('未配置 API Key', '[E_APIKEY_MISSING] 未配置 API Key：请点击右上角【设置】填写 API Key 后再重试。');
    const dialog = document.getElementById('settings-dialog');
    if (dialog && typeof dialog.showModal === 'function') {
        dialog.showModal();
        setTimeout(() => document.getElementById('apikey-input')?.focus(), 50);
    }
    return false;
}

export async function renderStoryboardPage(container, projectId) {
    rootContainer = container;
    state = {
        projectId,
        workspace: null,
        activePanel: PANELS.breakdown,
        assetTab: 'character',
        selectedShotId: null,
        selectedTakeByShot: {},
        scrollTopBySelector: {},
        decomposeText: '',
        llmModel: '',
        llmProvider: 'ark_default',
        llmBaseURL: 'https://ark.cn-beijing.volces.com/api/v3',
        llmApiKey: '',
        replaceExisting: true,
    };

    try {
        await loadWorkspace({ preserveSelection: false });
        renderPage();
    } catch (err) {
        container.innerHTML = `<div class="alert alert-error mt-4">加载工作台失败：${formatError(err)}</div>`;
    }
}

async function loadWorkspace({ preserveSelection = true } = {}) {
    if (loadWorkspaceInFlight) {
        return loadWorkspaceInFlight;
    }

    isLoadingWorkspace = true;
    loadWorkspaceInFlight = (async () => {
        const prevShotId = preserveSelection ? state.selectedShotId : null;
        const prevTakeByShot = preserveSelection ? { ...state.selectedTakeByShot } : {};

        const ws = await window.go.main.App.GetV1Workspace(state.projectId);
        state.workspace = ws;
        state.llmModel = state.llmModel || ws.llm_model_default || '';

        const shots = ws.storyboards || [];
        if (shots.length > 0) {
            state.selectedShotId = prevShotId && shots.some(s => s.id === prevShotId) ? prevShotId : shots[0].id;
            state.selectedTakeByShot = {};
            shots.forEach(shot => {
                const prevTake = prevTakeByShot[shot.id];
                const fallback = shot.active_take?.id || shot.takes?.[shot.takes.length - 1]?.id || null;
                const validPrev = prevTake && shot.takes?.some(t => t.id === prevTake) ? prevTake : fallback;
                if (validPrev) {
                    state.selectedTakeByShot[shot.id] = validPrev;
                }
            });
        } else {
            state.selectedShotId = null;
            state.selectedTakeByShot = {};
        }

        syncPolling();
    })();

    try {
        return await loadWorkspaceInFlight;
    } finally {
        isLoadingWorkspace = false;
        loadWorkspaceInFlight = null;
    }
}

function syncPolling() {
    const runningIds = collectRunningTakeIds();
    if (runningIds.length === 0) {
        if (pollTimer) {
            clearInterval(pollTimer);
            pollTimer = null;
        }
        return;
    }

    if (!pollTimer) {
        pollTimer = setInterval(async () => {
            const ids = collectRunningTakeIds();
            if (ids.length === 0) {
                clearInterval(pollTimer);
                pollTimer = null;
                return;
            }

            await Promise.all(ids.map(async (id) => {
                try {
                    await window.go.main.App.GetTakeStatus(id);
                } catch (err) {
                    console.warn('[poll] GetTakeStatus failed', id, err);
                    // Avoid alert spam while polling: show at most once per minute per error type.
                    reportErrorOnce('状态更新失败', err, {
                        ttlMs: 60000,
                    });
                }
            }));

            try {
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                console.warn('[poll] refresh failed', err);
            }
        }, 3500);
    }
}

function collectRunningTakeIds() {
    if (!state.workspace) return [];
    const ids = [];
    (state.workspace.storyboards || []).forEach(shot => {
        (shot.takes || []).forEach(take => {
            const st = (take.status || '').toLowerCase();
            if (st === 'running' || st === 'queued') {
                ids.push(take.id);
            }
        });
    });
    return ids;
}

function renderPage() {
    rememberScrollPositions();

    const ws = state.workspace;
    if (!ws) {
        rootContainer.innerHTML = '<div class="alert alert-error mt-4">加载项目失败</div>';
        return;
    }

    const project = ws.project;
    rootContainer.innerHTML = `
        <div class="cinema-page">
            <div class="cinema-topbar">
                <a href="#/" class="cinema-back">← 返回项目</a>
                <div class="cinema-title-wrap">
                    <h1 class="cinema-title">${escapeHtml(project.name)}</h1>
                    <div class="cinema-subtitle">Project #${project.id} · 比例锁定 ${escapeHtml(project.aspect_ratio || '16:9')}</div>
                </div>
                <button class="cinema-export-btn" data-export-project>导出 FCPXML</button>
            </div>

            <div class="cinema-panel-shell">
                ${renderActivePanel()}
            </div>

            <div class="cinema-bottom-nav">
                ${renderBottomNavButton(PANELS.breakdown, '分镜拆解')}
                ${renderBottomNavButton(PANELS.assets, '资产管理')}
                ${renderBottomNavButton(PANELS.workbench, '制作工作台')}
            </div>
        </div>
    `;

    attachCommonEvents();
    attachPanelEvents();
    applyLanguage();
    restoreScrollPositions();
}

function rememberScrollPositions() {
    if (!rootContainer) return;
    SCROLL_PRESERVE_SELECTORS.forEach(selector => {
        const el = rootContainer.querySelector(selector);
        if (!el) return;
        state.scrollTopBySelector[selector] = el.scrollTop;
    });
}

function restoreScrollPositions() {
    if (!rootContainer) return;
    SCROLL_PRESERVE_SELECTORS.forEach(selector => {
        const savedTop = state.scrollTopBySelector[selector];
        if (typeof savedTop !== 'number') return;
        const el = rootContainer.querySelector(selector);
        if (!el) return;
        el.scrollTop = savedTop;
    });
}

function renderBottomNavButton(key, label) {
    const active = state.activePanel === key ? 'active' : '';
    return `<button class="cinema-nav-btn ${active}" data-switch-panel="${key}">${label}</button>`;
}

function renderActivePanel() {
    switch (state.activePanel) {
        case PANELS.assets:
            return renderAssetsPanel();
        case PANELS.workbench:
            return renderWorkbenchPanel();
        case PANELS.breakdown:
        default:
            return renderBreakdownPanel();
    }
}

function renderBreakdownPanel() {
    const ws = state.workspace;
    const shots = ws.storyboards || [];

    const shotCards = shots.length === 0
        ? '<div class="cinema-empty">当前还没有分镜，请先在左侧导入并拆解，或点击右上角新建空白分镜。</div>'
        : shots.map((shot, index) => renderShotCard(shot, index)).join('');

    return `
        <div class="breakdown-grid">
            <section class="breakdown-import card-cinema">
                <div class="card-head">文本/Excel 导入</div>
                <div class="card-body">
                    <div class="form-row">
                        <label>API 提供商</label>
                        <select id="llm-provider-select">
                            <option value="ark_default" ${state.llmProvider === 'ark_default' ? 'selected' : ''}>全局 Ark（使用设置里的 API Key）</option>
                            <option value="ark_custom" ${state.llmProvider === 'ark_custom' ? 'selected' : ''}>自定义 Ark（独立 Key）</option>
                            <option value="openai_compatible" ${state.llmProvider === 'openai_compatible' ? 'selected' : ''}>OpenAI Compatible（独立 Key）</option>
                        </select>
                    </div>
                    <div class="form-row">
                        <label>LLM 模型</label>
                        <input type="text" id="llm-model-input" value="${escapeHtml(state.llmModel || '')}" placeholder="例如 doubao-seed-1-6-250615">
                    </div>
                    <div class="form-row">
                        <label>Base URL</label>
                        <input type="text" id="llm-base-url-input" value="${escapeHtml(state.llmBaseURL || '')}" placeholder="例如 https://ark.cn-beijing.volces.com/api/v3">
                    </div>
                    <div class="form-row">
                        <label>API Key（仅用于本次分镜拆解）</label>
                        <input type="password" id="llm-api-key-input" value="${escapeHtml(state.llmApiKey || '')}" placeholder="可留空（全局 Ark 模式）">
                    </div>
                    <div class="form-row checkbox-row">
                        <label>
                            <input type="checkbox" id="replace-existing-input" ${state.replaceExisting ? 'checked' : ''}>
                            覆盖当前分镜
                        </label>
                    </div>
                    <div class="form-row">
                        <label>分镜源文档（Markdown / Excel）</label>
                        <textarea id="decompose-source" rows="16" placeholder="粘贴 markdown 文本，或先点击“导入文件”">${escapeHtml(state.decomposeText || '')}</textarea>
                    </div>
                    <div class="btn-row">
                        <button class="btn-cinema secondary" data-load-source-file>导入文件</button>
                        <button class="btn-cinema" data-run-decompose>LLM 拆解为结构化 JSON</button>
                    </div>
                    <div class="hint-text">
                        输出字段固定包含：镜号/景别/运镜/画面内容/人物/场景/元素/风格/声音/时长(5或10秒)
                    </div>
                    <div class="hint-text">
                        你可以为“分镜拆解”单独指定 Provider、API Key 与模型，不影响全局 Settings。
                    </div>
                </div>
            </section>

            <section class="breakdown-shots card-cinema">
                <div class="card-head">
                    分镜编辑（可手动调整/拆分/合并/删除）
                    <button class="mini-btn" data-create-new-shot style="float:right;">+ 新建分镜</button>
                </div>
                <div class="card-body shot-scroll">
                    ${shotCards}
                </div>
            </section>
        </div>
    `;
}

function renderShotCard(shot, index) {
    return `
        <article class="shot-card" data-shot-card="${shot.id}">
            <div class="shot-card-head">
                <div class="shot-index">Shot ${index + 1}</div>
                <div class="shot-actions">
                    <button class="mini-btn" data-save-shot="${shot.id}">保存</button>
                    <button class="mini-btn" data-split-shot="${shot.id}">拆分</button>
                    <button class="mini-btn" data-merge-shot="${shot.id}">并入下一镜</button>
                    <button class="mini-btn danger" data-delete-shot="${shot.id}">删除</button>
                </div>
            </div>

            <div class="shot-grid-4">
                <label>镜号<input data-field="shot_no" value="${escapeHtml(shot.shot_no || '')}"></label>
                <label>景别<input data-field="shot_size" value="${escapeHtml(shot.shot_size || '')}"></label>
                <label>运镜<input data-field="camera_movement" value="${escapeHtml(shot.camera_movement || '')}"></label>
                <label>预估时长
                    <select data-field="estimated_duration">
                        <option value="5" ${Number(shot.estimated_duration) === 5 ? 'selected' : ''}>5 秒</option>
                        <option value="10" ${Number(shot.estimated_duration) === 10 ? 'selected' : ''}>10 秒</option>
                    </select>
                </label>
            </div>

            <label>画面内容<textarea data-field="frame_content" rows="3">${escapeHtml(shot.frame_content || '')}</textarea></label>
            <label>声音设计（可空）<textarea data-field="sound_design" rows="2">${escapeHtml(shot.sound_design || '')}</textarea></label>

            ${renderRefBlock(shot.id, 'characters', '人物', shot.characters || [])}
            ${renderRefBlock(shot.id, 'scenes', '场景', shot.scenes || [])}
            ${renderRefBlock(shot.id, 'elements', '特殊元素', shot.elements || [])}
            ${renderRefBlock(shot.id, 'styles', '风格', shot.styles || [])}
        </article>
    `;
}

function renderRefBlock(shotId, key, label, refs) {
    const rows = (refs.length > 0 ? refs : [{ id: '', name: '', prompt: '' }]).map(ref => `
        <div class="ref-row" data-ref-row>
            <input data-ref-field="id" placeholder="id" value="${escapeHtml(ref.id || '')}">
            <input data-ref-field="name" placeholder="名称" value="${escapeHtml(ref.name || '')}">
            <input data-ref-field="prompt" placeholder="参考图提示词" value="${escapeHtml(ref.prompt || '')}">
            <button class="mini-btn danger" data-remove-ref>×</button>
        </div>
    `).join('');

    return `
        <section class="ref-block" data-ref-block="${key}">
            <div class="ref-head">
                <strong>${label}</strong>
                <button class="mini-btn" data-add-ref="${key}" data-shot-id="${shotId}">+ 新增</button>
            </div>
            <div class="ref-rows">${rows}</div>
        </section>
    `;
}

function renderAssetsPanel() {
    const ws = state.workspace;
    const tabs = [
        { key: 'character', label: '角色库' },
        { key: 'scene', label: '场景库' },
        { key: 'element', label: '物品库' },
        { key: 'style', label: '风格参考' },
        { key: 'frames', label: '分镜首尾帧' },
    ];

    return `
        <div class="assets-shell card-cinema">
            <div class="card-head">资产管理（Good Take 优先 > 最新素材）</div>
            <div class="asset-tabs">
                ${tabs.map(tab => `<button class="asset-tab ${state.assetTab === tab.key ? 'active' : ''}" data-asset-tab="${tab.key}">${tab.label}</button>`).join('')}
            </div>
            <div class="card-body asset-body">
                ${state.assetTab === 'frames' ? renderFrameAssetTab(ws.storyboards || []) : renderCatalogAssetTab(ws.asset_catalogs || [], state.assetTab)}
            </div>
        </div>
    `;
}

function renderCatalogAssetTab(catalogs, tabKey) {
    const rows = catalogs.filter(a => a.asset_type === tabKey);
    if (rows.length === 0) {
        return '<div class="cinema-empty">该资产库还没有内容，先在分镜拆解里创建引用。</div>';
    }

    return rows.map(asset => {
        const activePath = asset.active?.image_path ? `/${asset.active.image_path}` : '';
        const versionList = (asset.versions || []).map(v => {
            const thumb = v.image_path ? `/${v.image_path}` : '';
            return `
                <button class="version-chip ${v.is_good ? 'good' : ''}" data-toggle-asset-good="${v.id}" title="V${v.version_no}">
                    V${v.version_no}${v.is_good ? '★' : ''}
                    ${thumb ? `<img src="${thumb}" alt="v${v.version_no}">` : ''}
                </button>
            `;
        }).join('');

        return `
            <article class="asset-row" data-asset-row="${asset.id}">
                <div class="asset-preview">
                    ${activePath ? `<img src="${activePath}" alt="${escapeHtml(asset.name)}">` : '<div class="asset-placeholder">No Ref</div>'}
                </div>
                <div class="asset-main">
                    <div class="asset-meta">
                        <span class="asset-id">${escapeHtml(asset.asset_code)}</span>
                        <input data-asset-name value="${escapeHtml(asset.name || '')}" placeholder="名称">
                    </div>
                    <textarea data-asset-prompt rows="2" placeholder="参考图提示词">${escapeHtml(asset.prompt || '')}</textarea>
                    <textarea data-asset-input-images rows="2" placeholder="输入图URL（可多行，多图输入单图输出）"></textarea>
                    <div class="asset-version-strip">${versionList || '<span class="hint-text">暂无版本</span>'}</div>
                    <div class="btn-row">
                        <button class="mini-btn" data-save-asset="${asset.id}">保存字段</button>
                        <button class="mini-btn" data-upload-asset="${asset.id}">上传参考图</button>
                        <button class="mini-btn" data-generate-asset="${asset.id}">AI 生成</button>
                        <button class="mini-btn" data-retry-asset="${asset.id}">重试抽卡</button>
                    </div>
                </div>
            </article>
        `;
    }).join('');
}

function renderFrameAssetTab(shots) {
    if (shots.length === 0) {
        return '<div class="cinema-empty">先在分镜拆解中创建分镜。</div>';
    }

    return shots.map(shot => `
        <article class="frame-shot-row" data-frame-shot="${shot.id}">
            <div class="frame-shot-head">
                <strong>${escapeHtml(shot.shot_no || `Shot ${shot.shot_order}`)}</strong>
                <span>${escapeHtml((shot.frame_content || '').slice(0, 80))}</span>
            </div>
            <div class="frame-grid-2">
                ${renderFrameCol(shot, 'start', '首帧')}
                ${renderFrameCol(shot, 'end', '尾帧')}
            </div>
        </article>
    `).join('');
}

function renderFrameCol(shot, frameType, label) {
    const list = frameType === 'start' ? (shot.start_frames || []) : (shot.end_frames || []);
    const active = frameType === 'start' ? shot.active_start_frame : shot.active_end_frame;
    const activeSrc = active?.image_path ? `/${active.image_path}` : '';

    const versions = list.map(v => {
        const src = v.image_path ? `/${v.image_path}` : '';
        return `
            <button class="version-chip ${v.is_good ? 'good' : ''}" data-toggle-frame-good="${v.id}">
                V${v.version_no}${v.is_good ? '★' : ''}
                ${src ? `<img src="${src}" alt="v${v.version_no}">` : ''}
            </button>
        `;
    }).join('');

    return `
        <section class="frame-col" data-frame-col="${frameType}">
            <div class="frame-title">${label}</div>
            <div class="frame-preview">${activeSrc ? `<img src="${activeSrc}" alt="${label}">` : '<div class="asset-placeholder">No Frame</div>'}</div>
            <textarea data-frame-prompt rows="2" placeholder="${label}提示词">${escapeHtml(shot.frame_content || '')}</textarea>
            <textarea data-frame-input-images rows="2" placeholder="输入图URL（可选，多图支持）"></textarea>
            <div class="asset-version-strip">${versions || '<span class="hint-text">暂无版本</span>'}</div>
            <div class="btn-row">
                <button class="mini-btn" data-upload-frame="${shot.id}" data-frame-type="${frameType}">上传</button>
                <button class="mini-btn" data-generate-frame="${shot.id}" data-frame-type="${frameType}">AI 生成</button>
                <button class="mini-btn" data-retry-frame="${shot.id}" data-frame-type="${frameType}">重试抽卡</button>
            </div>
        </section>
    `;
}

function renderWorkbenchPanel() {
    const ws = state.workspace;
    const shots = ws.storyboards || [];
    if (shots.length === 0) {
        return `
            <div class="card-cinema">
                <div class="card-body cinema-empty">
                    暂无分镜，先去"分镜拆解"导入并生成，或
                    <button class="mini-btn" data-create-new-shot style="margin-left:8px;">+ 新建分镜</button>
                </div>
            </div>
        `;
    }

    const selectedShot = getSelectedShot();
    const selectedTake = getSelectedTake(selectedShot);

    return `
        <div class="workbench-shell">
            <aside class="wb-left card-cinema">
                <div class="card-head">1. 资源/列表</div>
                <div class="card-body wb-left-body wb-scroll">
                    ${renderWorkbenchCharacterLibrary(ws)}
                    ${renderWorkbenchStoryboardTextList(shots)}
                </div>
            </aside>

            <section class="wb-center card-cinema">
                <div class="card-head">2. 监视器</div>
                <div class="card-body">
                    ${renderStagePreview(selectedShot, selectedTake)}
                </div>
            </section>

            <aside class="wb-right card-cinema">
                <div class="card-head">3. 参数</div>
                <div class="card-body">
                    ${renderTakeInspector(selectedShot, selectedTake)}
                </div>
            </aside>
        </div>
        ${renderTimeline(shots)}
    `;
}

function renderWorkbenchCharacterLibrary(ws) {
    const catalogs = ws.asset_catalogs || [];
    const chars = catalogs.filter(c => c.asset_type === 'character');
    const items = chars.slice(0, 12).map(c => {
        const p = c.active?.image_path || '';
        const src = p ? `/${String(p).replace(/^\//, '')}` : '';
        return `
            <div class="wb-resource-item" title="${escapeHtml(c.name || c.asset_code || '')}">
                ${src ? `<img src="${src}" alt="${escapeHtml(c.name || c.asset_code || '')}">` : '<div class="wb-resource-ph">角色</div>'}
                <span>${escapeHtml(c.name || c.asset_code || '')}</span>
            </div>
        `;
    }).join('');

    return `
        <section class="wb-section">
            <div class="wb-section-title">角色库</div>
            <div class="wb-resource-list">
                ${items || '<div class="hint-text">暂无角色素材</div>'}
            </div>
        </section>
    `;
}

function renderWorkbenchStoryboardTextList(shots) {
    const rows = (shots || []).map((shot, idx) => renderWorkbenchShotTextRow(shot, idx)).join('');
    return `
        <section class="wb-section">
            <div class="wb-section-title">
                文字分镜表
                <button class="mini-btn" data-create-new-shot style="float:right;font-size:11px;">+ 新建</button>
            </div>
            <div class="wb-story-list">
                ${rows}
            </div>
        </section>
    `;
}

function renderWorkbenchShotTextRow(shot, idx) {
    const selected = shot.id === state.selectedShotId ? 'selected' : '';
    const activeTakeId = state.selectedTakeByShot[shot.id] || shot.active_take?.id;
    const takeTabs = (shot.takes || []).map((take, index) => {
        const st = (take.status || '').toLowerCase();
        const running = st === 'running' || st === 'queued';
        const label = running ? `T${index + 1}…` : `T${index + 1}`;
        return `
            <button class="take-pill ${activeTakeId === take.id ? 'active' : ''}" data-select-take="${take.id}" data-shot-id="${shot.id}">
                ${label}${take.is_good ? '★' : ''}
            </button>
        `;
    }).join('');

    const no = shot.shot_no || `Shot ${idx + 1}`;
    const desc = (shot.frame_content || '').replace(/\s+/g, ' ').slice(0, 42);
    return `
        <div class="wb-story-row ${selected}" data-select-shot="${shot.id}">
            <div class="wb-story-line"><strong>${escapeHtml(no)}</strong><span>${escapeHtml(desc)}</span></div>
            <div class="take-pill-row">${takeTabs || '<span class="hint-text">暂无 Take</span>'}</div>
        </div>
    `;
}

function renderWorkbenchShotItem(shot, idx) {
    const selected = shot.id === state.selectedShotId ? 'selected' : '';
    const activeTakeId = state.selectedTakeByShot[shot.id] || shot.active_take?.id;
    const takeTabs = (shot.takes || []).map((take, index) => `
        <button class="take-pill ${activeTakeId === take.id ? 'active' : ''}" data-select-take="${take.id}" data-shot-id="${shot.id}">
            T${index + 1}${take.is_good ? '★' : ''}
        </button>
    `).join('');

    return `
        <div class="wb-shot-item ${selected}" data-select-shot="${shot.id}">
            <div class="wb-shot-line">
                <strong>${escapeHtml(shot.shot_no || `#${idx + 1}`)}</strong>
                <span>${escapeHtml(shot.shot_size || '')}</span>
            </div>
            <div class="wb-shot-content">${escapeHtml((shot.frame_content || '').slice(0, 60))}</div>
            <div class="take-pill-row">${takeTabs || '<span class="hint-text">暂无 Take</span>'}</div>
            <div class="mini-thumb-row">${renderShotAssetThumbs(shot)}</div>
        </div>
    `;
}

function renderShotAssetThumbs(shot) {
    const refs = [
        ...(shot.characters || []).slice(0, 2).map(r => ({ type: 'character', id: r.id })),
        ...(shot.scenes || []).slice(0, 1).map(r => ({ type: 'scene', id: r.id })),
        ...(shot.elements || []).slice(0, 1).map(r => ({ type: 'element', id: r.id })),
        ...(shot.styles || []).slice(0, 1).map(r => ({ type: 'style', id: r.id })),
    ];

    const thumbs = refs.map(ref => {
        const path = findActiveCatalogImage(ref.type, ref.id);
        if (!path) return '';
        return `<img src="/${path}" alt="${escapeHtml(ref.id)}">`;
    }).filter(Boolean);

    return thumbs.join('') || '<span class="hint-text">未绑定参考图</span>';
}

function renderStagePreview(shot, take) {
    if (!shot || !take) {
        return '<div class="cinema-empty">请选择一个分镜。</div>';
    }

    const prevShot = getPreviousShot(shot.id);
    const prevTail = prevShot?.active_end_frame?.image_path || prevShot?.active_take?.last_frame_path || prevShot?.active_take?.last_frame_url || '';
    const curStart = shot.active_start_frame?.image_path || take.first_frame_path || '';
    const curEnd = shot.active_end_frame?.image_path || take.last_frame_path || take.last_frame_url || '';

    const status = (take.status || '').toLowerCase();
    const monitor = status === 'succeeded'
        ? `<video controls class="stage-video" src="${escapeHtml(take.video_url || '')}"></video>`
        : status === 'running' || status === 'queued'
            ? '<div class="stage-loading">生成中...</div>'
            : '<div class="stage-loading">当前 Take 尚未生成视频</div>';

    return `
        <div class="stage-main">${monitor}</div>
        <div class="stage-compare">
            <div class="compare-col">
                <div class="compare-title">上一镜</div>
                ${renderSmallFrame(prevTail, '上一镜尾帧')}
            </div>
            <div class="compare-col">
                <div class="compare-title">当前镜</div>
                ${renderSmallFrame(curStart || curEnd, '当前镜首帧')}
            </div>
        </div>
        <div class="stage-version-bar">
            <span>Take #${findTakeIndex(shot, take.id)}</span>
            <button class="mini-btn ${take.is_good ? 'good' : ''}" data-toggle-good-take="${take.id}">${take.is_good ? '取消 Good' : '标记 Good Take'}</button>
            <button class="mini-btn" data-generate-take="${take.id}">${status === 'failed' ? '重试生成' : '生成视频'}</button>
        </div>
    `;
}

function renderSmallFrame(path, label) {
    if (!path) {
        return `<div class="small-frame"><div class="asset-placeholder">${label}</div></div>`;
    }
    const src = path.startsWith('http') ? path : `/${path.replace(/^\//, '')}`;
    return `<div class="small-frame"><img src="${src}" alt="${escapeHtml(label)}"><span>${label}</span></div>`;
}

function renderTakeInspector(shot, take) {
    if (!shot || !take) return '<div class="cinema-empty">请选择一个分镜。</div>';

    const modelOptions = (state.workspace.models || []).map(m => `<option value="${m.id}" ${m.id === take.model_id ? 'selected' : ''}>${escapeHtml(m.name)}</option>`).join('');
    const audioSupported = (state.workspace.audio_supported_models || []).includes(take.model_id);
    const isFlex = (take.service_tier || 'standard') === 'flex';
    const expiresAfter = Number(take.expires_after || 0) > 0 ? Number(take.expires_after) : 86400;

    return `
        <div class="inspector" data-inspector-shot="${shot.id}" data-inspector-take="${take.id}">
            <label>视频提示词
                <textarea id="wb-prompt" rows="6">${escapeHtml(take.prompt || '')}</textarea>
            </label>
            <div class="shot-grid-2">
                <label>目标模型
                    <select id="wb-model">${modelOptions}</select>
                </label>
                <label>推理模式
                    <select id="wb-service-tier">
                        <option value="standard" ${!isFlex ? 'selected' : ''}>在线推理 (standard)</option>
                        <option value="flex" ${isFlex ? 'selected' : ''}>离线推理 (flex)</option>
                    </select>
                </label>
            </div>
            <div class="shot-grid-2">
                <label>时长
                    <select id="wb-duration">
                        <option value="5" ${Number(take.duration) === 5 ? 'selected' : ''}>5 秒</option>
                        <option value="10" ${Number(take.duration) === 10 ? 'selected' : ''}>10 秒</option>
                    </select>
                </label>
                <label>离线超时（秒）
                    <input id="wb-execution-timeout" type="number" min="60" step="60" value="${expiresAfter}" ${isFlex ? '' : 'disabled'}>
                </label>
            </div>
            <div class="shot-grid-2">
                <label class="checkbox-inline">
                    <input type="checkbox" id="wb-chain-from-prev" ${take.chain_from_prev ? 'checked' : ''}>
                    接力上一分镜尾帧
                </label>
                <label class="checkbox-inline ${audioSupported ? '' : 'disabled'}">
                    <input type="checkbox" id="wb-generate-audio" ${take.generate_audio ? 'checked' : ''} ${audioSupported ? '' : 'disabled'}>
                    同步音效
                </label>
            </div>

            <div class="frame-quick-view">
                ${renderSmallFrame(shot.active_start_frame?.image_path || '', '资产首帧')}
                ${renderSmallFrame(shot.active_end_frame?.image_path || '', '资产尾帧')}
            </div>

            <div class="offline-note">
                <strong>离线推理说明</strong>
                <span>时延不敏感（小时级）建议使用 <code>flex</code>，成本约为在线的 50%。设置合理超时时间，超时任务会自动终止。</span>
            </div>

            <div class="hint-text">项目比例固定：${escapeHtml(state.workspace.project.aspect_ratio || '16:9')}（创建后不可更改）</div>

            <div class="btn-row">
                <button class="btn-cinema secondary" data-save-new-take="${shot.id}">保存为新 Take</button>
                <button class="btn-cinema" data-generate-take="${take.id}">生成当前 Take</button>
            </div>
        </div>
    `;
}

function renderTimeline(shots) {
    const clips = shots.map((shot, idx) => {
        const take = shot.active_take;
        const duration = Number(take?.duration || shot.estimated_duration || 5);
        const width = Math.max(90, duration * 28);
        const chained = take?.chain_from_prev && idx > 0;

		const st = (take?.status || '').toLowerCase();
		const running = st === 'running' || st === 'queued';
		const failed = st === 'failed';
		const statusText = running ? '正在生成...' : failed ? '生成失败' : '';

		const thumbPath = shot.active_end_frame?.image_path
			|| take?.local_last_frame_path
			|| take?.last_frame_path
			|| take?.last_frame_url
			|| '';
		const thumbSrc = thumbPath
			? (String(thumbPath).startsWith('http') ? thumbPath : `/${String(thumbPath).replace(/^\//, '')}`)
			: '';
        return `
            <div class="timeline-clip" style="width:${width}px" data-select-shot="${shot.id}">
                ${chained ? '<span class="chain-flag">🔗</span>' : ''}
                ${thumbSrc ? `<img class="timeline-thumb" src="${thumbSrc}" alt="thumb">` : '<div class="timeline-thumb placeholder"></div>'}
                <div class="timeline-meta">
                    <strong>${escapeHtml(shot.shot_no || `S${idx + 1}`)}</strong>
                    <span>${duration}s${statusText ? ' · ' + statusText : ''}</span>
                </div>
            </div>
        `;
    }).join('');

    return `
        <div class="timeline-shell card-cinema">
            <div class="card-head">4. 时间线</div>
            <div class="card-body">
                <div class="timeline-track">${clips}</div>
                <div class="timeline-export-wrap">
                    <button class="cinema-export-btn" data-export-project>导出 FCPXML</button>
                </div>
            </div>
        </div>
    `;
}

function attachCommonEvents() {
    rootContainer.querySelectorAll('[data-switch-panel]').forEach(btn => {
        btn.addEventListener('click', () => {
            state.activePanel = btn.dataset.switchPanel;
            renderPage();
        });
    });

    rootContainer.querySelectorAll('[data-export-project]').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                await window.go.main.App.ExportProject(state.projectId);
            } catch (err) {
                if (err) reportError('导出失败', err);
            }
        });
    });
}

function attachPanelEvents() {
    if (state.activePanel === PANELS.breakdown) {
        attachBreakdownEvents();
    } else if (state.activePanel === PANELS.assets) {
        attachAssetEvents();
    } else if (state.activePanel === PANELS.workbench) {
        attachWorkbenchEvents();
    }
}

function attachBreakdownEvents() {
    const providerSelect = document.getElementById('llm-provider-select');
    const baseUrlInput = document.getElementById('llm-base-url-input');
    const apiKeyInput = document.getElementById('llm-api-key-input');

    function refreshProviderFields() {
        const provider = state.llmProvider || 'ark_default';
        if (!baseUrlInput || !apiKeyInput) return;

        if (provider === 'ark_default') {
            baseUrlInput.disabled = true;
            apiKeyInput.disabled = true;
            baseUrlInput.placeholder = '使用内置 Ark 地址';
            apiKeyInput.placeholder = '使用 Settings 中的 API Key';
        } else if (provider === 'ark_custom') {
            baseUrlInput.disabled = false;
            apiKeyInput.disabled = false;
            if (!baseUrlInput.value.trim()) {
                baseUrlInput.value = 'https://ark.cn-beijing.volces.com/api/v3';
                state.llmBaseURL = baseUrlInput.value;
            }
            baseUrlInput.placeholder = '例如 https://ark.cn-beijing.volces.com/api/v3';
            apiKeyInput.placeholder = '输入本次调用使用的 Ark API Key';
        } else {
            baseUrlInput.disabled = false;
            apiKeyInput.disabled = false;
            baseUrlInput.placeholder = '例如 https://api.openai.com/v1';
            apiKeyInput.placeholder = '输入 OpenAI-compatible API Key';
        }
    }

    providerSelect?.addEventListener('change', (e) => {
        state.llmProvider = e.target.value;
        if (state.llmProvider === 'ark_default') {
            state.llmBaseURL = 'https://ark.cn-beijing.volces.com/api/v3';
            state.llmApiKey = '';
            if (baseUrlInput) baseUrlInput.value = state.llmBaseURL;
            if (apiKeyInput) apiKeyInput.value = '';
        } else if (state.llmProvider === 'openai_compatible' && !state.llmBaseURL) {
            state.llmBaseURL = 'https://api.openai.com/v1';
            if (baseUrlInput) baseUrlInput.value = state.llmBaseURL;
        }
        refreshProviderFields();
    });

    document.getElementById('llm-model-input')?.addEventListener('input', (e) => {
        state.llmModel = e.target.value.trim();
    });

    baseUrlInput?.addEventListener('input', (e) => {
        state.llmBaseURL = e.target.value.trim();
    });

    apiKeyInput?.addEventListener('input', (e) => {
        state.llmApiKey = e.target.value;
    });

    document.getElementById('replace-existing-input')?.addEventListener('change', (e) => {
        state.replaceExisting = !!e.target.checked;
    });

    document.getElementById('decompose-source')?.addEventListener('input', (e) => {
        state.decomposeText = e.target.value;
    });

    rootContainer.querySelector('[data-load-source-file]')?.addEventListener('click', async () => {
        try {
            const result = await window.go.main.App.SelectStoryboardSourceFile();
            if (!result || !result.content) return;
            state.decomposeText = result.content;
            renderPage();
        } catch (err) {
            reportError('导入失败', err);
        }
    });

    rootContainer.querySelector('[data-run-decompose]')?.addEventListener('click', async () => {
        const sourceText = (state.decomposeText || '').trim();
        if (!sourceText) {
            alert('请先输入分镜文案或导入文件');
            return;
        }

        if ((state.llmProvider || 'ark_default') === 'ark_default') {
            const ok = await ensureHasGlobalAPIKey();
            if (!ok) return;
        } else if ((state.llmProvider || '') === 'ark_custom' || (state.llmProvider || '') === 'openai_compatible') {
            if (!(state.llmApiKey || '').trim()) {
                alert('请先填写“API Key（仅用于本次分镜拆解）”');
                document.getElementById('llm-api-key-input')?.focus();
                return;
            }
            if ((state.llmProvider || '') === 'openai_compatible' && !(state.llmBaseURL || '').trim()) {
                alert('OpenAI Compatible 模式需要填写 Base URL');
                document.getElementById('llm-base-url-input')?.focus();
                return;
            }
        }

        try {
            await window.go.main.App.DecomposeStoryboardWithLLM({
                project_id: state.projectId,
                source_text: sourceText,
                llm_model_id: state.llmModel || state.workspace.llm_model_default,
                provider: state.llmProvider || 'ark_default',
                api_key: state.llmApiKey || '',
                base_url: state.llmBaseURL || '',
                replace_existing: state.replaceExisting,
            });
            await loadWorkspace({ preserveSelection: false });
            renderPage();
        } catch (err) {
            reportError('拆解失败', err);
        }
    });

    refreshProviderFields();

    rootContainer.querySelectorAll('[data-add-ref]').forEach(btn => {
        btn.addEventListener('click', () => {
            const shotId = Number(btn.dataset.shotId);
            const key = btn.dataset.addRef;
            const card = rootContainer.querySelector(`[data-shot-card="${shotId}"]`);
            const block = card?.querySelector(`[data-ref-block="${key}"] .ref-rows`);
            if (!block) return;
            block.insertAdjacentHTML('beforeend', `
                <div class="ref-row" data-ref-row>
                    <input data-ref-field="id" placeholder="id" value="">
                    <input data-ref-field="name" placeholder="名称" value="">
                    <input data-ref-field="prompt" placeholder="参考图提示词" value="">
                    <button class="mini-btn danger" data-remove-ref>×</button>
                </div>
            `);
            attachBreakdownEvents();
        });
    });

    rootContainer.querySelectorAll('[data-remove-ref]').forEach(btn => {
        btn.addEventListener('click', () => {
            const row = btn.closest('[data-ref-row]');
            if (row) row.remove();
        });
    });

    rootContainer.querySelectorAll('[data-save-shot]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.saveShot);
            const card = rootContainer.querySelector(`[data-shot-card="${shotId}"]`);
            if (!card) return;

            const payload = {
                storyboard_id: shotId,
                shot_no: getFieldValue(card, 'shot_no'),
                shot_size: getFieldValue(card, 'shot_size'),
                camera_movement: getFieldValue(card, 'camera_movement'),
                frame_content: getFieldValue(card, 'frame_content'),
                sound_design: getFieldValue(card, 'sound_design'),
                estimated_duration: Number(getFieldValue(card, 'estimated_duration') || 5),
                duration_fine: 0,
                characters: collectRefs(card, 'characters'),
                scenes: collectRefs(card, 'scenes'),
                elements: collectRefs(card, 'elements'),
                styles: collectRefs(card, 'styles'),
            };

            try {
                await window.go.main.App.UpdateShotMetadata(payload);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('保存失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-delete-shot]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.deleteShot);
            if (!confirm('确认删除这个分镜？')) return;
            try {
                await window.go.main.App.DeleteV1Shot(shotId);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('删除失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-merge-shot]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.mergeShot);
            try {
                await window.go.main.App.MergeShotWithNext(shotId);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('合并失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-split-shot]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.splitShot);
            const second = prompt('请输入拆分后“第二镜”的画面内容');
            if (second === null) return;
            try {
                await window.go.main.App.SplitShot({
                    storyboard_id: shotId,
                    first_content: '',
                    second_content: second,
                });
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('拆分失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-create-new-shot]').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                const newShotId = await window.go.main.App.CreateV1Shot({
                    project_id: state.projectId,
                    after_storyboard_id: 0,
                });
                await loadWorkspace({ preserveSelection: true });
                state.selectedShotId = newShotId;
                renderPage();
            } catch (err) {
                reportError('新建分镜失败', err);
            }
        });
    });
}

function attachAssetEvents() {
    rootContainer.querySelectorAll('[data-asset-tab]').forEach(btn => {
        btn.addEventListener('click', () => {
            state.assetTab = btn.dataset.assetTab;
            renderPage();
        });
    });

    rootContainer.querySelectorAll('[data-save-asset]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = Number(btn.dataset.saveAsset);
            const row = rootContainer.querySelector(`[data-asset-row="${id}"]`);
            if (!row) return;
            try {
                await window.go.main.App.UpdateAssetCatalog({
                    catalog_id: id,
                    name: row.querySelector('[data-asset-name]')?.value || '',
                    prompt: row.querySelector('[data-asset-prompt]')?.value || '',
                });
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('保存资产失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-upload-asset]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = Number(btn.dataset.uploadAsset);
            try {
                await window.go.main.App.UploadAssetImage(id);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('上传失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-generate-asset], [data-retry-asset]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = Number(btn.dataset.generateAsset || btn.dataset.retryAsset);
            const row = rootContainer.querySelector(`[data-asset-row="${id}"]`);
            if (!row) return;
            const prompt = row.querySelector('[data-asset-prompt]')?.value || '';
            const inputImages = parseMultilineList(row.querySelector('[data-asset-input-images]')?.value || '');

            const ok = await ensureHasGlobalAPIKey();
            if (!ok) return;

            try {
                await window.go.main.App.GenerateAssetImage({
                    catalog_id: id,
                    model_id: state.workspace.image_model_default,
                    prompt,
                    input_images: inputImages,
                });
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('生成失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-toggle-asset-good]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = Number(btn.dataset.toggleAssetGood);
            try {
                await window.go.main.App.ToggleAssetVersionGood(id);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('设置 Good 失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-upload-frame]').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                await window.go.main.App.UploadShotFrame({
                    storyboard_id: Number(btn.dataset.uploadFrame),
                    frame_type: btn.dataset.frameType,
                });
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('上传帧失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-generate-frame], [data-retry-frame]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.generateFrame || btn.dataset.retryFrame);
            const frameType = btn.dataset.frameType;
            const shotRow = rootContainer.querySelector(`[data-frame-shot="${shotId}"]`);
            const col = shotRow?.querySelector(`[data-frame-col="${frameType}"]`);
            if (!col) return;
            const prompt = col.querySelector('[data-frame-prompt]')?.value || '';
            const inputImages = parseMultilineList(col.querySelector('[data-frame-input-images]')?.value || '');

            const ok = await ensureHasGlobalAPIKey();
            if (!ok) return;

            try {
                await window.go.main.App.GenerateShotFrame({
                    storyboard_id: shotId,
                    frame_type: frameType,
                    model_id: state.workspace.image_model_default,
                    prompt,
                    input_images: inputImages,
                });
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('生成帧失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-toggle-frame-good]').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                await window.go.main.App.ToggleShotFrameGood(Number(btn.dataset.toggleFrameGood));
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('设置帧 Good 失败', err);
            }
        });
    });
}

function attachWorkbenchEvents() {
    rootContainer.querySelector('#wb-service-tier')?.addEventListener('change', (e) => {
        const timeoutInput = rootContainer.querySelector('#wb-execution-timeout');
        if (!timeoutInput) return;
        const isFlex = e.target.value === 'flex';
        timeoutInput.disabled = !isFlex;
        if (isFlex && (!timeoutInput.value || Number(timeoutInput.value) <= 0)) {
            timeoutInput.value = '86400';
        }
    });

    rootContainer.querySelectorAll('[data-select-shot]').forEach(el => {
        el.addEventListener('click', () => {
            state.selectedShotId = Number(el.dataset.selectShot);
            renderPage();
        });
    });

    rootContainer.querySelectorAll('[data-select-take]').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const shotId = Number(btn.dataset.shotId);
            const takeId = Number(btn.dataset.selectTake);
            state.selectedShotId = shotId;
            state.selectedTakeByShot[shotId] = takeId;
            renderPage();
        });
    });

    rootContainer.querySelectorAll('[data-toggle-good-take]').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                await window.go.main.App.ToggleGoodTake(Number(btn.dataset.toggleGoodTake));
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('标记 Good Take 失败', err);
            }
        });
    });

    rootContainer.querySelectorAll('[data-generate-take]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const takeId = Number(btn.dataset.generateTake);
            const originalText = btn.textContent;
            btn.disabled = true;
            btn.textContent = '提交中...';
            try {
                const ok = await ensureHasGlobalAPIKey();
                if (!ok) return;
                await window.go.main.App.GenerateTakeVideo(takeId);
                await loadWorkspace({ preserveSelection: true });
                renderPage();
            } catch (err) {
                reportError('生成失败', err);
            } finally {
                btn.disabled = false;
                btn.textContent = originalText;
            }
        });
    });

    rootContainer.querySelectorAll('[data-save-new-take]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const shotId = Number(btn.dataset.saveNewTake);
            const shot = (state.workspace.storyboards || []).find(s => s.id === shotId);
            if (!shot) return;

            const prompt = rootContainer.querySelector('#wb-prompt')?.value || '';
            const modelId = rootContainer.querySelector('#wb-model')?.value || '';
            const duration = Number(rootContainer.querySelector('#wb-duration')?.value || 5);
            const serviceTier = rootContainer.querySelector('#wb-service-tier')?.value || 'standard';
            const executionExpiresAfterRaw = Number(rootContainer.querySelector('#wb-execution-timeout')?.value || 0);
            const executionExpiresAfter = serviceTier === 'flex'
                ? Math.max(60, Math.floor(executionExpiresAfterRaw || 86400))
                : 0;
            const chainFromPrev = !!rootContainer.querySelector('#wb-chain-from-prev')?.checked;
            const generateAudio = !!rootContainer.querySelector('#wb-generate-audio')?.checked;

            const firstFrameFromAsset = shot.active_start_frame?.image_path || '';
            const lastFrameFromAsset = shot.active_end_frame?.image_path || '';

            try {
                await window.go.main.App.UpdateStoryboard({
                    storyboard_id: shotId,
                    prompt,
                    model_id: modelId,
                    ratio: state.workspace.project.aspect_ratio || '16:9',
                    duration,
                    generate_audio: generateAudio,
                    service_tier: serviceTier,
                    execution_expires_after: executionExpiresAfter,
                    first_frame_path: chainFromPrev ? '' : firstFrameFromAsset,
                    last_frame_path: lastFrameFromAsset,
                    delete_first_frame: false,
                    delete_last_frame: false,
                    chain_from_prev: chainFromPrev,
                });
                await loadWorkspace({ preserveSelection: true });
                const refreshedShot = (state.workspace.storyboards || []).find(s => s.id === shotId);
                const latestTake = refreshedShot?.takes?.[refreshedShot.takes.length - 1];
                if (latestTake) {
                    state.selectedTakeByShot[shotId] = latestTake.id;
                }
                renderPage();
            } catch (err) {
                reportError('保存新 Take 失败', err);
            }
        });
    });
}

function getSelectedShot() {
    const shots = state.workspace?.storyboards || [];
    if (!shots.length) return null;
    return shots.find(s => s.id === state.selectedShotId) || shots[0];
}

function getSelectedTake(shot) {
    if (!shot) return null;
    const takes = shot.takes || [];
    if (!takes.length) return null;
    const selectedId = state.selectedTakeByShot[shot.id];
    return takes.find(t => t.id === selectedId) || takes.find(t => t.id === shot.active_take?.id) || takes[takes.length - 1];
}

function getPreviousShot(shotId) {
    const shots = state.workspace?.storyboards || [];
    const idx = shots.findIndex(s => s.id === shotId);
    if (idx <= 0) return null;
    return shots[idx - 1];
}

function findTakeIndex(shot, takeId) {
    const idx = (shot.takes || []).findIndex(t => t.id === takeId);
    return idx >= 0 ? idx + 1 : '-';
}

function findActiveCatalogImage(assetType, assetCode) {
    const catalogs = state.workspace?.asset_catalogs || [];
    const catalog = catalogs.find(c => c.asset_type === assetType && c.asset_code === assetCode);
    return catalog?.active?.image_path || '';
}

function getFieldValue(card, field) {
    const el = card.querySelector(`[data-field="${field}"]`);
    return el ? el.value : '';
}

function collectRefs(card, key) {
    const rows = card.querySelectorAll(`[data-ref-block="${key}"] [data-ref-row]`);
    return Array.from(rows).map(row => ({
        id: row.querySelector('[data-ref-field="id"]')?.value?.trim() || '',
        name: row.querySelector('[data-ref-field="name"]')?.value?.trim() || '',
        prompt: row.querySelector('[data-ref-field="prompt"]')?.value?.trim() || '',
    })).filter(r => r.id || r.name || r.prompt);
}

function parseMultilineList(value) {
    return String(value || '')
        .split(/[\n,]/)
        .map(v => v.trim())
        .filter(Boolean);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text == null ? '' : String(text);
    return div.innerHTML;
}
