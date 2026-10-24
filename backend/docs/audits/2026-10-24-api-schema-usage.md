# API 与数据库使用审计（2026-10-24）

本清单以当前 Gin handler、`frontend/miniprogram/services/*.js`、应用读模型和 `000001` 至 `000004` 迁移为依据。分类含义：`keep` 为当前业务或安全/查询所需；`derive` 为可由其他字段可靠推导、进入新基线时删除；`delete` 为生产代码没有读写的预留设计。审计没有未确认项。

## HTTP API

除三个认证接口和 `/healthz` 外，以下接口均位于 JWT 保护的 `/v1` 路由组。响应继续使用 `{data: ...}`，错误继续使用 `{error: {code, message}}`。

| 领域 | 方法与路径 | 当前小程序调用方 | 请求字段 | 结论 |
| --- | --- | --- | --- | --- |
| system | `GET /healthz` | 部署探活 | 无 | keep |
| account | `POST /v1/auth/wechat` | `auth.loginWithWeChat` | `code` | keep |
| account | `POST /v1/auth/refresh` | HTTP client 自动刷新 | `refreshToken` | keep |
| account | `POST /v1/auth/logout` | `auth.logout` | `refreshToken` | keep |
| account | `GET /v1/users/me` | `users.getMe` | 无 | keep |
| account | `PUT /v1/users/me` | `users.updateProfile` | `displayName, avatarRef` | keep |
| account | `POST /v1/users/me/avatar-upload` | `users.requestAvatarUpload` | `contentType, size` | keep |
| couple | `POST /v1/couples/invitations` | `relationships.createInvitation` | 无 | keep |
| couple | `GET /v1/couples/invitations/:code` | `relationships.previewInvitation` | path `code` | keep |
| couple | `POST /v1/couples/bind` | `relationships.bind` | `code` | keep |
| couple | `GET /v1/couples/me` | `relationships.getCurrent` | 无 | keep |
| couple | `DELETE /v1/couples/me` | `relationships.end` | 无 | keep |
| couple | `POST /v1/couples/solo-test` | `relationships.enableSoloTest` | 无；仅测试配置开启时注册 | keep |
| companionship | `GET /v1/presence` | `companionship.getSnapshot` | 无 | keep |
| companionship | `PUT /v1/presence` | `companionship.updatePresence` | `preset, customText, expiresAt` | keep |
| companionship | `POST /v1/nudges` | `companionship.sendNudge` | `kind` | keep |
| companionship | `GET /v1/notification-preferences` | `notifications.get` | 无 | keep |
| companionship | `PUT /v1/notification-preferences` | `notifications.update` | `momentEnabled, presenceEnabled, nudgeEnabled, aiPersonalizationEnabled` | keep |
| media/moment | `POST /v1/uploads` | `uploads.requestGrant` | `contentType, size` | keep |
| moment | `POST /v1/moments` | `moments.publish` | `text, status, placeName, latitude, longitude, photoKeys, occurredAt, spaceType, visibility` | keep |
| moment | `PATCH /v1/moments/:id` | `moments.updateMoment` | 发布字段加 `version` | keep |
| moment | `DELETE /v1/moments/:id` | `moments.deleteMoment` | 无 | keep |
| moment | `GET /v1/moments` | `moments.getTimeline` | query `space, owner, cursor, limit` | keep |
| moment | `GET /v1/home` | `moments.getHome` | 无 | keep |
| moment | `POST /v1/moments/:id/share-to-couple` | `moments.shareToCouple` | 可选空对象 | keep |
| moment | `POST /v1/moments/:id/reactions` | `moments.react` | `type, text` | keep |
| moment | `DELETE /v1/moments/:id/reactions/:reactionId` | `moments.deleteReaction` | path IDs | keep |
| moment | `PUT /v1/moments/:id/perspective` | `moments.savePerspective` | `text, photoKeys` | keep |
| moment | `DELETE /v1/moments/:id/perspective` | `moments.deletePerspective` | 无 | keep |
| moment | `GET /v1/moments/calendar` | `moments.getCalendar` | query `month` | keep |
| moment | `GET /v1/moments/map` | `moments.getMap` | query `from, to` | keep |
| moment | `GET /v1/reports/monthly` | `moments.getMonthlyReport` | query `month` | keep |
| moment | `GET /v1/memories/resurface` | `moments.getResurfacedMemory` | 无 | keep |
| anniversary | `GET/POST /v1/anniversaries` | `together.list/createAnniversaries` | create: `title, date, countMode, repeatsYearly, pinned, coverRef, message` | keep |
| anniversary | `GET/PATCH/DELETE /v1/anniversaries/:id` | `together.get/update/deleteAnniversary` | update 同 create | keep |
| anniversary | `POST /v1/anniversaries/:id/cover-upload` | `together.requestCoverUpload` | `contentType, size` | keep |
| todo | `GET/POST /v1/todos` | `together.list/createTodos` | create: `title, dueAt` | keep |
| todo | `PATCH/DELETE /v1/todos/:id` | `together.setTodoCompleted/deleteTodo` | patch: `completed` | keep |
| wish | `GET/POST /v1/wishes` | `together.list/createWishes` | create: `title, description` | keep |
| wish | `PATCH/DELETE /v1/wishes/:id` | `together.setWishCompleted/deleteWish` | patch: `completed, linkedMomentId` | keep |
| dailyquestion | `GET /v1/daily-question` | `together.getDailyQuestion` | 无 | keep |
| dailyquestion | `PUT /v1/daily-question/answer` | `together.submitDailyAnswer` | `questionId, text` | keep |
| dailyquestion | `GET /v1/daily-questions` | `together.listDailyQuestions` | 无 | keep |

