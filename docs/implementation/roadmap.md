# 实施路线图

## Phase 0：工程底座

- [x] Monorepo 目录
- [x] Go API 与健康检查
- [x] Next.js Web
- [x] MySQL / Redis
- [x] 数据库迁移
- [x] 请求追踪
- [x] Docker Compose
- [x] CI
- [x] Manifest v1 与订单状态机草案

## Phase 1：身份、组织和安全基础

- [ ] 注册、登录、邮箱验证
- [ ] Access/Refresh Token
- [ ] RBAC
- [ ] API Key 哈希存储与 Scope
- [ ] Redis 限流
- [ ] Idempotency 中间件
- [ ] 管理审计日志

## Phase 2：Data Agent 与数据版本

- [ ] 跨平台目录扫描
- [ ] 分块哈希与 Merkle Root
- [ ] Manifest 序列化与签名
- [ ] 扫描断点恢复
- [ ] Agent 注册、心跳与在线状态

## Phase 3：市场、订单和佣金账本

- [ ] 数据集和不可变版本
- [ ] 审核与发布
- [ ] 搜索和筛选
- [ ] 订单状态机
- [ ] 佣金冻结、扣除和释放
- [ ] 支付 Provider 与回调幂等

## Phase 4：受控 P2P 交付

- [ ] 私有 Torrent
- [ ] Private Tracker
- [ ] Delivery Grant
- [ ] 设备公钥绑定
- [ ] 下载 Session、次数和有效期
