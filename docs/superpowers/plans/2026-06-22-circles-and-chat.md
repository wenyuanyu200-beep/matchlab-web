# Circles and Chat MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变现有生产闭环的前提下，交付圈子创建审核、加入圈子、默认频道轮询聊天、两人私聊、未读数和管理员审核的响应式 MVP。

**Architecture:** 后端增加彼此独立的 `circle` 与 `conversation` 业务包，通过现有 JWT/admin 中间件增量注册 API，并用幂等 PostgreSQL migration 建表。前端增加圈子与消息页面、共享轮询 hook 和导航未读 badge；桌面端圈子详情采用已确认的左右协作工作区，移动端切换为三个 Tab。

**Tech Stack:** Go 1.22、Gin、GORM、PostgreSQL、Next.js 16、React 19、TypeScript、Tailwind CSS 4、Vitest、Testing Library

---

## 文件职责与并行边界

**Agent 1（后端 API 与 migration）仅修改：**

- `database/migrations/20260622_circles_and_chat.sql`：可重复执行的七表 migration 和索引。
- `database/schema.sql`：新环境的等价基础结构。
- `backend/internal/circle/*`：圈子领域模型、仓储、服务、HTTP handler 与测试。
- `backend/internal/conversation/*`：私聊领域模型、仓储、服务、HTTP handler 与测试。
- `backend/internal/router/router.go`、`backend/internal/router/router_test.go`：依赖注入和路由注册。

**Agent 2（前端页面与交互）仅修改：**

- `frontend/lib/types.ts`、`frontend/lib/circles.ts`：共享类型、分类与状态文案。
- `frontend/hooks/usePollingMessages.ts` 及测试：可见性感知的增量轮询。
- `frontend/components/Navbar.tsx` 及测试：圈子、消息与未读 badge。
- `frontend/components/chat/*`：共享消息列表和输入组件。
- `frontend/app/circles/**`、`frontend/app/messages/**`：新页面及测试。
- `frontend/app/admin/page.tsx`、`frontend/app/admin/page.test.tsx`：审核区。
- `frontend/app/globals.css`：最少量响应式工作区样式。

**Agent 3（QA、权限、E2E、部署说明）仅修改：**

- `docs/API_DOC.md`：新接口契约与权限矩阵。
- `docs/DEPLOY.md`：migration、发布、回滚与人工验证步骤。
- `docs/FRONTEND.md`：新页面、轮询和响应式说明。
- `docs/qa/circles-and-chat-regression.md`：测试证据、风险清单与人工回归结果。

Agent 3 先做只读权限审计和基线测试，待 Agent 1/2 完成后运行全量验证；不直接改 Agent 1/2 的实现文件，发现问题交给主 Agent 分派修复，避免共享工作区冲突。

## Phase 0：分支与基线

### Task 1：确认安全起点

**Files:**
- Inspect: `.git/HEAD`
- Inspect: `scripts/production_e2e_recommend.sh`

- [ ] **Step 1: 确认分支与工作区**

Run: `git status --short --branch`

Expected: 当前分支为 `feature/circles-and-chat`，业务文件无未提交修改。

- [ ] **Step 2: 建立后端基线**

Run: `cd backend && go test ./...`

Expected: 所有现有 package PASS。

- [ ] **Step 3: 建立前端基线**

Run: `cd frontend && npm run test && npm run lint && npm run build`

Expected: Vitest、ESLint 和 Next.js production build 全部成功。

- [ ] **Step 4: 检查生产 E2E 的外部依赖**

Run: `bash -n scripts/production_e2e_recommend.sh && rg -n "BASE_URL|TOKEN|PASSWORD|curl" scripts/production_e2e_recommend.sh`

Expected: shell 语法通过；记录脚本需要的环境、账号和是否会写入生产数据，不在无授权时执行远程写操作。

## Phase 1：Agent 1 后端与数据库

### Task 2：幂等数据库结构

