# 小事成双对纵向领域 DDD 重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保持当前小程序 API 与业务行为的前提下，将 Go 后端改造为按领域纵向分包的 DDD 结构，合并 Moment 状态字段，并用干净基线重建可丢弃的测试数据库。

**Architecture:** 每个业务上下文拥有自己的 `domain / application / infrastructure / interfaces/http`，跨领域依赖只经由消费方定义的窄端口。按 `platform -> account -> couple -> media -> moment -> companionship -> anniversary -> todo -> wish -> dailyquestion` 顺序迁移，每个领域迁移后立即删除旧包；最后才清空本地数据库并切换到新基线。

**Tech Stack:** Go 1.23, Gin 1.11, MySQL 8, `database/sql`, `go-sqlmock`, Tencent COS SDK, WeChat Mini Program JavaScript, Node.js 20 `node:test`.

**Spec:** `docs/superpowers/specs/2026-10-24-vertical-slice-ddd-redesign.md`

## Global Constraints

- 保持现有 `/v1` URL、HTTP 方法、小程序使用中的 JSON 字段和错误 code；只删除经前端调用、文档和测试三方审计均未使用的项。
- 不引入微服务、消息队列、事件总线、CQRS 框架、DI 框架或通用 `Repository[T]`。
- `domain` 只依赖 Go 标准库；`application` 只经窄端口使用持久化与外部系统；领域之间不得导入对方的 `infrastructure` 或 `interfaces`。
- 公开类型用业务语义命名，不新增裸 `Repository`、`Store`、`Service`、`UseCases`。
- Moment 只保留一个 `status`：空值合法，trim 后最多 20 个 Unicode code point（Go rune），禁止控制字符。
- 实施期间使用现有 schema，Moment 的新语义暂时落到 `status_code`；全部领域迁移完成后再切换新 `000001_baseline`。
- 尊重未提交的 `backend/scripts/dev.sh` 和 `frontend/miniprogram/config.js`，不覆盖、不纳入重构提交。
- 每个任务遵循 TDD，并且产生一个可独立审查的提交。

## Review Focus

- 旧前端传入 `in_class / miss_you / tired` 时不得被静默转换成中文；新前端必须直接传展示文本，后端按单一字符串校验。
- 包含组合 emoji 的状态必须前后端都按 Unicode code point 计数，20 个通过、21 个拒绝。
- 每日一问双方解锁遇到并发重试时，只能有一条时间线投影，答案与投影必须同成同败。
- 个人小事、共同小事、分享副本和双视角在重构后必须继续遵守原有可见性和授权边界。
- 新基线必须在全空库上可 up/down/up，所有外键、唯一约束和分页索引可实际执行，不得只做 SQL 文本断言。

---

### Task 1: 锁定现有行为与架构边界

**Files:**
- Create: `backend/docs/audits/2026-10-24-api-schema-usage.md`
- Modify: `backend/internal/architecture/dependencies_test.go`
- Test: `backend/internal/architecture/dependencies_test.go`

**Interfaces:**
- Consumes: 现有 Go import graph、Gin 路由、前端 `services/*.js`、四份旧迁移。
- Produces: 已使用 API/字段/表索引清单；`TestVerticalSliceDependencies` 作为后续所有任务的架构门禁。

- [ ] **Step 1: 扩展架构测试**

  在 `dependencies_test.go` 新增 `TestVerticalSliceDependencies`，断言 `domain` 不引入项目内包、`application` 不引入 `infrastructure/interfaces/bootstrap`、一个领域不引入另一领域的 `infrastructure/interfaces`、旧全局分层目录在终局不存在。在迁移期用显式 `legacyRoots` 白名单记录尚未迁移的旧目录。

- [ ] **Step 2: 验证新门禁会抓到越界 import**

  Run: `cd backend && go test ./internal/architecture -run TestVerticalSliceDependencies -count=1`
  Expected: FAIL，报告当前旧目录或人工注入的越界样例。

