# 新增用例检查清单

## 1. 先写领域测试

在目标上下文的 `domain` 中先写失败测试，明确校验、状态迁移和越权规则。真实时间由调用方传入，领域层不要直接读系统时钟。

## 2. 定义用例输入与端口

在 `application` 中增加 Command/Query、具名用例和最小端口。端口按调用方需要定义，不要复制一整套通用仓储。

## 3. 实现适配器

- MySQL 适配器命名说明真实职责，例如 `MySQLThingWriter`、`MySQLThingReader`。
- 外部服务适配器放在对应上下文的 `infrastructure` 下。
- 涉及多条写入时，写明事务边界和失败回滚行为，并用 sqlmock 或临时库测试。

## 4. 增加 HTTP 路由

Handler 只绑定参数、读取 `httpapi.UserID(c)`、调用用例和映射 DTO。错误映射使用领域错误，不把 SQL 错误直接暴露给客户端。

## 5. 在 bootstrap 装配

在 `internal/bootstrap` 中创建具体存储和用例，再注入 Handler。不要让 `application` 导入 `infrastructure`，也不要让两个上下文互相导入具体适配器。

## 6. 变更数据库

当前项目从 `000001_baseline` 管理全新数据库。开发期新增结构时：

1. 新建下一个编号的 `up.sql` 和 `down.sql`，不要改已发布的迁移。
2. `down.sql` 按外键逆序删除。
3. 更新 `internal/architecture/migration_test.go` 的结构断言。
4. 运行 `go run ./cmd/migrate -direction up`。
5. 在隔离测试库运行 `up -> down -> up`。

## 7. 完成验证

```bash
cd backend
gofmt -w cmd internal
go test ./... -count=1
go test -race ./...
go vet ./...

cd ../frontend
npm test
```

还需要在微信开发者工具验证页面交互、返回键、真机网络和小程序权限声明。
