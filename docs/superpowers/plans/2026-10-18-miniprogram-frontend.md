# 「小事成双对」微信小程序前端实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个可在微信开发者工具和两台真机运行、对接现有 Go API 的「小事成双对」原生微信小程序 MVP。

**Architecture:** 使用原生 JavaScript、WXML、WXSS 和少量 TDesign Miniprogram。页面编排用户流程，业务组件只通过属性和事件通信，`services` 负责 API，轻量 store 管理会话；Token 刷新、上传队列和分页合并写成可由 Node 内置测试运行器验证的纯 JavaScript 模块。

**Tech Stack:** 微信小程序原生框架、JavaScript、TDesign Miniprogram、Node.js `node:test`、微信开发者工具。

**Spec:** `docs/superpowers/specs/2026-10-18-miniprogram-frontend-design.md`

## Global Constraints

- 小程序名称固定为「小事成双对」，主文案固定为“把平凡的小事，慢慢过成两个人的故事”。
- 技术方案固定为原生微信小程序 JavaScript、WXML、WXSS；不引入 TypeScript、Taro、UniApp 或全局状态框架。
- 本地自动化测试要求 Node.js 20 或更高版本；TDesign Miniprogram 固定使用 `1.17.0`。
- 仅选择性使用 TDesign Button、Dialog、Toast、Loading、Input 等基础组件；品牌组件必须自研。
- 视觉方向固定为「薰衣草梦境」；主色 `#9D87C1`，主文字 `#403946`，暖白到淡紫背景，月光黄只作少量强调。
- 记录文字最多 500 个 Unicode 字符，地点最多 80 个，照片最多九张；至少包含文字、状态、地点或照片之一。
- 照片直传必须携带申请授权时完全一致的 `Content-Type`；所有照片成功后才创建记录。
- 底部导航只有“首页”“时光”“我的”；发布入口使用独立悬浮按钮。
- 主要点击区域不小于约 44px；适配刘海屏和底部手势安全区。
- Token、微信 code 和签名 URL 不得写入日志。

## Review Focus

- 多个 API 同时收到 401：只发起一次 refresh，其余请求等待后重放；Task 2 用并发请求测试固定。
- 第九张与第十张照片边界：九张允许，第十张在进入上传队列前拒绝；Task 5 用纯函数测试固定。
- 中途一张照片上传失败：保留全部草稿和成功项，只重试失败项，不创建残缺记录；Task 5 用队列状态测试固定。
- 时间线两页含重复边界记录：按记录 ID 去重且保持倒序，游标只在成功后更新；Task 6 用分页合并测试固定。
- 解除关系或退出登录后返回旧页面：用户、关系、首页、时间线和草稿均不可泄露；Task 7 用 store 清理测试固定。

---

### Task 1：建立小程序骨架、主题和测试入口

**Files:**
- Create: `package.json`
- Create: `project.config.json`
- Create: `project.private.config.json.example`
- Create: `miniprogram/app.js`
- Create: `miniprogram/app.json`
- Create: `miniprogram/app.wxss`
- Create: `miniprogram/sitemap.json`
- Create: `miniprogram/config/index.js`
- Create: `miniprogram/styles/tokens.wxss`
- Create: `miniprogram/styles/animations.wxss`
- Create: `miniprogram/utils/runes.js`
- Create: `miniprogram/pages/welcome/{index.js,index.json,index.wxml,index.wxss}`（仅提供可编译的品牌启动占位页）
- Test: `miniprogram/tests/runes.test.js`
- Modify: `.gitignore`
- Modify: `Makefile`

**Interfaces:**
- Produces: `countRunes(value: string): number`、`truncateRunes(value: string, limit: number): string`。
- Produces: `APP_CONFIG.apiBaseURL`，开发默认 `http://127.0.0.1:8080`，可由 `project.private.config.json` 的环境约定覆盖。
- Produces: `make test-miniprogram`，从仓库根目录运行 `node --test miniprogram/tests/*.test.js`。

- [ ] **Step 1: 写 Unicode 长度 RED 测试**