**Files:**
- Create: `database/migrations/20260622_circles_and_chat.sql`
- Modify: `database/schema.sql`

- [ ] **Step 1: 写 migration 静态安全测试**

Run:

```powershell
$sql = Get-Content database/migrations/20260622_circles_and_chat.sql -Raw
if ($sql -match '(?i)\b(DROP|TRUNCATE)\b') { throw 'destructive SQL found' }
@('circles','circle_members','circle_channels','circle_messages','conversations','conversation_members','direct_messages') | ForEach-Object { if ($sql -notmatch "CREATE TABLE IF NOT EXISTS\s+$($_)") { throw "missing $_" } }
```

Expected: 在 migration 尚未创建时失败。

- [ ] **Step 2: 写最小幂等 migration**

每张表使用 UUID 主键和 `CREATE TABLE IF NOT EXISTS`；所有查询路径使用 `CREATE INDEX IF NOT EXISTS`。`conversations.direct_key` 可为空，但 direct 会话写入排序后的 `min(userID):max(userID)`，并建立 `WHERE type = 'direct'` 的唯一索引。消息正文增加 `char_length(content) BETWEEN 1 AND 1000` 检查；状态与角色增加白名单检查。

- [ ] **Step 3: 重跑静态测试并检查 schema 一致性**

Run: 上一步 PowerShell 检查，再运行 `git diff --check database/`。

Expected: 无破坏性 SQL、七张表齐全、无空白错误。

- [ ] **Step 4: 提交 migration**

Run: `git add database && git commit -m "feat: add circles and chat database schema"`

### Task 3：圈子领域的 TDD 实现

**Files:**
- Create: `backend/internal/circle/doc.go`
- Create: `backend/internal/circle/model.go`
- Create: `backend/internal/circle/repository.go`
- Create: `backend/internal/circle/repository_test.go`
- Create: `backend/internal/circle/service.go`
- Create: `backend/internal/circle/service_test.go`
- Create: `backend/internal/circle/handler.go`
- Create: `backend/internal/circle/handler_test.go`

- [ ] **Step 1: 用失败测试固定服务契约**

在 `service_test.go` 定义 fake repository，覆盖下列用例：创建时 trim/校验并在事务语义仓储方法中创建 owner 与 general；approved 列表过滤；pending/rejected 仅 creator/admin 可见；重复加入返回 `ErrAlreadyMember`；非成员不能列频道或消息；空白与 1001 字消息返回 `ErrInvalidMessage`。

核心接口固定为：

```go
type Repository interface {
    List(context.Context, ListFilter) ([]Circle, error)
    Create(context.Context, uuid.UUID, CreateInput) (CircleDetail, error)
    Detail(context.Context, uuid.UUID, uuid.UUID, string) (CircleDetail, error)
    Join(context.Context, uuid.UUID, uuid.UUID) error
    MyCircles(context.Context, uuid.UUID) ([]Circle, error)
    Members(context.Context, uuid.UUID, int) ([]Member, error)
    Channels(context.Context, uuid.UUID) ([]Channel, error)
    Messages(context.Context, uuid.UUID, uuid.UUID, MessageCursor) ([]Message, error)
    SendMessage(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (Message, error)
    AdminList(context.Context, string) ([]Circle, error)
    Review(context.Context, uuid.UUID, uuid.UUID, ReviewInput) (Circle, error)
}
```

- [ ] **Step 2: 运行圈子测试确认失败**

Run: `cd backend && go test ./internal/circle -run 'TestService' -v`

Expected: FAIL，因为类型和服务尚未实现。

- [ ] **Step 3: 实现模型、校验与服务**

`CreateInput` 限制 name 2–40、description ≤300、tags ≤8；`MessageCursor` 限制 `limit` 为 1–100 且默认 50；`ReviewInput` 仅允许 approve/reject，reject 必须有 trim 后原因。所有公开用户摘要仅含 `id/nickname/school`。

- [ ] **Step 4: 实现 GORM repository**

