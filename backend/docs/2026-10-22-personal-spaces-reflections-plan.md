# Personal Spaces and Shared Reflections Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task with review checkpoints.

**Goal:** 在不破坏现有共同小事时间线的前提下，交付个人空间、分享与回应、双视角回忆、每日一句、回忆浮现和统一手账动效。

**Architecture:** 延续全局分层 DDD：业务规则放在 `domain`，用例和端口放在 `application`，MySQL/COS/AI 实现放在 `infrastructure`，HTTP 只做协议转换，`bootstrap` 负责装配。空间访问统一通过应用层访问策略决策，仓储查询再使用明确的 owner/relationship 条件收窄数据，避免只依赖前端过滤。

**Tech Stack:** Go 1.23、Gin、MySQL 8、golang-migrate、sqlmock、微信原生小程序、TDesign MiniProgram、Node.js 20 `node:test`、OpenAI-compatible HTTP API。

**Spec:** `backend/docs/2026-10-22-personal-spaces-reflections-design.md`

## Global Constraints

- 严格保持 `interfaces → application → domain`，基础设施只实现 application 端口；执行 `go test ./internal/architecture` 验证。
- 现有共同小事全部迁移为 `shared + couple + standard`，旧记录、照片、发生时间和作者不可丢失。
- 所有读取、编辑、删除、图片签名、回应和双视角操作必须调用同一套可见性策略；前端过滤不能作为权限边界。
- `private` 内容、图片 URL、精确经纬度、令牌和密钥不得进入 AI 上下文或日志。
- 每个任务遵循 Red → Green → Refactor：先写一个会因缺少行为而失败的测试，再实现最小代码，再跑相关测试和全量测试。
- 保留工作区已有但与本计划无关的 `backend/scripts/dev.sh` 和 `frontend/miniprogram/config.js` 本地修改；每次提交只暂存本任务文件。
- 现有共同时间线在每个提交后都必须仍可发布、读取、编辑、删除和上传照片。

## Review Focus

- 访问权限是否在解绑、重新绑定、陌生用户和单人测试模式下仍正确。
- 数据库唯一约束与事务是否能抵御重复点击和并发首开。
- AI 降级是否在 3 秒内结束且永不阻塞首页。
- 新接口是否保持稳定错误码，前端是否给出可操作中文提示。
- 动效是否遵守 `reduceMotion`，并且不影响滚动、点击和低端真机性能。

---

## Task 1: 迁移数据模型并建立空间领域不变量

**Files:**

- Create: `backend/migrations/000004_personal_spaces_reflections.up.sql`
- Create: `backend/migrations/000004_personal_spaces_reflections.down.sql`
- Modify: `backend/internal/architecture/migration_test.go`
- Create: `backend/internal/domain/moment/space.go`
- Create: `backend/internal/domain/moment/space_test.go`
- Modify: `backend/internal/domain/moment/moment.go`
- Modify: `backend/internal/domain/moment/moment_test.go`
- Modify: `backend/internal/domain/moment/errors.go`

- [ ] **Step 1: 写迁移形状和数据保留失败测试**

  在 `migration_test.go` 增加 `TestPersonalSpacesMigrationShape` 和可连接 MySQL 时运行的 `TestPersonalSpacesMigrationPreservesExistingMoments`。先插入一条旧 Moment 与照片，应用迁移后断言它变为 `shared/couple/standard`，作者、照片、`occurred_at` 不变；回滚后新增表与列消失。

- [ ] **Step 2: 运行迁移测试并确认失败**

  Run: `cd backend && go test ./internal/architecture -run 'TestPersonalSpaces' -count=1`

  Expected: FAIL，提示缺少 `000004` 文件或字段。

- [ ] **Step 3: 实现可回滚迁移**

  `moments` 新增 `space_type`、`visibility`、`kind`、`source_moment_id`、`daily_question_id`；新增 `moment_perspectives`、`moment_perspective_photos`、`moment_reactions`、`daily_questions`、`daily_answers`。先用默认值回填历史数据，再建立索引、外键、唯一约束和检查约束。Down migration 按外键依赖逆序删除。

