# MatchLab API 文档

## 通用信息

- 本地基础地址：`http://127.0.0.1:8080`
- 公网基础地址：`http://139.224.119.187`
- API 前缀：`/api`
- 响应格式：JSON

## GET /api/health

检查 API 进程是否正在运行。该接口故意不检查 PostgreSQL，因此数据库离线时仍可用于进程存活探测。

请求：

```http
GET /api/health HTTP/1.1
Host: 127.0.0.1:8080
```

成功响应：`200 OK`

```json
{
  "ok": true,
  "message": "MatchLab API running"
}
```

示例：

```bash
curl -i http://127.0.0.1:8080/api/health
```

## 通用响应格式

成功响应使用：

```json
{"data": {}}
```

错误响应使用：

```json
{
  "error": "machine_readable_code",
  "message": "可读错误信息"
}
```

## POST /api/auth/register

创建邮箱账号。邮箱会去除首尾空格并转为小写；密码至少 8 个字符，并且 UTF-8 编码后不超过 bcrypt 的 72 字节限制。

```json
{
  "email": "test@example.com",
  "password": "12345678",
  "nickname": "测试用户",
  "school": "西南大学"
}
```

成功：`201 Created`

```json
{
  "data": {
    "user": {
      "id": "ab49ea1e-19bd-4c78-a8d1-95e0e827ba4d",
      "email": "test@example.com",
      "nickname": "测试用户",
      "role": "user",
      "school": "西南大学",
      "created_at": "2026-06-21T12:00:00Z",
      "updated_at": "2026-06-21T12:00:00Z"
    }
  }
}
```

可能状态：`400` 参数错误、`409` 邮箱已注册、`503` 数据库不可用。

```bash
curl -i -X POST http://127.0.0.1:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"12345678","nickname":"测试用户","school":"西南大学"}'
```

## POST /api/auth/login

校验邮箱和密码。邮箱不存在与密码错误统一返回 `401 Unauthorized`。

```json
{
  "email": "test@example.com",
  "password": "12345678"
}
```

成功：`200 OK`

```json
{
  "data": {
    "token": "eyJ...",
    "user": {
      "id": "ab49ea1e-19bd-4c78-a8d1-95e0e827ba4d",
      "email": "test@example.com",
      "nickname": "测试用户",
      "role": "user",
      "school": "西南大学",
      "created_at": "2026-06-21T12:00:00Z",
      "updated_at": "2026-06-21T12:00:00Z"
    }
  }
}
```

```bash
curl -i -X POST http://127.0.0.1:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"12345678"}'
```

可用 `jq` 提取 token：

```bash
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"12345678"}' | jq -r '.data.token')
```

## GET /api/me

返回 Bearer token 对应的当前用户。响应不包含 `password_hash`。

```bash
curl -i http://127.0.0.1:8080/api/me \
  -H "Authorization: Bearer $TOKEN"
```

成功：`200 OK`

```json
{
  "data": {
    "user": {
      "id": "ab49ea1e-19bd-4c78-a8d1-95e0e827ba4d",
      "email": "test@example.com",
      "nickname": "测试用户",
      "role": "user",
      "school": "西南大学",
      "created_at": "2026-06-21T12:00:00Z",
      "updated_at": "2026-06-21T12:00:00Z"
    }
  }
}
```

缺少、过期、签名错误或格式错误的 token 返回 `401 Unauthorized`。

## Activities and applications

The activity module supports campus activities and project collaboration. Responses never include password hashes.

### Create activity

```bash
curl -i -X POST http://127.0.0.1:8080/api/activities \
  -H "Authorization: Bearer $TOKEN_A" \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "智能硬件比赛组队",
    "type": "competition",
    "description": "准备参加电子设计竞赛，寻找会单片机和结构设计的队友。",
    "required_count": 3,
    "tags": ["电赛", "STM32", "硬件"],
    "preferred_tags": ["嵌入式", "焊接", "控制"],
    "time_text": "周末下午",
    "location_text": "西南大学"
  }'
```

### Activity list

Public. Defaults to `status=recruiting`. Supports `type`, `status`, and `keyword`.

```bash
curl -i 'http://127.0.0.1:8080/api/activities'
curl -i 'http://127.0.0.1:8080/api/activities?type=competition&keyword=STM32'
```

