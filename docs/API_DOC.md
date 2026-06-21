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