- [ ] **Step 4: 写领域构造失败测试**

  在 `space_test.go` 覆盖：`shared+couple` 合法，`personal+private/partner` 合法，其余组合返回 `ErrInvalidSpaceVisibility`；`daily_question` 只能是共同空间并关联问题；分享副本必须有 `sourceMomentID`。

- [ ] **Step 5: 实现空间值对象并扩展 Moment**

  新增 `SpaceType`、`Visibility`、`Kind` 常量和 `Valid`/组合校验。把 `NewMomentDetails` 改为接收一个 `CreateInput`，其中包含当前字段以及 `SpaceType`、`Visibility`、`Kind`、可空来源 ID/问题 ID；`RestoreMomentDetails` 同步恢复新字段，并提供只读 getter。

- [ ] **Step 6: 跑领域、迁移和架构测试**

  Run: `cd backend && go test ./internal/domain/moment ./internal/architecture -count=1`

  Expected: PASS（无 MySQL 时集成迁移测试允许按现有约定 Skip）。

- [ ] **Step 7: 提交**

  ```bash
  git add backend/migrations/000004_personal_spaces_reflections.* backend/internal/domain/moment backend/internal/architecture/migration_test.go
  git commit -m "feat: add personal space domain model"
  ```

## Task 2: 建立统一访问策略与三个空间查询

**Files:**

