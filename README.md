# 小事成双对

一款原生微信情侣小程序：双方可以记录文字、状态、位置和最多九张照片，在共同空间或个人空间保存日常，并通过轻回应、双视角、每日一句和回忆浮现一起珍藏小事。

## 目录地图

```text
frontend/                         微信小程序
  miniprogram/pages/              页面：接收用户操作、组织页面状态
  miniprogram/components/         可复用界面组件
  miniprogram/services/           后端 API 调用
  miniprogram/store/              会话、草稿、时间线等客户端状态
  miniprogram/utils/              无界面的纯逻辑与工具
  miniprogram/tests/              Node 自动化测试
  project.config.json             微信开发者工具项目配置

backend/                          Go 后端（DDD 分层）
  cmd/api/                        API 进程入口
  cmd/migrate/                    数据库迁移入口
  internal/interfaces/http/       HTTP 路由、参数与响应
  internal/application/           应用用例和端口接口
  internal/domain/                业务实体、规则与业务错误
  internal/infrastructure/        MySQL、COS、微信等适配器
  internal/bootstrap/             依赖装配
  migrations/                     版本化 SQL
  scripts/                        本地环境和启动脚本
```

依赖方向是：`interfaces → application → domain`。`infrastructure` 实现 application 定义的端口，最后由 `bootstrap` 组装。Domain 不应该引用 Gin、SQL、COS 或微信 SDK。

## 本地运行

需要：Go 1.23+、Node.js 20+、MySQL 8.x、微信开发者工具。

当前电脑已验证 Go `1.27.2` 和 Homebrew MySQL 可用。

### 后端

```bash
cd backend
cp .env.example .env
# 编辑 .env，填入微信与 COS 配置；.env 已被 Git 忽略

# 仅对全新的本地数据库执行一次
./scripts/setup-local-db.sh

# 启动 API
./scripts/dev.sh
```

若 MySQL 管理员有密码：

```bash
MYSQL_ADMIN_USER=root MYSQL_ADMIN_PASSWORD='你的本地密码' ./scripts/setup-local-db.sh
```

服务默认监听 `http://127.0.0.1:8080`。验证：

```bash
curl http://127.0.0.1:8080/healthz
# {"data":{"status":"ok"}}
```

已有数据库不要重复执行迁移脚本；先检查 `backend/migrations/` 和当前表结构。

本次功能依赖 `000004_personal_spaces_reflections` 迁移。它会保留旧的共同小事，并补齐 `shared/couple/standard` 默认字段，同时创建回应、双视角和每日问题表。部署前请先备份生产库，再按顺序执行迁移。

后端测试：

```bash
cd backend
bash scripts/scripts_test.sh
go test ./...
go test -race ./...
go vet ./...
```

### 前端

```bash
cd frontend
npm install
npm test
```

在微信开发者工具中直接导入 `frontend/`。本地联调时，将 `frontend/miniprogram/config.js` 的 `apiBaseURL` 设为：

- 开发者工具：`http://127.0.0.1:8080`
- 同一 Wi-Fi 的真机：`http://电脑局域网IP:8080`

`config.js` 的本地值、`project.private.config.json` 和 `backend/.env` 都不应提交。

## 新增功能与 API

- 三个时间线：`GET /v1/moments?space=shared`、`space=personal&owner=me`、`space=personal&owner=partner`。
- 发布目标和可见范围：`POST /v1/moments` 的 `spaceType` 与 `visibility`。
- 分享与回应：`POST /v1/moments/:id/share-to-couple`、`POST/DELETE /v1/moments/:id/reactions`。
- 双视角：`PUT/DELETE /v1/moments/:id/perspective`，最多三张照片。
- 每日一句：`GET /v1/daily-question`、`PUT /v1/daily-question/answer`、`GET /v1/daily-questions`。
- 回忆浮现：`GET /v1/memories/resurface`。

每日一句始终有内置题库可用。要启用可选的 OpenAI-compatible 个性化生成，在 `.env` 中同时配置：

```bash
AI_BASE_URL=https://你的服务地址/v1
AI_API_KEY=只保存在本机或服务器环境中的密钥
AI_MODEL=模型名称
```

三项缺任意一项都不会阻止服务启动，而是直接使用内置题库。个性化上下文只包含当前关系共同小事的截断主题、宽泛情绪和非精确场景；不会发送个人空间内容、图片地址、对象键、经纬度或用户凭证。用户也可以在“消息与个性化”中关闭此功能。

