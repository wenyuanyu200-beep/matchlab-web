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

## 计划中的认证接口

以下接口尚未实现，仅作为下一阶段边界：

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/users/me`

实现前将补充请求字段、状态码、错误码和 JWT 约定。