- Create: `backend/internal/application/moment/access_policy.go`
- Create: `backend/internal/application/moment/access_policy_test.go`
- Modify: `backend/internal/application/moment/ports.go`
- Modify: `backend/internal/application/moment/get_timeline.go`
- Modify: `backend/internal/application/moment/get_home.go`
- Modify: `backend/internal/application/moment/read_models.go`
- Modify: `backend/internal/application/moment/use_cases_test.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `backend/internal/interfaces/http/moment/handler.go`
- Modify: `backend/internal/interfaces/http/moment/handler_test.go`

- [ ] **Step 1: 写访问矩阵失败测试**

  为 `AccessPolicy.CanRead(moment, viewerID, activeRelationship)`、`CanMutate`、`CanInteract` 建表驱动测试：作者、当前伴侣、陌生人、解绑后、重新绑定后、shared/private/partner 全组合；单人测试只把本人当作者，不伪造伴侣权限。

- [ ] **Step 2: 实现纯应用层访问策略**

  策略仅接收领域对象与当前关系快照，不调用数据库。返回稳定错误 `ErrMomentNotVisible`、`ErrPartnerRequired`，供所有后续用例复用。

- [ ] **Step 3: 将仓储列表端口改为显式查询对象**

  在 `ports.go` 定义 `TimelineFilter{ViewerID, RelationshipID, OwnerID, SpaceType}` 和 `MomentLookup{ViewerID, RelationshipID, MomentID}`；把 `List`、`Recent`、`Find` 改为接收查询对象，避免含义不清的连续 `int64` 参数。

- [ ] **Step 4: 写 MySQL 查询失败测试**

  在 repository sqlmock 测试中分别断言：shared 只按当前 `couple_id`；personal/me 按 `author_user_id`；personal/partner 同时要求当前 partner 作者和 `visibility='partner'`；详情查询包含同等范围条件。

- [ ] **Step 5: 实现仓储扫描与三个空间查询**

  扩展 SELECT/INSERT/UPDATE/scan 处理新字段；保持 `(occurred_at,id)` 游标。`Recent` 只服务 shared 首页；日历、地图、月报继续只统计 shared 空间，SQL 显式增加 `space_type='shared'`。

- [ ] **Step 6: 扩展 timeline HTTP 契约**

  `GET /v1/moments` 解析 `space=shared|personal` 与 `owner=me|partner`。缺省保持 `space=shared`；非法组合返回 `INVALID_SPACE_QUERY`。DTO 增加 `spaceType`、`visibility`、`kind`、`sourceMomentId`、`dailyQuestionId`。

- [ ] **Step 7: 跑相关测试**

  Run: `cd backend && go test ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -count=1`

  Expected: PASS。

- [ ] **Step 8: 提交**

  ```bash
  git add backend/internal/application/moment backend/internal/infrastructure/persistence/mysql/moment backend/internal/interfaces/http/moment
  git commit -m "feat: query shared and personal moment spaces"
  ```

## Task 3: 打通按空间发布、修改与前端空间入口

**Files:**

- Modify: `backend/internal/application/moment/publish_moment.go`
- Modify: `backend/internal/application/moment/update_moment.go`
- Modify: `backend/internal/application/moment/delete_moment.go`
- Modify: `backend/internal/application/moment/use_cases_test.go`
- Modify: `backend/internal/interfaces/http/moment/handler.go`
- Modify: `backend/internal/interfaces/http/moment/handler_test.go`
- Modify: `frontend/miniprogram/services/moments.js`
- Modify: `frontend/miniprogram/store/timeline.js`
- Modify: `frontend/miniprogram/store/draft.js`
- Create: `frontend/miniprogram/components/space-switcher/index.{js,json,wxml,wxss}`
- Create: `frontend/miniprogram/components/visibility-picker/index.{js,json,wxml,wxss}`
- Create: `frontend/miniprogram/components/privacy-badge/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/compose/index.{js,json,wxml,wxss}`
- Create: `frontend/miniprogram/pages/personal-moments/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.json`
- Modify: `frontend/miniprogram/app.json`
- Modify: `frontend/miniprogram/utils/compose-guard.js`
- Modify: `frontend/miniprogram/tests/compose-guard.test.js`
- Create: `frontend/miniprogram/tests/space-model.test.js`

- [ ] **Step 1: 写后端发布/修改权限失败测试**

  覆盖默认 shared、个人 private、个人 partner、非法组合、没有伴侣时 shared/partner 被拒绝但 private 可发布、只有作者可修改可见性、共同记录不可改成个人记录。

- [ ] **Step 2: 实现用例和 HTTP 字段**

  `PublishMomentCommand`、`UpdateMomentCommand` 增加 `SpaceType`、`Visibility`。个人 private 发布不要求活跃关系；shared/partner 在保存事务前后都校验当前关系。HTTP 请求与错误映射增加稳定错误码。

- [ ] **Step 3: 写前端空间状态失败测试**

  新建纯函数 `normalizeSpaceSelection`、`publishButtonLabel`、`timelineQueryKey`；断言发布页每次新打开默认 shared，切 personal 默认 private，三个 feed 的分页状态不会串页，草稿恢复保留显式空间。

- [ ] **Step 4: 实现服务与隔离 store**

  `momentService.getTimeline({space,owner,cursor,limit})` 传递查询参数。`createTimelineStore(service, initialFilter)` 把 filter 纳入 state/generation，切换空间时清空游标和列表。

- [ ] **Step 5: 实现发布选择器和个人空间页**

  Compose 顶部加入 `space-switcher`；仅 personal 显示 `visibility-picker`；按钮文案随选择变化。首页增加“我们的小事 / 我的小事”，伴侣头像进入 `owner=partner`；个人页根据 owner 显示“我的小事 / TA 的小事”，卡片显示 `privacy-badge`。

- [ ] **Step 6: 跑后端与前端测试**

  Run: `cd backend && go test ./internal/application/moment ./internal/interfaces/http/moment -count=1`

  Run: `cd frontend && npm test`

  Expected: PASS。

- [ ] **Step 7: 提交**

  ```bash
  git add backend/internal/application/moment backend/internal/interfaces/http/moment frontend/miniprogram
  git commit -m "feat: publish and browse personal moments"
  ```

## Task 4: 实现“分享给我们”和轻量回应

**Files:**

- Create: `backend/internal/domain/moment/reaction.go`
- Create: `backend/internal/domain/moment/reaction_test.go`
- Create: `backend/internal/application/moment/share_to_couple.go`
- Create: `backend/internal/application/moment/react_to_moment.go`
- Modify: `backend/internal/application/moment/ports.go`
- Modify: `backend/internal/application/moment/use_cases_test.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `backend/internal/interfaces/http/moment/handler.go`
- Modify: `backend/internal/interfaces/http/moment/handler_test.go`
- Modify: `backend/internal/bootstrap/moment.go`
- Modify: `frontend/miniprogram/services/moments.js`
- Create: `frontend/miniprogram/components/reaction-bar/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/components/moment-card/index.{js,json,wxml,wxss}`
- Create: `frontend/miniprogram/tests/moment-actions.test.js`

