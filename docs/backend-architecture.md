# 后端 DDD 架构说明

后端按 `identity`（身份）、`relationship`（情侣关系）和 `moment`（共同记录）三个业务域组织。`platform` 只放数据库、配置、时钟、鉴权中间件等通用技术能力。

## 四层职责

每个业务域都遵循由内向外的依赖方向：

```text
interfaces/http → application → domain
infrastructure  → application → domain
cmd/api 负责把各层实现组装起来
```

- `domain`：实体、值对象、领域规则和领域错误。不能依赖 Gin、SQL、JWT、COS 等框架。
- `application`：一个用户动作对应一个用例，负责流程编排；通过端口接口访问数据库、微信和对象存储。
- `infrastructure`：端口的技术实现，例如 MySQL 仓储、微信登录、JWT、腾讯云 COS。
- `interfaces/http`：解析 HTTP 请求、调用应用用例、转换响应 DTO、映射稳定错误码。这里不写业务判断，也不直接查数据库。

`cmd/api/main.go` 是组合根。依赖的创建顺序是平台设施 → 仓储/第三方适配器 → 应用服务 → HTTP Handler → Router。

## 新功能应该放在哪里

- 新增“不允许发布第十张照片”这类规则：放在 `domain`，并写领域单元测试。
- 新增“发布后通知伴侣”这类业务流程：放在 `application`，通知服务以端口接口注入。
- 新增 Redis、短信或新的对象存储：在 `infrastructure` 实现应用层端口。
- 新增或修改请求字段、状态码：放在 `interfaces/http`，领域对象不直接序列化。
- 多个业务域都要使用的纯技术能力：放在 `platform`；不要把业务规则移入 `platform`。

## Moment 发布流程

1. 客户端请求上传授权；应用层校验 JPEG/PNG/WebP 和 10 MiB 上限，COS 返回十分钟有效的签名 URL。
2. 客户端把照片直传 COS，再用对象键发布记录。
3. 应用层确认用户仍有有效情侣关系，领域层校验内容及最多九张照片。
4. MySQL 仓储在一个事务中锁定并验证上传授权、写入记录和照片顺序、消费授权。
5. 时间线和首页只按当前有效关系查询，私有照片通过短期读取签名 URL 返回。

## 本地验证

在仓库根目录运行：

```bash
make test-backend
make test-race
make vet
```

Makefile 会优先使用仓库内的 Go 工具链，不存在时使用系统 `go`。
