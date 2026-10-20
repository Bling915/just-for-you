# 小事成双对后端全局四层 DDD 重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变现有 API、数据库结构和业务行为的前提下，把 Go 后端迁移为严格的全局四层 DDD 结构，并将综合应用服务拆成独立用例。

**Architecture:** 代码按 `interfaces/http -> application -> domain` 单向依赖，基础设施实现应用层或领域层端口，`bootstrap` 负责唯一的依赖装配入口。三个业务域仍然是同一个 Go 服务内的包边界，不拆微服务。

**Tech Stack:** Go 1.23、Gin、`database/sql`、MySQL 8、JWT、微信登录 HTTP API、腾讯云 COS、`go test`、`go vet`。

**Spec:** `docs/superpowers/specs/2026-10-20-global-ddd-backend-design.md`

## Global Constraints

- API 路径、HTTP 状态码、JSON 字段、稳定错误码和 MySQL 表结构必须保持不变。
- `domain` 只能依赖 Go 标准库；`application` 不得依赖 Gin、SQL 或具体基础设施。
- 不引入 ORM、依赖注入框架、CQRS 框架、消息总线或领域事件。
- 一个用户动作对应一个 `<Action>UseCase` 文件和 `Execute` 方法，不保留综合 `Service`。
- `bootstrap` 是唯一允许同时导入 interfaces、application 和 infrastructure 的包。
- `cmd/api/main.go` 只负责创建应用、处理信号和启动/关闭，目标约 30 行。
- 单人测试模式不在本计划范围内。
- 真实密钥、微信 code、访问令牌、刷新令牌原文和数据库连接信息不得写入日志、测试快照或 Git。

## Review Focus

- 旧包被新层意外反向导入：Task 1 的架构测试必须固定所有允许的依赖方向。
- 重构后登录、绑定、发布等错误码漂移：Task 2–4 的 handler 契约测试必须复用现有断言。
- refresh token 轮换因拆分用例而丢失原子语义：Task 2 测试必须覆盖旧 token 失效和新 token 生效。
- relationship 到 moment 的跨域依赖变宽：Task 4 必须只消费 `ActiveRelationshipReader` 窄接口。
- 九图、上传授权、照片顺序或删除权限回归：Task 4 必须保留现有领域和应用测试并新增 wiring 验证。

---

## 目标文件职责

```text
server/internal/
├── interfaces/http/{identity,relationship,moment,shared}
├── application/{identity,relationship,moment}
├── domain/{identity,relationship,moment}
├── infrastructure/
│   ├── persistence/mysql/{identity,relationship,moment,shared}
│   ├── identity/wechat
│   ├── security/jwt
│   ├── storage/cos
│   ├── config
│   └── clock
├── bootstrap
└── architecture
```

- `domain/*`：实体、值对象、领域规则和领域错误。
- `application/*`：独立用例、Command/Query/Result 和端口。
- `interfaces/http/*`：Gin handler、DTO、路由注册和错误映射。
- `infrastructure/*`：MySQL、微信、JWT、COS、配置与系统时钟实现。
- `bootstrap/*`：应用生命周期和三个业务域的依赖装配。

### Task 1: 建立全局四层骨架和强制依赖测试

**Files:**
- Modify: `server/internal/architecture/dependencies_test.go`
- Create: `server/internal/infrastructure/clock/system.go`
- Move: `server/internal/platform/config/config.go` → `server/internal/infrastructure/config/config.go`
- Move: `server/internal/platform/config/config_test.go` → `server/internal/infrastructure/config/config_test.go`
- Move: `server/internal/platform/database/mysql.go` → `server/internal/infrastructure/persistence/mysql/shared/database.go`
- Move: `server/internal/platform/httpserver/{auth.go,response.go,router.go,router_test.go}` → `server/internal/interfaces/http/shared/`
- Modify: imports throughout `server/` only as required to keep compilation green

**Interfaces:**
- Produces: `clock.Clock { Now() time.Time }` and `clock.System`.
- Produces: `config.Load() (config.Config, error)`.
- Produces: `shared.OpenMySQL(dsn string) (*sql.DB, error)` and existing HTTP response/router APIs.
- Produces: architecture assertions for all four global layers.

- [ ] **Step 1: Write the failing architecture tests**

Add `TestGlobalLayerDependencies` cases proving that domain imports only standard library, application imports only domain/standard library, HTTP interfaces do not import infrastructure, and only bootstrap/cmd may compose all layers. Add a temporary violating fixture in the test and remove it after confirming the assertion catches it.

