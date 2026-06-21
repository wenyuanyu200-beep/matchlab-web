# MatchLab 问卷画像与智能推荐模块设计

## 目标与范围

本阶段只完善 Go 后端，提供问卷提交、用户画像、活动规则推荐、最新推荐结果查询，以及可独立复用的 KM 最大权匹配算法。推荐定位为校园活动与项目协作，不调用大模型，不新增前端、聊天或社交关系功能。

## 模块边界

- `internal/questionnaire` 负责问卷模型、画像模型、画像生成规则、持久化和 HTTP 接口。
- `internal/match` 负责活动候选读取、规则评分、推荐理由、推荐结果 upsert、历史结果读取和 HTTP 接口。
- `internal/algorithm` 提供与业务无关的 KM 最大权匹配和 greedy 对比实现。
- `internal/router` 只组装依赖并注册受 JWT 保护的路由。

各模块通过 repository 接口隔离 GORM，以便 handler、service 和评分逻辑使用内存仓库或纯函数测试。

## 数据模型与迁移

### questionnaires

沿用现有表结构。每次提交创建一条新的已完成问卷记录，`answers` 使用 JSONB 保存，`mode` 作为 answers 之外的独立列保存。版本号按用户现有最大版本递增，从而保留用户的历次问卷输入，并让推荐结果关联最新问卷。

### profiles

保留现有 `display_name`、`interests`、`preferences` 等兼容字段，并增量增加：

- `profile_type VARCHAR(32) NOT NULL DEFAULT 'activity'`
- `tags JSONB NOT NULL DEFAULT '[]'`
- `scores JSONB NOT NULL DEFAULT '{}'`
- `summary TEXT NOT NULL DEFAULT ''`

`user_id` 继续保持唯一。提交问卷和生成画像在同一事务中完成，通过 `user_id` upsert，已有 profile 只更新画像字段，不创建重复记录。为满足旧表的 `display_name NOT NULL` 约束，首次生成 profile 时使用用户 nickname；nickname 为空时使用“校园协作者”。

### matches

沿用 `(user_id, activity_id, algorithm_version)` 唯一约束。`algorithm_version` 使用 `activity-rules-v1`。每次推荐通过 upsert 更新 `questionnaire_id`、`score`、`explanation`、`status` 和 `updated_at`。

`explanation` JSONB 结构为：

```json
{
  "detail_scores": {
    "interest": 26,
    "skill": 22,
    "type": 20,
    "time": 8,
    "goal": 12
  },
  "reason": "推荐原因：你的兴趣标签与该活动的电赛、STM32、硬件高度重合，适合参与该类竞赛组队。"
}
```

`GET /api/me/matches` 返回每个活动当前最新的已保存结果，不保存每次推荐运行的历史快照。

## 问卷与画像流程

`POST /api/questionnaires` 从 JWT 获取 user ID，校验 `mode` 和 answers JSON 对象，规范化 interests、skills 和 activity_types 中的字符串列表，保存问卷并生成画像。

画像 tags 依次合并 interests、skills、activity_types，去除空值和重复项并保持首次出现顺序。scores 使用固定、可解释的 MVP 规则：字段存在且非空时给予该维度的目标分值，否则给予较低基础分。summary 根据 mode、兴趣和技能标签生成自然中文，不使用交友、脱单或约会等词。

`GET /api/me/profile` 返回当前 profile；未提交问卷且无 profile 时返回 `404 profile_not_found`。

## 推荐评分

`POST /api/match/recommend` 仅支持 `target_type=activity`。`limit` 默认 10，最小 1，最大 50。服务读取当前 profile、最新完成问卷及 recruiting 活动，并排除当前用户创建的活动。

每个活动按以下上限计算整数分：

- 兴趣匹配 30：用户 interests 与 activity.tags 的交集比例。
- 技能匹配 25：用户 skills 与 activity.preferred_tags 的交集比例。
- 类型匹配 20：activity.type 存在于用户 activity_types 时得满分。
- 时间匹配 10：available_time 与 activity.time_text 经过中文及字母数字关键词切分后按交集比例评分；完全包含时得满分。
- 目标匹配 15：goal 与 activity title、description、tags 的关键词交集比例。

比较时忽略首尾空格和英文大小写。总分为五项之和，按总分降序、活动创建时间降序稳定排序，再应用 limit。

推荐理由优先描述实际命中的兴趣、技能、类型和时间信号；无明显交集时使用中性表达，说明活动类型与校园协作场景可能适合进一步了解。理由不得出现产品禁用的高风险词。

推荐成功后逐条 upsert matches，并返回：

```json
{
  "activity": {},
  "score": 88,
  "detail_scores": {},
  "reason": "推荐原因：……"
}
```

若 profile 不存在，返回 `400 profile_required`，提示先提交问卷。数据库不可用返回 `503 service_unavailable`，参数错误返回 `400 invalid_request`。

## KM 算法

`internal/algorithm` 暴露 KM 最大权匹配函数，输入二维整数权重矩阵，输出行列匹配对及总分。实现支持矩形矩阵；当行数大于列数时允许部分行不匹配。空矩阵返回空匹配和 0，行宽不一致返回错误。

greedy 对比函数按“从左到右处理每一行，每行选择尚未使用的最高权列”执行。指定矩阵：

```text
A1: 9, 8, 1
A2: 8, 1, 1
A3: 1, 8, 8
```

KM 选择 A1→2、A2→1、A3→3，总分 24；greedy 选择 A1→1、A2→2、A3→3，总分 18。

## 路由

以下路由统一使用现有 JWT middleware：

- `POST /api/questionnaires`
- `GET /api/me/profile`
- `POST /api/match/recommend`
- `GET /api/me/matches`

现有 health、auth、activity 和 application 路由保持不变。

## 测试策略

使用测试先行：先验证测试因缺少行为而失败，再写最小实现通过。

- 画像纯函数：tags 去重、scores、summary 和空字段行为。
- 问卷 service/handler：JWT user ID、事务性保存、profile upsert、错误映射。
- 推荐评分：各维度权重、排除本人活动、排序、limit、自然理由。
- 推荐持久化：唯一键 upsert 和 matches 查询结果映射。
- 路由：四个新增接口均受 JWT 保护。
- KM/greedy：指定 24/18 样例、矩形矩阵、空矩阵和非法矩阵。
- 回归：运行 `go test ./...`、`go vet ./...` 和 `go build ./cmd/server`。

## 文档与部署

更新 `docs/API_DOC.md`，加入四组接口 curl、完整用户 B 验证流程、matches 表检查 SQL 和 KM 测试命令。部署时先执行可重复运行的 `database/schema.sql` 增量迁移，再构建二进制并重启现有 systemd 服务。
