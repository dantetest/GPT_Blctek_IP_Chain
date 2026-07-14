# 系统架构概览

MVP 采用模块化单体。身份、KYC、订阅、数据集、订单、账务、交付、存证、分析和后台管理在同一 Go API 中以模块边界组织。

独立运行组件：Data Agent、Quality/Dedup、Private Tracker、MCP Server。

```text
Next.js Web
    |
Go API ---- MySQL
    |          |
    |          +-- Outbox / Ledger / Domain Data
    +------ Redis
    |
Async Worker

Data Agent <----> Private Tracker <----> Buyer Client
```

## 不变量

1. 原始数据不经过平台业务服务器。
2. 订单只购买具体的数据版本。
3. 金额以整数分记录。
4. 账务流水只追加。
5. 外部副作用必须可重试且幂等。
6. P2P 授权绑定订单、设备和有效期。
