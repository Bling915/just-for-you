# 小事成双对后端全局四层 DDD 重构设计

## 目标与背景

当前 Go 后端已经按 `identity`、`relationship`、`moment` 三个业务域实现了轻量 DDD，但每个业务域内部重复四层目录，应用层又由一个大型 `Service` 承担全部用例。对于熟悉公司项目中全局分层结构的维护者，这种组织方式阅读入口不明确，`main.go` 的依赖装配也过于集中。

本次重构把后端调整为严格的全局四层目录，并把应用服务拆成一个用户动作对应一个用例。重构只改变代码组织和依赖装配，不改变现有 API 路径、JSON 契约、数据库表结构或业务行为。完成后，维护者可以从 HTTP handler 进入对应 use case，再进入领域模型和仓储接口，最后找到基础设施实现。

## 范围

本次包含：

- 将 `identity`、`relationship`、`moment` 迁移到全局 `interfaces`、`application`、`domain`、`infrastructure` 四层。
- 将每个领域的综合 `Service` 拆为独立用例。
- 将配置和依赖装配集中到 `bootstrap`，把 `cmd/api/main.go` 控制在约 30 行。
- 完善架构依赖测试、用例测试和 HTTP 契约测试。
- 编写面向维护者的后端阅读与扩展文档。

本次不包含：

- 拆分微服务。
- 修改现有 MySQL 表或迁移历史数据。
- 修改小程序页面和交互。
- 引入 ORM、依赖注入框架、CQRS 框架、消息总线或领域事件。
- 实现单人测试模式；该功能在架构重构稳定后单独设计和实现。

## 目标目录

```text
server/
├── cmd/
│   ├── api/main.go
│   └── migrate/main.go
├── internal/
│   ├── interfaces/http/
│   │   ├── identity/
│   │   ├── relationship/
│   │   ├── moment/
│   │   └── shared/
│   ├── application/
│   │   ├── identity/
│   │   ├── relationship/
│   │   └── moment/
│   ├── domain/
│   │   ├── identity/
│   │   ├── relationship/
│   │   └── moment/
│   ├── infrastructure/
│   │   ├── persistence/mysql/
│   │   │   ├── identity/
│   │   │   ├── relationship/
│   │   │   ├── moment/
│   │   │   └── shared/
│   │   ├── identity/wechat/
│   │   ├── security/jwt/
│   │   ├── storage/cos/
│   │   ├── config/
│   │   └── clock/
│   ├── bootstrap/
│   │   ├── application.go
│   │   ├── identity.go
│   │   ├── relationship.go
│   │   └── moment.go
│   └── architecture/
└── migrations/
```

## 依赖规则

依赖方向固定为：

```text
interfaces/http -> application -> domain
infrastructure  -> application/domain
bootstrap       -> interfaces/application/infrastructure
cmd             -> bootstrap
```

- `domain` 只能依赖 Go 标准库，不得导入 Gin、SQL、JWT、微信或 COS 包，也不得反向依赖 `application`。
- `application` 可以依赖 `domain`，但不得导入 Gin、`database/sql` 或具体基础设施包。
- `interfaces/http` 只负责鉴权后的请求解析、DTO 转换、调用用例和错误映射，不直接查询数据库。
- `infrastructure` 实现领域层或应用层定义的端口，不包含 HTTP 表现逻辑。
- `bootstrap` 是唯一允许同时感知接口层、应用层和基础设施层的装配位置。
- `cmd/api/main.go` 只加载应用、处理启动错误并运行/关闭 HTTP 服务。

`internal/architecture` 的自动测试将通过 `go list -json` 检查这些 import 约束，防止后续开发把 SQL 或 Gin 重新写回业务层。

## 应用层用例

应用层按业务域分包，但不再提供包含所有操作的通用 `Service`。每个用户动作对应一个明确的用例类型：

```text
application/
├── identity/
│   ├── login.go
│   ├── refresh_session.go
│   ├── logout.go
│   ├── get_profile.go
│   ├── update_profile.go
│   └── ports.go
├── relationship/
│   ├── create_invitation.go
│   ├── preview_invitation.go
│   ├── bind_relationship.go
│   ├── get_relationship.go
│   ├── end_relationship.go
│   └── ports.go
└── moment/
    ├── request_photo_upload.go
    ├── publish_moment.go
    ├── get_timeline.go
    ├── get_home.go
    ├── delete_moment.go
    └── ports.go
```

用例命名采用 `<Action>UseCase`，统一提供 `Execute(ctx, command)`。有输入的写操作使用 `<Action>Command`，读取操作可使用 `<Action>Query`，输出使用 `<Action>Result` 或领域对象。不会为此引入 CQRS 框架；这些名称只是为了让输入、输出和职责一眼可见。

每个 HTTP handler 只持有自己需要的用例接口，而不是一个能够执行所有操作的庞大服务。用例依赖的 Repository、外部身份交换、Token、照片存储和 Clock 接口集中在本业务域的 `ports.go` 中。