### Activity detail

```bash
curl -i "http://127.0.0.1:8080/api/activities/$ACTIVITY_ID"
```

### My activities

```bash
curl -i http://127.0.0.1:8080/api/me/activities \
  -H "Authorization: Bearer $TOKEN_A"
```

### Apply to activity

```bash
curl -i -X POST "http://127.0.0.1:8080/api/activities/$ACTIVITY_ID/apply" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"reason":"我有 STM32 基础，也做过焊接，希望加入这个队伍。"}'
```

### My applications

```bash
curl -i http://127.0.0.1:8080/api/me/applications \
  -H "Authorization: Bearer $TOKEN_B"
```

### Activity applications

Only the activity creator can view applications.

```bash
curl -i "http://127.0.0.1:8080/api/activities/$ACTIVITY_ID/applications" \
  -H "Authorization: Bearer $TOKEN_A"
```

### Approve application

```bash
curl -i -X POST "http://127.0.0.1:8080/api/applications/$APPLICATION_ID/approve" \
  -H "Authorization: Bearer $TOKEN_A"
```

### Reject application

```bash
curl -i -X POST "http://127.0.0.1:8080/api/applications/$APPLICATION_ID/reject" \
  -H "Authorization: Bearer $TOKEN_A"
```

## Complete activity/application smoke test

Replace `BASE` with the public server when testing deployment, for example `http://139.224.119.187`.

```bash
BASE=http://127.0.0.1:8080

curl -i -X POST "$BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"creator_a@example.com","password":"12345678","nickname":"用户A","school":"西南大学"}'

TOKEN_A=$(curl -s -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"creator_a@example.com","password":"12345678"}' | jq -r '.data.token')

ACTIVITY_ID=$(curl -s -X POST "$BASE/api/activities" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "智能硬件比赛组队",
    "type": "competition",
    "description": "准备参加电子设计竞赛，寻找会单片机和结构设计的队友。",
    "required_count": 3,
    "tags": ["电赛", "STM32", "硬件"],
    "preferred_tags": ["嵌入式", "焊接", "控制"],
    "time_text": "周末下午",
    "location_text": "西南大学"
  }' | jq -r '.data.activity.id')

curl -i -X POST "$BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"applicant_b@example.com","password":"12345678","nickname":"用户B","school":"西南大学"}'

TOKEN_B=$(curl -s -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"applicant_b@example.com","password":"12345678"}' | jq -r '.data.token')

curl -i "$BASE/api/activities"
curl -i "$BASE/api/activities/$ACTIVITY_ID"

APPLICATION_ID=$(curl -s -X POST "$BASE/api/activities/$ACTIVITY_ID/apply" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"reason":"我有 STM32 基础，也做过焊接，希望加入这个队伍。"}' | jq -r '.data.application.id')

curl -i "$BASE/api/me/applications" \
  -H "Authorization: Bearer $TOKEN_B"

curl -i "$BASE/api/activities/$ACTIVITY_ID/applications" \
  -H "Authorization: Bearer $TOKEN_A"

curl -i -X POST "$BASE/api/applications/$APPLICATION_ID/approve" \
  -H "Authorization: Bearer $TOKEN_A"
```

## 问卷画像与活动推荐

以下接口均需要 Bearer JWT。推荐模块使用可解释的规则评分，不调用大模型；业务范围保持为校园活动与项目协作。

### POST /api/questionnaires

提交问卷后会创建一条新 questionnaire，并按当前 JWT 用户的 `user_id` 创建或更新唯一 profile。`answers` 使用 JSONB 保存。

```bash
curl -i -X POST "$BASE/api/questionnaires" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{
    "mode": "activity",
    "answers": {
      "interests": ["电赛", "STM32", "硬件"],
      "skills": ["嵌入式", "焊接", "控制"],
      "available_time": "周末下午",
      "activity_types": ["competition", "project"],
      "goal": "找队友一起参加比赛",
      "communication_style": "稳定沟通"
    }
  }'
```

成功返回 `201 Created`：

