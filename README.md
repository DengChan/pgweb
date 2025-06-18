# pgweb - Electron桌面客户端

基于[pgweb](https://github.com/sosedoff/pgweb)的Electron桌面客户端版本，将原有的Web应用重构为跨平台桌面应用。

[![Release](https://img.shields.io/github/release/sosedoff/pgweb.svg?label=Release)](https://github.com/sosedoff/pgweb/releases)
[![Linux Build](https://github.com/sosedoff/pgweb/actions/workflows/checks.yml/badge.svg)](https://github.com/sosedoff/pgweb/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/sosedoff/pgweb)](https://goreportcard.com/report/github.com/sosedoff/pgweb)

## 概述

pgweb是一个用Go语言编写的PostgreSQL数据库管理工具，本项目将其重构为Electron桌面应用，支持Windows、Linux和macOS平台。提供了原生桌面体验，同时保持了原有Web版本的所有功能。

## 功能特性

- **跨平台桌面应用**: Windows/Linux/macOS原生支持
- **内置Go后端**: 无需外部依赖，一键启动
- **PostgreSQL兼容**: 支持PostgreSQL 9.1+
- **SSH隧道支持**: 原生SSH连接支持
- **多数据库会话**: 同时管理多个数据库连接
- **SQL查询执行**: 执行和分析自定义SQL查询
- **数据导出**: 支持CSV/JSON/XML格式导出
- **查询历史**: 保存和管理查询记录
- **连接书签**: 保存常用数据库连接配置
- **现代UI**: 基于Bootstrap的响应式界面

## 项目架构

```
pgweb/
├── main.go                    # Go后端入口
├── pkg/                       # Go后端核心代码
├── static/                    # Web前端资源
├── electron/                  # Electron桌面客户端
│   ├── main.js               # Electron主进程
│   ├── preload.js            # 预加载脚本
│   └── go-backend/           # 编译后的Go后端 (构建时生成)
├── build.js                  # 构建脚本
└── package.json              # 项目配置
```

## 快速开始

### 环境要求

- Node.js 22+ 
- Go 1.20+


### 安装依赖

```bash
# 安装Electron依赖
cd electron
npm install


# 安装项目打包依赖
cd ..
npm install
```

### 开发模式运行

1. **编译Go后端**
   ```bash
   # Windows
   CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o electron/go-backend/pgweb.exe main.go
   
   # Linux
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o electron/go-backend/pgweb main.go
   
   # macos
   CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o electron/go-backend/pgweb main.go
   
   ```

2. **启动开发模式**
   ```bash
   cd electron
   npm start
   ```

### 构建打包

使用自动化构建脚本：

```bash
# 目前仅支持构建Windows版本
npm run build


# 或者使用构建脚本指定平台
# node build.js win        # Windows
# node build.js linux      # Linux
# node build.js all        # 所有平台
```

构建产物：
- `dist/win-unpacked/PgWeb.exe` - Windows可执行文件
- `dist/PgWeb Setup 1.0.0.exe` - Windows安装程序
- `dist/linux-unpacked/` - Linux可执行文件
- `dist/pgweb_1.0.0_amd64.deb` - Linux DEB包

### 手动构建步骤

如果需要手动构建：

1. **编译Go后端**
   ```bash
   go build -ldflags "-s -w" -o dist/pgweb.exe main.go
   ```

2. **打包Electron应用**
   ```bash
   # Windows
   npx electron-builder --win
   
   # Linux
   npx electron-builder --linux deb
   ```

## 开发说明

### 目录结构说明

- **`main.go`**: Go后端入口文件
- **`pkg/`**: Go后端核心代码，包含API、数据库客户端、书签管理等
- **`static/`**: Web前端资源，包含HTML、CSS、JavaScript
- **`electron/`**: Electron桌面客户端代码
- **`build.js`**: 自动化构建脚本，处理Go编译和Electron打包

### 开发流程

1. **后端开发**: 修改`pkg/`下的Go代码
2. **前端开发**: 修改`static/`下的Web资源
3. **桌面客户端**: 修改`electron/`下的Electron代码
4. **测试**: 运行`npm build && npm start`进行开发测试
5. **构建**: 运行`npm install && npm run build`生成发布版本

### 日志和调试

- **开发模式**: 日志输出到控制台，自动打开开发者工具
- **生产模式**: 日志保存到用户数据目录的`logs/`文件夹
- **后端日志**: Go后端的stdout/stderr会显示在Electron控制台
- **全局日志系统**: 
  - 日志文件路径: `data/logs/pgweb.log`
  - 自动日志滚动: 每天自动滚动，保留30天历史
  - 同时输出到文件和控制台
  - JSON格式便于日志分析
  - 单个日志文件最大100MB

## 与原版pgweb的区别

| 特性 | 原版pgweb | Electron版本 |
|------|-----------|-------------|
| 部署方式 | Web服务器 | 桌面应用 |
| 安装方式 | 二进制文件 | 安装程序 |
| 启动方式 | 命令行启动 | 双击启动 |
| 用户体验 | 浏览器访问 | 原生桌面 |
| 自动更新 | 手动更新 | 支持自动更新 |
| 系统集成 | 无 | 任务栏、通知等 |

## 测试

运行Go后端测试：

```bash
make test
```

### 开发规范

- Go代码遵循标准Go格式化规范
- JavaScript代码使用ES6+标准
- 提交信息使用清晰的描述
- 添加必要的测试用例

## 许可证

MIT License - 详见[LICENSE](LICENSE)文件

## 相关链接

- [原版pgweb项目](https://github.com/sosedoff/pgweb)