稳定响应字段包括认证的 `userId, accessToken, refreshToken, accessExpiresAt`，上传的 `objectKey, uploadURL/uploadUrl, contentType, expiresAt`，时间线的 `moments, nextCursor`，以及日历/地图/月报字段 `date, count, momentId, latitude, longitude, occurredAt, placeName, month, momentCount, photoCount, topPlace, commonStatus`。内部 DTO 的 `json:"-"` 字段不属于 HTTP 契约。

## 数据表与字段

| 表 | 字段分类 | 结论与原因 |
| --- | --- | --- |
| `users` | 全部 keep | 微信身份、显示名、头像和审计时间均被 account 使用。 |
| `refresh_tokens` | 全部 keep | refresh rotation、撤销和过期检查需要。 |
| `couples` | 全部 keep | 活跃关系、结束时间和生成列唯一约束防止一人同时处于两个活跃关系。 |
| `invite_codes` | 全部 keep | 邀请预览、过期和一次消费需要。 |
| `upload_grants` | 全部 keep | `purpose/resource_id` 防止头像、小事照片、纪念日封面跨用途或跨资源消费。 |
| `moments` | 除下列两项外 keep | `space_type` 与 `visibility` 含义不同；`source_moment_id` 保留分享来源；`daily_question_id` 保留投影幂等；三个时间字段含义不同；`deleted_at` 支持软删除。 |
| `moments.status_code` | keep 后 rename | 当前适配器继续使用；新基线重命名为单一 `status VARCHAR(80) NOT NULL DEFAULT ''`。 |
| `moments.kind` | derive | `daily_question_id IS NOT NULL` 可完整推导每日一问投影；新基线删除。 |
| `moment_photos` | 全部 keep | 保持最多九图及稳定排序。 |
| `moment_perspectives` | 全部 keep | 双视角正文和每作者唯一性。 |
| `moment_perspective_photos` | 全部 keep | 双视角最多三图及稳定排序。 |
| `moment_reactions` | 全部 keep | 回应类型、文字回应和用户级幂等。 |
| `presences` | 全部 keep | “此刻状态”独立于 Moment 发布状态，含失效时间。 |
| `nudges` | 全部 keep | 轻互动、收发双方和冷却查询需要。 |
| `notification_preferences` | 全部 keep | 四项用户偏好均有 API 读写。 |
| `notification_deliveries` | 全表 delete | 生产代码没有任何读写，仅为未实现的投递记录预留。 |
| `anniversaries` | 全部 keep | 完整纪念日、置顶、计数方式、封面、寄语和系统纪念日需要。 |
| `todos` | 全部 keep | 创建者、完成人、到期时间与完成状态需要。 |
| `wishes` | 全部 keep | 完成人及关联实现愿望的 Moment 需要。 |
| `daily_questions` | 全部 keep | 每情侣每日唯一问题、来源、上下文版本与解锁时间需要。 |
| `daily_answers` | 全部 keep | 每题每用户唯一答案和提交时间需要。 |

## 索引与约束

以下均进入新基线；命名可在不改变用途的前提下统一：

- 身份/会话：`uq_users_openid`、`uq_refresh_hash`、`idx_refresh_user`。
- Couple/邀请：`chk_couple_order`、`uq_active_user_a/b`、`idx_couple_a/b_status`、`uq_invite_code` 及所有用户外键。
- 上传：`uq_upload_object`、`idx_upload_user`、`idx_upload_purpose_resource`。
- Moment：共享/个人时间线与日历组合索引、`idx_moments_source`、`uq_moment_daily_question`、空间/可见性检查、来源/作者/Couple/每日问题外键。
- 照片/双视角/回应：照片顺序唯一和数量检查、`uq_moment_perspective_author`、`idx_perspectives_author_updated`、`uq_moment_reaction_type`、`idx_reactions_moment_created`、文字回应检查及级联删除外键。
- 陪伴：`idx_presences_couple_updated`、`idx_presences_user_updated`、`idx_nudges_couple_created`、`idx_nudges_sender_created`、`uq_notification_preference_user`。
- 共同生活：`idx_anniversaries_couple_date`、`idx_todos_couple_status`、`idx_wishes_couple_status` 及创建者/完成人/关联 Moment 外键。
- 每日一问：`uq_daily_question_couple_date`、`idx_daily_questions_couple_created`、`uq_daily_answer_question_user`、`idx_daily_answers_user_submitted`。

删除 `notification_deliveries` 时同时删除其索引和外键。删除 `moments.kind` 时同时删除 `chk_moment_kind_links`，并以 `daily_question_id` 的 nullable 外键加唯一键表达标准小事与每日问题投影的区别。