- [ ] **Step 2: Run the RED check**

Run: `cd server && go test ./internal/architecture`

Expected: FAIL because current domain-first paths and platform imports do not satisfy the new global layer rules.

- [ ] **Step 3: Move shared adapters and update imports**

Use `git mv`; preserve exported behavior. Keep temporary compatibility only within this task, then ensure all packages compile with the new shared paths.

- [ ] **Step 4: Run focused and full tests**

Run: `cd server && go test ./internal/architecture ./internal/infrastructure/... ./internal/interfaces/http/shared/... && go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server
git commit -m "refactor: establish global DDD layers"
```

### Task 2: 迁移 Identity 并拆分五个用例

**Files:**
- Move: `server/internal/identity/domain/*` → `server/internal/domain/identity/`
- Create: `server/internal/application/identity/{ports.go,session.go,login.go,refresh_session.go,logout.go,get_profile.go,update_profile.go}`
- Create: `server/internal/application/identity/use_cases_test.go`
- Move: `server/internal/identity/infrastructure/mysql_repository.go` → `server/internal/infrastructure/persistence/mysql/identity/repository.go`
- Move: `server/internal/identity/infrastructure/wechat.go` → `server/internal/infrastructure/identity/wechat/code_exchanger.go`
- Move: `server/internal/identity/infrastructure/jwt.go` → `server/internal/infrastructure/security/jwt/token_service.go`
- Move and split: `server/internal/identity/interfaces/http/*` → `server/internal/interfaces/http/identity/{handler.go,dto.go,error_mapper.go,handler_test.go}`
- Delete: `server/internal/identity/`

**Interfaces:**
- Produces: `LoginUseCase.Execute(context.Context, LoginCommand) (Session, error)`.
- Produces: `RefreshSessionUseCase.Execute(context.Context, RefreshSessionCommand) (Session, error)`.
- Produces: `LogoutUseCase.Execute(context.Context, LogoutCommand) error`.
- Produces: `GetProfileUseCase.Execute(context.Context, GetProfileQuery) (domain.User, error)`.
- Produces: `UpdateProfileUseCase.Execute(context.Context, UpdateProfileCommand) (domain.User, error)`.
- Produces: `identity.UseCases` bundle containing only these five concrete use cases for bootstrap-to-handler wiring.

- [ ] **Step 1: Write use-case contract tests before moving implementation**

Test names and assertions: `TestLoginExecuteReusesExistingUserAndIssuesSession`, `TestRefreshSessionExecuteRotatesRefreshToken`, `TestLogoutExecuteRevokesRefreshToken`, `TestGetProfileExecuteReturnsUser`, `TestUpdateProfileExecuteValidatesAndPersistsUser`. Preserve the current 15-minute access and 30-day refresh lifetimes.

- [ ] **Step 2: Run RED**

Run: `cd server && go test ./internal/application/identity`

Expected: FAIL because use cases do not exist.

- [ ] **Step 3: Move the domain and implement minimal independent use cases**

Extract shared session creation into unexported application helper code without recreating a generic Service. Keep repository and token interfaces in `ports.go`; each use case stores only the ports it needs.

- [ ] **Step 4: Move Identity adapters and HTTP handler**

Keep `/v1/auth/wechat`, `/v1/auth/refresh`, `/v1/auth/logout`, `GET /v1/users/me`, and `PUT /v1/users/me` unchanged. Handler fields depend on five narrow executor interfaces or the explicit `UseCases` bundle; do not expose adapter types.

- [ ] **Step 5: Verify Identity behavior and old-path removal**

Run: `cd server && go test ./internal/domain/identity ./internal/application/identity ./internal/infrastructure/persistence/mysql/identity ./internal/interfaces/http/identity && go test ./... && ! rg 'internal/identity/' .`

Expected: PASS and no old Identity import path.

- [ ] **Step 6: Commit**

```bash
git add server
git commit -m "refactor: split identity into global DDD layers"
```

### Task 3: 迁移 Relationship 并拆分五个用例

**Files:**
- Move: `server/internal/relationship/domain/*` → `server/internal/domain/relationship/`
- Create: `server/internal/application/relationship/{ports.go,create_invitation.go,preview_invitation.go,bind_relationship.go,get_relationship.go,end_relationship.go}`
- Create: `server/internal/application/relationship/use_cases_test.go`
- Move: `server/internal/relationship/infrastructure/*` → `server/internal/infrastructure/persistence/mysql/relationship/`
- Move and split: `server/internal/relationship/interfaces/http/*` → `server/internal/interfaces/http/relationship/{handler.go,dto.go,error_mapper.go,handler_test.go}`
- Delete: `server/internal/relationship/`

