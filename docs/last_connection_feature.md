# 自动保存连接配置功能

## 功能概述

pgweb Electron版本现在支持自动保存和恢复最后使用的连接配置，让用户每次打开应用时无需重新输入连接信息。

## 功能特性

### 1. 自动保存
- 当成功建立数据库连接时，系统会自动保存连接配置
- 保存的信息包括：主机地址、端口、用户名、数据库名、SSL模式
- 如果使用SSH隧道，也会保存SSH连接信息
- 如果使用书签连接，会记录书签ID

### 2. 自动加载
- 打开连接窗口时，系统会自动填充上次使用的连接配置
- 如果上次使用的是书签，会自动选择对应的书签
- 如果上次使用的是SSH连接，会自动切换到SSH模式并填充SSH信息

### 3. 数据存储
- 连接配置保存在 `data/bookmarks/last_connection.toml` 文件中
- 开发环境：保存在项目根目录的 `data/bookmarks/` 下
- 生产环境：保存在用户数据目录的 `data/bookmarks/` 下

## 配置文件格式

```toml
host = "localhost"
port = 5432
user = "postgres"
database = "mydb"
ssl_mode = "disable"
last_used = 2024-01-01T12:00:00Z
bookmark_id = ""  # 如果使用书签则填入书签ID

# SSH配置（可选）
[ssh]
host = "ssh.example.com"
port = "22"
user = "sshuser"
```

## 使用方法

1. **正常使用连接**：
   - 输入数据库连接信息并成功连接
   - 系统自动保存连接配置

2. **下次启动应用**：
   - 点击"Connect"按钮打开连接窗口
   - 系统自动填充上次的连接信息
   - 直接点击"Connect"即可连接

3. **使用书签连接**：
   - 选择书签并连接
   - 系统记住使用的书签
   - 下次启动时自动选择该书签

## API接口

### 获取最后连接配置
```
GET /api/last_connection
```

响应：
```json
{
  "last_connection": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "database": "mydb",
    "ssl_mode": "disable",
    "last_used": "2024-01-01T12:00:00Z",
    "bookmark_id": ""
  }
}
```

### 手动保存连接配置
```
POST /api/last_connection
Content-Type: application/json

{
  "host": "localhost",
  "port": 5432,
  "user": "postgres",
  "database": "mydb",
  "ssl_mode": "disable"
}
```

## 注意事项

1. **密码不会保存**：出于安全考虑，密码信息不会被保存
2. **覆盖保存**：每次成功连接都会覆盖之前的配置
3. **权限要求**：需要对数据目录有写入权限
4. **跨平台兼容**：配置文件在Windows、Linux、macOS之间通用

## 故障排除

### 连接配置没有自动填充
1. 检查 `data/bookmarks/` 目录是否存在
2. 检查 `last_connection.toml` 文件是否存在
3. 查看控制台是否有错误信息

### 配置文件损坏
1. 删除 `data/bookmarks/last_connection.toml` 文件
2. 重新建立连接以生成新的配置文件

### 权限问题
1. 确保应用对数据目录有读写权限
2. 在Windows上可能需要以管理员身份运行

## 技术实现

- **后端**：使用Go的TOML库进行配置文件的读写
- **前端**：通过AJAX调用API获取和显示连接信息
- **存储**：使用TOML格式存储，便于人工编辑和调试
- **安全**：密码等敏感信息不存储在配置文件中 