创建和加入使用 `db.Transaction`；加入先插入成员，成功后 `member_count = member_count + 1`；消息查询必须同时校验 circle/channel 关系和 active membership；admin 可读频道但不能以普通成员身份发言。

- [ ] **Step 5: 实现 handler 并用失败测试固定 HTTP 状态**

测试 401、403、404、409、422 和 503 映射，以及列表查询 `category/q/limit/offset`、消息 `after_id/after_time/limit`。响应统一为 `{"data": ...}`，不得序列化邮箱。

- [ ] **Step 6: 运行并格式化**

Run: `cd backend && gofmt -w internal/circle && go test ./internal/circle -v`

Expected: PASS。

- [ ] **Step 7: 提交圈子领域**

Run: `git add backend/internal/circle && git commit -m "feat: add circle services and APIs"`

### Task 4：私聊领域的 TDD 实现

**Files:**
- Create: `backend/internal/conversation/doc.go`
- Create: `backend/internal/conversation/model.go`
- Create: `backend/internal/conversation/repository.go`
- Create: `backend/internal/conversation/repository_test.go`
- Create: `backend/internal/conversation/service.go`
- Create: `backend/internal/conversation/service_test.go`
- Create: `backend/internal/conversation/handler.go`
- Create: `backend/internal/conversation/handler_test.go`

- [ ] **Step 1: 写失败服务测试**

覆盖自聊拒绝、source_type 白名单、排序 direct key、并发唯一冲突后读取已有会话、非成员读写拒绝、空白/超长消息、会话列表最后消息、`created_at > last_read_at` 未读统计和只更新当前成员已读时间。

核心接口固定为：

```go
type Repository interface {
    Direct(context.Context, uuid.UUID, uuid.UUID, DirectInput) (Conversation, error)
    List(context.Context, uuid.UUID) ([]ConversationSummary, error)
    Messages(context.Context, uuid.UUID, uuid.UUID, MessageCursor) ([]Message, error)
    Send(context.Context, uuid.UUID, uuid.UUID, string) (Message, error)
    MarkRead(context.Context, uuid.UUID, uuid.UUID) error
    UnreadCount(context.Context, uuid.UUID) (int64, error)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/conversation -run 'TestService' -v`

Expected: FAIL，因为实现不存在。

- [ ] **Step 3: 实现 service、repository 与 handler**

只允许两人成员；每个 repository 读写都将 `user_id` 加入查询条件；direct 成功返回会话 id；列表返回对方 nickname/school、last message、unread_count、updated_at；消息响应不含邮箱。

- [ ] **Step 4: 运行并格式化**

Run: `cd backend && gofmt -w internal/conversation && go test ./internal/conversation -v`

Expected: PASS。

- [ ] **Step 5: 提交私聊领域**

Run: `git add backend/internal/conversation && git commit -m "feat: add direct conversations and unread counts"`

### Task 5：注册后端路由并做权限回归

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`

- [ ] **Step 1: 扩充失败路由测试**

将写接口、我的资源、频道和私聊路由列入未认证 401 测试；将 admin circles 三条路由列入普通用户 403、管理员可达测试；公开列表与 approved 详情不得被认证中间件拦截。

- [ ] **Step 2: 运行路由测试确认失败**

Run: `cd backend && go test ./internal/router -run 'TestCircle|TestConversation|TestAdmin' -v`

Expected: 新路由返回 404，测试失败。

- [ ] **Step 3: 注入 repository、service、handler 并注册路由**

公开：`GET /circles`、`GET /circles/:id`、`GET /circles/:id/members`。登录保护：创建、加入、我的圈子、频道、频道消息、所有 conversation 接口。管理员审核注册到现有 `adminRoutes`。

- [ ] **Step 4: 运行后端全量测试与构建**

Run: `cd backend && gofmt -w internal/router && go test ./... && go build -o bin/matchlab-api cmd/server/main.go`

Expected: 所有测试 PASS，二进制构建成功。

- [ ] **Step 5: 提交路由集成**

Run: `git add backend/internal/router && git commit -m "feat: register circles and chat routes"`

## Phase 2：Agent 2 前端页面与交互

### Task 6：共享类型、文案与轮询 hook

**Files:**
- Modify: `frontend/lib/types.ts`
- Create: `frontend/lib/circles.ts`
- Create: `frontend/hooks/usePollingMessages.ts`
- Create: `frontend/hooks/usePollingMessages.test.tsx`

- [ ] **Step 1: 写 fake timer 失败测试**

测试 hook 首次立即请求、3000ms 后增量请求、`document.hidden` 时暂停、恢复可见时立即请求、卸载清理计时器、请求错误只更新轻量 error。使用 `vi.useFakeTimers()` 和可配置 `document.hidden`。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npm run test -- hooks/usePollingMessages.test.tsx`

