# 后端 API（MVP）

服务默认监听 `http://localhost:8080`。除登录、刷新和退出外，请求都需要：

```http
Authorization: Bearer <accessToken>
Content-Type: application/json
```

成功响应统一为 `{"data": ...}`；失败响应为 `{"error":{"code":"...","message":"..."}}`。时间均为 RFC 3339 格式。

## 身份

| 方法与路径 | 请求体 | 成功 |
| --- | --- | --- |
| `POST /v1/auth/wechat` | `{"code":"wx.login 返回值"}` | 会话（accessToken、refreshToken） |
| `POST /v1/auth/refresh` | `{"refreshToken":"..."}` | 新会话 |
| `POST /v1/auth/logout` | `{"refreshToken":"..."}` | `204` |
| `GET /v1/users/me` | 无 | 当前用户 |
| `PUT /v1/users/me` | `{"displayName":"南方","avatarRef":"..."}` | 更新后的用户 |
| `POST /v1/users/me/avatar-upload` | `{"contentType":"image/jpeg","size":245760}` | 头像专用上传授权 |

头像支持 JPEG、PNG、WebP，最大 5 MiB。客户端必须先 PUT 到授权地址，再把返回的 `objectKey` 作为 `avatarRef` 更新资料；若任一步失败，应继续保留旧头像。用户和伴侣 DTO 同时返回私有短期读取地址 `avatarUrl`。

## 情侣关系

| 方法与路径 | 请求体/参数 | 成功 |
| --- | --- | --- |
| `POST /v1/couples/invitations` | 无 | `201`，邀请码和过期时间 |
| `GET /v1/couples/invitations/:code` | 路径邀请码 | 邀请预览 |
| `POST /v1/couples/bind` | `{"code":"ABC123"}` | `201`，情侣关系 |
| `POST /v1/couples/solo-test` | 无 | 仅 `ENABLE_SOLO_TEST=true` 时注册；创建虚拟搭子并返回 `201` |
| `GET /v1/couples/me` | 无 | 当前情侣关系 |
| `DELETE /v1/couples/me` | 无 | `204` |

## 照片与共同记录

### 获取照片上传授权

`POST /v1/uploads`

```json
{"contentType":"image/jpeg","size":245760}
```

返回 `objectKey`、`contentType`、`expiresAt` 和 `uploadUrl`。客户端用 `PUT` 将原始文件上传到 `uploadUrl`，并设置与申请时完全一致的 `Content-Type` 请求头。支持 JPEG、PNG、WebP，单张不超过 10 MiB；发布前服务端还会从 COS 验证对象确实存在，以及实际类型和大小。

### 发布记录

`POST /v1/moments`

```json
{
  "text":"今天也认真上课了",
  "status":"in_class",
  "placeName":"图书馆",
  "photoKeys":["users/1/2026/10/uuid.jpg"]
}
```

`status` 可为 `in_class`、`miss_you`、`tired` 或空字符串。四类内容至少填写一项；文字最多 500 个 Unicode 字符，地点最多 80 个，照片最多九张，顺序按 `photoKeys` 保留。

### 时间线、首页和删除

- `GET /v1/moments?cursor=<游标>&limit=20`：倒序返回当前情侣关系的记录；默认 20，服务端最大 50。响应含 `moments` 和可选 `nextCursor`。
- `GET /v1/home`：返回 `partnerLatestStatus` 和最近五条 `recent`。
- `DELETE /v1/moments/:id`：仅作者可删除，成功返回 `204`。

照片项包含对象键、显示顺序和短期有效的私有读取 `url`。

## 纪念日

| 方法与路径 | 用途 |
| --- | --- |
| `GET /v1/anniversaries` | 获取纪念日列表（置顶优先） |
| `POST /v1/anniversaries` | 新建纪念日 |
| `GET /v1/anniversaries/:id` | 获取纪念日详情 |
| `PATCH /v1/anniversaries/:id` | 更新标题、日期、模式、留言、置顶及封面引用 |
| `POST /v1/anniversaries/:id/cover-upload` | 获取该纪念日封面的上传授权 |
| `DELETE /v1/anniversaries/:id` | 删除非系统纪念日 |

创建/更新字段包括 `title`、`date`（`YYYY-MM-DD`）、`countMode`（`elapsed` 或 `countdown`）、`repeatsYearly`、`pinned`、`message`、`coverRef`。系统生成的“在一起”纪念日不可删除。闰日按每年重复时，非闰年在 2 月 28 日纪念。

## 主要错误码

| HTTP | code | 含义 |
| --- | --- | --- |
| 400 | `INVALID_REQUEST` / `INVALID_ARGUMENT` | JSON 或参数不合法 |
| 400 | `EMPTY_MOMENT` | 记录没有任何有效内容 |
| 400 | `TEXT_TOO_LONG` / `PLACE_TOO_LONG` | 字数超限 |
| 400 | `INVALID_STATUS` | 快捷状态不支持 |
| 400 | `TOO_MANY_PHOTOS` | 超过九张照片 |
| 400 | `INVALID_UPLOAD` | 上传授权无效、过期、已消费或不属于当前用户 |
| 400 | `INVITE_INVALID` / `INVITE_EXPIRED` / `INVITE_USED` | 邀请码不可用 |
| 400 | `CANNOT_BIND_SELF` / `ALREADY_BOUND` | 不能绑定自己或已有关系 |
| 401 | `UNAUTHORIZED` / `INVALID_REFRESH` | 登录已失效 |
| 403 | `NOT_AUTHOR` | 试图删除他人的记录 |
| 404 | `NO_ACTIVE_COUPLE` | 没有有效情侣关系 |
| 404 | `MOMENT_NOT_FOUND` / `USER_NOT_FOUND` | 资源不存在 |