```json
{
  "data": {
    "questionnaire": {
      "id": "...",
      "user_id": "...",
      "mode": "activity",
      "version": 1,
      "answers": {},
      "status": "completed"
    },
    "profile": {
      "user_id": "...",
      "profile_type": "activity",
      "tags": ["电赛", "STM32", "硬件", "嵌入式", "焊接", "控制", "competition", "project"],
      "scores": {
        "interest_score": 80,
        "skill_score": 75,
        "time_score": 70,
        "goal_score": 80,
        "communication_score": 75
      },
      "summary": "该用户偏向竞赛组队和项目协作，关注电赛、STM32、硬件，具备嵌入式、焊接相关兴趣，适合参与校园活动与项目协作。"
    }
  }
}
```

### GET /api/me/profile

```bash
curl -i "$BASE/api/me/profile" \
  -H "Authorization: Bearer $TOKEN_B"
```

未生成画像时返回 `404 profile_not_found`。

### POST /api/match/recommend

目前只支持 `target_type=activity`。`limit` 省略时默认为 10，允许范围为 1–50。候选活动必须为 `recruiting`，并自动排除当前用户自己创建的活动。

```bash
curl -i -X POST "$BASE/api/match/recommend" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"target_type":"activity","limit":10}'
```

每条结果包含活动、总分、五项明细和自然语言理由：

```json
{
  "activity": {
    "id": "...",
    "title": "智能硬件比赛组队",
    "type": "competition"
  },
  "score": 88,
  "detail_scores": {
    "interest": 26,
    "skill": 22,
    "type": 20,
    "time": 8,
    "goal": 12
  },
  "reason": "推荐原因：你的兴趣标签与该活动的‘电赛、STM32、硬件’高度重合，同时你的技能倾向与‘嵌入式、焊接’相符，适合参与该类校园竞赛或项目协作。"
}
```

尚未提交问卷时返回 `400 profile_required`。重复请求会按 `(user_id, activity_id, algorithm_version)` upsert，不会为同一算法版本创建重复记录。

### GET /api/me/matches

返回当前用户每个活动的最新已保存推荐结果，不保留每次推荐运行的快照。

```bash
curl -i "$BASE/api/me/matches" \
  -H "Authorization: Bearer $TOKEN_B"
```

## 用户 B 完整推荐测试

以下命令假设数据库中已有其他用户创建的 `recruiting` 活动。将 `BASE` 改为本地或公网地址。

```bash
BASE=http://127.0.0.1:8080

TOKEN_B=$(curl -s -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"applicant_b@example.com","password":"12345678"}' | jq -r '.data.token')

USER_B_ID=$(curl -s "$BASE/api/me" \
  -H "Authorization: Bearer $TOKEN_B" | jq -r '.data.user.id')

curl -s -X POST "$BASE/api/questionnaires" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{
    "mode":"activity",
    "answers":{
      "interests":["电赛","STM32","硬件"],
      "skills":["嵌入式","焊接","控制"],
      "available_time":"周末下午",
      "activity_types":["competition","project"],
      "goal":"找队友一起参加比赛",
      "communication_style":"稳定沟通"
    }
  }' | jq '.data | {questionnaire, profile}'

curl -s "$BASE/api/me/profile" \
  -H "Authorization: Bearer $TOKEN_B" | jq '.data.profile'

curl -s -X POST "$BASE/api/match/recommend" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H 'Content-Type: application/json' \
  -d '{"target_type":"activity","limit":10}' \
  | jq '.data.recommendations[] | {activity: .activity.title, score, detail_scores, reason}'

curl -s "$BASE/api/me/matches" \
  -H "Authorization: Bearer $TOKEN_B" \
  | jq '.data.matches[] | {activity: .activity.title, score, detail_scores, reason, updated_at}'
```

在服务器 PostgreSQL 中检查持久化结果：

```sql
SELECT user_id,
       activity_id,
       target_id,
       target_type,
       algorithm,
       algorithm_version,
       score,
       detail_scores,
       reason,
       updated_at
FROM matches
WHERE user_id = '<USER_B_ID>'
ORDER BY score DESC, updated_at DESC;
```

再次调用推荐接口后，以下查询中每组记录数仍应为 1，且 `updated_at` 会刷新：

```sql
SELECT user_id, activity_id, algorithm_version, COUNT(*)
FROM matches
WHERE user_id = '<USER_B_ID>'
GROUP BY user_id, activity_id, algorithm_version
ORDER BY activity_id;
```

