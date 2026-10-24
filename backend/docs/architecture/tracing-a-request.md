# 跟踪一次请求：发布一条小事

以 `POST /v1/moments` 为例，完整路径如下。

1. `internal/bootstrap/router.go` 创建 Gin 路由，JWT 中间件把用户 ID 写入请求上下文。
2. `internal/moment/interfaces/http/handler.go` 读取用户 ID，绑定请求，转换成 `PublishMomentCommand`。
3. `internal/moment/application/publish_moment.go` 查询当前情侣、校验上传授权、创建 `domain.Moment`，再调用 `MomentWriter`。
4. `internal/moment/domain/moment.go` 校验文本、状态、地点、时间、坐标和最多九张照片。
5. `internal/moment/infrastructure/mysql/moment_store.go` 在事务中写 `moments`、`moment_photos`，并消费上传授权。
6. Handler 把领域对象映射为 JSON；浏览器或小程序只看到 API DTO。

## 责任边界

- HTTP 层不判断“能不能在个人空间关联对方的小事”，这类规则在 domain 或 application。
- MySQL 适配器不替调用方做业务决策，但可以通过数据库唯一键兜底并发约束。
- `bootstrap` 是唯一知道 `cos.NewObjectStorage`、`momentmysql.NewMySQLMomentStore` 等具体类型的地方。

## 每日一问的原子边界

提交每日回答时，`SubmitDailyAnswer` 只编排关系查询和命令构造。`MySQLAnswerSubmissionStore` 在一个事务里保存答案、判断双方是否已回答、更新解锁时间，并用 `daily_question_id` 唯一键插入 Moment 投影。Moment 语义由 `daily_question_id IS NOT NULL` 推导，不再维护单独的 `kind` 字段。