Expected: FAIL，因为 hook 不存在。

- [ ] **Step 3: 实现共享契约**

定义 `Circle`、`CircleDetail`、`CircleMember`、`Channel`、`ChatMessage`、`ConversationSummary`；`usePollingMessages<T>` 接收 `enabled/intervalMs/fetchNewMessages/getID`，返回 `items/error/refresh/append/setItems`，使用 Set 按 id 去重并按时间排序。

- [ ] **Step 4: 运行测试并提交**

Run: `cd frontend && npm run test -- hooks/usePollingMessages.test.tsx && git add lib hooks && git commit -m "feat: add circle types and polling hook"`

Expected: PASS。

### Task 7：圈子广场与创建申请

**Files:**
- Create: `frontend/app/circles/page.tsx`
- Create: `frontend/app/circles/page.test.tsx`
- Create: `frontend/app/circles/create/page.tsx`
- Create: `frontend/app/circles/create/page.test.tsx`

- [ ] **Step 1: 写页面失败测试**

广场测试 approved 卡片、中文分类、搜索、分类筛选、登录/未登录加入行为和空状态；创建页测试未登录跳转、2–40 字名称、300 字描述、最多 8 标签、成功提示和 pending 状态。

- [ ] **Step 2: 运行确认失败**

Run: `cd frontend && npm run test -- app/circles/page.test.tsx app/circles/create/page.test.tsx`

Expected: FAIL，因为页面不存在。

- [ ] **Step 3: 实现页面**

广场使用现有 `page-shell/card/tag/button-*` 风格，查询参数发送到 `/circles`；加入成功就地更新 `joined/member_count`。创建页 POST `/circles`，标签用逗号分隔、trim、去重后提交。

- [ ] **Step 4: 运行并提交**

Run: `cd frontend && npm run test -- app/circles && git add app/circles && git commit -m "feat: add circles discovery and creation pages"`

Expected: PASS。

### Task 8：响应式圈子工作区

**Files:**
- Create: `frontend/components/chat/MessageList.tsx`
- Create: `frontend/components/chat/MessageComposer.tsx`
- Create: `frontend/components/chat/MessageComposer.test.tsx`
- Create: `frontend/app/circles/[id]/page.tsx`
- Create: `frontend/app/circles/[id]/page.test.tsx`
- Modify: `frontend/app/globals.css`

- [ ] **Step 1: 写失败交互测试**

覆盖四种状态：访客显示登录引导；登录未加入不请求频道消息并显示加入按钮；已加入加载 general、轮询、发送和私聊；pending/rejected 仅显示审核状态。另测空白不发送、1000 字上限、发送中禁用、失败保留草稿、本人无私聊按钮、他人私聊成功跳转 `/messages/:id`。

- [ ] **Step 2: 运行确认失败**

Run: `cd frontend && npm run test -- app/circles/[id]/page.test.tsx components/chat/MessageComposer.test.tsx`

Expected: FAIL，因为组件不存在。

- [ ] **Step 3: 实现桌面协作工作区**