在 `runes.test.js` 覆盖中文、emoji 和组合字符串，断言 `countRunes("我想你🌙") === 4`，并断言截断不会拆分 surrogate pair。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/runes.test.js`

Expected: FAIL，模块或函数尚不存在。

- [ ] **Step 3: 实现最小纯函数和项目骨架**

实现 `countRunes` 与 `truncateRunes`；创建 `app.json` 页面声明、欢迎占位页、TDesign npm 配置、主题变量和基础全局样式。`package.json` 固定 `tdesign-miniprogram@1.17.0` 并提供 `test` 脚本。尚未实现的页面先不加入 `app.json`，由对应任务创建后再加入；tabBar 在三个目标页面全部存在的 Task 7 配置。

- [ ] **Step 4: 验证骨架**

Run: `make test-miniprogram && node -e "JSON.parse(require('fs').readFileSync('project.config.json'))"`

Expected: Unicode 测试 PASS，项目配置 JSON 可解析。

- [ ] **Step 5: 提交**

```bash
git add package.json project.config.json project.private.config.json.example miniprogram .gitignore Makefile
git commit -m "feat: scaffold native miniprogram"
```

### Task 2：实现网络层、会话 store 与单次 Token 刷新

**Files:**
- Create: `miniprogram/utils/storage.js`
- Create: `miniprogram/store/session.js`
- Create: `miniprogram/services/http-client.js`
- Create: `miniprogram/services/auth.js`
- Create: `miniprogram/services/users.js`
- Test: `miniprogram/tests/http-client.test.js`
- Test: `miniprogram/tests/session.test.js`
- Modify: `miniprogram/app.js`

**Interfaces:**
- Consumes: `APP_CONFIG.apiBaseURL` from Task 1。
- Produces: `createHttpClient({ request, storage, onSessionExpired, baseURL })`，返回 `{ get, post, put, delete }`，每个方法解析 `{data}` 或抛出 `{status, code, message}`。
- Produces: `sessionStore`，方法为 `hydrate()`、`setSession(session)`、`setUser(user)`、`clear()`、`snapshot()`。
- Produces: `authService.loginWithWeChat(code)`、`refresh(refreshToken)`、`logout(refreshToken)`；`userService.getMe()`、`updateProfile(input)`。

- [ ] **Step 1: 写网络和会话 RED 测试**

测试正常响应解包、业务错误转换、Authorization 头、三个并发 401 只调用一次 `/v1/auth/refresh`、刷新成功后三个请求均重放、刷新失败只调用一次 `onSessionExpired`。会话测试断言 `clear()` 同时清除内存和持久化 token。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/http-client.test.js miniprogram/tests/session.test.js`

Expected: FAIL，client/store 尚不存在。

- [ ] **Step 3: 实现会话与请求层**

使用闭包保存单个 `refreshPromise`；刷新接口本身不得触发递归刷新。请求错误只暴露稳定字段，日志不得包含 token、code 或 URL 查询签名。

- [ ] **Step 4: 接入 App 启动恢复**

`App.onLaunch` 调用 `sessionStore.hydrate()`；有 token 时获取当前用户，完全失效时清除会话。不要在 `app.js` 中写页面跳转业务。

- [ ] **Step 5: 验证并提交**

Run: `node --test miniprogram/tests/http-client.test.js miniprogram/tests/session.test.js`

Expected: PASS，且并发刷新断言次数为 1。

```bash
git add miniprogram/app.js miniprogram/services miniprogram/store miniprogram/utils miniprogram/tests
git commit -m "feat: add miniprogram session and API client"
```

### Task 3：完成登录、资料与情侣绑定流程

**Files:**
- Modify: `miniprogram/pages/welcome/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/pages/profile-setup/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/pages/bind/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/pages/invitation/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/services/relationships.js`
- Create: `miniprogram/utils/error-copy.js`
- Test: `miniprogram/tests/error-copy.test.js`

**Interfaces:**
- Consumes: Task 2 的 `authService`、`userService`、`sessionStore` 和 HTTP client。
- Produces: `relationshipService.createInvitation()`、`previewInvitation(code)`、`bind(code)`、`getCurrent()`、`end()`。
- Produces: `messageForError(code: string): string`，至少覆盖 `INVITE_INVALID`、`INVITE_EXPIRED`、`INVITE_USED`、`CANNOT_BIND_SELF`、`ALREADY_BOUND`、`NO_ACTIVE_COUPLE`、`TEMPORARY_UNAVAILABLE`。

- [ ] **Step 1: 写错误文案 RED 测试**

