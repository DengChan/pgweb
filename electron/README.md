# Pgweb Electron 桌面客户端

## 运行

1. 编译 Go 后端到 go-backend/ 目录
   - Windows: `go build -o electron/go-backend/pgweb.exe main.go`
   - Linux: `go build -o electron/go-backend/pgweb main.go`

2. 安装依赖
   ```bash
   cd electron
   npm install
   ```

3. 启动 Electron 桌面客户端
   ```bash
   npm start
   ```

## 打包

后续可用 electron-builder 打包为 Windows、Linux 安装包。 