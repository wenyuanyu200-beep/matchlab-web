# 圈子与聊天 MVP 回归记录

## 权限与安全验收

执行前准备访客浏览器会话、普通用户 A/B/C 和 admin；使用全新测试数据，记录 circle、channel、conversation ID 与每步 HTTP 状态。生产环境仅执行经授权的 smoke，以下写操作优先在测试环境完成。

1. 访客打开圈子广场和 approved 圈子详情；确认只看到 approved，成员摘要无 email，加入/发送引导登录。
2. A 创建圈子，输入前后空格；确认响应为 pending，名称被 trim，`/api/me/circles` 可见，访客/B 请求详情为 404。
3. Admin 列出 pending 圈子；无 token 为 401，B 为 403，admin 可见且响应成员资料无 email。
4. Admin 拒绝该圈子；A 可见 rejected 与原因，访客/B 仍为 404，广场不出现。
5. A 再创建圈子并由 admin 通过；确认广场与详情出现，默认 `general` 频道存在，A 为 owner 且计数为 1。
6. B 加入 approved 圈子；确认重复加入返回稳定冲突且成员数不重复增长。C 不加入。
7. C 分别读取 channels、读取/发送频道消息，确认 403；B 和 A 可读写，admin 若不是成员仅可读不可写。
8. A/B 分别发送空白、1001 字符和前后空格消息；确认前两者为 4xx，合法内容被 trim 且响应不含 email。
9. B 从成员入口发起与 A 的 direct conversation；重复请求复用同一 ID，自聊请求为 4xx。
10. C 使用已知 conversation ID 读取、发送、标记已读，均为 403/404；A/B 可读写，admin 不能旁路访问。
11. A 发消息后核对 B 的 conversation 未读数和总未读数增加；B 进入详情并 read 后归零，A 的读状态不受影响。
12. 在桌面和窄屏各走一遍：双栏/三 Tab 无横向溢出；隐藏页面超过 3 秒无轮询，恢复立即拉取；发送失败保留草稿；最后复跑旧活动、报名、推荐与 admin 页面回归。

## 自动化证据

| 阶段 | 命令 | 结果 |
| --- | --- | --- |
| 基线 | `git status --short --branch` | `feature/circles-and-chat`；开始时已有 `backend/internal/router/router_test.go` 修改 |
| 基线 | `cd backend && go test ./...` | 未运行：本机无 `go`/`go.exe`，三种定位方式均无结果 |
| 基线 | `cd frontend && npm run test` | PASS：14 files、24 tests |
| 基线 | `cd frontend && npm run lint` | PASS |
| 基线 | `cd frontend && npm run build` | PASS；当时构建为功能合入前旧路由集 |
| 基线 | `bash -n scripts/production_e2e_recommend.sh` | FAIL exit 2：CRLF 导致第 8 行 `require_tool() {\r` 语法错误 |

`production_e2e_recommend.sh` 会写入用户、问卷、活动、报名、审批和推荐数据，本次没有执行远程写操作。

## 风险表

| 风险 | 严重级 | 检测/缓解 |
| --- | --- | --- |
| migration 与后端版本不同步 | 高 | 备份；migration 先行并重复执行；发布后检查七表和 health/smoke |
| pending/rejected 或 email 泄露 | 高 | 访客/非 creator 统一 404；逐个响应搜索 `email`；权限自动化覆盖 |
| 非成员读取频道或越权私聊 | 高 | 使用已知 UUID 做负向 GET/POST/read；admin 不拥有私聊旁路 |
| 未读数或成员数并发漂移 | 中 | 唯一约束与事务；并发重复请求；用明细聚合值对账 |
| 3 秒轮询造成数据库负载 | 中 | 增量游标和索引；后台暂停；监控 QPS、慢查询、连接数 |
| `(created_at,id)` 游标边界漏消息/重复 | 中 | 同时间戳多消息测试；前端 id 去重并稳定排序 |
| CRLF 导致生产 E2E 无法启动 | 中 | 转 LF 后 `bash -n`；未修复前禁止把脚本当发布门禁 |
| QA 主机缺 Go，后端未获独立验证 | 高 | 在 CI 或装有 Go 1.22+ 的发布机运行 `go test ./...` 和 build |
| 应用回滚误删新表 | 高 | 只回滚应用，明确保留新表；数据库恢复必须单独审批 |

## 最终结论

待 Agent 1/2 合入完成后补录最终全量测试、实现审查缺陷与人工回归结果。
