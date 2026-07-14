# Phase 1：身份认证与安全基线

## 已实现

- Argon2id 密码哈希和密码复杂度规则
- 邮箱标准化与唯一性约束
- 用户注册与可配置的开发环境自动邮箱验证
- 邮箱验证 Token 哈希存储和 Outbox 邮件事件
- 登录失败计数、临时锁定和自动解锁
- JWT Access Token（HS256、issuer、audience、session ID）
- 旋转式 Refresh Token，数据库仅保存 SHA-256 哈希
- 登出撤销 Refresh Session
- `/me` 认证接口
- 全局角色 `USER` / `ADMIN` 与 RBAC 中间件
- Redis 固定窗口限流，Redis 故障时 fail-open

## 安全不变量

1. 密码明文、Refresh Token 和邮箱验证 Token 永不写入数据库。
2. Access Token 必须包含用户、会话、角色和套餐快照。
3. Refresh Token 每次使用后立即轮换，旧 Token 不能再次使用。
4. 登录错误响应不区分“邮箱不存在”和“密码错误”。
5. 生产环境 JWT Secret 不足 32 字符时服务拒绝启动。
6. 邮箱发送使用 Outbox，注册事务和通知事件原子提交。

## 下一步

- API Key 创建、列表、撤销和 Scope 鉴权
- 通用 Idempotency-Key 中间件
- Redis 分套餐 API 限流
- 邮件 Worker 和验证码重发
- 密码重置流程
