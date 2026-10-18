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

## 情侣关系

| 方法与路径 | 请求体/参数 | 成功 |
| --- | --- | --- |
| `POST /v1/couples/invitations` | 无 | `201`，邀请码和过期时间 |
| `GET /v1/couples/invitations/:code` | 路径邀请码 | 邀请预览 |
| `POST /v1/couples/bind` | `{"code":"ABC123"}` | `201`，情侣关系 |
| `GET /v1/couples/me` | 无 | 当前情侣关系 |
| `DELETE /v1/couples/me` | 无 | `204` |

## 照片与共同记录

### 获取照片上传授权

`POST /v1/uploads`

```json
{"contentType":"image/jpeg","size":245760}
```

返回 `objectKey`、`contentType`、`expiresAt` 和 `uploadUrl`。客户端用 `PUT` 将原始文件上传到 `uploadUrl`。支持 JPEG、PNG、WebP，单张不超过 10 MiB。

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