- [ ] **Step 3: 添加迁移期白名单并生成审计文档**

  文档逐项列出所有当前路由、前端调用方、JSON 字段、表/字段/索引，标注 `keep / derive / delete / unknown`；明确 `kind` 和 `notification_deliveries` 为 delete。

- [ ] **Step 4: 运行架构测试**

  Run: `cd backend && go test ./internal/architecture -count=1`
  Expected: PASS，文档中没有 `unknown`。

- [ ] **Step 5: Commit**

  `git add backend/docs/audits backend/internal/architecture && git commit -m "test: lock vertical slice boundaries"`

### Task 2: 建立 Platform 技术内核

**Files:**
- Create: `backend/internal/platform/{clock,config,database,auth,httpresponse}/*.go`
- Modify: `backend/internal/interfaces/http/shared/{auth,response,router}.go`
- Modify: `backend/internal/bootstrap/application.go`
- Test: `backend/internal/platform/config/config_test.go`
- Test: `backend/internal/platform/auth/token_service_test.go`
- Test: `backend/internal/platform/httpresponse/response_test.go`
- Delete after callers move: `backend/internal/application/shared/clock.go`, `backend/internal/infrastructure/{clock,config,security}`, `backend/internal/infrastructure/persistence/mysql/shared`, `backend/internal/interfaces/http/shared/{auth,response}.go`

**Interfaces:**
- Consumes: 现有 config/JWT/database/HTTP response 行为。
- Produces: `clock.Clock { Now() time.Time }`、`config.Load() (config.Config, error)`、`database.OpenMySQL(dsn string) (*sql.DB, error)`、`auth.TokenService`、`httpresponse.WriteData/WriteError`。

- [ ] **Step 1: 为 platform 公开契约编写失败测试**

  复制并改名现有 config/JWT/response 用例，额外断言缺少必需密钥时 `config.Load` 返回 error，而非启动后 panic。

- [ ] **Step 2: 确认新包尚不存在**

  Run: `cd backend && go test ./internal/platform/... -count=1`
  Expected: FAIL with `directory prefix internal/platform` or missing symbols.

- [ ] **Step 3: 移动实现并保持行为**

  只做技术能力归位和明确命名；`router.go` 暂时保留作为迁移桥接，其返回格式与 auth middleware 对外行为不变。

- [ ] **Step 4: 更新所有调用方并删除旧技术包**

  Run: `cd backend && go test ./internal/platform/... ./internal/interfaces/http/shared ./internal/bootstrap -count=1`
  Expected: PASS.

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: establish platform kernel"`

### Task 3: 迁移 Account 领域

**Files:**
- Create: `backend/internal/account/domain/{user,refresh_token,errors}.go`
- Create: `backend/internal/account/application/{login,refresh_session,logout,get_profile,update_profile,request_avatar_upload,ports,session}.go`
- Create: `backend/internal/account/infrastructure/mysql/*.go`
- Create: `backend/internal/account/infrastructure/wechat/code_exchanger.go`
- Create: `backend/internal/account/interfaces/http/{handler,dto}.go`
- Modify: `backend/internal/bootstrap/{application,identity}.go`
- Move tests beside the corresponding new files.
- Delete: `backend/internal/{domain,application}/identity`, `backend/internal/infrastructure/{identity,persistence/mysql/identity}`, `backend/internal/interfaces/http/identity`

**Interfaces:**
- Consumes: `platform.clock.Clock`, `platform.auth.TokenService`，以及 Task 5 前临时由 account infrastructure 保持的头像上传适配。
- Produces: `application.FindAccount { GetAccount(context.Context, int64) (domain.User, error) }`，以及与现有 auth/profile API 等价的 `http.Handler`。

- [ ] **Step 1: 把 identity 用例测试迁入 account，改用新业务名**

  测试覆盖登录、refresh rotation、logout、资料修改、头像 grant 绑定，并断言原 JSON 契约不变。

- [ ] **Step 2: 确认新 account 契约未实现**

  Run: `cd backend && go test ./internal/account/... -count=1`
  Expected: FAIL due to missing domain/application implementations.

