# Just for You MVP 实施计划

> **供智能代理执行：** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`，逐项实施本计划。每一步使用复选框跟踪状态。

**目标：** 在七天内交付一个可供两位真实用户在微信真机上登录、绑定、分享最多九张照片的日常记录、查看共同时间线并解除关系的体验版。

**架构：** 原生微信小程序通过 HTTPS 调用单体 Go/Gin API。服务端使用 MySQL 保存业务数据，用腾讯云 COS 私有桶保存照片，并由 Nginx 提供 TLS 和反向代理。业务授权全部在服务层完成，HTTP 处理器不直接依赖关系 ID 判断权限。

**技术栈：** 原生微信小程序（WXML、WXSS、JavaScript）、Go 1.22+、Gin、`database/sql`、MySQL 8.0、腾讯云 COS、Nginx、Docker Compose（本地依赖与部署）。

**规格文档：** `docs/superpowers/specs/2026-10-18-just-for-you-mvp-design.md`

## 全局约束

- 单条记录最多九张照片，照片顺序必须保持不变；十张及以上必须由前端阻止、后端拒绝。
- 一条记录至少包含文字、快捷状态、地点名称或一张照片中的一项。
- 每个用户最多有一段有效情侣关系；邀请码为六位、短时有效、仅可使用一次。
- 时间线和照片访问权限必须由当前登录用户的有效情侣关系推导，不能信任客户端传入的关系 ID。
- 用户只能删除自己的记录；解除关系后双方立即失去历史记录访问权限。
- 不记录微信登录 `code`、访问令牌、刷新令牌原文或微信会话密钥。
- MVP 不实现聊天、评论、通知、精确定位、公开内容、支付及管理后台。

## 评审重点

- 两人同时使用同一邀请码：仅一个绑定事务成功，另一个收到 `INVITE_USED` 或 `ALREADY_BOUND`。
- 第十张照片及伪造的对象键：接口返回校验错误，不创建记录。
- 照片部分上传失败：前端保留表单和已选照片，只重试失败项，不创建残缺记录。
- 关系在翻页或删除请求前已解除：接口立即返回 `NO_ACTIVE_COUPLE`，不泄露旧数据。
- 登录令牌过期或刷新令牌被撤销：只允许重新认证，不出现无限刷新循环。

---

## 文件结构

```text
Just for you/
├── .env.example                    # 本地和服务器环境变量说明
├── Makefile                        # 开发、测试、迁移命令
├── docker-compose.yml              # 本地 MySQL 与 API
├── deploy/
│   ├── nginx.conf                  # HTTPS 反向代理模板
│   └── just-for-you.service        # Go 服务 systemd 单元
├── server/
│   ├── cmd/api/main.go             # 组装依赖并启动 HTTP 服务
│   ├── internal/config/config.go    # 环境变量解析与校验
│   ├── internal/httpapi/            # 路由、处理器、中间件、统一响应
│   ├── internal/auth/               # 微信交换、令牌与认证服务
│   ├── internal/user/               # 用户资料
│   ├── internal/couple/             # 邀请和情侣关系
│   ├── internal/moment/             # 记录、首页和时间线
│   ├── internal/storage/            # COS 上传/读取授权
│   ├── internal/testutil/           # 测试数据库和固定时钟
│   └── migrations/                  # MySQL 迁移
└── miniprogram/
    ├── app.js / app.json / app.wxss # 小程序入口与全局样式
    ├── config.js                    # API 基础地址
    ├── services/                    # HTTP、登录、情侣、记录、上传客户端
    ├── utils/                       # 纯函数：状态、照片队列、时间格式
    ├── pages/login/                 # 登录与资料
    ├── pages/bind/                  # 创建和输入邀请码
    ├── pages/home/                  # 最近状态与最近记录
    ├── pages/compose/               # 发布最多九张照片
    ├── pages/timeline/              # 共同时间线
    ├── pages/settings/              # 解除关系
    └── tests/                       # 可由 Node 执行的纯函数测试
