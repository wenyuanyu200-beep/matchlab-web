# MatchLab API 部署指南

## 圈子与聊天增量发布顺序

圈子与聊天新增表和路由，推荐按以下顺序发布，避免新应用先于表结构上线：

1. 记录发布 commit、当前应用版本并备份 PostgreSQL；验证备份可读取。
2. 用 `ON_ERROR_STOP=1` 执行幂等 migration `database/migrations/20260622_circles_and_chat.sql`，然后再次执行验证幂等性。
3. 检查七张新表与索引存在；migration 不得包含 `DROP` 或 `TRUNCATE`。
4. 构建、测试并发布后端，重启 `matchlab-api`。
5. 先检查 `/api/health`，再用无副作用 smoke 检查公开圈子列表、未认证保护和 admin 保护。
6. 使用 `NEXT_PUBLIC_API_BASE_URL=/api` 构建并发布前端，重启 `matchlab-frontend`。
7. 执行 `docs/qa/circles-and-chat-regression.md` 的人工回归并观察 API、PostgreSQL、Nginx 指标。

示例数据库步骤：

```bash
STAMP=$(date +%Y%m%d-%H%M%S)
sudo -u postgres pg_dump -Fc matchlab > "matchlab-$STAMP.dump"
sudo -u postgres pg_restore --list "matchlab-$STAMP.dump" >/dev/null

sudo -u postgres psql -d matchlab -v ON_ERROR_STOP=1 \
  -f database/migrations/20260622_circles_and_chat.sql
sudo -u postgres psql -d matchlab -v ON_ERROR_STOP=1 \
  -f database/migrations/20260622_circles_and_chat.sql
sudo -u postgres psql -d matchlab -c \
  "SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN ('circles','circle_members','circle_channels','circle_messages','conversations','conversation_members','direct_messages') ORDER BY 1;"
```

后端 smoke 建议只读执行：

```bash
curl -fsS http://127.0.0.1:8080/api/health | jq .
curl -fsS http://127.0.0.1:8080/api/circles | jq .
test "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/api/me/circles)" = 401
test "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/api/admin/circles)" = 401
```

### 回滚

回滚前端与后端应用到上一版本并重启服务；保留七张新表及其中数据，不执行逆向 DROP。旧应用不会访问新表，保留结构可避免聊天与审核数据丢失，并允许之后再次发布。若 migration 后、应用发布前出现故障，直接恢复旧应用即可。只有经过单独审批的数据恢复流程才可从发布前备份整体恢复数据库。

### production E2E 安全说明

`scripts/production_e2e_recommend.sh` 依赖 Bash、`curl`、`jq` 和 `date`。它会注册两个用户、提交问卷、创建活动、报名、审批并生成推荐，属于写数据测试；不要把未知 `BASE_URL` 指向生产后直接运行。当前 Windows 工作区的脚本为 CRLF，`bash -n` 会在 `require_tool() {` 处因 `\r` 失败，必须在受控改动中转换为 LF 并再次完成语法检查后才能执行。本次 QA 不执行任何远程写操作。

目标环境：阿里云 Ubuntu 22.04、Go 1.22.5、Nginx、PostgreSQL，公网 IP `139.224.119.187`。

## 1. 上传项目

将 `matchlab-web` 上传到服务器任意构建目录，例如当前用户的 home：

```bash
cd ~/matchlab-web
go version
nginx -v
psql --version
```

## 2. 配置数据库

为已有用户设置强密码并授予 schema 创建权限：

```bash
sudo -u postgres psql
```

```sql
ALTER USER matchlab_user WITH PASSWORD '替换为高强度密码';
GRANT CONNECT ON DATABASE matchlab TO matchlab_user;
\connect matchlab
GRANT USAGE, CREATE ON SCHEMA public TO matchlab_user;
\quit
```

导入 schema：