## 领域层

领域层按业务语言分为三个包：

- `domain/identity`：`User`、`RefreshToken` 以及身份相关业务错误。
- `domain/relationship`：`Relationship`、`Invitation`、绑定和解除规则。
- `domain/moment`：`Moment`、`Photo`、`Status` 和记录内容规则。

新业务对象通过 `New...` 构造函数创建并执行不变量校验；从持久化数据恢复已有对象时使用 `Restore...`，避免读取数据库时误触发“创建新对象”的规则。领域对象不携带 JSON 或 SQL 标签。

现有规则继续由领域层保证，包括昵称长度、邀请码有效性、记录至少包含一种内容、文字与地点长度、状态合法性、照片最多九张、照片不重复和只能由作者删除。业务错误继续使用稳定的 sentinel error，以便接口层保持现有 API 错误码。

Repository 接口如果表达领域聚合的持久化能力，放在相应 `domain` 包；只为某个用例提供的外部端口，例如微信身份交换、Token 签发、上传签名和 Clock，则放在 `application/<domain>/ports.go`。

## 基础设施层

基础设施层按技术能力组织：

- `persistence/mysql`：Repository 的 MySQL 实现、事务和数据库连接。
- `identity/wechat`：微信 `code` 换取外部身份。
- `security/jwt`：访问令牌签发与验证。
- `storage/cos`：照片上传和读取签名。
- `config`：环境变量解析与校验。
- `clock`：生产环境系统时间实现。

继续使用 `database/sql` 和现有 SQL，不引入 ORM。数据库行模型只存在于 MySQL 实现内部，通过明确的 mapper 恢复领域实体。关系绑定和发布记录所需的原子操作继续在 Repository 实现内部使用事务完成，事务对象不得暴露给应用层。

## 接口层和错误处理

HTTP 层按业务域分别注册路由，公共部分提供认证中间件、响应格式和领域错误到 API 错误码的映射。API 路径、HTTP 状态码、字段名称和现有错误码保持不变。

请求流程为：

```text
Gin Router
  -> Handler 解析并校验传输格式
  -> UseCase.Execute 执行业务流程
  -> Domain 校验业务规则
  -> Repository/Port 执行外部操作
  -> Handler 将结果映射为现有 JSON 契约
```

可预期的业务错误由共享 error mapper 转换；未知错误返回通用内部错误且不泄露密钥、微信 code、访问令牌、刷新令牌或数据库细节。

## Bootstrap 与启动

`bootstrap.Application` 负责：

1. 加载配置并连接 MySQL。
2. 创建 Clock、微信、JWT、COS 和 Repository 实现。
3. 创建每个独立用例。
4. 创建 HTTP handlers、认证中间件和 Router。
5. 暴露 `Run(ctx)` 和 `Close()` 生命周期方法。

三个业务装配文件分别负责各域依赖，避免形成新的超大 `application.go`。跨域能力通过窄接口连接，例如 moment 用例只读取当前有效关系，不依赖 relationship 的全部用例集合。

## 迁移策略

重构采用按域迁移，并保持每一步都可编译、可测试：

1. 建立全局目录、共享基础设施、bootstrap 骨架和更严格的架构测试。
2. 迁移 Identity，保持登录、刷新、退出和个人资料 API 不变。
3. 迁移 Relationship，保持邀请码、绑定、读取和解除 API 不变。
4. 迁移 Moment，拆分上传、发布、首页、时间线和删除用例。
5. 删除旧目录和兼容代理，精简 `main.go`，更新维护文档。

不在新旧目录长期保留两套实现。每个域迁移完成、测试通过后删除该域旧代码，防止双重入口继续增加理解成本。

## 测试与验收

- 领域测试：业务不变量和领域错误。
- 用例测试：使用内存 fake 验证编排、授权、事务边界外的失败行为。
- Repository 测试：使用 SQL mock 固定查询、事务和映射行为。
- HTTP 契约测试：固定现有路由、状态码、JSON 字段和错误码。
- 架构测试：固定四层 import 方向和禁止依赖。
- 全量验证：`go test ./...`、关键业务包 `go test -race`、`go vet ./...`。

验收标准：

- 所有现有 Go 测试迁移后通过，新增架构测试通过。
- 小程序无需改动即可继续调用后端。
- `cmd/api/main.go` 约 30 行，不包含具体 Repository 或 use case 装配。
- 不再存在 `internal/identity/{application,domain,...}` 这种按域嵌套四层的旧目录。
- 应用层不再存在包揽多个用户动作的通用 `Service`。
- 维护文档能够回答“新增接口、用例、业务规则和 MySQL 查询分别放在哪里”。

## 后续功能

单人测试模式作为下一项独立功能。全局分层完成后，它可以通过新增明确的用例和关系策略接入，而不需要在 handler 或 MySQL 代码中散落测试模式判断。
