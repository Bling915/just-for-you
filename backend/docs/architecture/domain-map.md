# 领域地图

后端采用纵向切片，而不是全局的 `domain/application/infrastructure/interfaces` 目录。每个业务上下文自己拥有四层，依赖只能从外向内。`bootstrap` 是唯一实例化 MySQL、COS、微信和 AI 具体实现的位置。

| 上下文 | 拥有的数据与规则 | HTTP 前缀 |
| --- | --- | --- |
| `account` | 微信登录、用户资料、头像授权 | `/auth`、`/users/me` |
| `couple` | 双人关系、邀请码、单人测试模式 | `/couples` |
| `media` | 上传授权与对象存储端口 | 无独立路由 |
| `moment` | 共同/个人小事、照片、状态、回应、双视角、回忆 | `/moments`、`/home` |
| `companionship` | 此刻状态、轻互动、通知偏好 | `/presence`、`/nudges` |
| `anniversary` | 纪念日、封面、置顶和日期规则 | `/anniversaries` |
| `todo` | 共同待办与完成/重开 | `/todos` |
| `wish` | 共同愿望与关联小事 | `/wishes` |
| `dailyquestion` | 每日一问、回答、解锁和 Moment 投影 | `/daily-question` |

## 边界规则

- `domain` 不导入 Gin、SQL、COS、微信 SDK，也不读取环境配置。
- `application` 只依赖自己定义的端口、同上下文领域和 `platform` 技术接口。
- `infrastructure` 实现 `application` 端口，不可被 `application` 反向导入。
- `interfaces/http` 只做鉴权用户读取、JSON 绑定、调用用例和 DTO 映射。
- 不同上下文之间通过窄端口通信，不共享一个总仓储或总服务。

## 跨域端口

- 所有需要情侣关系的上下文都定义自己的 `ActiveCoupleReader`，由 `bootstrap` 用 Couple 用例适配。
- Wish 通过 `MomentReader` 校验关联小事是否属于当前关系。
- Moment 首页通过 `AnniversaryHighlightsReader` 读取置顶纪念日，`bootstrap` 负责 DTO 转换。
- Daily Question 通过 `AnswerSubmissionStore` 原子写入回答、解锁状态和唯一的 Moment 投影。
- 账号头像、Moment 照片和纪念日封面分别通过各自的 `MediaStorage`/`CoverStorage` 端口访问对象存储。

## 目录选择

新增业务时先判断它是否拥有独立规则和生命周期。独立规则应在新上下文中建立四层；单纯的服务端流程辅助代码放到 `platform`。不要恢复旧的总 `application`、`domain` 或 `infrastructure` 目录。
