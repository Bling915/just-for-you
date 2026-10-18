# Just for You 后端 DDD 重构实施计划

> **供智能代理执行：** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`，逐项实施本计划。步骤使用复选框跟踪。

**目标：** 将现有 Go 后端迁移为按业务域组织的实用型 DDD 架构，并在新架构中完成九图记录、首页和共同时间线 API。

**架构：** 服务端按 `identity`、`relationship`、`moment` 三个业务域组织，每个域分为 `domain`、`application`、`infrastructure` 和 `interfaces/http`。公共配置、数据库、HTTP 响应和时钟位于 `platform`；领域层不依赖 Gin、MySQL、JWT、微信或 COS。

**技术栈：** Go 1.22+、Gin、`database/sql`、MySQL 8.0、JWT、腾讯云 COS 私有桶。

**规格文档：** `docs/superpowers/specs/2026-10-18-just-for-you-mvp-design.md`

## 全局约束

- 依赖方向必须为 `interfaces → application → domain`；基础设施只能实现领域层或应用层定义的端口。
- `domain` 包不得导入 Gin、`database/sql`、JWT、微信 SDK 或 COS SDK。
- 不引入领域事件、CQRS、Unit of Work 框架、依赖注入框架或 ORM。
- API 路径和现有 JSON 契约保持不变，避免 DDD 重构迫使小程序同步返工。
- 单条记录至少包含文字、状态、地点或照片之一；文字最多 500 个 Unicode 字符，地点最多 80 个 Unicode 字符，照片最多九张。
- 权限从登录用户的有效关系推导；不能信任客户端关系 ID；用户只能删除自己的记录。
- 真实密钥、微信登录 `code`、访问令牌、刷新令牌原文和微信会话密钥不得写入日志或 Git。

## 评审重点

- 领域包意外导入基础设施依赖：架构测试必须失败并指出违规 import。
- 重构前后 API 错误码漂移：HTTP 契约测试必须固定登录、邀请码和关系错误码。
- 两人并发接受同一邀请码：MySQL 事务只能产生一段有效关系。
- 第十张照片或不属于当前用户的对象键：发布用例拒绝且不创建记录。
- 解除关系后使用旧时间线游标或删除旧记录：返回 `NO_ACTIVE_RELATIONSHIP`，不泄露历史数据。

---

## 目标文件结构

```text
server/internal/
├── identity/
│   ├── domain/{user.go,refresh_token.go,errors.go}
│   ├── application/{ports.go,service.go,service_test.go}
│   ├── infrastructure/{mysql_repository.go,jwt.go,wechat.go}
│   └── interfaces/http/{handler.go,handler_test.go}
├── relationship/
│   ├── domain/{relationship.go,invitation.go,errors.go}
│   ├── application/{ports.go,service.go,service_test.go}
│   ├── infrastructure/{mysql_repository.go,mysql_repository_test.go}
│   └── interfaces/http/{handler.go,handler_test.go}
├── moment/
│   ├── domain/{moment.go,photo.go,status.go,errors.go,moment_test.go}
│   ├── application/{ports.go,service.go,service_test.go}
│   ├── infrastructure/{mysql_repository.go,cos_storage.go,mysql_repository_test.go}
│   └── interfaces/http/{handler.go,handler_test.go}
└── platform/
    ├── clock/clock.go
    ├── config/config.go
    ├── database/mysql.go
    └── httpserver/{auth.go,response.go,router.go,router_test.go}
