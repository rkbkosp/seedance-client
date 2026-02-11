export const i18nData = {
    "en": {
        "app.title": "Spark (火种)",
        "nav.settings": "Settings",
        "settings.title": "Settings",
        "settings.language": "Language",
        "settings.apikey": "API Key",
        "settings.apikey.placeholder": "Enter new API Key",
        "settings.apikey.hint": "Updating this will restart the client connection.",
        "btn.cancel": "Cancel",
        "btn.save": "Save",
        "btn.delete": "Delete",
        "btn.download": "Download",
        "footer.text": "&copy; 2026 Spark",

        "projects.title": "Projects",
        "projects.subtitle": "Manage your video generation campaigns",
        "projects.create.placeholder": "New Project Name",
        "projects.create.btn": "Create",
        "projects.open": "Open",

        "sb.back": "Back to Projects",
        "sb.scene": "Storyboard",
        "sb.prompt": "Prompt",
        "sb.first_frame": "First Frame",
        "sb.last_frame": "Last Frame",
        "sb.model": "Model",
        "sb.ratio": "Ratio",
        "sb.duration": "Duration",
        "sb.generating": "Generating...",
        "sb.failed": "Generation Failed",
        "sb.start_gen": "Generate Video",
        "sb.new": "Add New Storyboard",
        "sb.desc_placeholder": "Describe your scene in detail...",
        "sb.desc_hint": "Detailed descriptions yield better results.",
        "sb.upload_img": "Select Image",
        "sb.add_btn": "Add Storyboard",
        "sb.confirm_delete": "Delete this storyboard?",
        "sb.retry": "Retry",
        "sb.edit": "Edit",
        "sb.edit_title": "Edit Storyboard",
        "sb.update_btn": "Update",
        "sb.download_video": "Download Video",
        "sb.download_lastframe": "Download Last Frame",
        "sb.use_as_firstframe": "Use as First Frame",
        "sb.export": "Export Bundle",
        "sb.mode": "Mode",
        "sb.audio": "Audio"
    },
    "zh": {
        "app.title": "Spark (火种)",
        "nav.settings": "设置",
        "settings.title": "设置",
        "settings.language": "界面语言",
        "settings.apikey": "API Key",
        "settings.apikey.placeholder": "输入新的 API Key",
        "settings.apikey.hint": "更新此项将重置客户端连接。",
        "btn.cancel": "取消",
        "btn.save": "保存",
        "btn.delete": "删除",
        "btn.download": "下载视频",
        "footer.text": "&copy; 2026 Spark",

        "projects.title": "项目列表",
        "projects.subtitle": "管理您的视频生成任务",
        "projects.create.placeholder": "新项目名称",
        "projects.create.btn": "创建",
        "projects.open": "打开",

        "sb.back": "返回项目列表",
        "sb.scene": "分镜",
        "sb.prompt": "提示词",
        "sb.first_frame": "首帧 (起始)",
        "sb.last_frame": "尾帧 (结束)",
        "sb.model": "模型",
        "sb.ratio": "比例",
        "sb.duration": "时长",
        "sb.generating": "生成中...",
        "sb.failed": "生成失败",
        "sb.start_gen": "开始生成",
        "sb.new": "添加新分镜",
        "sb.desc_placeholder": "详细描述您的分镜内容...",
        "sb.desc_hint": "描述得越详细，效果越好。",
        "sb.upload_img": "选择图片",
        "sb.add_btn": "添加分镜",
        "sb.confirm_delete": "确定删除该分镜吗？",
        "sb.retry": "重试",
        "sb.edit": "编辑",
        "sb.edit_title": "编辑分镜",
        "sb.update_btn": "更新",
        "sb.download_video": "下载视频",
        "sb.download_lastframe": "下载尾帧",
        "sb.use_as_firstframe": "用作首帧",
        "sb.export": "导出素材包",
        "sb.mode": "模式",
        "sb.audio": "音频"
    }
};

export function getLanguage() {
    let lang = localStorage.getItem("seedance_lang");
    if (!lang) {
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang.toLowerCase().startsWith("zh")) {
            lang = "zh";
        } else {
            lang = "en";
        }
    }
    return lang;
}

export function setLanguage(lang) {
    localStorage.setItem("seedance_lang", lang);
}

export function t(key) {
    const lang = getLanguage();
    const data = i18nData[lang] || i18nData["en"];
    return data[key] || key;
}

export function applyLanguage() {
    const lang = getLanguage();
    const data = i18nData[lang] || i18nData["en"];

    document.querySelectorAll("[data-i18n]").forEach(el => {
        const key = el.getAttribute("data-i18n");
        if (data[key]) {
            if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
                el.placeholder = data[key];
            } else {
                el.innerHTML = data[key];
            }
        }
    });

    // Update language selector if present
    const sel = document.getElementById("lang-selector");
    if (sel) sel.value = lang;
}
