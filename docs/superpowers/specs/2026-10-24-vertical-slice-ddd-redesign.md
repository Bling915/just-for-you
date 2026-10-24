# 小事成双对：纵向领域 DDD 与数据库基线重构设计

## 1. 背景与意图

当前后端在目录形式上使用 `domain / application / infrastructure / interfaces` 全局分层，但实际阅读一个业务流程时需要跨多个顶层目录跳转；`together` 同时包含纪念日、待办、愿望和每日一问，`moment` 的端口与 MySQL 仓储也承担过多职责。项目所有者本身是后端工程师，希望可以快速定位业务、理解依赖并独立维护，而不是只能依靠原作者继续修改。

本次重构的核心不是增加抽象，而是让代码按业务领域聚合、命名直接表达用途，并删除已经确认无价值的兼容层和冗余结构。当前数据库均为测试数据，允许清空重建，因此数据库迁移可以合并为一份干净基线。

## 2. 目标

- 改为按业务领域纵向组织，每个领域内部保留标准 DDD 分层。
- 一个开发者查看某项业务时，只需进入一个领域目录即可看到规则、用例、适配器和接口。
- 拆除含糊的 `together` 大包和巨型 Moment 仓储。
- 用业务语义命名端口、实现、Command、Query 和 Handler，避免裸 `Repository`、`Service`、`UseCases`。
- 审计全部数据库表、字段、索引和外键，只保留当前功能实际需要的结构。
- 合并现有迁移为一份可从空库建立完整结构的基线迁移。
- 小事仅保留一个 `status` 字段，同时支持快捷状态和最多 20 个 Unicode 字符的自定义状态。
- 默认保持当前 HTTP URL、主要请求字段和响应字段兼容；允许同步删除确认无用的接口或字段。
- 保留现有功能行为和自动化测试覆盖。
- 更新维护文档，使项目所有者可以沿固定路径阅读和新增功能。

## 3. 非目标

- 不引入微服务、消息队列、事件总线、CQRS 框架或依赖注入框架。
- 不为了“纯 DDD”给每张表创建无业务价值的聚合、工厂或接口。
- 不重做小程序视觉设计。
- 不保留测试数据库数据，也不提供旧迁移链到新基线的线上无损升级路径。
- 不改变微信登录、COS、JWT、AI 服务的供应商协议，除非审计发现现有实现错误。

## 4. 选择的架构

### 4.1 考虑过的方案

1. **保留全局分层并改善命名**：改动较小，但阅读一个功能仍需跨四个顶层目录，不能解决项目所有者最核心的维护痛点。
2. **按领域纵向分包，每个领域内部分层**：改动较大，但每项业务完整聚合，最符合项目所有者的阅读习惯。已选择。
3. **完全重写**：可以一次获得整洁结构，但容易遗漏已有并发、隐私和授权规则，风险大于收益，因此不采用。

### 4.2 目标目录

```text
backend/internal/
  account/
    domain/             用户、刷新会话等业务对象与规则
    application/        登录、刷新、退出、资料、头像用例与端口
    infrastructure/     MySQL、微信登录、COS 头像适配器
    interfaces/http/    account HTTP handler 与 DTO
  couple/
    domain/             情侣关系、邀请码、单人测试规则
    application/
    infrastructure/
    interfaces/http/
  moment/
    domain/             小事、照片、空间、可见性、回应、双视角
    application/        发布、编辑、删除、时间线、分享、回应、回忆查询
    infrastructure/     MySQL 与 COS 照片适配器
    interfaces/http/
  companionship/
    domain/             此刻状态、轻互动
    application/
    infrastructure/     MySQL、微信订阅消息
    interfaces/http/
  anniversary/
  todo/
  wish/
  dailyquestion/
  media/
    domain/             上传授权及用途规则
    application/        请求上传、确认上传、消费授权
    infrastructure/     upload_grants MySQL 与 COS 对象存储
  platform/
    clock/              系统时钟适配器
    config/             环境配置
    database/           数据库连接
    auth/               JWT 与 HTTP 鉴权中间件
    httpresponse/       统一响应与错误映射
  bootstrap/            唯一允许了解所有具体实现并完成装配的位置
```

`account` 替换含糊的 `identity`，`couple` 替换抽象的 `relationship`。`together` 被拆成 `anniversary / todo / wish / dailyquestion`。`platform` 只包含无业务含义的技术能力，禁止放业务 DTO、通用仓储或“以后可能复用”的辅助函数。

`companionship` 只拥有“此刻状态、想你/抱抱/加油等轻互动、通知偏好和通知发送”。Moment 的发布状态、回应和双视角仍归 `moment`，不因界面上相邻而混入 companionship。

