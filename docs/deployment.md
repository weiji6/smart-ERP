# SmartERP 部署说明

## 1. 部署文件

部署时至少需要以下内容放在同一个目录下：

- `smarterp` 或 `smarterp.exe`：编译后的后端程序。
- `web/`：前端页面和静态资源。
- `config/`：AI 和数据库配置文件。

程序会按相对路径读取 `web/index.html`、`web/assets/`、`config/ai.local.json`、`config/db.local.json`，所以不要只复制单个二进制文件。

## 2. AI 配置

复制示例文件：

```bash
cp config/ai.example.json config/ai.local.json
```

需要修改：

- `apiKey`：你的真实 API Key。
- `baseURL`：模型服务地址，例如 OpenRouter 使用 `https://openrouter.ai/api/v1`，OpenAI 官方使用 `https://api.openai.com/v1`。
- `model`：实际模型名，例如 `openai/gpt-4o-mini` 或 `gpt-4o-mini`。

检查运行中服务是否读取成功：

```bash
curl http://服务器IP:端口/api/v1/ai/status
```

## 3. MySQL 配置

复制示例文件：

```bash
cp config/db.example.json config/db.local.json
```

需要修改 `dsn`：

```json
{
  "driver": "mysql",
  "dsn": "用户名:密码@tcp(数据库地址:3306)/数据库名?charset=utf8mb4&parseTime=true&loc=Local",
  "autoMigrate": true,
  "maxOpenConns": 10,
  "maxIdleConns": 5
}
```

建库示例：

```sql
CREATE DATABASE smarterp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'smarterp'@'%' IDENTIFIED BY 'smarterp_password';
GRANT ALL PRIVILEGES ON smarterp.* TO 'smarterp'@'%';
FLUSH PRIVILEGES;
```

检查运行中服务是否启用 MySQL：

```bash
curl http://服务器IP:端口/api/v1/db/status
```

返回 `storageMode=mysql` 表示已落库；返回 `storageMode=memory` 表示还在使用内存模式。

## 4. 端口配置

服务监听地址通过环境变量 `ADDR` 配置。

本机监听：

```bash
ADDR=:8081 ./smarterp
```

服务器对外监听：

```bash
ADDR=0.0.0.0:8081 ./smarterp
```

如果前面有 Nginx，建议让 Nginx 反代到 `127.0.0.1:8081`。

## 5. Linux systemd 示例

假设部署目录为 `/opt/smarterp`：

```ini
[Unit]
Description=SmartERP
After=network.target

[Service]
WorkingDirectory=/opt/smarterp
ExecStart=/opt/smarterp/smarterp
Environment=ADDR=0.0.0.0:8081
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

保存为 `/etc/systemd/system/smarterp.service` 后执行：

```bash
systemctl daemon-reload
systemctl enable smarterp
systemctl start smarterp
systemctl status smarterp
```

## 6. 常见部署检查

```bash
curl http://服务器IP:8081/health
curl http://服务器IP:8081/api/v1/ai/status
curl http://服务器IP:8081/api/v1/db/status
```

浏览器打开：

```text
http://服务器IP:8081/
```
