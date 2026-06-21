# MatchLab 问卷推荐 Service 分层与兼容迁移设计

## 背景与根因

本地 `main` 与 `origin/main` 已包含问卷、推荐、KM 和四条路由的实际实现。服务器仅出现 `doc.go` 且接口返回 404，说明服务器代码目录或运行中的二进制仍是旧版本。除重新部署正确提交外，本次按照最新规格补充明确的 service 层，并把 matches 的推荐明细从旧 `explanation` JSONB 迁移到独立字段。

## 目标

- 保留现有 health、auth、activity、application 行为。
- 提供可编译的 questionnaire 与 match model/repository/service/handler 分层。
- 四条 JWT 路由固定为 `/api/questionnaires`、`/api/me/profile`、`/api/match/recommend`、`/api/me/matches`。
- 对现有数据库执行可重复、无删表、无清空的兼容升级。
- 代码统一读写 matches 新字段，旧 `explanation` 与 `status` 暂时保留。

## 分层设计

### Questionnaire

- `model.go` 定义问卷、画像、JSONB 类型和请求信号。
- `repository.go` 只负责数据库事务：锁定用户、创建递增版本问卷、按 `user_id` upsert profile、读取 profile。
- `service.go` 负责 mode 校验、answers 规范化、画像生成和 repository 调用。
- `handler.go` 只负责从 JWT context 读取 user ID、JSON 绑定、HTTP 状态码和响应结构。公开方法为 `Submit` 与 `Profile`。

### Match

- `model.go` 定义五项得分、matches 新持久化字段和 API 返回模型。
- `repository.go` 负责读取画像与最新问卷、读取 recruiting 活动、upsert matches、读取当前算法版本结果。
- `service.go` 负责 target type 与 limit 校验、排除本人活动、规则评分、排序、持久化。
- `handler.go` 只负责 JWT、JSON 绑定、HTTP 错误和响应。公开方法为 `Recommend` 与 `MyMatches`。

评分继续采用 interest 30、skill 25、type 20、time 10、goal 15，总分 100；推荐理由只描述校园活动与项目协作。

## Matches 兼容迁移

`CREATE TABLE IF NOT EXISTS matches` 的新表定义包含：

- `target_id UUID`
- `target_type TEXT NOT NULL DEFAULT 'activity'`
- `algorithm TEXT NOT NULL DEFAULT 'rules'`
- `algorithm_version TEXT NOT NULL DEFAULT 'activity-rules-v1'`
- `detail_scores JSONB NOT NULL DEFAULT '{}'`
- `reason TEXT NOT NULL DEFAULT ''`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

对已有表使用 `ALTER TABLE matches ADD COLUMN IF NOT EXISTS` 增量增加相同字段，不删除 `explanation`、`status` 或任何表。

回填规则：

- `target_id = activity_id`
- `target_type = 'activity'`
- `algorithm = 'rules'`
- `detail_scores = COALESCE(explanation->'detail_scores', '{}'::jsonb)`
- `reason = COALESCE(explanation->>'reason', '')`
- 空 algorithm_version 回填为 `activity-rules-v1`
- 空 updated_at 回填为 created_at 或当前时间

建立唯一索引前，用窗口函数按 `(user_id, activity_id, algorithm_version)` 和 `updated_at DESC, created_at DESC, id DESC` 排序。每组最新记录保留原版本；其余重复记录不删除，而是将 algorithm_version 改为 `legacy-` 加 UUID 前缀。随后创建：

```sql
CREATE UNIQUE INDEX IF NOT EXISTS matches_user_activity_algorithm_uq
ON matches (user_id, activity_id, algorithm_version);
```

该策略保留所有历史行，并确保脚本重复执行时不会再次修改已处理记录。

## 持久化行为

新推荐统一写入：

- `target_id = activity_id`
- `target_type = activity`
- `algorithm = rules`
- `algorithm_version = activity-rules-v1`
- `score`
- `detail_scores`
- `reason`
- `updated_at`

upsert 冲突键为 `(user_id, activity_id, algorithm_version)`。冲突时更新上述推荐字段与 `questionnaire_id`，不创建重复行。`GET /api/me/matches` 只读取当前 `activity-rules-v1` 结果。

## 路由与错误

所有接口复用现有 JWT middleware：

```go
api.POST("/questionnaires", middleware.RequireAuth(tokens), questionnaireHandler.Submit)
api.GET("/me/profile", middleware.RequireAuth(tokens), questionnaireHandler.Profile)
api.POST("/match/recommend", middleware.RequireAuth(tokens), matchHandler.Recommend)
api.GET("/me/matches", middleware.RequireAuth(tokens), matchHandler.MyMatches)
```

无 profile 返回 404（画像查询）或 400 `profile_required`（推荐）；数据库不可用返回 503；非法 target type 或 limit 返回 400。

## 测试与部署验收

- service 单元测试覆盖画像生成、profile upsert 调用、本人活动排除、评分排序、limit 和持久化。
- handler 测试覆盖 JWT user ID、响应结构和错误映射。
- router 测试覆盖四条路由受 JWT 保护，并用有效 JWT 验证路由不会返回 404。
- KM 测试继续保证指定矩阵 KM=24、greedy=18。
- 最终运行 `go test ./...`。
- 按用户指定命令运行 `go build -o bin/matchlab-api cmd/server/main.go`。
- 服务器部署必须核对 `git rev-parse HEAD`、实际文件列表、schema 执行结果、二进制更新时间、systemd 重启时间和四条 curl。
