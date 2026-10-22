# 小事成双对：仓库整理、本地 Go 环境与代码导读设计

## 目标

将现有单仓库整理为清晰的 `frontend/` 与 `backend/` 两个独立目录，在不丢失 WorkBuddy 未提交修改的前提下，使 Go 后端可在本机连接 MySQL、执行迁移并启动，同时保留完整测试能力，并提供面向后端开发者的中文代码导读。

## 当前状态

- 完整项目位于 Git 工作树 `.worktrees/just-for-you-mvp`，主检出目录仅包含早期设计文档。
- 当前功能分支比远端超前，且包含 WorkBuddy 留下的前后端未提交修改；这些修改必须先验证并建立可恢复快照。
- 本机已安装 Go `1.27.2 darwin/arm64`，无需再次安装。
- 本机可以调用 MySQL 客户端，尚未检测到 Docker；本地启动不依赖 Docker。
- 微信小程序、Go 服务、测试、文档、IDE 文件和生成物当前混放在仓库根目录。

## 目标目录

```text
Just for you/
├── frontend/
│   ├── miniprogram/
│   ├── package.json
│   ├── package-lock.json
│   ├── project.config.json
│   └── project.private.config.json.example
├── backend/
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
├── README.md
└── .gitignore
```

测试文件与生产代码放在各自前后端目录中。数据库迁移、依赖锁文件、示例配置和微信项目配置属于可运行项目的一部分，必须保留。

## 清理规则

删除可重新生成或与交付无关的内容：`.DS_Store`、IDE 私有状态、`node_modules/`、Superpowers/WorkBuddy 执行记录、临时产物和废弃目录。

保留：`.git`、业务代码、测试、迁移、依赖清单与锁文件、`.gitignore`、精简后的 `README.md`、`.env.example`、必要的构建与微信开发者工具配置。

真实 `.env`、微信私有配置和密钥继续保持不提交。删除操作前必须确认目标清单，并以 Git 提交保存 WorkBuddy 修改，使所有业务代码均可恢复。

## 本地开发环境

后端使用本机 Homebrew Go，不在仓库内保存 Go 二进制。MySQL 优先使用 Homebrew 本地服务；创建开发数据库和最小权限开发账号后执行全部迁移。后端从 `backend/.env` 读取配置，并提供统一启动命令。

启动成功必须满足：

1. Go 依赖下载和编译成功；
2. MySQL 连接成功并完成全部迁移；
3. API 进程稳定监听本地端口；
4. `/healthz` 返回成功；
5. 不要求真实微信或 COS 凭据的自动化测试全部通过。

微信登录和 COS 上传依赖外部平台，若没有可用凭据则明确标记为人工验收项，不用假数据伪造通过。

## 路径与工具调整

- 前端测试命令在 `frontend/` 内独立运行。
- Go 命令在 `backend/` 内独立运行，module path 保持不变，避免无意义的 import 全量改名。
- 微信 `project.config.json` 调整 `miniprogramRoot`，保证开发者工具仍能直接导入 `frontend/`。
- 根目录 README 仅承担项目入口职责，分别列出前端、后端的安装、测试和启动命令。
- 不保留依赖旧目录的 Makefile；需要的命令改为前后端各自可执行的简单命令或根 README 指令。

## 验证策略

整理前：运行 WorkBuddy 变更涉及的 Node 与 Go 测试，并检查 Git diff，确定基线。

整理后：运行前端全量 Node 测试、后端 `go test ./...`、`go test -race ./...`、`go vet ./...`、`git diff --check`，并在微信开发者工具中执行普通编译。启动 MySQL、迁移和 API 后，以 HTTP 请求验证健康检查。

最终 `git status` 只能包含明确列出的本地私有配置；任何业务代码不得处于未跟踪或未提交状态。

## 代码导读

代码导读以用户熟悉的后端视角展开：

1. 先解释新的前后端目录与启动入口；
2. 以一个真实接口追踪 `interfaces/http → application → domain → infrastructure`；
3. 解释 DDD 各层能依赖什么、不能依赖什么；
4. 说明 MySQL repository、COS/微信适配器和依赖装配的位置；
5. 再解释小程序的 `pages → services → store/utils → components` 数据流；
6. 给出新增一个接口、新增一个页面和排查请求失败的操作模板。

导读写入根 README 的简短地图，并另提供一次中文交接说明，避免产生大量重复文档。

## 非目标

- 不重写现有业务功能或改变 API 语义。
- 不把前后端拆成两个 Git 仓库。
- 不删除测试、迁移或锁文件来追求目录表面简洁。
- 不在本轮部署生产环境、配置正式域名或上传微信体验版。
