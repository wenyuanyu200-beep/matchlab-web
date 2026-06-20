# MatchLab API 部署指南

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

部署脚本会构建 Linux amd64 二进制、创建 `matchlab` 系统用户、安装 systemd/Nginx 配置并启动服务：

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

生产建议值：

```dotenv
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
GIN_MODE=release
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=matchlab
DB_USER=matchlab_user
DB_PASSWORD=替换为高强度密码
DB_SSLMODE=disable
```

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
- 服务启动较慢：数据库配置存在但 PostgreSQL不可达；health 不依赖数据库，但启动会先尝试连接。
