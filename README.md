# Seedance Client

基于 [火山引擎 Seedance](https://www.volcengine.com/docs/6791/1347773) API 的视频生成管理桌面客户端。

## 功能特性

- 📁 **项目管理**：创建和管理多个视频项目，项目创建时固定画幅（如 16:9 / 9:16）。
- 🧭 **三面板工作流（达芬奇式）**：
  - **分镜拆解**：支持 Markdown / TXT / CSV / TSV / XLSX 导入，调用 LLM 自动拆解为结构化分镜 JSON。
  - **资产管理**：角色/场景/元素/风格与分镜首尾帧统一管理，支持版本化、Good 标记、上传与 AI 生成（可重试抽卡）。
  - **制作工作台**：三栏制作视图（分镜与 Take 列表、舞台预览、参数面板）+ 时间线可视化 + 导出。
- 🤖 **分镜解析 Provider 可选**：分镜拆解时可按次指定 API 提供商、API Key、Base URL、模型，不影响全局设置。
- 🔢 **Take 版本管理**：每个分镜支持多版本生成，对比并标记 Good Take。
- 🔗 **镜头接力**：支持“接力上一分镜尾帧”以增强连续性。
- 🚀 **在线/离线推理**：支持 `standard`（在线）与 `flex`（离线）模式，离线模式支持超时秒数配置。
- 🎵 **声画同步**：支持生成音效（仅支持音频的模型可用）。
- 📦 **本地缓存**：成功视频与尾帧自动下载缓存，优先使用本地文件预览。
- 🎬 **FCPXML 导出**：导出项目成功片段为 ZIP，包含 `project.fcpxml`（兼容 DaVinci Resolve / FCP 工作流）。
- 🌐 **多语言**：中英文界面切换。
- 🖥️ **桌面应用**：基于 Wails，支持 macOS / Windows / Linux。

## 技术栈

- **后端**：Go + GORM + [Wails v2](https://wails.io/)
- **前端**：Vanilla JS + Vite + TailwindCSS + DaisyUI
- **数据库**：SQLite
- **模型调用**：Volcengine Ark Runtime SDK + OpenAI-compatible HTTP 接口（用于分镜拆解）

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
  - `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- API Key（至少需要火山视觉模型权限）

### 从源码构建

```bash
# 克隆项目
git clone <your-repo-url>
cd seedance-client

# 开发模式（热重载）
wails dev

# 构建桌面应用
wails build
```

首次启动后可在右上角 **Settings** 配置全局 `ARK_API_KEY`。

## 1.x 工作流说明

### 1) 分镜拆解

- 输入来源：Markdown / TXT / CSV / TSV / XLSX（首个工作表）。
- 输出结构字段（可手动修改）：
  - 镜号、景别、运镜、画面内容
  - 人物（`id/name/prompt`）
  - 场景（`id/name/prompt`，不同机位角度可视为不同场景）
  - 特殊元素（`id/name/prompt`）
  - 风格（`id/name/prompt`）
  - 声音设计（可空）
  - 预估时长（5 或 10 秒）
- 编辑操作：保存、拆分、并入下一镜、删除。

### 2) 资产管理

- 资产分类：角色库 / 场景库 / 物品库 / 风格参考 / 分镜首尾帧。
- 每个资产（或首尾帧）支持：
  - 多版本
  - Good 标记（优先级高于“最新版本”）
  - 本地上传
  - AI 生成（单图/多图输入到单图输出）
  - 重试抽卡

### 3) 制作工作台

- 左栏：分镜列表 + Take 切换 + 引用素材缩略图。
- 中栏：主预览、上一镜尾帧/本镜首尾帧、Take 状态与 Good 标记。
- 右栏：参数编辑并保存新 Take（模型、时长、推理模式、音效、接力、离线超时等）。
- 底部：时间线轨道（接力关系显示 `🔗`）+ 常驻导出按钮。

## 分镜拆解 Provider 配置

分镜拆解时支持以下 Provider（按次生效）：

- `ark_default`
  - 使用全局 Settings 中的 API Key。
  - Base URL 与 Key 在拆解面板中只读。
- `ark_custom`
  - 在拆解面板填入独立 Ark API Key / Base URL / 模型。
- `openai_compatible`
  - 在拆解面板填入兼容 OpenAI 的 Base URL / API Key / 模型。

说明：这些配置仅用于“分镜拆解”，不会覆盖全局 Settings。

## 视频生成参数说明

### 推理模式（Service Tier）

- `standard`（在线推理）：低时延、适合实时性场景。
- `flex`（离线推理）：成本更低，适合时延不敏感场景。

实现细节：

- 当 `service_tier = standard` 时，请求中不会提交 `service_tier` 字段。
- 当 `service_tier = flex` 时，会提交：
  - `service_tier = "flex"`
  - `execution_expires_after`（离线任务超时秒数）

### 音效生成（Generate Audio）

- 默认关闭。
- 仅当目标模型支持音频时可开启。

### 画幅与时长

- 画幅在项目创建时固定，后续不允许修改。
- 时长当前为 5 秒或 10 秒。

## 项目结构

```text
seedance-client/
├── main.go                     # Wails 应用入口
├── app.go                      # 核心应用逻辑（项目/分镜/Take/生成/导出）
├── app_v1_workspace.go         # 1.x 三面板工作流接口（分镜拆解/资产管理/工作台）
├── wails.json                  # Wails 配置
├── config/
│   ├── config.go               # 模型配置加载
│   └── models.json             # 模型定义和定价
├── models/
│   ├── models.go               # Project/Storyboard/Take/AssetCatalog/ShotFrameVersion 等
│   └── setup.go                # 数据库初始化与迁移
├── services/
│   ├── volcengine_service.go   # 火山引擎 API 封装
│   ├── download_service.go     # 资源下载缓存
│   └── export_service.go       # ZIP + FCPXML 导出
├── frontend/
│   ├── index.html
│   ├── package.json
│   └── src/
│       ├── main.js             # 应用壳 + 路由
│       ├── projects.js         # 项目页
│       ├── storyboard.js       # 1.x 三面板 UI
│       ├── storyboard_v2.js    # v2.0 占位页
│       ├── i18n.js
│       └── style.css
├── uploads/                    # 用户上传图片
└── downloads/                  # 视频与帧缓存
```

## 支持的模型（示例）

- `doubao-seedance-1-5-pro-251215`：Seedance 1.5 Pro（支持音频）
- `doubao-seedance-1-0-pro-fast-251015`：Seedance 1.0 Pro Fast

以 `config/models.json` 为准。

## 注意事项

1. 请确认 API Key 具备对应模型权限。
2. `flex` 离线推理建议设置合理超时时间，超时任务会自动终止。
3. 缓存目录：上传资源在 `uploads/`，下载资源在 `downloads/`。

## License

MIT License
