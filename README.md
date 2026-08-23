# 代表建议办理与督办平台

面向代表、联名人、承办单位和人大督办人员的离线可运行履职系统。建议从草稿、联名、提交快照进入分办和签收，经过办理计划、材料、答复与评价；不满意评价在同一事务内保留原答复并创建重办轮次和督办通知。统一工作日策略驱动期限、红灯和每周催办，档案与排名保留统计周期、分子和分母。

## 架构与目录

- `cmd/server`：HTTP 与后台 outbox worker 生命周期、优雅停机。
- `cmd/migrate`：版本化迁移与幂等种子入口。
- `internal/domain`：身份、建议、分办、办理、答复、评价、督办与档案聚合及状态机。
- `internal/application`：提交、分办、办理、答复、评价、督办、排名用例和调用方端口。
- `internal/repository/postgres`：pgx 持久化、事务、乐观并发与 outbox claim。
- `internal/repository/memory`：测试和确定性离线演示使用的可回滚本地适配器。
- `internal/transport/http`、`internal/middleware`：Gin API、RBAC、资源身份、稳定错误、限流与安全头。
- `internal/platform`：时钟、ID、幂等、文件、通知、指标和可取消 outbox worker。
- `migrations`、`api/openapi`：数据库约束与 OpenAPI 3.0 契约。
- `web`：Vue 3、TypeScript、Vite、Pinia 与 Element Plus 工作台。
- `tests/integration`：PostgreSQL 与事务集成测试。

## 本地启动

需要 Go 1.24+、Node.js 22+ 和 PostgreSQL 17。所有依赖均从本地安装，不使用 CDN 或云服务。

```bash
cp .env.example .env
docker compose up -d postgres
go run ./cmd/migrate up
go run ./cmd/migrate seed
go run ./cmd/server
cd web && npm ci && npm run desk:dev
```

服务提供 `/healthz`、`/readyz` 和 `/api/v1`；前端默认通过 Vite 代理访问 `localhost:8080`。文件内容保存在 `FILE_STORAGE_ROOT`，下载必须经授权 API，响应不会暴露物理路径。

## 容器启动

```bash
docker compose up --build
```

镜像为多阶段 amd64/arm64 兼容构建，运行阶段使用 UID 10001 的非 root 用户、只读根文件系统和独立文件卷。应用启动前执行 `go run ./cmd/migrate up && go run ./cmd/migrate seed`，或在部署流水线中执行同一命令。

## 配置

关键变量见 `.env.example`。时间在数据库中统一为 UTC，界面按 `APP_TIMEZONE` 展示。`JWT_SIGNING_KEY` 至少 32 字符；访问令牌有效期 15 分钟，刷新令牌只保存 SHA-256 摘要且可按版本撤销。日志不得写入口令、令牌、证件原文、签名正文或附件内容。

演示账号为 `representative`、`unit_officer`、`supervisor`，本地种子统一使用仅供演示的密码哈希。生产环境必须替换密码、签名密钥和数据库凭据。

## 核心状态机

`draft -> submitted -> assigned -> accepted -> handling -> answered -> evaluated -> archived`

提交后 `proposal_snapshots` 只增不改；正文和联名名单不再编辑，后续内容进入补充材料。分办后承办单位必须先签收才能建计划或答复。四类答复按轮次唯一保存，评价不覆盖答复。不满意评价以事务写入评价、重办轮次、督办 case、outbox 事件和审计记录，任一写入失败则全部回滚。

## 接口示例

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/proposals?page=1&page_size=20&sort=submitted_at"

curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reply_id":"reply-1","rating":"unsatisfied","comment":"关键问题尚未落实"}' \
  http://localhost:8080/api/v1/proposals/proposal-1/evaluations
```

错误响应固定包含 `code`、`message`、`field_errors` 和 `request_id`。列表支持页码、游标、排序与状态、分类、承办单位、所有者白名单筛选。

## 验证

```bash
gofmt -l .
go build ./...
go test ./...
go test -race ./...
go vet ./...
cd web && npm run test:unit -- --run
cd web && npm run check:types
cd web && npm run desk:bundle
```

项目生成时执行上述命令；PostgreSQL 集成测试在设置 `TEST_DATABASE_URL` 时运行，否则明确跳过。Docker 验证使用 Compose 中的 PostgreSQL 17 与当前平台镜像。

## 限制

本项目不接入真实短信、推送或云文件服务；通知使用确定性本地适配器，生产可在不改应用用例的前提下替换端口实现。节假日数据由部署方配置，默认仅按周末和显式调休表计算。金额字段若后续扩展，必须继续使用最小货币单位或定点数。
