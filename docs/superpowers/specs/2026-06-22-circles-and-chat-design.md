# MatchLab 圈子与聊天 MVP 设计规格

## 目标与边界

在现有注册登录、活动发布、报名审批、画像推荐、管理员后台和生产推荐 E2E 闭环不变的前提下，新增可演示的圈子、默认频道聊天和两人私聊系统。圈子定位为校园兴趣、学习与项目协作空间，不引入交友、约会或陌生人社交叙事。

MVP 包含圈子创建审核、公开广场、加入圈子、默认 `general` 频道、频道轮询聊天、两人私聊、私聊未读数和管理员审核。明确不做邮箱验证码、WebSocket、图片/文件消息、表情包、撤回、禁言、复杂权限和多频道管理后台。

## 总体架构

沿用现有 Next.js + TypeScript 前端、Go + Gin 后端、GORM + PostgreSQL 数据层和 JWT 中间件。新增 `circle` 与 `conversation` 两个后端业务包，各自包含模型、仓储、服务、处理器和测试；路由只在现有 `/api` 与受保护的 `/api/admin` 分组中增量注册，不改变现有活动、报名、推荐接口。

聊天采用数据库持久化与 3000ms 增量轮询。客户端使用 `after_time` 作为主游标并按 `(created_at, id)` 去重、排序；API 同时接受 `after_id`，以便后续优化。页面隐藏时暂停轮询，重新可见后立即拉取，发送成功后将服务端返回记录追加到列表。

## 数据库设计与安全迁移

新增可重复执行的 `database/migrations/20260622_circles_and_chat.sql`，只使用 `CREATE TABLE IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS` 和安全约束，不使用 `DROP` 或 `TRUNCATE`。同时将等价表结构追加到 `database/schema.sql`，确保新环境与增量部署一致。

新增七张表：

- `circles`：圈子资料、创建者、审核状态、审核信息和成员计数。
- `circle_members`：成员、角色和成员状态，唯一约束 `(circle_id, user_id)`。
- `circle_channels`：频道；每个圈子至少有一个 `general` 文本频道。
- `circle_messages`：频道消息，按频道和时间建立轮询索引。
- `conversations`：私聊会话及来源上下文；direct 会话保存标准化双方身份生成的 `direct_key` 并设置唯一索引。
- `conversation_members`：会话参与人和 `last_read_at`，唯一约束 `(conversation_id, user_id)`。
- `direct_messages`：私聊消息，按会话和时间建立轮询索引。

所有主键使用 UUID。消息内容由后端 trim 并限制为 1–1000 字符。圈子名称为 2–40 字符，描述不超过 300 字符，标签最多 8 个。创建圈子、写入 owner 成员和创建默认频道放在一个事务中；加入圈子与增加 `member_count` 也放在一个事务中，依赖唯一约束避免并发重复加入。直接会话通过标准化的双方身份键保证同一对用户只存在一个 direct conversation，避免并发重复创建。

## 后端接口与权限

圈子接口：

- `GET /api/circles`：允许访客或登录用户查看 approved 圈子；登录用户额外返回 `joined`。
- `POST /api/circles`：登录用户创建 pending 圈子。
- `GET /api/circles/:id`：approved 圈子可公开查看基础信息；pending/rejected 仅创建者和管理员可见。
- `POST /api/circles/:id/join`：登录用户加入 approved 圈子，重复加入返回稳定的业务错误。
- `GET /api/me/circles`：返回当前用户加入的 approved 圈子和自己创建的待审/拒绝圈子。
- `GET /api/circles/:circleId/members`：基础资料可展示 approved 圈子的前 8 位成员；只有已加入成员才获得私聊入口所需的用户 id。
- `GET /api/circles/:circleId/channels`：仅活跃成员或管理员可读。
- `GET|POST /api/circles/:circleId/channels/:channelId/messages`：仅活跃成员或管理员可读；只有活跃成员可发送。

私聊接口：

- `POST /api/conversations/direct`：登录用户创建或复用两人会话，拒绝与自己私聊；`source_type` 仅允许 `circle/activity/match/manual`。
- `GET /api/me/conversations`：返回对方公开资料、最后一条消息、会话未读数和更新时间。
- `GET|POST /api/conversations/:id/messages`：仅会话成员可读写。
- `POST /api/conversations/:id/read`：仅会话成员更新自己的 `last_read_at`。
- `GET /api/me/unread-count`：返回当前用户所有私聊未读消息总数。

