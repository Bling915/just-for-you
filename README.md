# 小事成双对

一款原生微信情侣小程序。双方可以记录文字、状态、位置和最多九张照片，在共同空间或个人空间保存日常，并通过轻回应、双视角、每日一问和回忆浮现一起珍藏小事。

## 技术栈

- 小程序：原生微信小程序、Node.js 测试
- 后端：Go、Gin、MySQL 8
- 文件存储：腾讯云 COS
- 登录：微信小程序登录
- 架构：按业务上下文组织的纵向 DDD 切片

## 目录

```text
frontend/
  miniprogram/pages/         页面与交互
  miniprogram/components/    可复用组件
  miniprogram/services/      后端 API
  miniprogram/store/         页面间状态
  miniprogram/utils/         可测试的纯逻辑
  miniprogram/tests/         Node 测试

backend/
  cmd/api/                   API 进程
  cmd/migrate/               迁移命令行
  internal/account/          账号上下文
  internal/couple/           情侣关系上下文
  internal/media/            上传授权与 COS 端口
  internal/moment/           小事、回应、双视角和回忆
  internal/companionship/    此刻状态与轻互动
  internal/anniversary/      纪念日
  internal/todo/             共同待办
  internal/wish/             共同愿望
  internal/dailyquestion/    每日一问
  internal/platform/         鉴权、配置、数据库、HTTP 等通用技术能力
  internal/bootstrap/        唯一依赖装配层与路由入口
  migrations/                数据库基线
  docs/architecture/         架构与维护说明
```

每个业务上下文拥有自己的 `domain/application/infrastructure/interfaces`。依赖方向是外层向内层，具体实现只在 `bootstrap` 创建。

- [领域地图](backend/docs/architecture/domain-map.md)
- [跟踪一次发布请求](backend/docs/architecture/tracing-a-request.md)
- [新增用例检查清单](backend/docs/architecture/adding-a-use-case.md)

## 本地启动

需要 Go 1.23+、Node.js 20+、MySQL 8.x 和微信开发者工具。

### 数据库

首次运行会创建本地 `just_for_you` 数据库、应用账号并执行迁移：

```bash
cd backend
./scripts/setup-local-db.sh
```

如果 MySQL root 有密码：

```bash
MYSQL_ADMIN_USER=root MYSQL_ADMIN_PASSWORD='你的本地密码' ./scripts/setup-local-db.sh
```

迁移只使用一个受 `schema_migrations` 跟踪的基线。重复执行不会再次应用：

```bash
cd backend
go run ./cmd/migrate -direction up
```

### 后端

```bash
cd backend
cp .env.example .env
# 填入微信与 COS 配置；.env 不进入 Git
./scripts/dev.sh
```

默认监听 `http://127.0.0.1:8080`：

```bash
curl http://127.0.0.1:8080/healthz
```

后端验证：

```bash
cd backend
go test ./... -count=1
go test -race ./...
go vet ./...
bash scripts/scripts_test.sh
```

### 小程序

```bash
cd frontend
npm install
npm test
```

在微信开发者工具中导入 `frontend/`。本地 `apiBaseURL`：

- 开发者工具：`http://127.0.0.1:8080`
- 同一 Wi-Fi 真机：`http://电脑局域网IP:8080`

不要把 `backend/.env`、`frontend/miniprogram/config.js` 的本地地址、`project.private.config.json` 或任何云密钥提交到 Git。

## 主要 API

- 账号：`POST /v1/auth/wechat`、`POST /v1/auth/refresh`、`GET/PUT /v1/users/me`
- 关系：`POST /v1/couples/invitations`、`POST /v1/couples/bind`、`GET/DELETE /v1/couples/me`
- 小事：`GET/POST /v1/moments`、`PATCH/DELETE /v1/moments/:id`
- 回应与双视角：`POST /v1/moments/:id/reactions`、`PUT /v1/moments/:id/perspective`
- 陪伴：`GET/PUT /v1/presence`、`POST /v1/nudges`
- 共同生活：`/v1/anniversaries`、`/v1/todos`、`/v1/wishes`
- 每日一问：`GET /v1/daily-question`、`PUT /v1/daily-question/answer`

## 数据与状态约定

- 小事只有一个 `status` 字符串，快捷状态和自定义状态共享该字段。
- 个人小事只属于作者；`private` 仅自己可见，`partner` 对伴侣可见。
- 每日一问的 Moment 投影由 `daily_question_id IS NOT NULL` 判断，数据库唯一键保证每题最多一条投影。
- 已发布的迁移不再修改；结构变更新增下一个编号的 `up/down` 文件。

## 排查“服务暂不可用”

1. 检查 `curl http://127.0.0.1:8080/healthz`。
2. 查看 `./scripts/dev.sh` 输出中的配置或 MySQL 错误。
3. 确认 `apiBaseURL` 对当前设备可访问；真机不能用 `127.0.0.1`。
4. 运行 `go run ./cmd/migrate -direction up`，确认迁移记录和表结构存在。
5. 对照微信开发者工具 Network 面板与后端日志中的状态码。