## 本期验收

- 后端全量 `go test ./... -count=1` 与前端 `npm test` 已通过。
- 在两个独立临时 MySQL 数据库完成 `000001` 至 `000004` 升级、旧数据保留检查和完整降级往返。
- 使用独立端口启动当前后端，`/healthz` 返回 200，受保护的新接口未登录时正确返回 401。
- 微信开发者工具和双真机流程仍需在你的两个真实微信账号上完成最终交互验收，尤其是弱网、九图、双视角三图和解绑后的立即失权。

## 用一个发布日记请求理解后端 DDD

以 `POST /v1/moments` 为例：

1. `backend/internal/interfaces/http/moment/handler.go` 注册路由，读取 JSON 和当前用户 ID，把输入转换成 `PublishMomentCommand`。
2. `backend/internal/application/moment/publish_moment.go` 执行完整用例：查询情侣关系、校验照片上传授权、创建 Moment、调用仓储保存。它只依赖 `ports.go` 中的接口。
3. `backend/internal/domain/moment/` 定义 Moment、Photo、Status 以及“最多九图、内容不能为空”等业务规则。这里不认识 HTTP 或 MySQL。
4. `backend/internal/infrastructure/persistence/mysql/moment/` 实现仓储，把领域对象保存到 MySQL；`infrastructure/storage/cos/` 负责对象存储。
5. `backend/internal/bootstrap/moment.go` 创建具体实现，并把它们注入 application 用例与 HTTP handler。
6. Handler 最后只负责把领域结果转成 JSON；业务判断不应堆在 handler 中。

阅读其他功能也按这条路线：先找 HTTP 路由，再找它调用的 `Execute`，接着看 domain 规则，最后看 repository/外部适配器。

## 理解小程序的数据流

仍以发布日记为例：

1. `frontend/miniprogram/pages/compose/` 响应输入、选图和发布按钮。
2. 页面调用 `services/uploads.js` 上传图片，再调用 `services/moments.js` 发布日记。
3. `services/http-client.js` 统一拼接 API 地址、携带 Access Token、处理错误；并发遇到 401 时只刷新一次会话。
4. `store/` 保存跨页面状态，例如登录会话、草稿和时间线缓存；纯计算放在 `utils/`，方便 Node 测试。
5. 请求成功后页面更新自己的 data/store，WXML 再把数据传给 `components/` 渲染。

简单判断文件应该放哪：

- 只属于某个页面的交互：`pages/`
- 多个页面复用的界面：`components/`
- 后端请求：`services/`
- 跨页面状态：`store/`
- 不依赖微信界面的纯逻辑：`utils/`

## 日常维护模板

### 新增一个后端接口

1. 先在 domain 写规则和测试。
2. 在 application 定义 Command/Query、端口和用例测试。
3. 在 infrastructure 实现端口。
4. 在 interfaces/http 增加薄 handler。
5. 在 bootstrap 完成装配。
6. 运行 `go test ./...`、race 和 vet。

### 新增一个小程序页面

1. 创建 `pages/名称/index.{js,json,wxml,wxss}`。
2. 在 `app.json` 注册页面；二级页使用原生导航栏以获得返回键。
3. API 调用写入 `services/`，可测试逻辑写入 `utils/`。
4. 在 `miniprogram/tests/` 先写失败测试，再实现。
5. 运行 `npm test`，随后在微信开发者工具普通编译。

### 页面显示“服务暂不可用”

按顺序排查：

1. `curl http://127.0.0.1:8080/healthz` 是否成功。
2. `backend/scripts/dev.sh` 终端是否有数据库或配置错误。
3. `config.js` 是否使用了当前设备能访问的地址；真机不能使用 `127.0.0.1` 指向电脑。
4. MySQL 是否运行，三期迁移字段是否存在。
5. 微信开发者工具 Network 面板中的状态码和后端日志是否一致。

### 密钥安全

- 不提交 `.env`、AccessKey、Secret、JWT、Token、签名 URL。
- 泄露过的微信、COS 或云厂商密钥必须在控制台轮换，不能只从 Git 删除。
- 提交前运行 `git diff --cached`，确认暂存内容不包含密钥和本地 IP。