- [ ] **Step 3: 迁移生产代码并重命名含糊类型**

  用 `Login`、`RefreshSession`、`UpdateProfile`、`MySQLAccountReader`、`MySQLSessionStore` 替换 `*UseCase` 和 `MySQLRepository`；不在此任务改变业务逻辑。

- [ ] **Step 4: 切换 bootstrap/router，删除 identity 旧包**

  Run: `cd backend && go test ./internal/account/... ./internal/bootstrap ./internal/architecture -count=1`
  Expected: PASS，`rg 'internal/(domain|application)/identity|interfaces/http/identity' backend --glob '*.go'` 无结果。

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: move identity into account context"`

### Task 4: 迁移 Couple 领域

**Files:**
- Create: `backend/internal/couple/domain/{couple,invitation,errors}.go`
- Create: `backend/internal/couple/application/{create_invitation,preview_invitation,bind_couple,get_current_couple,end_couple,enable_solo_test,ports}.go`
- Create: `backend/internal/couple/infrastructure/mysql/*.go`
- Create: `backend/internal/couple/interfaces/http/{handler,dto}.go`
- Modify: `backend/internal/bootstrap/{application,relationship}.go`
- Move tests beside new files.
- Delete: `backend/internal/{domain,application}/relationship`, `backend/internal/infrastructure/persistence/mysql/relationship`, `backend/internal/interfaces/http/relationship`

**Interfaces:**
- Consumes: `platform.clock.Clock`.
- Produces: `infrastructure/mysql.ActiveCoupleLookup.GetActiveCouple(ctx context.Context, userID int64) (ActiveCoupleRecord, error)` 作为 bootstrap 可注入适配器；各消费领域仍自行定义所需的 `ActiveCoupleReader` 窄端口和本地 read model。

- [ ] **Step 1: 迁移邀请、绑定、解绑和单人测试用例**

  额外断言非成员调用 `PartnerOf` 返回明确错误，单人测试仍只在原有配置允许时注册路由。

- [ ] **Step 2: 运行新测试确认失败**

  Run: `cd backend && go test ./internal/couple/... -count=1`
  Expected: FAIL due to missing new packages/types.

- [ ] **Step 3: 迁移实现并切换消费方到 `ActiveCoupleReader`**

  Moment/Companionship/Together 各自在 application 包定义 `ActiveCoupleReader`；bootstrap 用 Couple 的 MySQL lookup 适配这些端口。消费领域不得引入 couple infrastructure。

- [ ] **Step 4: 删除 relationship 旧包并验证**

  Run: `cd backend && go test ./internal/couple/... ./internal/application/... ./internal/bootstrap ./internal/architecture -count=1`
  Expected: PASS，旧 relationship import 为零。

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: create couple bounded context"`

### Task 5: 建立 Media 上传授权领域

**Files:**
- Create: `backend/internal/media/domain/{purpose,grant,errors}.go`
- Create: `backend/internal/media/application/{request_upload,consume_grant,ports}.go`
- Create: `backend/internal/media/infrastructure/mysql/upload_grant_store.go`
- Create: `backend/internal/media/infrastructure/cos/object_storage.go`
- Move: `backend/internal/infrastructure/storage/cos/photo_storage.go`
- Modify: account/moment/anniversary application ports and bootstrap wiring.
- Test: `backend/internal/media/domain/grant_test.go`
- Test: `backend/internal/media/application/request_upload_test.go`
- Test: `backend/internal/media/infrastructure/mysql/upload_grant_store_test.go`

**Interfaces:**
- Consumes: `platform.clock.Clock`, COS SDK, current `upload_grants` schema.
- Produces: `application.UploadRequester.Request(ctx, RequestUploadCommand) (UploadGrant, error)` and consumer-owned adapters for avatar, moment photo, anniversary cover.

- [ ] **Step 1: 写上传用途与资源绑定测试**

  断言 `avatar / moment_photo / anniversary_cover` 用途、content type、过期时间、owner/resource ID 校验，且 grant 不能二次消费。

