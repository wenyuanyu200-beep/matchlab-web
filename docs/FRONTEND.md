# MatchLab 前端开发与部署

前端位于 `frontend/`，技术栈为 Next.js 16、TypeScript、React 19 和 Tailwind CSS 4。

## 本地运行

```bash
cd frontend
npm install
cp .env.example .env.local
npm run dev
```

Windows PowerShell：

```powershell
Set-Location frontend
npm install
Copy-Item .env.example .env.local
npm run dev
```

浏览器访问 `http://localhost:3000`。

## API 地址

`.env.example` 默认配置公网后端：

```dotenv
NEXT_PUBLIC_API_BASE_URL=http://139.224.119.187/api
```

当地址为完整 HTTP URL 时，浏览器会请求 Next.js 的同源 `/api-proxy`，再由 Next.js 转发到公网 API，避免本地开发被浏览器 CORS 策略拦截。

代码未配置环境变量时默认使用同源 `/api`，不会绑定固定公网 IP。`.env.example` 的完整 URL 仅用于本地开发连接当前公网后端。

生产环境前后端位于同一域名时建议配置：

```dotenv
NEXT_PUBLIC_API_BASE_URL=/api
```

`NEXT_PUBLIC_` 变量会在 `npm run build` 时写入浏览器包，修改后必须重新构建。

## 测试与构建

```bash
npm test
npm run lint
npm run build
npm run start
```

生产服务默认监听 `0.0.0.0:3000`。动态活动详情 `/activities/[id]` 需要 Node.js 运行 Next.js，不能作为未知 ID 的纯静态文件导出。

## Ubuntu systemd

将项目放到 `/opt/matchlab/frontend`，安装 Node.js 20+，执行：

```bash
cd /opt/matchlab/frontend
cp .env.example .env.production
sed -i 's#http://139.224.119.187/api#/api#' .env.production
npm ci
npm run build
sudo chown -R matchlab:matchlab /opt/matchlab/frontend
```

复制仓库中已审计的 systemd unit：

```bash
sudo cp /opt/matchlab/deploy/matchlab-frontend.service /etc/systemd/system/
sudo mkdir -p /opt/matchlab/frontend/.next/cache
sudo chown -R matchlab:matchlab /opt/matchlab/frontend/.next
```

unit 内容包括：

```ini
[Unit]
Description=MatchLab Next.js frontend
After=network.target

[Service]
Type=simple
User=matchlab
Group=matchlab
WorkingDirectory=/opt/matchlab/frontend
Environment=NODE_ENV=production
Environment=PORT=3000
ExecStart=/usr/bin/npm run start
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/matchlab/frontend/.next/cache

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now matchlab-frontend
sudo systemctl status matchlab-frontend --no-pager
```

## Nginx 反向代理

前端页面交给 Next.js，现有 `/api/` 继续转发到 Go 服务：

```nginx
server {
    listen 80;
    server_name 139.224.119.187;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
sudo nginx -t
sudo systemctl reload nginx
curl -I http://127.0.0.1:3000
curl http://127.0.0.1:8080/api/health
curl http://139.224.119.187/api/health
```

若后续启用域名和 HTTPS，只需更新 Nginx 证书配置；前端保持 `NEXT_PUBLIC_API_BASE_URL=/api`，不会产生混合内容问题。