`media` 拥有跨领域使用的上传授权和 COS 访问。Account、Moment、Anniversary 各自在 application 中定义所需的窄接口，Media 提供适配器实现；业务领域不得直接查询 `upload_grants` 或调用 COS SDK。

### 4.3 依赖规则

单个领域内部：

```text
interfaces -> application -> domain
infrastructure -> application + domain
bootstrap -> 所有具体实现
```

- `domain` 仅依赖标准库。
- `application` 通过按消费者定义的端口使用数据库和外部服务。
- 一个领域不得直接导入另一个领域的 infrastructure 或 interfaces。
- 跨领域读取通过窄端口完成，例如 Moment 应用层定义 `ActiveCoupleReader`，由 Couple 的 MySQL 查询适配器实现。
- 不创建通用 `Repository[T]`；端口按用途命名并保持最小接口。

## 5. 命名规范

### 5.1 应用用例

- 写操作：`PublishMoment`, `UpdateMoment`, `DeleteMoment`, `BindCouple`。
- 读操作：`GetMomentTimeline`, `GetCurrentCouple`, `ListAnniversaries`。
- 输入：`PublishMomentCommand`、`GetMomentTimelineQuery`。
- 用例对象只有在需要依赖时才定义，例如 `type PublishMoment struct { ... }`；不再统一追加含糊的 `UseCase`。
- HTTP 端按业务动作拆文件，但同一领域只暴露一个 `Handler` 负责注册路由。

### 5.2 端口

- `MomentWriter`：保存、更新、软删除小事。
- `MomentReader`：按 ID 读取小事。
- `TimelineReader`：读取分页时间线。
- `ReactionStore`、`PerspectiveStore`：分别处理回应与双视角。
- `PhotoStorage`、`UploadGrantStore`：照片与上传授权。
- `ActiveCoupleReader`：读取当前有效情侣关系。

禁止仅叫 `Repository`、`Store`、`Service` 的公开类型。MySQL 实现使用与端口对应的明确名称，例如 `MySQLMomentWriter`、`MySQLTimelineReader`。

### 5.3 领域语言

- Go 领域使用 `Moment`、`Couple`、`Account`、`Anniversary` 等业务词。
- 数据库表保留复数蛇形命名；HTTP JSON 保留现有 camelCase。
- API DTO 与领域对象明确分开，DTO 不进入 domain。

## 6. Moment 状态统一设计

小事只有一个状态值，不再区分状态码和自定义状态。

- 数据库列：`moments.status VARCHAR(80) NOT NULL DEFAULT ''`。
- API 字段：继续使用 `status`。
- Go 值对象：`type Status string`，创建时执行 trim；空值允许；最多 20 个 Unicode code point（Go rune）；拒绝控制字符。组合 emoji 可能计为多个字符，前后端使用同一计数规则。
- 快捷状态由前端提供：`在上课`、`想你了`、`有点累`。
- 用户选择快捷状态时，请求直接发送中文展示值。
- 用户选择自定义状态时，清空当前快捷选中并发送自定义文字。
- 切回快捷状态时，清空自定义输入。
- 卡片、首页最近状态和月报直接展示 `status` 原文，不维护第二套 label map。
- 编辑页和草稿使用同一字段，旧的 `in_class / miss_you / tired` 测试数据无需迁移，因为数据库会重建。

## 7. 数据库审计与新基线

### 7.1 策略

删除现有 `000001` 至 `000004` 增量迁移，新增一组干净基线：

```text
000001_baseline.up.sql
000001_baseline.down.sql
```

本地和测试环境删除 `just_for_you` 数据库后重新执行基线。迁移工具增加 `schema_migrations` 记录，后续新功能从 `000002` 开始，禁止再次手工猜测某个版本是否执行。

### 7.2 审计分类

实现计划开始前，为每张表和字段生成清单并归入以下类别：

- **业务必需**：被领域规则、查询或 API 明确使用，保留。
- **技术必需**：外键、幂等、并发控制、分页或安全所需，保留并写明原因。
- **派生可删除**：能够可靠由其他字段计算且没有性能依据，删除。
- **已废弃**：前后端、用例和查询均不使用，删除。
- **用途不明**：在确认前不进入新基线，并通过 Git 历史追踪来源。

### 7.3 已确认的处理