```

### 任务 1：服务端骨架、配置与数据库迁移

**文件：**
- 创建：`server/go.mod`
- 创建：`server/cmd/api/main.go`
- 创建：`server/internal/config/config.go`
- 创建：`server/internal/config/config_test.go`
- 创建：`server/internal/httpapi/router.go`
- 创建：`server/internal/httpapi/response.go`
- 创建：`server/internal/httpapi/router_test.go`
- 创建：`server/migrations/000001_init.up.sql`
- 创建：`server/migrations/000001_init.down.sql`
- 创建：`.env.example`
- 创建：`docker-compose.yml`
- 创建：`Makefile`

**接口：**
- 产出：`config.Load() (config.Config, error)`，读取并校验 `HTTP_ADDR`、`MYSQL_DSN`、`JWT_SECRET`、`WECHAT_APP_ID`、`WECHAT_APP_SECRET`、COS 配置。
- 产出：`httpapi.NewRouter(httpapi.Dependencies) *gin.Engine`，初始提供 `GET /healthz`。
- 产出：统一响应 `{ "data": ... }` 与 `{ "error": { "code": string, "message": string } }`。

- [ ] **步骤 1：编写失败测试**

在 `config_test.go` 验证缺少 `MYSQL_DSN` 或短于 32 字节的 `JWT_SECRET` 时返回明确错误；在 `router_test.go` 验证 `GET /healthz` 返回 200 和 `{"data":{"status":"ok"}}`。

- [ ] **步骤 2：运行测试并确认失败**

运行：`cd server && go test ./internal/config ./internal/httpapi`

预期：因包和函数尚不存在而失败。

- [ ] **步骤 3：实现最小服务骨架**

初始化 Go 模块，引入 Gin 和 MySQL 驱动，实现配置加载、统一响应、健康检查与 `main` 依赖组装。迁移创建 `users`、`couples`、`invite_codes`、`moments`、`moment_photos`、`refresh_tokens`、`upload_grants`；使用外键、唯一索引和 `(moment_id, sort_order)` 唯一约束。

- [ ] **步骤 4：验证迁移和测试**

运行：`docker compose up -d mysql && make migrate-up && cd server && go test ./...`

预期：迁移成功，全部测试通过，`curl http://localhost:8080/healthz` 返回 `status=ok`。

- [ ] **步骤 5：提交**

```bash
git add .env.example Makefile docker-compose.yml server
git commit -m "feat: bootstrap Go API and MySQL schema"
```

### 任务 2：微信登录、令牌与用户资料

**文件：**
- 创建：`server/internal/auth/model.go`
- 创建：`server/internal/auth/repository.go`
- 创建：`server/internal/auth/service.go`
- 创建：`server/internal/auth/service_test.go`
- 创建：`server/internal/auth/wechat.go`
- 创建：`server/internal/auth/token.go`
- 创建：`server/internal/httpapi/auth_handler.go`
- 创建：`server/internal/httpapi/auth_middleware.go`
- 创建：`server/internal/httpapi/auth_handler_test.go`
- 创建：`server/internal/user/service.go`
- 创建：`server/internal/httpapi/user_handler.go`
- 修改：`server/internal/httpapi/router.go`

**接口：**
- 消费：任务 1 的数据库连接、统一响应和路由。
- 产出：`auth.WeChatExchanger.Exchange(ctx context.Context, code string) (openid string, sessionKey string, err error)`。
- 产出：`auth.Service.Login(ctx context.Context, code string) (auth.Session, error)`、`Refresh(ctx, rawToken string) (auth.Session, error)`、`Revoke(ctx, rawToken string) error`。
- 产出：`user.Service.UpdateMe(ctx context.Context, userID int64, input user.UpdateInput) (user.User, error)`。
- 端点：`POST /v1/auth/wechat`、`POST /v1/auth/refresh`、`POST /v1/auth/logout`、`GET /v1/users/me`、`PUT /v1/users/me`。

- [ ] **步骤 1：编写登录与令牌失败测试**

覆盖：空 `code` 被拒绝；伪造 exchanger 返回的同一 `openid` 始终对应同一用户；访问令牌包含用户 ID 且 15 分钟过期；刷新令牌仅以 SHA-256 哈希入库；已撤销或过期刷新令牌不能换新；日志测试确认响应和错误不包含 `code`、令牌原文或 `sessionKey`。