断言所有绑定错误码都有中文可执行提示，未知 5xx 返回“服务暂时不可用，请稍后重试”，不得显示后端英文错误。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/error-copy.test.js`

Expected: FAIL，映射尚不存在。

- [ ] **Step 3: 实现错误映射和关系 service**

路径和 DTO 必须与 `docs/backend-api.md` 一致。邀请码输入统一去空格、转大写并限制六位。

- [ ] **Step 4: 实现四个页面**

欢迎页使用品牌名称和固定主文案；登录按钮防重复点击。资料页使用微信头像选择能力和昵称输入。绑定页提供“创建邀请码”“输入邀请码”两条路径；邀请页显示邀请方预览和确认操作。

- [ ] **Step 5: 微信开发者工具页面验收**

验证四页可编译、加载/失败/禁用状态存在，键盘不会遮挡确认按钮，邀请码可复制且过期提示明确。

- [ ] **Step 6: 验证并提交**

Run: `make test-miniprogram`

Expected: 全部 PASS。

```bash
git add miniprogram/pages miniprogram/services/relationships.js miniprogram/utils/error-copy.js miniprogram/tests
git commit -m "feat: add login and couple binding flows"
```

### Task 4：实现薰衣草品牌组件与首页

**Files:**
- Create: `miniprogram/components/status-orb/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/components/moment-card/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/components/photo-grid/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/components/empty-state/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/components/skeleton-card/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/components/floating-compose-button/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/pages/home/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/services/moments.js`
- Create: `miniprogram/utils/date.js`
- Test: `miniprogram/tests/date.test.js`

**Interfaces:**
- Consumes: Task 2 HTTP client；Task 3 relationship service。
- Produces: `momentService.getHome()`、`getTimeline({cursor, limit})`、`publish(input)`、`deleteMoment(id)`。
- Produces: `formatRelativeTime(iso, now): string`。
- Produces: `moment-card` properties `moment`、`currentUserId`，events `delete`、`previewphotos`。

- [ ] **Step 1: 写相对时间 RED 测试**

覆盖“刚刚”“12 分钟前”“今天 14:20”“昨天 22:10”和跨年完整日期；无效时间返回空字符串而不是抛异常。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/date.test.js`

Expected: FAIL。

- [ ] **Step 3: 实现 date 工具与 moment service**

只使用后端返回的短期照片 URL 显示图片，对象键不得拼成公开地址。

- [ ] **Step 4: 实现品牌组件**

全部颜色、圆角、阴影和动效读取 Task 1 主题变量。`status-orb` 使用 3–4 秒轻微呼吸；卡片进入仅淡入上移；组件提供无数据和图片加载失败状态。

- [ ] **Step 5: 实现首页状态机**

首页具有 `loading`、`ready`、`empty`、`error-with-cache` 和 `error-empty` 状态。展示伴侣最近状态和最近五条；下拉刷新；悬浮按钮进入发布页。

- [ ] **Step 6: 微信开发者工具视觉验收**

检查 320px、375px、430px 三种宽度，安全区无遮挡，状态球滚动时无明显掉帧，空状态和图片失败不破版。

- [ ] **Step 7: 验证并提交**

Run: `make test-miniprogram`

Expected: PASS。

```bash
git add miniprogram/components miniprogram/pages/home miniprogram/services/moments.js miniprogram/utils/date.js miniprogram/tests
git commit -m "feat: add lavender home experience"
```

### Task 5：实现九图选择、上传队列与发布页

**Files:**
- Create: `miniprogram/components/photo-picker/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/pages/compose/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/services/uploads.js`
- Create: `miniprogram/store/draft.js`
- Create: `miniprogram/utils/upload-queue.js`
- Test: `miniprogram/tests/upload-queue.test.js`
- Test: `miniprogram/tests/draft.test.js`

**Interfaces:**
- Consumes: Task 2 HTTP client；Task 4 `momentService.publish(input)`。
- Produces: `uploadService.requestGrant({contentType, size})`、`putFile({filePath, uploadURL, contentType, onStateChange})`。
- Produces: `createUploadQueue({requestGrant, putFile, concurrency: 2})`，方法 `add(files)`、`start()`、`retryFailed()`、`snapshot()`。
- Produces: `draftStore`，方法 `load()`、`patch(input)`、`clear()`、`snapshot()`。

- [ ] **Step 1: 写上传队列与草稿 RED 测试**

