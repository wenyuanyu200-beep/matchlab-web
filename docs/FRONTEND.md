# MatchLab 前端开发与部署

## 圈子与聊天页面

- `/circles`：approved 圈子广场，支持关键词与分类筛选。
- `/circles/create`：登录用户创建 pending 圈子申请。
- `/circles/[id]`：圈子资料、成员与 `general` 频道工作区；未加入用户看到加入引导，pending/rejected 只对 creator/admin 显示审核状态。
- `/messages`：当前用户 direct conversations 与未读数。
- `/messages/[id]`：两人私聊详情，进入页面和收到新消息后同步已读状态。
- `/admin`：admin 的 pending 圈子列表与通过/拒绝操作。

桌面端圈子详情使用资料侧栏和频道聊天双栏布局；移动端改为“圈子信息 / 频道聊天 / 成员”三个 Tab，默认频道聊天。每个 Tab 独立滚动，页面不得产生横向溢出。消息按本人靠右、他人靠左显示；成员摘要不显示 email。

### 轮询与发送

频道和私聊复用 `usePollingMessages`：启用时立即加载，之后每 3000ms 携带 `after_time` 和 `after_id` 增量拉取。消息以 id 去重并按 `(created_at,id)` 排序。页面进入后台时监听 `visibilitychange` 清理计时器；恢复可见后立即拉取并重新计时，组件卸载时也必须清理。

输入框 trim 后为空时不能发送，并设置 `maxLength=1000`。发送期间禁用按钮；成功后把服务端记录合并进消息流，失败时保留草稿并在聊天区域显示错误。私聊页进入以及收到新消息后调用 read 接口；导航未读 badge 低频刷新，并与本地已读操作同步。

访客点击加入或发送应跳转 `/login`；登录非成员不请求完整频道消息。接口返回 401 时沿用全局 token 清理与登录跳转，403/404 则在当前工作区显示受限或不存在状态。

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