```bash
sudo -u postgres psql -d matchlab -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
sudo -u postgres psql -d matchlab -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
PGPASSWORD='替换为高强度密码' psql \
  -h 127.0.0.1 -U matchlab_user -d matchlab \
  -v ON_ERROR_STOP=1 -f database/schema.sql
```

确认八张表：

```bash
PGPASSWORD='替换为高强度密码' psql \
  -h 127.0.0.1 -U matchlab_user -d matchlab -c '\dt'
```

## 3. 自动安装服务

部署脚本会先以 `ON_ERROR_STOP=1` 重复安全地应用 `database/schema.sql`，再构建 Linux amd64 二进制、创建 `matchlab` 系统用户并安装 systemd/Nginx 配置。这样可避免新二进制连接旧 matches 表导致推荐接口 500：

```bash
sudo sh deploy/deploy.sh
```

首次运行会从 `.env.example` 创建 `/opt/matchlab/backend/.env`。立即编辑：

```bash
sudo nano /opt/matchlab/backend/.env
sudo chown matchlab:matchlab /opt/matchlab/backend/.env
sudo chmod 600 /opt/matchlab/backend/.env
sudo systemctl restart matchlab-api
```

脚本会安装前端 systemd unit；完成 `docs/FRONTEND.md` 的前端构建和文件安装后启用：

```bash
sudo systemctl enable --now matchlab-frontend
```

生产建议值：

```dotenv
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
GIN_MODE=release
JWT_SECRET=替换为至少32字节的随机值
CORS_ALLOWED_ORIGINS=http://139.224.119.187
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=matchlab
DB_USER=matchlab_user
DB_PASSWORD=替换为高强度密码
DB_SSLMODE=disable
```

生成生产 JWT 密钥：

```bash
openssl rand -base64 48
```

`GIN_MODE=release` 时，服务会拒绝开发默认值、示例占位值、短于 32 字符的 `JWT_SECRET`，也会拒绝不完整的数据库配置；PostgreSQL 连接失败时进程直接退出，避免出现 health 正常但业务全部 503 的伪健康部署。`CORS_ALLOWED_ORIGINS` 使用逗号分隔的完整 Origin，不要填写路径。

## 4. 验证服务

先绕过 Nginx 验证 Go 服务：

```bash
curl -i http://127.0.0.1:8080/api/health
sudo systemctl status matchlab-api --no-pager
sudo journalctl -u matchlab-api -n 100 --no-pager
```

再验证 Nginx：

```bash
sudo nginx -t
curl -i http://139.224.119.187/api/health
```

如果公网访问失败，确认阿里云安全组允许 TCP 80；无需向公网开放 8080 或 5432。

## 5. 更新发布

上传新代码后重新运行脚本：

```bash
cd ~/matchlab-web
sudo sh deploy/deploy.sh
sudo systemctl restart matchlab-api
curl -f http://127.0.0.1:8080/api/health
```

脚本不会覆盖已有 `/opt/matchlab/backend/.env`。

## 6. 手动回滚

部署前保留旧二进制：

```bash
sudo cp /opt/matchlab/backend/bin/matchlab-api \
  /opt/matchlab/backend/bin/matchlab-api.previous
```

需要回滚时：

```bash
sudo cp /opt/matchlab/backend/bin/matchlab-api.previous \
  /opt/matchlab/backend/bin/matchlab-api
sudo systemctl restart matchlab-api
```

## 7. 常见排查

- `status=203/EXEC`：检查二进制路径和执行权限。
- 环境文件读取失败：检查 `.env` 所有者为 `matchlab` 且权限为 `600`。
- 数据库认证失败：核对密码、`pg_hba.conf` 及 `DB_HOST=127.0.0.1`。
- Nginx 返回 502：先用本机 curl 检查 8080，再查看 systemd 日志。
- release 模式启动失败：查看日志并检查 JWT、数据库六项配置和 PostgreSQL 连通性。
- 浏览器出现 CORS 错误：确认请求的协议、域名和端口与 `CORS_ALLOWED_ORIGINS` 完全一致。