断言 0/1/9 张允许、添加第十张拒绝；并发上传不超过 2；PUT 携带授权时一致的 Content-Type；一项失败时不调用 publish；`retryFailed()` 只重试失败项；页面卸载后文字、状态、地点和照片草稿仍存在。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/upload-queue.test.js miniprogram/tests/draft.test.js`

Expected: FAIL。

- [ ] **Step 3: 实现上传队列和草稿 store**

每个文件状态固定为 `pending | granting | uploading | success | failed`，保存 `objectKey` 和安全的展示信息，不持久化签名 URL。使用文件系统读取二进制并通过 `wx.request` 发送 PUT；整体进度按“已完成文件数 / 总文件数”显示真实完成状态，不伪造字节百分比。失败不清除其他成功项。

- [ ] **Step 4: 实现 photo-picker**

使用 `wx.chooseMedia`，选择前计算剩余名额；支持预览、删除、拖动排序、单张进度和失败重试。触发 `change` 时按展示顺序返回文件列表。

- [ ] **Step 5: 实现发布页**

快捷状态固定为 `in_class`、`miss_you`、`tired`；使用 Task 1 Unicode 工具显示 500/80 字计数。空内容不提交；上传全部完成后按当前排序传 `photoKeys`；成功才清空草稿。

- [ ] **Step 6: 微信开发者工具网络验收**

模拟慢网、单张 500、COS PUT 失败和重复点击。确认真实进度显示、草稿保留、只重试失败照片、没有残缺记录请求。

- [ ] **Step 7: 验证并提交**

Run: `make test-miniprogram`

Expected: 全部 PASS。

```bash
git add miniprogram/components/photo-picker miniprogram/pages/compose miniprogram/services/uploads.js miniprogram/store/draft.js miniprogram/utils/upload-queue.js miniprogram/tests
git commit -m "feat: add resilient nine-photo publishing"
```

### Task 6：实现共同时间线、分页、预览与删除

**Files:**
- Create: `miniprogram/pages/timeline/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/store/timeline.js`
- Create: `miniprogram/utils/merge-page.js`
- Test: `miniprogram/tests/merge-page.test.js`
- Test: `miniprogram/tests/timeline-store.test.js`
- Modify: `miniprogram/components/moment-card/index.js`

**Interfaces:**
- Consumes: Task 4 `momentService`、`moment-card`、`photo-grid`。
- Produces: `mergeMomentPage(current, incoming): Moment[]`，按 ID 去重并保持 `(createdAt, id)` 倒序。
- Produces: `timelineStore`，方法 `reset()`、`loadFirst()`、`loadMore()`、`remove(id)`、`snapshot()`。

- [ ] **Step 1: 写分页 RED 测试**

覆盖两页边界重复 ID、相同时间戳不同 ID、加载失败不推进 cursor、连续触底只发一个请求、删除后本地列表同步移除。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/merge-page.test.js miniprogram/tests/timeline-store.test.js`

Expected: FAIL。

- [ ] **Step 3: 实现分页合并和 store**

`loadMore()` 使用进行中的 Promise 防重；只有成功响应才写入 `nextCursor`。`reset()` 清空列表、游标、错误和进行中状态。

- [ ] **Step 4: 实现时间线页面**

支持首屏骨架、空状态、错误重试、下拉刷新、触底分页和没有更多提示。图片预览按记录内顺序调用 `wx.previewImage`。

- [ ] **Step 5: 实现作者删除交互**

只有 `authorId === currentUserId` 时显示删除入口；使用确认 Dialog，成功后平滑移除。`NOT_AUTHOR` 或 `MOMENT_NOT_FOUND` 时刷新时间线并显示明确反馈。

- [ ] **Step 6: 验证并提交**

Run: `make test-miniprogram`

Expected: 全部 PASS。

```bash
git add miniprogram/pages/timeline miniprogram/store/timeline.js miniprogram/utils/merge-page.js miniprogram/components/moment-card miniprogram/tests
git commit -m "feat: add shared moment timeline"
```

### Task 7：实现“我的”、退出与解除关系后的数据隔离

**Files:**
- Create: `miniprogram/pages/me/{index.js,index.json,index.wxml,index.wxss}`
- Create: `miniprogram/store/app-state.js`
- Test: `miniprogram/tests/app-state.test.js`
- Modify: `miniprogram/store/session.js`
- Modify: `miniprogram/store/draft.js`
- Modify: `miniprogram/store/timeline.js`
- Modify: `miniprogram/pages/home/index.js`
- Modify: `miniprogram/app.json`

**Interfaces:**
- Consumes: Task 2 session、Task 3 relationship service、Task 5 draft、Task 6 timeline。
- Produces: `appState.clearPrivateData()`，同步清除 session user、relationship、home、timeline 和 compose draft；`logout()` 和 `endRelationship()` 必须调用相应范围的清理函数。

- [ ] **Step 1: 写隐私清理 RED 测试**