- [ ] **Step 2: 确认新 media 测试失败**

  Run: `cd backend && go test ./internal/media/... -count=1`
  Expected: FAIL due to missing implementation.

- [ ] **Step 3: 实现 Media 并用窄适配器接入三个消费领域**

  Account/Moment/Anniversary 只保留自己需要的 `AvatarStorage / PhotoStorage / CoverStorage` 窄端口，不直接查 `upload_grants`。

- [ ] **Step 4: 验证上传流程**

  Run: `cd backend && go test ./internal/media/... ./internal/account/... ./internal/application/moment/... ./internal/application/together/... -count=1`
  Expected: PASS.

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: centralize upload grants in media"`

### Task 6: 迁移 Moment 领域并合并状态语义

**Files:**
- Create: `backend/internal/moment/domain/*.go`
- Create: `backend/internal/moment/application/{publish_moment,update_moment,delete_moment,get_timeline,get_home,share_to_couple,react_to_moment,save_perspective,get_resurfaced_memory,read_models,ports}.go`
- Create: `backend/internal/moment/infrastructure/mysql/{moment_writer,moment_reader,timeline_reader,reaction_store,perspective_store,report_reader}.go`
- Create: `backend/internal/moment/interfaces/http/{handler,dto}.go`
- Modify: `backend/internal/bootstrap/{application,moment}.go`
- Move all existing Moment tests beside the new files.
- Delete: old global moment domain/application/mysql/http packages.

**Interfaces:**
- Consumes: Moment application 自定义的 `ActiveCoupleReader`、`PhotoStorage`，以及 `platform.clock.Clock`；bootstrap 注入具体适配器。
- Produces: `domain.NewStatus(string) (Status, error)`; `MomentWriter`, `MomentReader`, `TimelineReader`, `ReactionStore`, `PerspectiveStore`; unchanged Moment HTTP contract.

- [ ] **Step 1: 先为单一 `Status` 值对象写边界测试**

  断言 trim、空值、`\u5728\u4e0a\u8bfe`、20 rune、21 rune、组合 emoji 按 rune 计数、`\n` 等控制字符被拒绝；断言 `Kind` 不再出现在新 `Moment` 结构。

- [ ] **Step 2: 迁移并拆分 application/mysql 的巨型端口**

  每个端口只包含对应用例所需方法；对 `personal/shared/partner`、分享副本、reaction、perspective 权限的旧测试必须原样迁入。

- [ ] **Step 3: 在旧 schema 上运行新状态语义**

  MySQL adapter 暂时读写 `status_code`，但业务层和 API 只见 `status`；快捷状态不做 code/label 映射。

- [ ] **Step 4: 运行 Moment 与架构测试**

  Run: `cd backend && go test ./internal/moment/... ./internal/bootstrap ./internal/architecture -count=1`
  Expected: PASS，`rg 'internal/(domain|application)/moment|persistence/mysql/moment|interfaces/http/moment' backend --glob '*.go'` 不再匹配旧 import。

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: create moment bounded context"`

### Task 7: 前端接入单一自定义状态

**Files:**
- Modify: `frontend/miniprogram/pages/compose/{index.js,index.wxml,index.wxss}`
- Modify: `frontend/miniprogram/pages/moment-edit/{index.js,index.wxml,index.wxss}`
- Modify: `frontend/miniprogram/store/draft.js`
- Modify: `frontend/miniprogram/utils/{compose-guard,runes}.js`
- Modify: `frontend/miniprogram/components/moment-card/*`
- Test: `frontend/miniprogram/tests/{draft,compose-guard,runes,moment-actions}.test.js`

**Interfaces:**
- Consumes: Moment API `status: string` from Task 6.
- Produces: 快捷状态与自定义文本共用的单一 draft/payload `status` 字段。

- [ ] **Step 1: 先把前端测试期望改为展示文本**

  将 `miss_you/tired/in_class` 期望改为 `想你了/有点累/在上课`；新增快捷与自定义互斥、草稿恢复、20/21 rune 边界测试。