## KM 最大权匹配测试

KM 和按行 greedy 对比实现位于 `backend/internal/algorithm`。运行：

```bash
cd backend
go test ./internal/algorithm -run TestKMOutperformsRowGreedy -v
```

测试矩阵为：

```text
A1: 9, 8, 1
A2: 8, 1, 1
A3: 1, 8, 8
```

期望结果：KM 总分 `24`，greedy 总分 `18`。

## 数据库增量升级

本次必须在发布新二进制前执行更新后的 `database/schema.sql`。脚本使用 `IF NOT EXISTS` 和兼容性回填，可重复执行；不需要删除数据库，也不需要重新导入 users、activities 或 applications 数据。

迁移会为 matches 增加 `target_id`、`target_type`、`algorithm`、`detail_scores`、`reason` 等独立字段。旧 `explanation` 和 `status` 字段继续保留；已有重复推荐行不会删除，而是转换为独立的 legacy 算法版本后再建立唯一索引。

## 服务器发布版本核验

服务器出现接口 404 时，先确认正在构建的目录确实包含本模块，而不是直接重复重启旧二进制：

```bash
cd ~/matchlab-web

git fetch origin
git checkout main
git pull --ff-only origin main
git rev-parse HEAD

find backend/internal/questionnaire \
     backend/internal/match \
     backend/internal/algorithm \
     -maxdepth 1 -type f -print | sort
```

文件列表必须包含 questionnaire 与 match 的 `model.go`、`repository.go`、`service.go`、`handler.go`，以及 algorithm 的 `km.go`、`km_test.go`。

先升级数据库，再按要求测试和构建：

```bash
sudo -u postgres psql \
  -d matchlab \
  -v ON_ERROR_STOP=1 \
  -f database/schema.sql

cd backend
go test ./...
mkdir -p bin
go build -o bin/matchlab-api cmd/server/main.go
ls -l --time-style=long-iso bin/matchlab-api
```

安装并重启服务：

```bash
cd ..
sudo sh deploy/deploy.sh
sudo systemctl restart matchlab-api
sudo systemctl status matchlab-api --no-pager
sudo journalctl -u matchlab-api -n 100 --no-pager

curl -i http://127.0.0.1:8080/api/health
curl -i http://139.224.119.187/api/health
```

## 管理员后台 API

以下接口都需要 `admin` 角色。先在 PostgreSQL 中将测试账号设为管理员：

```sql
UPDATE users SET role = 'admin' WHERE email = 'test@example.com';
```

角色更新后需要重新登录，以便新签发的 JWT 包含 `admin` 角色：

```bash
BASE="http://139.224.119.187"

TOKEN_A=$(curl -s -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"YOUR_PASSWORD"}' \
  | jq -r '.data.token')
```

获取统计数据：

```bash
curl -s "$BASE/api/admin/stats" \
  -H "Authorization: Bearer $TOKEN_A" | jq
```

获取用户、活动、报名和反馈列表：

```bash
curl -s "$BASE/api/admin/users?keyword=test&role=admin&limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN_A" | jq

curl -s "$BASE/api/admin/activities?keyword=电赛&type=competition&status=recruiting&limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN_A" | jq

curl -s "$BASE/api/admin/applications?status=approved&limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN_A" | jq

curl -s "$BASE/api/admin/feedbacks?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN_A" | jq
```

修改用户角色（`USER_ID` 替换为目标用户 UUID）：

```bash
curl -s -X POST "$BASE/api/admin/users/$USER_ID/role" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H 'Content-Type: application/json' \
  -d '{"role":"admin"}' | jq
```

普通用户访问管理员接口应返回 `403`：

```bash
TOKEN_B=$(curl -s -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"userb@example.com","password":"YOUR_PASSWORD"}' \
  | jq -r '.data.token')

curl -i "$BASE/api/admin/stats" \
  -H "Authorization: Bearer $TOKEN_B"
```

管理员模块复用现有表结构，本组任务无需修改或重新执行 `database/schema.sql`。

最后使用本文“用户 B 完整推荐测试”中的四个接口命令验证。只要有效 JWT 请求不再返回 404，就说明新路由已经进入运行中的二进制；随后再根据 2xx、4xx 或 5xx 响应检查数据库和请求数据。