- [ ] **步骤 2：运行目标测试并确认失败**

运行：`cd server && go test ./internal/auth ./internal/httpapi -run 'Test(Login|Refresh|Auth|Secrets)'`

预期：接口尚未实现而失败。

- [ ] **步骤 3：实现认证服务和中间件**

使用 15 分钟 JWT 访问令牌和 30 天随机刷新令牌；刷新时轮换并撤销旧令牌。中间件把 `userID int64` 写入请求上下文。微信交换封装在接口后，单元测试禁止访问真实微信服务。

- [ ] **步骤 4：实现用户资料接口**

昵称去除首尾空白后长度为 1–30 个 Unicode 字符；头像仅保存经过授权的对象键或微信临时头像 URL。处理器只调用服务接口并映射稳定错误码。

- [ ] **步骤 5：运行验证并提交**

运行：`cd server && go test ./...`

预期：全部通过。

```bash
git add server/internal server/cmd/api/main.go
git commit -m "feat: add WeChat authentication and profiles"
```

### 任务 3：邀请码、绑定与解除关系

**文件：**
- 创建：`server/internal/couple/model.go`
- 创建：`server/internal/couple/repository.go`
- 创建：`server/internal/couple/service.go`
- 创建：`server/internal/couple/service_test.go`
- 创建：`server/internal/couple/integration_test.go`
- 创建：`server/internal/httpapi/couple_handler.go`
- 创建：`server/internal/httpapi/couple_handler_test.go`
- 修改：`server/internal/httpapi/router.go`

**接口：**
- 消费：任务 2 的认证用户 ID。
- 产出：`couple.Service.CreateInvite(ctx context.Context, userID int64) (couple.Invite, error)`，邀请码使用不含易混淆字符的六位大写字母和数字，10 分钟过期。
- 产出：`PreviewInvite(ctx, userID int64, code string) (couple.InvitePreview, error)`。
- 产出：`AcceptInvite(ctx context.Context, userID int64, code string) (couple.Couple, error)`。
- 产出：`GetActive(ctx context.Context, userID int64) (couple.Couple, error)`、`Unbind(ctx context.Context, userID int64) error`。
- 端点：`POST /v1/couples/invitations`、`GET /v1/couples/invitations/:code`、`POST /v1/couples/bind`、`GET /v1/couples/me`、`DELETE /v1/couples/me`。

- [ ] **步骤 1：编写业务规则失败测试**

覆盖六位格式、10 分钟过期、一次性消费、不能绑定自己、任一用户已有有效关系时拒绝、解除后查询返回 `NO_ACTIVE_COUPLE`。加入评审重点测试：两个 goroutine 同时接受同一邀请码，数据库中仅产生一段有效关系。

- [ ] **步骤 2：运行目标测试并确认失败**

运行：`cd server && go test ./internal/couple -run 'Test(CreateInvite|AcceptInvite|Concurrent|Unbind)'`

预期：服务尚未实现而失败。

- [ ] **步骤 3：实现事务化情侣服务**

`AcceptInvite` 在一个 MySQL 事务中用 `SELECT ... FOR UPDATE` 锁定邀请码和双方用户的有效关系检查，成功后写入关系并标记邀请码已用。错误映射为 `INVITE_INVALID`、`INVITE_EXPIRED`、`INVITE_USED`、`CANNOT_BIND_SELF`、`ALREADY_BOUND`。

- [ ] **步骤 4：实现 HTTP 接口并验证**

运行：`cd server && go test ./internal/couple ./internal/httpapi && go test -race ./internal/couple`

预期：全部通过；并发测试只有一次成功。

- [ ] **步骤 5：提交**

```bash
git add server/internal/couple server/internal/httpapi
git commit -m "feat: add couple invitations and binding"
```

### 任务 4：照片授权、记录、首页与时间线