管理员接口全部注册到现有 `RequireAuth + RequireAdmin` 路由组：

- `GET /api/admin/circles?status=pending`
- `POST /api/admin/circles/:id/approve`
- `POST /api/admin/circles/:id/reject`

普通响应只暴露用户 id、nickname、school、role/圈子角色和 initials 所需信息，不返回邮箱。pending/rejected 圈子对非创建者、非管理员统一返回 404，避免泄露资源存在性。

## 前端页面与交互

新增 `/circles` 圈子广场、`/circles/create` 创建申请、`/circles/[id]` 圈子工作区、`/messages` 私聊列表和 `/messages/[id]` 私聊详情；扩展 `/admin` 圈子审核，并在导航栏加入“圈子”“消息”和登录后的未读 badge。

圈子详情采用已确认的方案 A：

- 桌面端为左右两栏。左侧窄栏显示圈子名称、中文分类 badge、描述、标签、成员数、加入状态、最多 8 位成员和私聊入口；创建者或管理员还能看到审核状态。右侧为默认频道工作区，顶部显示频道名称和说明，中间是消息流，底部是固定输入区。
- 移动端不保留并排布局，使用“圈子信息 / 频道聊天 / 成员”三个顶部 Tab，默认进入频道聊天；各 Tab 独立滚动且不产生横向溢出。
- 当前用户消息靠右，其他用户消息靠左；消息展示 nickname、school、时间和非本人“私聊”入口。视觉沿用浅色卡片、蓝紫主按钮和柔和 badge，不引入重型社区工具风格。

状态规则：

- 未登录用户可以查看 approved 圈子基础信息，加入或发送动作跳转 `/login`，聊天区显示登录引导。
- 已登录但未加入用户只看基础资料和有限成员摘要，聊天区显示加入引导，不请求完整消息。
- 已加入用户可读写频道消息并向其他成员发起私聊。
- pending/rejected 仅创建者或管理员可见，页面展示待审或拒绝原因，不开放普通聊天。

私聊按钮调用 `POST /api/conversations/direct`，传 `source_type: "circle"` 与当前圈子 id，成功后跳转 `/messages/[id]`。当前用户本人旁边不显示按钮。

## 轮询与错误处理

封装共享 `usePollingMessages` hook，供频道和私聊详情复用。hook 在启用时立即加载，随后每 3000ms 增量拉取；监听 `visibilitychange`，隐藏时清理计时器，恢复时立即拉取并重建计时器。客户端通过消息 id 去重，防止发送后立即追加与下一次轮询产生重复。

空消息不发送，输入限制 `maxLength=1000`，发送中禁用按钮。发送失败保留草稿，只在聊天区域显示轻量错误，不使用连续 alert。私聊详情进入和收到新消息后调用 read 接口，导航未读数通过低频刷新与本地读状态同步。

## 测试策略

后端采用服务层单元测试、仓储 SQL mock/数据库边界测试和路由权限测试，重点覆盖：审核可见性、重复加入、非成员读写频道、direct 会话复用、非成员访问私聊、自聊拒绝、空白/超长消息、已读与未读计数、管理员中间件。

前端使用 Vitest + Testing Library 覆盖广场筛选、创建提交、详情页登录/未加入/已加入/待审状态、移动 Tab、轮询暂停恢复、发送失败保留草稿、私聊跳转、未读 badge 和管理员审核。完成后运行：

- `cd backend && go test ./...`
- `cd backend && go build -o bin/matchlab-api cmd/server/main.go`
- `cd frontend && npm install`
- `cd frontend && npm run test`
- `cd frontend && npm run lint`
- `cd frontend && npm run build`

最后按 A 创建圈子、管理员审核、B 加入、双方频道聊天、B 发起私聊、双方私聊、未加入用户受限的流程人工回归，并复跑 `scripts/production_e2e_recommend.sh`。生产脚本若需要真实环境凭据或会改变生产数据，只做语法/依赖检查并在交付说明中列出安全执行命令，不擅自操作生产环境。

## 部署与兼容性

部署顺序为：备份数据库并记录版本、执行幂等 migration、部署并重启后端、验证 health 与新接口、部署前端、执行新旧人工回归。旧接口路径、现有表结构和现有环境变量均不改变；新表与新路由仅做增量扩展。最大风险是 migration 与应用发布不同步、聊天轮询造成额外数据库负载、成员计数并发漂移和旧生产数据库约束差异，分别通过先迁移后发布、增量游标与索引、事务更新和部署前结构检查控制。