分别填充用户、关系、首页、时间线和草稿，断言退出登录全部清除；解除关系保留登录用户但清除关系、首页、时间线和草稿；旧页面再次 `onShow` 时不能读到旧数据。

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test miniprogram/tests/app-state.test.js`

Expected: FAIL。

- [ ] **Step 3: 实现 app-state 编排**

各 store 暴露自己的 `clear/reset`，`app-state` 只做组合调用，不直接修改其他模块私有字段。

- [ ] **Step 4: 实现“我的”页面**

展示头像、昵称、关系摘要、资料编辑、退出和解除关系。退出与解除都二次确认；解除文案明确历史立即不可见。操作成功后使用 `reLaunch`，防止返回旧页面。此时在 `app.json` 配置“首页”“时光”“我的”三项 tabBar。

- [ ] **Step 5: 补齐所有页面 onShow 权限守卫**

首页、时间线和发布页在 `onShow` 检查有效会话与关系；无关系进入绑定页，不先绘制旧缓存。

- [ ] **Step 6: 验证并提交**

Run: `make test-miniprogram`

Expected: 全部 PASS。

```bash
git add miniprogram/pages/me miniprogram/pages/home miniprogram/store miniprogram/tests
git commit -m "feat: add account and privacy cleanup flows"
```

### Task 8：完善动效降级、文档并导入微信开发者工具验收

**Files:**
- Create: `docs/miniprogram-architecture.md`
- Create: `docs/miniprogram-development.md`
- Create: `docs/miniprogram-manual-test.md`
- Modify: `miniprogram/styles/animations.wxss`
- Modify: `miniprogram/app.wxss`
- Modify: `Makefile`
- Modify: `README.md`（若仓库不存在则创建）

**Interfaces:**
- Consumes: Tasks 1–7 的完整小程序。
- Produces: `make test` 同时运行后端与小程序测试；`make verify` 运行小程序测试、Go 测试、race 和 vet。
- Produces: 可由微信开发者工具直接导入的仓库根 `project.config.json`。

- [ ] **Step 1: 添加动效和设备降级检查**

统一循环动画开关；系统或设备不适合动效时为根节点添加 `reduce-motion`，停止状态球、流光和装饰性动画，同时保留操作反馈。

- [ ] **Step 2: 编写开发与架构文档**

说明目录职责、API base URL 配置、安装依赖、微信开发者工具构建 npm、后端启动和常见调试方法。文档不得包含真实 AppID、Token 或密钥。

- [ ] **Step 3: 编写人工验收清单**

覆盖两用户登录、资料、绑定、首页状态、纯文字/状态/1图/9图发布、失败重试、时间线分页、图片预览、作者删除、Token 刷新、退出和解除关系后的拒绝。

- [ ] **Step 4: 执行自动化验收**

Run: `make verify`

Expected: Node 测试、Go 全量测试、三个业务域 race 和 `go vet` 全部退出码 0。

- [ ] **Step 5: 使用微信开发者工具导入并编译**

导入仓库根目录，配置用户自己的 AppID，执行“工具 → 构建 npm”。修复所有编译错误、控制台未处理异常和缺失组件。检查 320px、375px、430px 模拟器。

- [ ] **Step 6: 真机预览准备**

生成预览二维码前确认 API HTTPS 域名已加入微信 request/uploadFile/downloadFile 合法域名，COS 域名已配置，且 `.env` 和私有项目配置未进入 Git。

- [ ] **Step 7: 提交**

```bash
git add miniprogram Makefile README.md docs project.config.json package.json
git commit -m "docs: finish miniprogram development workflow"
```

### Task 9：最终分支审查与双真机验收交接

**Files:**
- Modify: 仅限审查发现问题对应的文件
- Update: `docs/miniprogram-manual-test.md` 的实际验收结果区

**Interfaces:**
- Consumes: 完整前后端 MVP。
- Produces: 可供两位用户实际试用的候选版本，以及明确列出的云端配置或人工验收剩余项。

- [ ] **Step 1: 请求独立代码审查**

审查整个前端分支，重点检查会话并发、隐私缓存、上传失败、分页重复、组件边界、无障碍和小程序平台限制。

- [ ] **Step 2: 用 TDD 修复 Critical 和 Important 问题**

每个行为问题先增加失败测试，再修改实现；视觉问题提供修改前后开发者工具截图或明确检查结果。

- [ ] **Step 3: 再次执行最终验证**

Run: `make verify`

Expected: 全部退出码 0，Git 工作树只包含有意提交的文件。

- [ ] **Step 4: 双真机验收交接**

由用户在 iPhone 与 Android 上完成登录扫码和真实账号操作；实现者协助观察控制台、请求和后端日志。不能在缺少微信/COS/服务器凭证时声称云端流程已验证。

- [ ] **Step 5: 提交验收修复**

```bash
git add miniprogram docs
git commit -m "test: complete miniprogram MVP acceptance"
```