**Interfaces:**
- Produces: `CreateInvitationUseCase.Execute(context.Context, CreateInvitationCommand) (domain.Invitation, error)`.
- Produces: `PreviewInvitationUseCase.Execute(context.Context, PreviewInvitationQuery) (domain.Invitation, error)`.
- Produces: `BindRelationshipUseCase.Execute(context.Context, BindRelationshipCommand) (domain.Relationship, error)`.
- Produces: `GetRelationshipUseCase.Execute(context.Context, GetRelationshipQuery) (domain.Relationship, error)`.
- Produces: `EndRelationshipUseCase.Execute(context.Context, EndRelationshipCommand) error`.
- Produces: `ActiveRelationshipReader.GetActiveRelationship(context.Context, int64) (ActiveRelationship, error)` for Moment only.

- [ ] **Step 1: Write failing independent use-case tests**

Cover code normalization, ten-minute invitation expiry, self-binding rejection, existing relationship rejection, transactional acceptance, reading active relationship, and ending at the injected clock time.

- [ ] **Step 2: Run RED**

Run: `cd server && go test ./internal/application/relationship`

Expected: FAIL because use cases do not exist.

- [ ] **Step 3: Move domain and implement use cases**

Each use case holds only `Repository`, `Clock`, or code generator dependencies needed by its flow. Keep transaction semantics inside the MySQL repository.

- [ ] **Step 4: Move MySQL and HTTP adapters**

Preserve `/v1/couples/invitations`, `/v1/couples/invitations/:code`, `/v1/couples/bind`, `/v1/couples/me` and all existing relationship error codes.

- [ ] **Step 5: Verify Relationship behavior and old-path removal**

Run: `cd server && go test ./internal/domain/relationship ./internal/application/relationship ./internal/infrastructure/persistence/mysql/relationship ./internal/interfaces/http/relationship && go test ./... && ! rg 'internal/relationship/' .`

Expected: PASS and no old Relationship import path.

- [ ] **Step 6: Commit**

```bash
git add server
git commit -m "refactor: split relationship into global DDD layers"
```

### Task 4: 迁移 Moment 并拆分五个用例

**Files:**
- Move: `server/internal/moment/domain/*` → `server/internal/domain/moment/`
- Create: `server/internal/application/moment/{ports.go,request_photo_upload.go,publish_moment.go,get_timeline.go,get_home.go,delete_moment.go,photo_urls.go}`
- Create: `server/internal/application/moment/use_cases_test.go`
- Move: `server/internal/moment/infrastructure/mysql_repository.go` → `server/internal/infrastructure/persistence/mysql/moment/repository.go`
- Move: `server/internal/moment/infrastructure/relationship_reader.go` → `server/internal/infrastructure/persistence/mysql/moment/relationship_reader.go`
- Move: `server/internal/moment/infrastructure/cos_storage.go` → `server/internal/infrastructure/storage/cos/photo_storage.go`
- Move and split: `server/internal/moment/interfaces/http/*` → `server/internal/interfaces/http/moment/{handler.go,dto.go,error_mapper.go,handler_test.go}`
- Delete: `server/internal/moment/`

**Interfaces:**
- Produces: `RequestPhotoUploadUseCase.Execute(context.Context, RequestPhotoUploadCommand) (UploadGrant, error)`.
- Produces: `PublishMomentUseCase.Execute(context.Context, PublishMomentCommand) (domain.Moment, error)`.
- Produces: `GetTimelineUseCase.Execute(context.Context, GetTimelineQuery) (TimelinePage, error)`.
- Produces: `GetHomeUseCase.Execute(context.Context, GetHomeQuery) (HomeResult, error)`.
- Produces: `DeleteMomentUseCase.Execute(context.Context, DeleteMomentCommand) error`.

- [ ] **Step 1: Write failing independent use-case tests**

Cover allowed content types and 10 MiB limit, no active relationship, 0/1/9/10 photos, upload ownership and expiry, photo verification, stable photo order, default/max timeline limits, signed read URLs, missing/non-author deletion, and deletion after relationship ends.

- [ ] **Step 2: Run RED**

Run: `cd server && go test ./internal/application/moment`

Expected: FAIL because use cases do not exist.

- [ ] **Step 3: Move domain and implement use cases**

