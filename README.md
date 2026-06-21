# MatchLab Web MVP

MatchLab 面向市场的 Web 版 MVP。当前包含 Go API 基础、邮箱注册登录、JWT 鉴权、PostgreSQL 表结构和 Ubuntu 部署配置；前端尚未实现。

## 当前能力

- Gin HTTP 服务，默认监听 `127.0.0.1:8080`
- `GET /api/health` 存活检查
- 邮箱注册、登录和 `GET /api/me`
- bcrypt 密码哈希与 7 天 JWT access token
- godotenv 环境配置
- GORM + PostgreSQL 连接入口
- 数据库不可用时仍可启动并响应 health
- PostgreSQL 基础 schema
- Nginx、systemd 和 Ubuntu 部署脚本

## 目录

```text
matchlab-web/
├─ backend/       Go API
├─ frontend/      下一阶段的 Next.js 前端目录
├─ database/      PostgreSQL schema
├─ deploy/        Nginx、systemd、部署脚本
└─ docs/          MVP、API 和部署文档
```

## 本地运行

要求 Go 1.22+。仅测试 health 时不需要启动 PostgreSQL。

```bash
cd matchlab-web/backend
go mod tidy
go run ./cmd/server
```

服务默认地址为 `http://127.0.0.1:8080`。如果需要自定义端口，可创建 `.env`：

```dotenv
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
GIN_MODE=debug
JWT_SECRET=local-development-secret
```

配置数据库时复制示例并替换密码：

```bash
cp .env.example .env
```

Windows PowerShell：

```powershell
Set-Location .\matchlab-web\backend
Copy-Item .env.example .env
go mod tidy
go run .\cmd\server
```

不要提交 `.env`；它已被 `backend/.gitignore` 忽略。

## 测试 health

curl：

```bash
curl -i http://127.0.0.1:8080/api/health
```

PowerShell：

```powershell
Invoke-RestMethod http://127.0.0.1:8080/api/health
```

预期 HTTP 200：

```json
{
  "ok": true,
  "message": "MatchLab API running"
}
```

自动化验证：

```bash
cd matchlab-web/backend
go test ./...
go vet ./...
go build ./cmd/server
```

## PostgreSQL 配置

服务器已存在数据库 `matchlab` 和用户 `matchlab_user` 时，用 PostgreSQL 管理员执行：

```sql
ALTER USER matchlab_user WITH PASSWORD '替换为高强度密码';
GRANT CONNECT ON DATABASE matchlab TO matchlab_user;
\connect matchlab
GRANT USAGE, CREATE ON SCHEMA public TO matchlab_user;
```

然后以应用用户导入表结构：

```bash
cd matchlab-web
sudo -u postgres psql -d matchlab -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
sudo -u postgres psql -d matchlab -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
PGPASSWORD='替换为高强度密码' psql \
  -h 127.0.0.1 -U matchlab_user -d matchlab \
  -v ON_ERROR_STOP=1 -f database/schema.sql
```

生产环境把相同配置写入 `/opt/matchlab/backend/.env`：

```dotenv
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
GIN_MODE=release
JWT_SECRET=替换为至少32字节的随机值
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=matchlab
DB_USER=matchlab_user
DB_PASSWORD=替换为高强度密码
DB_SSLMODE=disable
```

## Ubuntu 22.04 部署

将整个 `matchlab-web` 上传到服务器后：

```bash
cd matchlab-web
sudo sh deploy/deploy.sh
sudo nano /opt/matchlab/backend/.env
sudo chown matchlab:matchlab /opt/matchlab/backend/.env
sudo chmod 600 /opt/matchlab/backend/.env
sudo systemctl restart matchlab-api
```

验证：

```bash
curl http://127.0.0.1:8080/api/health
curl http://139.224.119.187/api/health
sudo systemctl status matchlab-api --no-pager
sudo journalctl -u matchlab-api -n 100 --no-pager
```

完整步骤见 [docs/DEPLOY.md](docs/DEPLOY.md)。

## JWT 安全要求

未设置 `JWT_SECRET` 时，服务会使用开发默认值并打印警告。该默认值只便于本地运行，生产环境必须替换：

```bash
openssl rand -base64 48
```

将结果写入 `/opt/matchlab/backend/.env` 的 `JWT_SECRET`，不要提交真实 `.env`。完整接口及 curl 示例见 [docs/API_DOC.md](docs/API_DOC.md)。