在 `lg` 断点使用 `grid-template-columns: minmax(240px, 320px) minmax(0, 1fr)`；左侧显示资料、状态和前 8 位成员，右侧使用 `min-width:0`、可滚动消息区和底部输入框。消息本人靠右、他人靠左，展示 nickname/school/time。

- [ ] **Step 4: 实现移动 Tab**

`lg:hidden` 显示“圈子信息 / 频道聊天 / 成员”按钮，默认频道聊天；一次只渲染一个面板，正文使用 `overflow-wrap:anywhere`，不得出现固定最小宽度。

- [ ] **Step 5: 接入轮询和私聊**

只在 `joined && approved && activeTab === 'chat'` 时启用轮询。私聊 POST `/conversations/direct`，body 为 `target_user_id/source_type:'circle'/source_id`，成功后 `router.push('/messages/' + id)`。

- [ ] **Step 6: 运行并提交**

Run: `cd frontend && npm run test -- app/circles/[id] components/chat && git add app/circles/[id] components/chat app/globals.css && git commit -m "feat: add responsive circle workspace"`

Expected: PASS。

### Task 9：私聊页面、未读导航与管理员审核

**Files:**
- Create: `frontend/app/messages/page.tsx`
- Create: `frontend/app/messages/page.test.tsx`
- Create: `frontend/app/messages/[id]/page.tsx`
- Create: `frontend/app/messages/[id]/page.test.tsx`
- Modify: `frontend/components/Navbar.tsx`
- Modify: `frontend/components/Navbar.test.tsx`
- Modify: `frontend/app/admin/page.tsx`
- Modify: `frontend/app/admin/page.test.tsx`

- [ ] **Step 1: 写失败测试**

测试会话列表的对方、最后消息、未读 badge；详情页加载、3000ms 轮询、进入标记已读、空状态和发送失败保留草稿；Navbar 登录后请求未读数并显示“圈子/消息”；admin 加载 pending 圈子、通过、拒绝原因必填和成功后移除条目。

- [ ] **Step 2: 运行确认失败**

Run: `cd frontend && npm run test -- app/messages components/Navbar.test.tsx app/admin/page.test.tsx`

Expected: FAIL，新功能尚未存在。

- [ ] **Step 3: 实现私聊页面**

复用 chat 组件和 polling hook；进入详情后 POST `/conversations/:id/read`，成功后触发自定义 unread 事件让 Navbar 立即刷新；轮询仅在页面可见时运行。

- [ ] **Step 4: 实现 Navbar 与审核区**

Navbar 仅登录后拉取 `/me/unread-count`，失败静默；admin 并行加载 `/admin/circles?status=pending`，approve/reject 成功后本地更新列表并刷新待审数。

- [ ] **Step 5: 前端全量验证并提交**

Run: `cd frontend && npm run test && npm run lint && npm run build`

Expected: 全部成功。

Run: `git add frontend && git commit -m "feat: add direct messages unread badge and circle review"`

## Phase 3：Agent 3 QA、文档与回归

### Task 10：权限矩阵与契约审计

**Files:**
- Modify: `docs/API_DOC.md`
- Create: `docs/qa/circles-and-chat-regression.md`

- [ ] **Step 1: 建立权限矩阵**

逐接口记录访客、已登录非成员、圈子成员、创建者、管理员的预期 200/401/403/404；特别检查 pending/rejected、频道读写、私聊成员关系、自聊、邮箱泄露和 admin middleware。

- [ ] **Step 2: 对照实现与测试证据**

Run: `rg -n "circles|channels|conversations|unread" backend/internal/router backend/internal/circle backend/internal/conversation`

Expected: 每个设计接口有路由、handler 和至少一个服务/路由测试。

- [ ] **Step 3: 记录缺陷而不跨边界改实现**

将发现写入 QA 文档的“待修复”表，字段为严重级、接口/页面、复现命令、预期、实际；立即通知主 Agent 分派给对应实现 Agent。