Share only the unexported `signReadURLs` helper between home and timeline. `PublishMomentUseCase` consumes the narrow `ActiveRelationshipReader`; it must not depend on relationship handlers or all relationship use cases.

- [ ] **Step 4: Move MySQL/COS adapters and HTTP handler**

Preserve `POST /v1/uploads`, `POST/GET /v1/moments`, `GET /v1/home`, `DELETE /v1/moments/:id`, all DTO fields, upload rules and error codes.

- [ ] **Step 5: Verify Moment behavior and old-path removal**

Run: `cd server && go test ./internal/domain/moment ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/infrastructure/storage/cos ./internal/interfaces/http/moment && go test ./... && ! rg 'internal/moment/' .`

Expected: PASS and no old Moment import path.

- [ ] **Step 6: Commit**

```bash
git add server
git commit -m "refactor: split moment into global DDD layers"
```

### Task 5: 建立 Bootstrap、精简入口并完成文档

**Files:**
- Create: `server/internal/bootstrap/{application.go,identity.go,relationship.go,moment.go}`
- Modify: `server/cmd/api/main.go`
- Modify: `server/cmd/migrate/main.go`
- Create: `server/internal/bootstrap/application_test.go`
- Create: `docs/backend-architecture.md`
- Modify: `README.md`
- Modify: `Makefile`

**Interfaces:**
- Produces: `bootstrap.New() (*Application, error)`.
- Produces: `(*Application).Run(context.Context) error` and `(*Application).Close() error`.
- Consumes: all repositories, external adapters, use cases, handlers and shared router created in Tasks 1–4.

- [ ] **Step 1: Write failing bootstrap wiring tests**

Test `TestApplicationBuildsAllRoutes` with injected config/adapters or a test constructor, asserting health, auth, profile, relationship, upload, moment and home routes are registered. Test `TestCloseClosesDatabaseOnce` for lifecycle safety.

- [ ] **Step 2: Run RED**

Run: `cd server && go test ./internal/bootstrap`

Expected: FAIL because bootstrap does not exist.

- [ ] **Step 3: Implement per-domain composition and application lifecycle**

Keep `identity.go`, `relationship.go`, and `moment.go` responsible for constructing their repositories, adapters, use cases and handlers. `application.go` owns shared resources, Router, HTTP server and cleanup.

- [ ] **Step 4: Replace API entry point and update migration imports**

`main.go` creates a signal-aware context, calls `bootstrap.New`, defers `Close`, and calls `Run`. It must not import application, repository, JWT,微信 or COS packages directly.

- [ ] **Step 5: Write maintenance documentation**

Explain the four layers, dependency diagram, request reading path, exact placement for a new endpoint/use case/rule/query, local commands, and why the project remains a modular monolith rather than microservices.

- [ ] **Step 6: Run complete verification**

Run: `cd server && gofmt -w . && go test ./... && go test -race ./internal/domain/... ./internal/application/... && go vet ./...`

Run: `rg -n 'internal/(identity|relationship|moment|platform)/' server docs README.md || true`

Run: `wc -l server/cmd/api/main.go`

Expected: tests and vet PASS; no production import references old paths; `main.go` is approximately 30 lines.

- [ ] **Step 7: Commit**

```bash
git add Makefile README.md docs server
git commit -m "refactor: compose backend through bootstrap"
```

### Task 6: 整体审查、兼容性验收与分支交付

**Files:**
- Modify only files required by review findings
- Review: all changes from `63f2346` through HEAD

**Interfaces:**
- Consumes: the completed global-layer backend.
- Produces: a clean branch with evidence that behavior and architecture are preserved.

- [ ] **Step 1: Review the complete branch against the spec**

Check for dependency violations, leaked adapter types, generic services, duplicated DTO mapping, dead compatibility code, secrets, missing errors and mismatched names.

- [ ] **Step 2: Fix review findings with targeted regression tests**

Every behavior fix begins with a failing focused test; formatting-only or documentation findings need no new test.

- [ ] **Step 3: Run repository-wide acceptance**

Run: `make test-backend && make test-race && make vet`

Run: `git diff --check 63f2346..HEAD && git status --short`

Expected: all commands exit 0 and worktree is clean.

- [ ] **Step 4: Commit any review fixes**

```bash
git add server docs Makefile README.md
git commit -m "fix: address global DDD refactor review"
```

- [ ] **Step 5: Report the new reading path**

Provide clickable links to one representative handler, use case, domain entity, repository and bootstrap file, plus verification results and deployment impact. Do not deploy automatically; deployment remains a separate explicit action.
