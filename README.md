# Seedance Client

基于 [火山引擎 Seedance](https://www.volcengine.com/docs/6791/1347773) API 的视频生成管理客户端。

## 功能特性

- 📁 **项目管理** - 创建和管理多个视频生成项目
- 🎬 **分镜管理** - 为每个项目添加多个分镜，支持首帧/尾帧控制
- 🎨 **拖拽上传** - 支持拖拽图片到上传区域
- 🔄 **实时状态** - 自动轮询视频生成状态
- 🔁 **失败重试** - 生成失败时可一键重试
- ✏️ **编辑分镜** - 支持修改已创建的分镜参数
- 🌐 **多语言** - 支持中英文界面切换

## 技术栈

- **后端**: Go + Gin + GORM
- **前端**: HTML + TailwindCSS + Material Design 3
- **数据库**: SQLite
- **API**: 火山引擎 Ark Runtime SDK

## 快速开始

### 环境要求

- Go 1.21+
- 火山引擎 API Key

### 安装

```bash
# 克隆项目
git clone <your-repo-url>
cd seedance-client

# 安装依赖
go mod download

# 设置环境变量（可选）
export ARK_API_KEY="your-api-key-here"

# 编译运行
go build -o seedance-client .
./seedance-client
```

### 访问

打开浏览器访问 [http://localhost:23313](http://localhost:23313)

## 项目结构

```
seedance-client/
├── main.go              # 入口文件
├── handlers/            # HTTP 处理器
│   ├── project.go       # 项目相关接口
│   └── storyboard.go    # 分镜相关接口
├── models/              # 数据模型
│   ├── models.go        # Project, Storyboard 模型
│   └── db.go            # 数据库初始化
├── services/            # 服务层
│   └── volcengine_service.go  # 火山引擎 API 封装
├── templates/           # HTML 模板
│   ├── header.html      # 公共头部（含导航、样式、i18n）
│   ├── projects.html    # 项目列表页
│   └── storyboard.html  # 分镜详情页
├── static/              # 静态资源
└── uploads/             # 上传的图片文件
```

## API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/` | 项目列表 |
| POST | `/projects` | 创建项目 |
| POST | `/projects/delete/:id` | 删除项目 |
| GET | `/projects/:id` | 项目分镜详情 |
| POST | `/projects/:id/storyboards` | 创建分镜 |
| POST | `/storyboards/delete/:sid` | 删除分镜 |
| POST | `/storyboards/:sid/generate` | 开始生成视频 |
| GET | `/storyboards/:sid/status` | 查询生成状态 |
| POST | `/storyboards/:sid/update` | 更新分镜 |
| POST | `/settings/apikey` | 更新 API Key |

## 支持的模型

- `doubao-seedance-1-5-pro-251215` - Seedance 1.5 Pro（推荐）
- `doubao-seedance-1-0-pro-fast-251015` - Seedance 1.0 Pro Fast

## 视频参数

- **比例**: 16:9, 9:16, adaptive（根据上传的首尾帧自适应）
- **时长**: 5秒, 10秒
- **首帧**: 可选，控制视频起始画面
- **尾帧**: 可选，控制视频结束画面

## 注意事项

1. API Key 需要有火山引擎视觉模型的调用权限
2. 视频生成需要一定时间，请耐心等待
3. 上传的图片会保存在 `uploads/` 目录

## License

MIT License