- [ ] **Step 2: 确认旧实现测试失败**

  Run: `cd frontend && npm test`
  Expected: FAIL on old status codes and missing custom input behavior.

- [ ] **Step 3: 实现发布/编辑状态互斥与原文展示**

  选快捷项直接写中文到 `status`；输入自定义值清空快捷选中；切回快捷值替换自定义值；发布按钮对 21 rune 禁用并显示明确文案。

- [ ] **Step 4: 运行全部前端测试**

  Run: `cd frontend && npm test`
  Expected: PASS，且 `rg "in_class|miss_you|tired" frontend/miniprogram --glob '!tests/*'` 无结果。

- [ ] **Step 5: Commit**

  `git add frontend/miniprogram ':!frontend/miniprogram/config.js' && git commit -m "feat: support one custom moment status"`

### Task 8: 迁移 Companionship 领域

**Files:**
- Create: `backend/internal/companionship/{domain,application,infrastructure/mysql,infrastructure/wechat,interfaces/http}/*.go`
- Modify: `backend/internal/bootstrap/companionship.go`
- Move current companionship tests beside new packages.
- Delete: old global companionship packages and `backend/internal/infrastructure/notification/wechat`.

**Interfaces:**
- Consumes: consumer-defined `ActiveCoupleReader`, `platform.clock.Clock`.
- Produces: `UpdatePresence`, `GetSnapshot`, `SendNudge`, `GetNotificationPreferences`, `UpdateNotificationPreferences` and the same presence/nudge/preferences routes.

- [ ] **Step 1: 迁移测试并补齐领域边界**

  断言过期 presence 不生效、nudge 只能发给 partner、冷却限制不变，Moment status/reaction/perspective 不得进入本领域。

- [ ] **Step 2: 迁移实现并把 `NotificationPreferencesService` 拆成明确用例**

  MySQL 实现命名为 `MySQLPresenceStore`、`MySQLNudgeStore`、`MySQLNotificationPreferencesStore`，不再暴露一个总 `Repository`。

- [ ] **Step 3: 切换 bootstrap 并删除旧包**

  Run: `cd backend && go test ./internal/companionship/... ./internal/bootstrap ./internal/architecture -count=1`
  Expected: PASS.

- [ ] **Step 4: Commit**

  `git add backend/internal && git commit -m "refactor: isolate companionship context"`

### Task 9: 拆分 Anniversary、Todo 与 Wish 领域

**Files:**
- Create: `backend/internal/anniversary/{domain,application,infrastructure/mysql,interfaces/http}/*.go`
- Create: `backend/internal/todo/{domain,application,infrastructure/mysql,interfaces/http}/*.go`
- Create: `backend/internal/wish/{domain,application,infrastructure/mysql,interfaces/http}/*.go`
- Modify: `backend/internal/bootstrap/{application,together}.go`
- Move related together tests into each new context.

**Interfaces:**
- Consumes: each context owns an `ActiveCoupleReader`; Anniversary owns a narrow `CoverStorage`.
- Produces: existing anniversary/todo/wish HTTP routes with separate handlers; `anniversary/application.PinnedAnniversaryReader` for Moment home highlights.

- [ ] **Step 1: 按三个领域拆分现有 together 测试**

  Anniversary 覆盖置顶、循环日期、相框封面；Todo 覆盖完成/重开/删除权；Wish 覆盖完成与 linked moment。

- [ ] **Step 2: 确认三个新包的测试先失败**

  Run: `cd backend && go test ./internal/anniversary/... ./internal/todo/... ./internal/wish/... -count=1`
  Expected: FAIL due to missing implementations.

- [ ] **Step 3: 迁移 domain/application/MySQL，把总 `Service` 拆为具名用例**

  例如 `CreateAnniversary`、`ListTodos`、`CompleteWish`；三个领域不共享一个 Together repository。