```

### Task 1：建立 Platform 层和架构边界测试

**文件：**
- 创建：`server/internal/platform/clock/clock.go`
- 移动并修改：`server/internal/config/config.go` → `server/internal/platform/config/config.go`
- 移动并修改：`server/internal/config/config_test.go` → `server/internal/platform/config/config_test.go`
- 创建：`server/internal/platform/database/mysql.go`
- 移动并修改：`server/internal/httpapi/response.go` → `server/internal/platform/httpserver/response.go`
- 创建：`server/internal/architecture/dependencies_test.go`
- 修改：`server/cmd/api/main.go`
- 修改：`.gitignore`
- 删除：上一架构未提交的 `server/internal/moment/service_test.go`、`server/internal/storage/service_test.go`

**接口：**
- 产出：`clock.Clock` 接口 `Now() time.Time`，以及 `clock.System`。
- 产出：`config.Load() (config.Config, error)`。
- 产出：`database.OpenMySQL(dsn string) (*sql.DB, error)`。
- 产出：`httpserver.WriteData(...)` 与 `httpserver.WriteError(...)`。

- [ ] **步骤 1：写失败的架构测试**

`TestDomainPackagesDoNotImportInfrastructure` 使用 `go list -json` 检查所有 `*/domain` 包，断言依赖列表不含 Gin、`database/sql`、JWT、微信或 COS；先创建一个临时违规 fixture，证明测试会失败，再移除 fixture。

- [ ] **步骤 2：运行 RED**

运行：`cd server && go test ./internal/architecture ./internal/platform/...`

预期：Platform 包尚不存在，测试失败。

- [ ] **步骤 3：实现 Platform 层并清理工作区噪音**

迁移配置和响应代码；创建数据库与时钟适配器；`.gitignore` 加入 `.idea/`；删除仅属于旧架构且尚未提交的 Moment/Storage 测试。更新 `main.go` import，但临时保留旧业务包。

- [ ] **步骤 4：验证并提交**

运行：`cd server && go test ./...`

预期：全部通过。

```bash
git add .gitignore server
git commit -m "refactor: establish DDD platform boundaries"
```

### Task 2：迁移 Identity 身份域

**文件：**
- 创建：`server/internal/identity/domain/{user.go,refresh_token.go,errors.go}`
- 创建：`server/internal/identity/domain/user_test.go`
- 创建：`server/internal/identity/application/{ports.go,service.go,service_test.go}`
- 创建：`server/internal/identity/infrastructure/{mysql_repository.go,jwt.go,wechat.go}`
- 创建：`server/internal/identity/interfaces/http/{handler.go,handler_test.go}`
- 修改：`server/internal/platform/httpserver/{auth.go,router.go}`
- 修改：`server/cmd/api/main.go`
- 删除：`server/internal/auth/`、`server/internal/user/`、`server/internal/httpapi/auth_*`、`server/internal/httpapi/user_handler.go`

**接口：**
- 产出：`domain.NewUser(id int64, openID string) User`、`User.UpdateProfile(name, avatar string) error`。
- 产出：`application.IdentityProvider.Exchange(ctx, code) (ExternalIdentity, error)`。
- 产出：`application.Repository` 的用户查找、刷新令牌保存/轮换/撤销方法。
- 产出：`application.Service.LoginWithWeChat`、`RefreshSession`、`Logout`、`UpdateProfile`。
- 产出：`httpserver.Authenticate(tokens identityapplication.AccessTokenVerifier) gin.HandlerFunc`。

- [ ] **步骤 1：写 Identity 领域和用例失败测试**

覆盖昵称去空白、1–30 个 Unicode 字符边界、空微信 `code`、同一 OpenID 复用用户、15 分钟访问令牌、30 天刷新令牌、刷新轮换和撤销后拒绝。测试断言领域错误，不断言 Gin 或 SQL 行为。

- [ ] **步骤 2：运行 RED**

运行：`cd server && go test ./internal/identity/...`

预期：Identity 包尚未实现而失败。

- [ ] **步骤 3：实现领域、用例和基础设施适配器**

把现有可用逻辑迁移到对应层；应用服务只依赖端口和 `clock.Clock`。MySQL、JWT、微信 HTTP 调用分别放入基础设施文件。

- [ ] **步骤 4：迁移 HTTP 契约并删除旧包**

保持 `/v1/auth/wechat`、`/v1/auth/refresh`、`/v1/auth/logout`、`GET/PUT /v1/users/me` 及原错误码不变。删除旧 `auth`、`user` 代码后运行 `rg 'internal/(auth|user)' server`，预期无结果。

- [ ] **步骤 5：验证并提交**

运行：`cd server && go test ./... && go test -race ./internal/identity/...`

```bash
git add server
git commit -m "refactor: migrate identity domain to DDD"
```

### Task 3：迁移 Relationship 关系域

**文件：**
- 创建：`server/internal/relationship/domain/{relationship.go,invitation.go,errors.go}`
- 创建：`server/internal/relationship/domain/invitation_test.go`
- 创建：`server/internal/relationship/application/{ports.go,service.go,service_test.go}`
- 创建：`server/internal/relationship/infrastructure/{mysql_repository.go,mysql_repository_test.go}`
- 创建：`server/internal/relationship/interfaces/http/{handler.go,handler_test.go}`
- 修改：`server/internal/platform/httpserver/router.go`
- 修改：`server/cmd/api/main.go`
- 删除：`server/internal/couple/`、`server/internal/httpapi/couple_handler.go`

**接口：**
- 产出：`domain.NewInvitation(code string, creatorID int64, expiresAt time.Time) (Invitation, error)`、`Invitation.CanBeAcceptedBy(userID int64, now time.Time) error`。
- 产出：`application.Repository.AcceptInvitation(ctx, code, consumerID, now) (domain.Relationship, error)`，事务和并发锁由基础设施实现。
- 产出：`application.Service.CreateInvitation`、`PreviewInvitation`、`BindRelationship`、`GetRelationship`、`EndRelationship`。

- [ ] **步骤 1：写领域和应用失败测试**

覆盖六位非易混淆字符、十分钟过期、一次性使用、禁止自绑、已有关系拒绝、解除后无有效关系，以及两个 goroutine 接受同一邀请码时只允许一次成功。

- [ ] **步骤 2：运行 RED**

运行：`cd server && go test ./internal/relationship/...`

预期：Relationship 包尚未实现而失败。

- [ ] **步骤 3：实现领域与应用服务**

领域对象执行不依赖数据库的规则；应用服务生成邀请码、调用仓储并返回领域对象。仓储接口不暴露 `*sql.Tx`。

- [ ] **步骤 4：实现 MySQL 事务和 HTTP 适配器**

MySQL `AcceptInvitation` 使用 `SELECT ... FOR UPDATE` 锁邀请码并检查双方有效关系；路由保持 `/v1/couples/...` 不变。加入 `INVITE_EXPIRED`、`INVITE_USED`、`CANNOT_BIND_SELF`、`ALREADY_BOUND`、`NO_ACTIVE_COUPLE` 契约测试。

- [ ] **步骤 5：删除旧包、验证并提交**

运行：`cd server && go test ./... && go test -race ./internal/relationship/...`

```bash
git add server
git commit -m "refactor: migrate relationship domain to DDD"
```

### Task 4：实现 Moment 领域和发布用例

**文件：**
- 创建：`server/internal/moment/domain/{moment.go,photo.go,status.go,errors.go,moment_test.go}`
- 创建：`server/internal/moment/application/{ports.go,service.go,service_test.go}`

**接口：**
- 消费：Identity 提供的当前用户 ID；Relationship 的 `ActiveRelationshipReader`。
- 产出：`domain.NewMoment(authorID, relationshipID int64, text string, status domain.Status, place string, photoKeys []string, createdAt time.Time) (domain.Moment, error)`。
- 产出：`application.UploadAuthorizer`、`MomentRepository`、`ActiveRelationshipReader` 和 `PhotoStorage` 端口。
- 产出：`application.Service.RequestPhotoUpload`、`PublishMoment`、`GetHome`、`GetTimeline`、`DeleteMoment`。

- [ ] **步骤 1：写 Moment 领域 RED 测试**

覆盖空内容、500/501 字文字、80/81 字地点、合法/非法状态、0/1/9/10 张照片、重复对象键和照片顺序。期望使用领域错误 `ErrEmptyMoment`、`ErrTextTooLong`、`ErrPlaceTooLong`、`ErrInvalidStatus`、`ErrTooManyPhotos`。

- [ ] **步骤 2：实现最小领域模型并验证 GREEN**

运行：`cd server && go test ./internal/moment/domain`

预期：全部通过，且包依赖检查通过。

- [ ] **步骤 3：写发布、权限与失败原子性 RED 测试**

内存端口覆盖：无有效关系、伪造/过期/他人照片授权、九张照片顺序、保存失败不消费照片授权、非作者删除、解除关系后旧游标和删除均返回 `ErrNoActiveRelationship`。

- [ ] **步骤 4：实现应用用例并验证**

运行：`cd server && go test ./internal/moment/application`

预期：全部通过。

- [ ] **步骤 5：提交**

```bash
git add server/internal/moment/domain server/internal/moment/application
git commit -m "feat: add moment domain and use cases"
```

### Task 5：实现 Moment 基础设施和 HTTP API

**文件：**
- 创建：`server/internal/moment/infrastructure/{mysql_repository.go,mysql_repository_test.go,cos_storage.go}`
- 创建：`server/internal/moment/interfaces/http/{handler.go,handler_test.go}`
- 修改：`server/internal/platform/httpserver/router.go`
- 修改：`server/cmd/api/main.go`

**接口：**
- 消费：Task 4 定义的全部端口和应用服务。
- 端点：`POST /v1/uploads`、`POST /v1/moments`、`GET /v1/moments`、`GET /v1/home`、`DELETE /v1/moments/:id`。
- 时间线：默认 20 条、最多 50 条，游标由 `(created_at, id)` 编码。
- 首页：返回伴侣最近状态和最近五条未删除记录。

- [ ] **步骤 1：写 MySQL 和 HTTP RED 测试**

覆盖记录与 `moment_photos` 事务保存、1–9 排序、游标同时间戳无重复遗漏、软删除、关系过滤、第三方访问、十张照片的双重校验，以及各领域错误到稳定 API 错误码的映射。

- [ ] **步骤 2：实现 MySQL 仓储与 COS 适配器**

对象键格式为 `users/{userID}/{yyyy}/{mm}/{uuid}.{ext}`；只允许 JPEG、PNG、WebP，单张最大 10 MiB，授权十分钟。创建记录时在同一事务验证并消费全部授权。

- [ ] **步骤 3：实现 HTTP 适配器和路由**

Handler 只解析 DTO、调用用例并映射错误；不查询数据库、不实现关系判断。私有照片读取返回短期签名 URL。

- [ ] **步骤 4：全量验证并提交**

运行：`cd server && go test ./... && go test -race ./internal/identity/... ./internal/relationship/... ./internal/moment/... && go vet ./...`

预期：全部通过，`rg 'gin-gonic|database/sql|golang-jwt|cos-go' server/internal/*/domain` 无结果。

```bash
git add server
git commit -m "feat: add DDD moment adapters and API"
```

### Task 6：后端文档与人工 API 验收

**文件：**
- 创建：`docs/backend-architecture.md`
- 创建：`docs/backend-api.md`
- 创建：`docs/backend-manual-test.md`
- 修改：`Makefile`

**接口：**
- 产出：`make test-backend`、`make test-race`、`make vet`。

- [ ] **步骤 1：编写可执行的后端检查命令**

Makefile 命令必须从仓库根目录运行，并使用本地 Go 工具链或系统 `go`。`test-backend` 运行全部 Go 测试；`test-race` 运行三个业务域；`vet` 运行 `go vet ./...`。

- [ ] **步骤 2：编写架构和 API 文档**

架构文档解释四层职责、依赖方向及“新功能放在哪里”；API 文档列出请求、响应和错误码；人工测试文档覆盖两用户登录、绑定、九图发布、第三方拒绝、作者删除和解除后拒绝。

- [ ] **步骤 3：执行最终后端验收**

运行：`make test-backend && make test-race && make vet`

预期：全部退出码为 0。

- [ ] **步骤 4：提交**

```bash
git add Makefile docs
git commit -m "docs: explain DDD backend and API"
```