- [ ] **Step 1: 写领域回应失败测试**

  覆盖预设类型、最多 120 Unicode 字短文本、未知类型、空自定义文本；同一用户同种预设回应定义为幂等 upsert。

- [ ] **Step 2: 写用例权限与复制语义失败测试**

  分享仅允许个人记录作者且要求活跃关系；生成新 `shared/couple/standard` Moment 并保留 `sourceMomentID`，不改变源记录；回应前必须 `CanRead`，只能撤回自己的回应。

- [ ] **Step 3: 实现事务仓储端口**

  增加 `ShareToCouple`、`UpsertReaction`、`DeleteReaction`，在事务内锁定源 Moment 并复核关系。shared 副本复用源照片 object key，不重复上传；删除源记录不级联删除已生成的共同副本，只允许 `source_moment_id` 置空或受保护保留。

- [ ] **Step 4: 注册 HTTP 路由**

  实现 `POST /v1/moments/:id/share-to-couple`、`POST /v1/moments/:id/reactions`、`DELETE /v1/moments/:id/reactions/:reactionId`，DTO 返回聚合后的回应以及 `reactedByMe`。

- [ ] **Step 5: 实现前端操作与乐观回滚**

  个人本人卡片显示“分享给我们”；可见卡片显示 reaction bar。点击预设回应先更新本地计数，请求失败时回滚并提示；短文本使用轻量弹层且前端同样限制 120 字。

- [ ] **Step 6: 验证并提交**

  Run: `cd backend && go test ./internal/domain/moment ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -count=1`

  Run: `cd frontend && npm test`

  ```bash
  git add backend/internal/domain/moment backend/internal/application/moment backend/internal/infrastructure/persistence/mysql/moment backend/internal/interfaces/http/moment backend/internal/bootstrap/moment.go frontend/miniprogram
  git commit -m "feat: share and react to little moments"
  ```

## Task 5: 实现共同小事双视角与三图上传

**Files:**

