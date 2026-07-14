# BlctekIP IP-Chain

AI 训练数据合规确权与交易平台。平台采用“数据留在提供者本地”的设计，通过本地 Data Agent 生成可复现的数据清单与 Merkle Root，平台负责身份、登记、审核、存证、交易撮合和受控 P2P 交付。

## 当前状态

这是从零重构的工程基线，当前包含：

- Go 1.22 + Gin API 服务
- Next.js 14 Web 应用
- MySQL 8.0、Redis 7
- 数据库迁移命令
- 统一健康检查与请求追踪
- Docker Compose 开发环境
- GitHub Actions 基础 CI
- Manifest v1、订单状态机和架构决策文档

## 快速开始

1. 复制环境变量：

```bash
cp .env.example .env
```

2. 启动依赖与应用：

```bash
docker compose up --build
```

3. 访问：

- Web: http://localhost:3000
- API: http://localhost:8080
- Health: http://localhost:8080/api/v1/health

## 本地开发

### API

```bash
cd apps/api
go mod download
go run ./cmd/migrate up
go run ./cmd/server
```

### Web

```bash
cd apps/web
npm install
npm run dev
```

## 目录

```text
apps/api            Go API 与异步任务基础
apps/web            Next.js 前端
packages/openapi    API 契约
docs                产品、架构、ADR、协议与状态机
deployments         容器与部署配置
```

## 开发原则

- 订单必须绑定不可变的数据版本、价格和许可快照。
- 金额统一使用整数分，不使用浮点数。
- 外部副作用通过 Outbox 可靠投递。
- AI 搜索只能生成受限 Search Intent，禁止直接执行模型生成的 SQL。
- P2P 交付使用私有 Tracker 和订单级授权，不公开永久 Magnet。