- [ ] **Step 4: 拆 HTTP handler 并更新 bootstrap**

  Run: `cd backend && go test ./internal/anniversary/... ./internal/todo/... ./internal/wish/... ./internal/moment/... ./internal/bootstrap -count=1`
  Expected: PASS，现有 API 路径不变。

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: split shared-life contexts"`

### Task 10: 迁移 Daily Question 并锁定原子投影

**Files:**
- Create: `backend/internal/dailyquestion/domain/*.go`
- Create: `backend/internal/dailyquestion/application/{get_daily_question,submit_daily_answer,list_daily_questions,question_context,ports}.go`
- Create: `backend/internal/dailyquestion/infrastructure/mysql/{question_store,answer_submission_store}.go`
- Create: `backend/internal/dailyquestion/infrastructure/ai/{fallback_questions,openai_compatible}.go`
- Create: `backend/internal/dailyquestion/interfaces/http/{handler,dto}.go`
- Modify: `backend/internal/bootstrap/together.go`
- Test: `backend/internal/dailyquestion/infrastructure/mysql/answer_submission_store_test.go`
- Delete: remaining `backend/internal/{domain,application}/together`, `backend/internal/infrastructure/{ai,persistence/mysql/together}`, `backend/internal/interfaces/http/together`

**Interfaces:**
- Consumes: Daily Question application 自定义的 `ActiveCoupleReader`；Moment 对外暴露的状态/文本校验函数，不导入 Moment infrastructure。
- Produces: `application.AnswerSubmissionStore.Submit(ctx context.Context, command SubmitAnswerTransaction) (domain.DailyQuestion, error)`; unchanged daily-question routes.

- [ ] **Step 1: 先写 AnswerSubmissionStore 事务测试**

  用 sqlmock 断言：保存答案、双方已答检查、`daily_question_id` 唯一投影插入在同一 transaction；任一 SQL 失败必须 rollback；重试不产生第二条 Moment。

- [ ] **Step 2: 确认新事务测试失败**

  Run: `cd backend && go test ./internal/dailyquestion/infrastructure/mysql -run AnswerSubmission -count=1`
  Expected: FAIL due to missing store.

- [ ] **Step 3: 迁移每日一问和 AI 适配器**

  `SubmitDailyAnswer` 负责编排，`MySQLAnswerSubmissionStore` 是唯一允许跨 daily question 表和 Moment 投影行的适配器，不暴露通用 Moment repository。

- [ ] **Step 4: 删除最后的 Together 旧包并验证**

  Run: `cd backend && go test ./internal/dailyquestion/... ./internal/architecture ./internal/bootstrap -count=1`
  Expected: PASS，`rg 'internal/(domain|application)/together|interfaces/http/together|persistence/mysql/together' backend --glob '*.go'` 无结果。

- [ ] **Step 5: Commit**

  `git add backend/internal && git commit -m "refactor: isolate daily question workflow"`

### Task 11: 用干净基线重建数据库

**Files:**
- Create: `backend/migrations/000001_baseline.up.sql`
- Create: `backend/migrations/000001_baseline.down.sql`
- Modify: `backend/cmd/migrate/main.go`
- Modify: all new MySQL adapters that still use `status_code` or `kind`.
- Test: `backend/internal/architecture/migration_test.go`
- Delete: `backend/migrations/000001_init.*` through `000004_personal_spaces_reflections.*`

**Interfaces:**
- Consumes: Task 1 audit and all migrated contexts' actual SQL queries.
- Produces: tracked migration runner using `schema_migrations`; Moment column `status VARCHAR(80) NOT NULL DEFAULT ''`; no `kind` and no `notification_deliveries`.

- [ ] **Step 1: 把 migration 测试改为真实空库 up/down/up**

  测试使用独立测试库，执行基线后逐项查 `information_schema`；实际插入 account/couple/moment/daily answer 最小数据，验证外键、唯一投影、状态长度和分页索引。