**文件：**
- 创建：`server/internal/storage/service.go`
- 创建：`server/internal/storage/cos.go`
- 创建：`server/internal/storage/service_test.go`
- 创建：`server/internal/moment/model.go`
- 创建：`server/internal/moment/repository.go`
- 创建：`server/internal/moment/service.go`
- 创建：`server/internal/moment/service_test.go`
- 创建：`server/internal/moment/integration_test.go`
- 创建：`server/internal/httpapi/storage_handler.go`
- 创建：`server/internal/httpapi/moment_handler.go`
- 创建：`server/internal/httpapi/moment_handler_test.go`
- 修改：`server/internal/httpapi/router.go`

**接口：**
- 消费：任务 2 的认证用户 ID；任务 3 的有效关系查询。
- 产出：`storage.Service.CreateUploadGrant(ctx context.Context, userID int64, input storage.UploadInput) (storage.UploadGrant, error)`；只允许 `image/jpeg`、`image/png`、`image/webp`，单张最大 10 MiB，授权 10 分钟有效。
- 产出：`moment.Service.Create(ctx context.Context, userID int64, input moment.CreateInput) (moment.Moment, error)`，其中 `PhotoKeys []string` 长度为 0–9。
- 产出：`List(ctx context.Context, userID int64, cursor string, limit int) (moment.Page, error)`，`limit` 默认 20、最大 50。
- 产出：`Home(ctx context.Context, userID int64) (moment.Home, error)`，返回伴侣最近状态和最近 5 条记录。
- 产出：`Delete(ctx context.Context, userID, momentID int64) error`。
- 端点：`POST /v1/uploads`、`POST /v1/moments`、`GET /v1/moments`、`GET /v1/home`、`DELETE /v1/moments/:id`。

- [ ] **步骤 1：编写上传授权和记录规则失败测试**

覆盖：空记录被拒绝；文字最多 500 个 Unicode 字符；地点最多 80 个字符；快捷状态仅允许 `in_class`、`miss_you`、`tired`；第十张照片被拒绝；伪造、过期或属于其他用户的对象键被拒绝；照片排序 1–9 保持输入顺序；无有效关系不能创建或查看记录。

- [ ] **步骤 2：编写权限和分页失败测试**

覆盖：第三方不能查看；非作者不能删除；作者可以软删除；解除关系后旧游标查询和删除均返回 `NO_ACTIVE_COUPLE`；相同时间戳使用 `(created_at, id)` 游标仍无重复或遗漏。

- [ ] **步骤 3：运行目标测试并确认失败**

运行：`cd server && go test ./internal/storage ./internal/moment ./internal/httpapi -run 'Test(Upload|CreateMoment|List|Delete|Home)'`

预期：服务尚未实现而失败。

- [ ] **步骤 4：实现上传授权与记录事务**

上传对象键格式为 `users/{userID}/{yyyy}/{mm}/{uuid}.{ext}`。创建记录时在一个事务内验证全部 `upload_grants`，写入 `moments` 和按顺序写入 `moment_photos`，随后标记授权已消费。读取照片时生成短期私有下载 URL。

- [ ] **步骤 5：实现时间线、首页和删除接口**

所有查询先由当前用户解析有效关系；不接受客户端提供的关系 ID。时间线使用稳定游标分页，首页只返回伴侣最近状态与最近 5 条未删除记录。

- [ ] **步骤 6：运行验证并提交**

运行：`cd server && go test ./... && go test -race ./internal/moment ./internal/couple`

预期：全部通过。

```bash
git add server/internal
git commit -m "feat: add private moments and photo uploads"
```

### 任务 5：小程序基础、登录与会话恢复

**文件：**
- 创建：`miniprogram/project.config.json`
- 创建：`miniprogram/app.js`
- 创建：`miniprogram/app.json`
- 创建：`miniprogram/app.wxss`
- 创建：`miniprogram/config.js`
- 创建：`miniprogram/services/http.js`
- 创建：`miniprogram/services/auth.js`
- 创建：`miniprogram/services/user.js`
- 创建：`miniprogram/utils/session.js`
- 创建：`miniprogram/pages/login/login.{js,json,wxml,wxss}`
- 创建：`miniprogram/package.json`
- 创建：`miniprogram/tests/session.test.js`

**接口：**
- 消费：任务 2 的认证和用户端点。
- 产出：`request({ path, method, data, auth }) -> Promise<any>`；401 时最多刷新一次，失败则清理会话并跳转登录。
- 产出：`loginWithWeChat() -> Promise<Session>`、`restoreSession() -> Promise<Session|null>`、`logout() -> Promise<void>`。