- `moments.status_code` 改为单列 `status`。
- `space_type / visibility` 保留，它们分别表达所属空间与可见范围，并非重复字段。
- `source_moment_id` 保留，用于个人小事分享到共同空间后的来源关联。
- `daily_question_id` 保留，用于每日回答解锁后生成的特殊共同记录；删除 `kind`，因为是否为每日问题记录可以由 `daily_question_id IS NOT NULL` 完整推导。领域对象通过查询方法表达该语义，不再持久化重复分类。
- `occurred_at / created_at / updated_at` 保留，分别表达业务发生时间、创建时间和并发版本。
- `deleted_at` 保留以支持软删除和历史引用完整性。
- `upload_grants.purpose / resource_id` 保留，当前分别区分小事照片、头像和纪念日封面，并约束封面所属资源，这是上传授权校验所需而非冗余。
- 删除 `notification_deliveries`；当前生产代码没有写入或读取它，它只是未实现通知投递记录的预留表。
- 所有索引根据真实查询的 `WHERE + ORDER BY` 重新建立，不机械照搬旧迁移。

## 8. HTTP 兼容边界

- 保持现有 `/v1` URL 和 HTTP 方法。
- 保持小程序正在使用的 JSON 字段，包括 `spaceType`、`visibility`、`status`、`photoKeys`。
- 内部目录和类型重命名不泄漏到 API。
- 可以删除确认没有前端调用、没有文档承诺、没有测试覆盖的接口或响应字段；删除项必须在实现计划中列清单。
- 错误 code 默认保持兼容，内部 Go 错误改名后仍映射到原 code。

## 9. 重构顺序与安全边界

采用按领域逐个迁移的方式，而不是一次移动全部文件：

1. 建立新架构依赖测试和目录骨架。
2. 生成数据库字段/接口/前端调用审计清单。
3. 先生成并验证干净基线草案，但旧代码迁移期间仍使用当前 schema，避免为了中间状态增加兼容代码。
4. 依次迁移 `account -> couple -> media -> moment -> companionship -> anniversary -> todo -> wish -> dailyquestion`。Moment 迁移时先让单一状态语义运行在旧列 `status_code` 上。
5. 每迁移一个领域，删除对应旧包并运行全量测试，禁止新旧实现长期并存。
6. 所有代码完成纵向迁移后，正式切换到新基线，把 `status_code` 更名为 `status`、删除 `kind` 和废弃表，并清空重建本地数据库。
7. Moment 迁移同时完成前端自定义状态；切换基线前后使用相同 API 行为。
8. 更新 bootstrap、README、领域地图和“如何追踪一个请求”文档。
9. 在空数据库执行完整建库、启动、API smoke test 和前端测试。

每个阶段都必须保持编译通过。重构不与新产品功能混做，除已明确要求的自定义状态。

### 9.1 跨领域流程与事务

“每日一问双方回答完成后生成共同时间线记录”由 `dailyquestion/application/SubmitDailyAnswer` 编排。保存答案、判断解锁和创建时间线投影必须在同一 MySQL 事务中完成，防止出现“已解锁但时间线缺记录”或重复记录。

为避免引入通用 Unit of Work，`dailyquestion` 定义一个用例专用端口 `AnswerSubmissionStore`，其 MySQL 实现执行该原子事务。它可以写入 daily question 表和用于时间线展示的 Moment 投影行，但输入必须是明确的应用 DTO，并复用 Moment 提供的状态/文本校验函数；这是受测试约束的跨领域投影适配器，不允许扩展成任意 Moment 仓储。其他跨领域调用默认不共享事务，通过窄读写端口完成。

## 10. 测试与验收

- 架构测试验证每个领域的四层依赖方向及跨领域规则。
- 每个领域保留 domain 单元测试、application 用例测试、infrastructure SQL 测试、HTTP handler 测试。
- 基线迁移测试验证空库 up/down、外键、索引和关键约束。
- Moment 状态测试覆盖空状态、三个快捷中文状态、20 字自定义状态、超长与控制字符。
- 前端测试覆盖发布、编辑、草稿恢复、快捷/自定义互斥和卡片原文展示。
- 全量运行 `go test ./... -count=1`、`go test -race ./...`、`go vet ./...`、`npm test`。
- 用新建本地数据库启动 API，验证登录、绑定、发布共同/个人小事、回应、纪念日、每日一问主要流程。

## 11. 文档交付

README 增加：

- 领域目录地图。
- 每个领域的职责与依赖。
- 从 HTTP 路由追踪到用例、领域规则和 MySQL 的固定阅读方法。
- 新增领域、新增用例、新增数据库字段的模板。
- 数据库重建步骤和后续迁移规范。

重构完成后，项目所有者应能仅凭目录名和类型名回答：“这个功能在哪个领域、入口是什么、业务规则在哪、数据库实现在哪”。

## 12. 留给实现者的局部决定

- 同一领域内按文件还是按 `command / query` 子目录组织 application；优先选择文件数量更少、跳转更短的方案。
- SQL 是否使用独立 query 文件；当前规模下默认保留 Go 常量，除非单个查询明显过长。
- 领域内部的小型 DTO 名称与私有辅助函数，只要遵守本设计命名原则即可。
