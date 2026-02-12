export function formatError(err) {
    if (err == null) return '未知错误';

    if (typeof err === 'string') return err;

    // Wails sometimes returns Error-like objects
    if (typeof err === 'object') {
        if (err instanceof Error) {
            return stripErrorCode(err.message || String(err));
        }
        if (typeof err.message === 'string' && err.message.trim()) {
            return stripErrorCode(err.message);
        }
        if (err.error != null) {
            return formatError(err.error);
        }
        try {
            return stripErrorCode(JSON.stringify(err));
        } catch {
            return stripErrorCode(String(err));
        }
    }

    return stripErrorCode(String(err));
}

export function withPrefix(prefix, err) {
    const p = (prefix || '').trim();
    const msg = formatError(err);
    return p ? `${p}: ${msg}` : msg;
}

function stripErrorCode(message) {
    const s = String(message || '');
    // Internal convention: "[E_SOMETHING] actual message"
    return s.replace(/^\s*\[[A-Z0-9_]+\]\s*/i, '').trim();
}

function extractErrorCode(message) {
    const m = String(message || '').match(/^\s*\[([A-Z0-9_]+)\]/i);
    return m ? m[1].toUpperCase() : '';
}

function buildHints(rawMessage) {
    const code = extractErrorCode(rawMessage);
    const msg = String(rawMessage || '');
    const lower = msg.toLowerCase();

    if (code === 'E_APIKEY_MISSING' || msg.includes('未配置 API Key') || msg.includes('API Key')) {
        return '建议：打开右上角【设置】填写 API Key；确认 Key 有效且未过期。';
    }

    if (code === 'E_HTTP_401' || lower.includes('401') || lower.includes('unauthorized') || msg.includes('鉴权') || msg.includes('无权限')) {
        return '建议：API Key 可能无效/无权限；请在【设置】更新正确的 Key。';
    }

    if (code === 'E_HTTP_429' || lower.includes('429') || lower.includes('rate limit') || msg.includes('限流')) {
        return '建议：请求过于频繁/被限流；稍等 30-60 秒后重试，或降低并发操作。';
    }

    if (lower.includes('timeout') || lower.includes('deadline exceeded') || msg.includes('超时')) {
        return '建议：网络可能不稳定；请检查代理/网络后重试（离线 flex 任务可适当提高超时）。';
    }

    if (lower.includes('connection refused') || lower.includes('network') || msg.includes('网络')) {
        return '建议：检查网络/代理配置，确认可以访问服务端地址。';
    }

    return '';
}

export function reportError(prefix, err, options = {}) {
    const raw = typeof err === 'string' ? err : (err && err.message ? err.message : String(err));
    const msg = withPrefix(prefix, err);
    const hint = buildHints(raw);
    const finalMsg = hint ? `${msg}\n\n${hint}` : msg;
    alert(finalMsg);

    const code = extractErrorCode(raw);
    const shouldOpenSettings = options.openSettingsOnAPIKey !== false && (code === 'E_APIKEY_MISSING' || raw.includes('未配置 API Key'));
    if (shouldOpenSettings) {
        const dialog = document.getElementById('settings-dialog');
        if (dialog && typeof dialog.showModal === 'function') {
            dialog.showModal();
            setTimeout(() => document.getElementById('apikey-input')?.focus(), 50);
        }
    }
}

const dedupeShownAt = new Map();

function defaultDedupeKey(prefix, err) {
    const raw = typeof err === 'string' ? err : (err && err.message ? err.message : String(err));
    const code = extractErrorCode(raw);
    if (code) return `${prefix}::${code}`;
    const normalized = stripErrorCode(raw).slice(0, 120);
    return `${prefix}::${normalized}`;
}

export function reportErrorOnce(prefix, err, options = {}) {
    const ttlMs = Number(options.ttlMs || 60000);
    const key = options.dedupeKey || defaultDedupeKey(prefix, err);

    const now = Date.now();
    const last = dedupeShownAt.get(key) || 0;
    if (ttlMs > 0 && now-last < ttlMs) {
        if (options.logSuppressed) {
            console.warn('[error] suppressed', key, err);
        }
        return;
    }
    dedupeShownAt.set(key, now);
    reportError(prefix, err, options);
}