- [ ] **步骤 1：编写会话失败测试**

使用 Node 内置测试运行器测试纯会话状态机：有效访问令牌直接请求；401 只刷新一次；刷新失败清空令牌；两个同时 401 的请求共享一次刷新 Promise，防止刷新风暴和无限循环。

- [ ] **步骤 2：运行测试并确认失败**

运行：`cd miniprogram && node --test tests/session.test.js`

预期：模块尚未实现而失败。

- [ ] **步骤 3：实现入口、请求层和登录页**

登录页调用 `wx.login`，再调用 `/v1/auth/wechat`；资料未完整时提交昵称和头像。`app.js` 恢复会话后，根据 `GET /v1/couples/me` 跳转绑定页或首页。

- [ ] **步骤 4：运行自动测试并在开发者工具验证**

运行：`cd miniprogram && node --test tests/*.test.js`

预期：全部通过；开发者工具中使用测试后端可登录、退出并恢复会话。

- [ ] **步骤 5：提交**

```bash
git add miniprogram
git commit -m "feat: add mini program login flow"
```

### 任务 6：情侣绑定和设置页面

**文件：**
- 创建：`miniprogram/services/couple.js`
- 创建：`miniprogram/pages/bind/bind.{js,json,wxml,wxss}`
- 创建：`miniprogram/pages/settings/settings.{js,json,wxml,wxss}`
- 创建：`miniprogram/utils/invite.js`
- 创建：`miniprogram/tests/invite.test.js`
- 修改：`miniprogram/app.json`

**接口：**
- 消费：任务 3 的情侣端点和任务 5 的 `request`。
- 产出：`normalizeInviteCode(value) -> string`，转大写、去空格，仅保留允许字符，最长六位。
- 产出：`createInvite()`、`previewInvite(code)`、`acceptInvite(code)`、`getActiveCouple()`、`unbind()`。

- [ ] **步骤 1：编写邀请码输入失败测试**

验证小写转大写、空格清除、易混淆字符被拒绝、超过六位截断；绑定提交期间不能重复调用；错误码分别映射为“已过期”“已使用”“已绑定”等中文提示。

- [ ] **步骤 2：运行测试并确认失败**

运行：`cd miniprogram && node --test tests/invite.test.js`

预期：纯函数尚未实现而失败。

- [ ] **步骤 3：实现绑定与解除界面**

绑定页支持生成、复制、输入、预览和确认；设置页在二次确认后解除关系，并清空页面缓存后跳转绑定页。邀请码有效期必须展示。

- [ ] **步骤 4：验证并提交**

运行：`cd miniprogram && node --test tests/*.test.js`

预期：全部通过；开发者工具中可完成生成、预览、绑定和解除。

```bash
git add miniprogram
git commit -m "feat: add couple binding experience"
```

### 任务 7：首页、九图发布和共同时间线

**文件：**
- 创建：`miniprogram/services/moment.js`
- 创建：`miniprogram/services/upload.js`
- 创建：`miniprogram/utils/photo-queue.js`
- 创建：`miniprogram/utils/moment.js`
- 创建：`miniprogram/tests/photo-queue.test.js`
- 创建：`miniprogram/tests/moment.test.js`
- 创建：`miniprogram/pages/home/home.{js,json,wxml,wxss}`
- 创建：`miniprogram/pages/compose/compose.{js,json,wxml,wxss}`
- 创建：`miniprogram/pages/timeline/timeline.{js,json,wxml,wxss}`
- 修改：`miniprogram/app.json`

**接口：**
- 消费：任务 4 的上传、记录、首页和时间线端点。
- 产出：`createPhotoQueue(files, max = 9)`、`removePhoto(queue, id)`、`movePhoto(queue, from, to)`、`markUploadResult(queue, id, result)`、`failedPhotos(queue)`。
- 产出：`validateMomentDraft({ text, status, placeName, photos }) -> { valid, error }`。
- 产出：`uploadPhotos(queue, onProgress) -> Promise<string[]>`，返回保持选择顺序的对象键数组。

