# 后端架构说明

后端是一个按业务域划分的模块化单体，使用全局四层 DDD。它保持一次部署和一个数据库，避免现阶段引入微服务的运维成本。

## 阅读一条请求

以发布记录为例，按下面顺序阅读：

1. `server/internal/interfaces/http/moment/handler.go`：解析 HTTP 请求并转换 DTO。
2. `server/internal/application/moment/publish_moment.go`：编排发布用例。
3. `server/internal/domain/moment/moment.go`：校验记录的业务规则。
4. `server/internal/infrastructure/persistence/mysql/moment/repository.go`：执行 SQL 和事务。
5. `server/internal/bootstrap/moment.go`：把上述实现装配起来。

依赖方向固定为：

```text
interfaces/http -> application -> domain
infrastructure  -> application/domain
bootstrap       -> 所有层
cmd             -> bootstrap
```

`server/internal/architecture/dependencies_test.go` 会自动阻止反向依赖。

## 新代码放在哪里

- 新增 HTTP 接口：在 `interfaces/http/<业务域>` 添加 handler 和 DTO。
- 新增用户动作：在 `application/<业务域>` 新建一个 `<action>.go`，类型命名为 `<Action>UseCase`，入口统一为 `Execute`。
- 新增业务规则：放入 `domain/<业务域>` 的实体或值对象，不得导入 Gin、SQL 或云服务 SDK。
- 新增 SQL：放入 `infrastructure/persistence/mysql/<业务域>`，事务不得暴露到应用层。
- 新增第三方服务：在 `application` 定义窄接口，在 `infrastructure` 提供实现。
- 新增依赖装配：修改 `bootstrap/<业务域>.go`，不要把装配写回 `main.go`。

## 本地检查

从仓库根目录运行：

```bash
make test-backend
make test-race
make vet
make build-server
```

本项目暂不拆微服务。只有当某个业务域需要独立扩缩容、独立发布或形成清晰的数据所有权时，才评估拆分；目录长得像微服务不是拆分理由。
