# Phase 1B：API Key 与请求幂等

## API Key

- 密钥格式：`bkip_<12位前缀>_<随机密钥>`。
- 数据库只保存 SHA-256 哈希，完整密钥只在创建响应中返回一次。
- 基础版不可创建 API Key。
- 专业版最多 5 个启用密钥，每月每个密钥 10,000 次调用。
- 企业版密钥数量和月调用次数不设平台上限。
- 支持有效期、撤销、最后使用时间、月额度自动重置。
- 创建数量通过锁定用户记录串行化，避免并发绕过套餐上限。

### Scope

- `datasets:read`
- `orders:read`
- `orders:write`
- `assets:read`
- `assistant:query`

## 幂等

需要幂等保护的写接口必须提交 `Idempotency-Key`：

1. 服务端按用户、HTTP 方法和路由建立作用域。
2. 请求体、用户、方法和路由共同生成 SHA-256 请求摘要。
3. 相同 Key 和相同请求返回首次响应，并添加 `Idempotency-Replayed: true`。
4. 相同 Key 用于不同请求返回 `409 IDEMPOTENCY_CONFLICT`。
5. 并发中的重复请求返回 `409 IDEMPOTENCY_IN_PROGRESS`。
6. 5xx 响应不缓存，允许客户端重试。
7. 请求体上限为 2 MiB，记录默认保留 24 小时。