- [ ] **步骤 1：编写九图队列和表单失败测试**

覆盖 0、1、9、10 张照片；删除和调整顺序；部分失败后仅返回失败项用于重试；成功结果仍按用户选择顺序排列；空记录、501 字文字、81 字地点和非法状态被拒绝。

- [ ] **步骤 2：运行测试并确认失败**

运行：`cd miniprogram && node --test tests/photo-queue.test.js tests/moment.test.js`

预期：工具函数尚未实现而失败。

- [ ] **步骤 3：实现发布页**

使用 `wx.chooseMedia`，限制剩余可选数量；提供预览、删除和排序；逐项显示上传状态与整体进度。全部照片上传成功后才调用 `POST /v1/moments`；失败时保留文字、状态、地点和照片队列，只重试失败照片；提交期间禁用按钮。

- [ ] **步骤 4：实现首页与时间线**

首页显示伴侣最近状态、更新时间和最近 5 条记录；时间线支持下拉刷新和游标加载更多。多图使用 1、2–4、5–9 张自适应宫格，点击调用 `wx.previewImage`，删除操作只对自己的记录显示并要求确认。

- [ ] **步骤 5：运行验证并提交**

运行：`cd miniprogram && node --test tests/*.test.js`

预期：全部通过；开发者工具可分别发布 1 张和 9 张照片，选择第 10 张时被阻止，展示顺序一致。

```bash
git add miniprogram
git commit -m "feat: add shared timeline and nine-photo moments"
```

### 任务 8：部署、联调与双人真机验收

**文件：**
- 创建：`deploy/nginx.conf`
- 创建：`deploy/just-for-you.service`
- 创建：`docs/deployment.md`
- 创建：`docs/manual-acceptance.md`
- 修改：`.env.example`
- 修改：`Makefile`
- 修改：`miniprogram/config.js`

**接口：**
- 消费：前述全部任务产出的 API 和小程序页面。
- 产出：`make test`、`make migrate-up`、`make build-server` 和 `make smoke BASE_URL=https://...`。

- [ ] **步骤 1：编写部署冒烟检查**

在 Makefile 中定义 `smoke`，依次检查 `/healthz`、无令牌访问 `/v1/users/me` 返回 401、HTTPS 证书有效，并确保响应不暴露服务器堆栈或密钥。

- [ ] **步骤 2：配置服务器与私有 COS**

安装 MySQL 8.0、Nginx 和编译后的 Go 二进制；创建最小权限数据库账号和 COS 子账号；配置 HTTPS、请求体大小、日志轮转、防火墙和 systemd 自动重启。真实密钥只写服务器环境文件，不提交 Git。

- [ ] **步骤 3：执行全量自动验证**

运行：`make test && make build-server && make smoke BASE_URL=https://实际域名`

预期：Go 测试、竞态测试、小程序纯函数测试、构建和 HTTPS 冒烟检查全部通过。

- [ ] **步骤 4：完成微信体验版配置**

在微信公众平台配置合法 request/upload/download 域名，在开发者工具中关闭“不校验合法域名”后重新验证，再上传体验版并添加伴侣为体验成员。

- [ ] **步骤 5：双人真机验收**

按 `docs/manual-acceptance.md` 使用两个微信账号逐项执行规格文档中的 12 项人工验收，记录手机型号、微信版本、时间和结果。任何失败先创建复现步骤，再修复并重新执行相关自动测试和人工步骤。

- [ ] **步骤 6：最终提交**

```bash
git add deploy docs .env.example Makefile miniprogram/config.js
git commit -m "ops: add deployment and acceptance workflow"
```

## 建议的一周节奏

- 第 1 天：任务 1–2，打通健康检查和微信登录。
- 第 2 天：任务 3，完成双人邀请码绑定。
- 第 3 天：任务 4，完成记录、权限、九图授权和时间线 API。
- 第 4 天：任务 5–6，完成小程序登录、绑定和解除。
- 第 5 天：任务 7，完成首页、九图发布和共同时间线。
- 第 6 天：任务 8 的部署、域名、COS 和首次真机联调。
- 第 7 天：双人验收、修复阻塞问题并上传稳定体验版。