- [ ] **Step 2: 确认旧迁移无法满足新断言**

  Run: `cd backend && go test ./internal/architecture -run Migration -count=1`
  Expected: FAIL on `status`/removed `kind`/migration tracking assertions.

- [ ] **Step 3: 写新 baseline 和幂等迁移记录**

  Runner 默认执行所有未应用版本，事务内记录 `schema_migrations`；down 按外键逆序删表。

- [ ] **Step 4: 切换 SQL 到 `status`，删除 `kind` 读写**

  Daily question Moment 语义通过 `daily_question_id IS NOT NULL` 推导；旧 `status_code` 不得留在 Go/SQL/测试中。

- [ ] **Step 5: 删除旧迁移、清空并重建本地库**

  先校验目标 DSN 的 database name 精确为本地 `just_for_you`，再 drop/create 该单库，不使用宽泛环境变量或 glob 作删除目标。

- [ ] **Step 6: 验证基线与全部 MySQL 测试**

  Run: `cd backend && go test ./internal/architecture ./internal/... -count=1`
  Expected: PASS，`rg 'status_code|notification_deliveries|moments.*kind|kind.*moments' backend --glob '*.{go,sql}'` 无结果。

- [ ] **Step 7: Commit**

  `git add backend/migrations backend/cmd/migrate backend/internal && git commit -m "refactor: replace migrations with clean baseline"`

### Task 12: 收口 Bootstrap、文档与端到端验证

**Files:**
- Modify: `backend/internal/bootstrap/*.go`
- Modify: `backend/internal/interfaces/http/shared/router.go` or move it to `backend/internal/bootstrap/router.go`
- Modify: `backend/README.md` or repository `README.md` according to the existing canonical location.
- Create: `backend/docs/architecture/domain-map.md`
- Create: `backend/docs/architecture/tracing-a-request.md`
- Create: `backend/docs/architecture/adding-a-use-case.md`
- Test: `backend/internal/bootstrap/application_test.go`
- Test: `backend/internal/architecture/dependencies_test.go`

**Interfaces:**
- Consumes: all contexts and platform packages from Tasks 2–11.
- Produces: bootstrap as the only composition root, final maintainability documentation, runnable clean-database application.

- [ ] **Step 1: 写最终目录和路由装配失败测试**

  断言每个领域 handler 只注册一次、所有 Task 1 审计中 keep 的路由存在、全局 legacy 目录和 `legacyRoots` 白名单为空。

- [ ] **Step 2: 简化 bootstrap 装配**

  每个 context 只暴露一个组装函数和 HTTP handler；bootstrap 是唯一实例化 MySQL/COS/WeChat/AI 具体实现的位置。

- [ ] **Step 3: 编写维护文档**

  `domain-map.md` 写清边界和跨域端口；`tracing-a-request.md` 用“发布一条小事”展示 HTTP -> application -> domain -> MySQL；`adding-a-use-case.md` 给出新增用例/字段/迁移的检查清单。

- [ ] **Step 4: 运行格式化和后端静态/并发验证**

  Run: `cd backend && gofmt -w cmd internal && go test ./... -count=1 && go test -race ./... && go vet ./...`
  Expected: all PASS with no vet findings.

- [ ] **Step 5: 运行前端全量测试**

  Run: `cd frontend && npm test`
  Expected: all PASS.

- [ ] **Step 6: 用空库做 API smoke test**

  启动 API 后依次验证 `/healthz`、微信登录的可控 fake/provider 路径、单人测试绑定、发布共同/个人小事、自定义状态、reaction、perspective、纪念日、每日一问。

- [ ] **Step 7: 检查工作区与秘密泄漏**

  Run: `git status --short && git diff --check && git diff --cached --check`
  Expected: 只保留用户本地的 `backend/scripts/dev.sh` 和 `frontend/miniprogram/config.js` 未提交改动；无密钥、DSN 密码或云凭据进入 diff。

- [ ] **Step 8: Commit**

  `git add README.md backend/README.md backend/docs backend/internal ':!backend/scripts/dev.sh' && git commit -m "docs: explain vertical slice architecture"`