### Task 11：部署与人工回归说明

**Files:**
- Modify: `docs/DEPLOY.md`
- Modify: `docs/FRONTEND.md`
- Modify: `docs/API_DOC.md`
- Modify: `docs/qa/circles-and-chat-regression.md`

- [ ] **Step 1: 写部署步骤**

顺序固定为数据库备份 → 执行 `20260622_circles_and_chat.sql` → 重启后端 → `/api/health` 与新接口 smoke → 部署前端 → 新旧人工回归。回滚只回滚应用版本并保留新表，不提供 DROP migration。

- [ ] **Step 2: 写 12 步人工测试清单**

覆盖 A 创建、admin 审核、B 发现并加入、A/B general 聊天、3 秒增量、B 发起私聊、双方私聊、未读数、未加入限制，以及活动/报名/推荐/admin 回归。

- [ ] **Step 3: 提交文档**

Run: `git add docs && git commit -m "docs: add circles chat API deployment and QA guide"`

## Phase 4：集成修复与最终验证

### Task 12：合并检查与缺陷闭环

**Files:**
- Modify: only files tied to a reproduced defect

- [ ] **Step 1: 检查共享工作区和提交历史**

Run: `git status --short --branch && git log --oneline --decorate -12`

Expected: 所有三条工作流提交可见，无未解释的临时文件。

- [ ] **Step 2: 运行格式和静态检查**

Run: `git diff --check HEAD~10..HEAD && cd backend && gofmt -l .`

Expected: `git diff --check` 无输出，`gofmt -l` 无输出。

- [ ] **Step 3: 对每个 QA 缺陷先补失败测试再修复**

Run: 由缺陷所属 package/page 的精确测试命令验证先 FAIL；最小修复后重复同一命令直到 PASS。每个修复单独提交为 `fix: <observable behavior>`。

### Task 13：完成前证据验证

**Files:**
- Update: `docs/qa/circles-and-chat-regression.md`

- [ ] **Step 1: 后端最终验证**

Run: `cd backend && go test ./... && go build -o bin/matchlab-api cmd/server/main.go`

Expected: exit code 0，所有 package PASS，生成 `backend/bin/matchlab-api`。

- [ ] **Step 2: 前端最终验证**

Run: `cd frontend && npm install && npm run test && npm run lint && npm run build`

Expected: exit code 0，Vitest 全绿、ESLint 无 error、Next.js build 成功。

- [ ] **Step 3: 生产 E2E 安全验证**

Run: `bash -n scripts/production_e2e_recommend.sh`

若用户已提供测试环境与可写测试账号，再运行：`bash scripts/production_e2e_recommend.sh`。否则在 QA 文档明确标记“远程生产写入未执行”，提供命令和所需环境变量，不将语法检查误报为 E2E 通过。

- [ ] **Step 4: 浏览器响应式人工验证**

在 1280×720 和 390×844 检查 `/circles`、`/circles/:id`、`/messages`、`/messages/:id`、`/admin`：无横向溢出；移动端 Tab 可用；输入框、错误提示、加入引导和私聊跳转正确。

- [ ] **Step 5: 更新 QA 证据并提交**

记录每条命令、时间、exit code、失败与修复、未执行项及原因。

Run: `git add docs/qa/circles-and-chat-regression.md && git commit -m "test: record circles and chat verification"`

### Task 14：交付摘要

**Files:**
- Inspect: `git diff main...HEAD --stat`
- Inspect: `git log main..HEAD --oneline`

- [ ] **Step 1: 汇总交付内容**

按新增表/migration、后端接口、前端页面、审核流程、轮询实现、私聊权限、测试结果、原功能影响、风险和部署步骤九项输出。

- [ ] **Step 2: 明确剩余风险**

只报告有证据的风险：生产 migration 尚未实跑、远程 E2E 未获凭据、轮询负载尚未压测或人工浏览器流程尚未完成；不得把未执行检查描述为通过。