- Create: `backend/internal/domain/moment/perspective.go`
- Create: `backend/internal/domain/moment/perspective_test.go`
- Create: `backend/internal/application/moment/save_perspective.go`
- Modify: `backend/internal/application/moment/ports.go`
- Modify: `backend/internal/application/moment/use_cases_test.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `backend/internal/interfaces/http/moment/handler.go`
- Modify: `backend/internal/interfaces/http/moment/handler_test.go`
- Modify: `backend/internal/bootstrap/moment.go`
- Modify: `frontend/miniprogram/services/moments.js`
- Create: `frontend/miniprogram/components/perspective-panel/index.{js,json,wxml,wxss}`
- Create: `frontend/miniprogram/pages/perspective-edit/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/components/moment-card/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/app.json`
- Create: `frontend/miniprogram/tests/perspective.test.js`

- [ ] **Step 1: 写 Perspective 不变量测试**

  断言只允许 shared Moment 的另一方添加；最多三张照片；文字长度与空内容规则明确；只有 Perspective 作者能更新或删除。

- [ ] **Step 2: 写用例和仓储并发测试**

  `SavePerspectiveCommand` 复用上传授权校验和 COS 对象验证。唯一键 `(moment_id, author_user_id)` 保证重复提交更新同一份视角；事务同时消费照片 grant，关系解绑后拒绝写入。

- [ ] **Step 3: 实现 API 与聚合 DTO**

  添加 `PUT/DELETE /v1/moments/:id/perspective`。Moment DTO 带可空 `perspective`，其照片读取 URL 与原 Moment 在一次签名流程中生成。

- [ ] **Step 4: 实现前端编辑页和组合卡**

  Perspective 上传队列复用现有压缩与上传模块，但 hard limit 为 3。卡片按原记录 + TA 视角组合，只有当前 Perspective 作者看到编辑/删除入口。

- [ ] **Step 5: 验证并提交**

  Run: `cd backend && go test ./internal/domain/moment ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -count=1`

  Run: `cd frontend && npm test`

  ```bash
  git add backend/internal frontend/miniprogram
  git commit -m "feat: add dual perspective memories"
  ```

## Task 6: 交付每日一句领域、题库降级与双向解锁

**Files:**

- Create: `backend/internal/domain/together/daily_question.go`
- Create: `backend/internal/domain/together/daily_question_test.go`
- Create: `backend/internal/application/together/daily_question.go`
- Modify: `backend/internal/application/together/ports.go`
- Modify: `backend/internal/application/together/use_cases_test.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/together/repository.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/together/repository_test.go`
- Create: `backend/internal/infrastructure/ai/fallback_questions.go`
- Create: `backend/internal/infrastructure/ai/fallback_questions_test.go`
- Modify: `backend/internal/interfaces/http/together/handler.go`
- Modify: `backend/internal/interfaces/http/together/handler_test.go`
- Modify: `backend/internal/bootstrap/together.go`
- Modify: `frontend/miniprogram/services/together.js`
- Create: `frontend/miniprogram/store/daily-question.js`
- Create: `frontend/miniprogram/components/daily-question-card/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.json`
- Create: `frontend/miniprogram/pages/daily-questions/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/app.json`
- Create: `frontend/miniprogram/tests/daily-question.test.js`

- [ ] **Step 1: 写领域解锁失败测试**

  覆盖：每天每段关系一道题；一方答案隐藏对方字段；解锁前本人可改；双方提交后冻结；第二份答案使问题解锁并产生一次 `daily_question` shared Moment；历史题可补答。

- [ ] **Step 2: 定义端口与事务用例**

  在 application 定义 `QuestionGenerator.Generate(ctx, QuestionContext) (GeneratedQuestion,error)`、`DailyQuestionRepository`。`GetToday` 使用上海自然日，先查缓存，缺失时生成并依赖唯一键处理并发；`SubmitAnswer` 在一个事务中 upsert、判断解锁、创建系统 Moment。

- [ ] **Step 3: 实现内置题库**

  提供经过固定 seed/日期选择的题库，禁止诊断、攻击性或只有一方能回答的问题；生成器不可用时立即返回 `source=fallback`。

- [ ] **Step 4: 实现 HTTP 隐藏规则**

  注册 `GET /v1/daily-question`、`PUT /v1/daily-question/answer`、`GET /v1/daily-questions`。未解锁 DTO 不序列化 partner answer；解锁后同时返回两份答案。

- [ ] **Step 5: 实现前端状态卡与历史页**

  首页卡片支持未答、等待、解锁三态；提交后失效并刷新 shared feed。历史页分页展示，未解锁问题只显示本人的状态。

- [ ] **Step 6: 验证并提交**

  Run: `cd backend && go test ./internal/domain/together ./internal/application/together ./internal/infrastructure/persistence/mysql/together ./internal/infrastructure/ai ./internal/interfaces/http/together -count=1`

  Run: `cd frontend && npm test`

  ```bash
  git add backend/internal frontend/miniprogram
  git commit -m "feat: add daily question unlock flow"
  ```

## Task 7: 接入隐私安全的 OpenAI-compatible 个性化生成

**Files:**

- Modify: `backend/.env.example`
- Modify: `backend/internal/infrastructure/config/config.go`
- Modify: `backend/internal/infrastructure/config/config_test.go`
- Create: `backend/internal/application/together/question_context.go`
- Create: `backend/internal/application/together/question_context_test.go`
- Create: `backend/internal/infrastructure/ai/openai_compatible.go`
- Create: `backend/internal/infrastructure/ai/openai_compatible_test.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/together/repository.go`
- Modify: `backend/internal/bootstrap/together.go`
- Modify: `backend/internal/application/companionship/notification_preferences.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/companionship/notification_repository.go`
- Modify: `backend/internal/interfaces/http/companionship/handler.go`

- [ ] **Step 1: 写上下文隐私测试**

  构造 shared、partner、private Moment 和带照片/经纬度的数据，断言 `QuestionContextBuilder` 只输出允许文本的主题、宽泛情绪、相对时间和非精确场景，不包含 private 文本、object key、URL、经纬度、用户凭证。

- [ ] **Step 2: 写 AI 客户端失败测试**

  用 `httptest.Server` 覆盖成功、超过 3 秒、非 2xx、空输出、多问题、过长、敏感/诊断性输出；失败统一返回可分类但不含原始上下文的错误，调用方随后使用 fallback。

- [ ] **Step 3: 扩展可选配置**

  `Config` 增加 `AIBaseURL`、`AIAPIKey`、`AIModel`；三项任一缺失不阻止服务启动，而是装配 fallback-only。真实 key 只来自环境变量，不写入仓库。

- [ ] **Step 4: 实现生成器与降级组合器**

  OpenAI-compatible 客户端使用注入的 `http.Client` 和 context timeout；请求日志只记录 provider 状态类别。组合器优先 AI，任何错误/无效输出同请求内落到题库。

- [ ] **Step 5: 接入个性化开关**

  在现有 notification preferences 增加 `aiPersonalizationEnabled`，默认开启；关闭时不构建/发送 AI 上下文，直接题库。

- [ ] **Step 6: 验证并提交**

  Run: `cd backend && go test ./internal/application/together ./internal/infrastructure/ai ./internal/infrastructure/config ./internal/application/companionship ./internal/infrastructure/persistence/mysql/companionship ./internal/interfaces/http/companionship -count=1`

  ```bash
  git add backend/.env.example backend/internal
  git commit -m "feat: generate privacy-safe daily questions"
  ```

## Task 8: 实现“那天的我们”回忆浮现

**Files:**

- Create: `backend/internal/application/moment/get_resurfaced_memory.go`
- Create: `backend/internal/application/moment/get_resurfaced_memory_test.go`
- Modify: `backend/internal/application/moment/ports.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `backend/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `backend/internal/interfaces/http/moment/handler.go`
- Modify: `backend/internal/interfaces/http/moment/handler_test.go`
- Modify: `backend/internal/bootstrap/moment.go`
- Modify: `frontend/miniprogram/services/moments.js`
- Create: `frontend/miniprogram/components/memory-resurface-card/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.{js,json,wxml,wxss}`
- Modify: `frontend/miniprogram/pages/home/index.json`
- Create: `frontend/miniprogram/tests/memory-resurface.test.js`

- [ ] **Step 1: 写候选排序和权限失败测试**

  固定 clock，覆盖去年今日、一个月前、首次共同记录、相同地点、纪念日前后；定义确定性优先级并去重。只返回当前活跃关系 shared Moment，绝不返回 personal/private 或旧关系数据；无候选返回空而非错误。

- [ ] **Step 2: 实现仓储候选查询和用例**

  Repository 只取小范围候选与必要字段，应用层打分并选一条；照片签名复用现有逻辑。接口 `GET /v1/memories/resurface` 返回可空 `memory` 与 `reason`。

- [ ] **Step 3: 实现首页回忆卡**

  首页并行请求但不让回忆失败影响主内容；无候选不渲染。卡片可进入原 Moment 所在时间线位置或详情视图，并显示“去年今天 / 一个月前 / 我们的第一件小事”等原因。

- [ ] **Step 4: 验证并提交**

  Run: `cd backend && go test ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -count=1`

  Run: `cd frontend && npm test`

  ```bash
  git add backend/internal frontend/miniprogram
  git commit -m "feat: resurface shared memories"
  ```

## Task 9: 统一手账漂浮组件动效与减少动态效果

**Files:**

- Modify: `frontend/miniprogram/styles/tokens.wxss`
- Modify: `frontend/miniprogram/styles/animations.wxss`
- Modify: `frontend/miniprogram/app.wxss`
- Modify: `frontend/miniprogram/components/moment-card/index.wxss`
- Modify: `frontend/miniprogram/components/daily-question-card/index.wxss`
- Modify: `frontend/miniprogram/components/memory-resurface-card/index.wxss`
- Modify: `frontend/miniprogram/components/perspective-panel/index.wxss`
- Create: `frontend/miniprogram/utils/motion.js`
- Create: `frontend/miniprogram/tests/motion.test.js`
- Modify: `frontend/miniprogram/pages/home/index.js`
- Modify: `frontend/miniprogram/pages/timeline/index.js`
- Modify: `frontend/miniprogram/pages/personal-moments/index.js`

- [ ] **Step 1: 写动效降级失败测试**

  纯函数根据全局 `reduceMotion` 生成 motion class/style：正常模式允许轻微位移、旋转和贴纸弹跳；减少动态效果时禁止循环、旋转和弹跳，只允许 `<=160ms` 淡入。列表 stagger 必须有上限，避免长列表延迟。

- [ ] **Step 2: 扩展统一 token**

  增加 easing、浮动距离、旋转角度、纸层阴影、拍立得边框和 stagger token。组件不得私自定义另一套时长；交互态继续保留清晰焦点/按压反馈。

- [ ] **Step 3: 应用手账视觉**

  Moment、每日问题、回忆和双视角卡使用受控的奇偶轻旋转、纸张层和照片边框；新发布成功用一次性贴纸弹跳。动画只作用于 transform/opacity，不动画布局属性。

- [ ] **Step 4: 验证前端测试和静态检查**

  Run: `cd frontend && npm test`

  Run: `cd frontend && npm ls --depth=0`

  Expected: PASS，无额外未声明依赖。

- [ ] **Step 5: 提交**

  ```bash
  git add frontend/miniprogram
  git commit -m "feat: add scrapbook motion system"
  ```

## Task 10: 全量回归、真实迁移与手工验收准备

**Files:**

- Modify: `README.md`
- Modify: `backend/docs/2026-10-22-personal-spaces-reflections-design.md`
- Modify: tests found deficient during review only

- [ ] **Step 1: 全量自动化测试**

  Run: `cd backend && go test ./... -count=1`

  Run: `cd frontend && npm test`

  Expected: 全部 PASS；MySQL 不可用时只有明确的迁移集成测试 Skip，不允许普通单元测试 Skip。

- [ ] **Step 2: 真实 MySQL 迁移往返**

  使用独立临时数据库执行 `000001` 至 `000004`，插入旧格式数据后升级，核对 Moment/照片数量与关键字段；再在另一个临时库验证完整 down 顺序。绝不在开发库直接做破坏性回滚。

- [ ] **Step 3: 启动后端并做 API smoke test**

  Run: `cd backend && ./scripts/dev.sh`

  在另一终端验证 `/healthz`，再用两个测试账号覆盖 shared/private/partner 发布、三个 feed、分享、回应、双视角、每日题解锁和回忆接口。AI 环境变量分别测试“未配置”和 mock provider 两种模式。

- [ ] **Step 4: 微信开发者工具与真机验收**

  覆盖九图发布、Perspective 三图、弱网重试、草稿恢复、返回键、解绑立即失权、重新绑定不继承旧关系、单人测试不越权，以及 reduce motion。记录任何真机差异并补自动化回归测试后修复。

- [ ] **Step 5: 自检规格覆盖与安全边界**

  对照设计文档逐项勾选；执行 `rg -n "TODO|FIXME|REPLACE_WITH|AI_API_KEY|COS_SECRET|WECHAT_APP_SECRET" backend frontend`，逐项确认命中只来自 `.env.example` 的变量名/示例值，不存在真实密钥或未完成实现；审查所有 Moment 读取 SQL 是否包含空间/owner/relationship 条件。

- [ ] **Step 6: 更新说明并最终提交**

  README 补充新 API、迁移、AI 可选配置、本地测试步骤和隐私边界；设计文档状态改为“已实现并验收”。

  ```bash
  git add README.md backend/.env.example backend/docs backend/internal backend/migrations frontend/miniprogram
  git status --short
  git commit -m "docs: finalize personal spaces release"
  ```

- [ ] **Step 7: 最终分支复核**

  Run: `git diff --check HEAD~10..HEAD`

  Run: `git status --short`

  Expected: diff 无空白错误；只剩用户原有的 `backend/scripts/dev.sh` 与 `frontend/miniprogram/config.js` 本地修改，不把它们误提交。
