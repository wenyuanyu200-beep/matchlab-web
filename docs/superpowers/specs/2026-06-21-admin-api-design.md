# MatchLab 管理员后台 API 设计

## 目标与范围

本阶段只新增 Go 后端管理员 API，为后续后台页面提供用户、活动、报名、推荐、问卷和反馈数据。不实现前端或复杂权限系统，不改变已有普通用户接口。

## 管理员鉴权

在 `internal/middleware` 增加 `RequireAdmin()`。管理员路由先经过现有 `RequireAuth(tokens)`：JWT 无效或缺失由认证 middleware 返回 401；认证成功后，`RequireAdmin` 从 Gin context 读取 JWT role，role 不为 `admin` 时返回 403 `forbidden`。

角色来自 JWT，因此通过数据库或管理员接口变更角色后，用户必须重新登录才能获得反映新角色的 token。该设计避免每个管理员请求额外查询数据库。

## 模块结构

`internal/admin` 使用现有项目的四层结构：

- `model.go`：统计、筛选条件、安全用户、活动、报名、反馈和角色请求模型。
- `repository.go`：GORM 统计、过滤分页列表、JOIN 查询和用户角色更新。
- `service.go`：分页规范化、role 校验、自我降级保护和 repository 编排。
- `handler.go`：Gin query/body/UUID 解析、当前管理员 ID 获取、HTTP 响应与错误映射。

## 接口

以下路由统一位于 `/api/admin`，并同时使用认证和管理员 middleware：

- `GET /api/admin/stats`
- `GET /api/admin/users`
- `GET /api/admin/activities`
- `GET /api/admin/applications`
- `GET /api/admin/feedbacks`
- `POST /api/admin/users/:id/role`

成功列表响应统一位于 `data` 下，分别使用 `users`、`activities`、`applications`、`feedbacks`。列表无数据时返回空数组，不返回 null。

## 统计

Stats 返回：

- users_count
- activities_count
- applications_count
- matches_count
- questionnaires_count
- feedbacks_count
- recruiting_activities_count
- pending_applications_count
- approved_applications_count

所有值使用 64 位整数。repository 对对应表和状态执行 count，任一查询失败则整体返回 500；数据库未配置返回 503。

## 列表与过滤

分页默认 `limit=20`、最大 100、最小 1；`offset` 默认 0。非法数字、超出范围的 limit 或负 offset 返回 400。

Users：

- keyword 对 email、nickname、school 使用 `ILIKE`。
- role 为空或 `user/admin`；非法 role 返回 400。
- 只选择 id、email、nickname、role、school、created_at，模型中不存在 password_hash JSON 字段。

Activities：

- keyword 对 title、description 使用 `ILIKE`。
- type 使用精确匹配。
- status 为空或 `recruiting/full/closed`；非法状态返回 400。
- JOIN users 返回 creator nickname/school。

Applications：

- status 为空或 `pending/approved/rejected/cancelled`；非法状态返回 400。
- JOIN activities 和 users 返回 activity title、applicant nickname/school。

Feedbacks：

- 返回 id、user_id、activity_id、match_id、rating、comment、created_at。
- 支持 limit/offset。

所有列表按 created_at 降序、id 降序稳定排序。

## 角色变更

`POST /api/admin/users/:id/role` 仅接受 `user` 或 `admin`。目标用户不存在返回 404；非法 UUID 或 role 返回 400。若当前管理员将自己的角色设置为 user，service 返回 400 `self_demotion_forbidden`。重复设置为当前角色允许成功并返回安全用户模型。

## 数据库

不修改 `database/schema.sql`。现有 users.role、活动/报名状态字段和各表已有索引足够支持 MVP 数据量。关键词 ILIKE 暂不增加全文索引。

## 测试

- middleware：无 role、user role 返回 403，admin role 放行；无效 JWT 仍由 RequireAuth 返回 401。
- service：分页默认/边界、过滤值校验、自我降级、角色更新。
- handler：安全响应、空数组、UUID/body/query 错误映射。
- repository 查询通过 service/handler fake 验证契约；GORM 实现使用明确 Select，防止密码泄漏。
- router：六条路由无 token 返回 401、普通 token 返回 403、admin token 能进入 handler 而不是 404。
- 回归：`go test -count=1 ./...`、`go vet ./...`、`go build -o bin/matchlab-api cmd/server/main.go`。
