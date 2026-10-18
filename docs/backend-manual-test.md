# 后端人工验收清单

## 准备

1. 复制 `.env.example` 并填入 MySQL、微信小程序和腾讯云 COS 配置。
2. 创建空数据库，运行 `make migrate-up`，再运行 `make build-server`。
3. 启动 API，并准备微信测试账号 A、B、C。A 与 B 用来绑定，C 用于越权检查。
4. 每个账号调用 `wx.login` 后请求 `POST /v1/auth/wechat`，分别保存 access token。

每次请求都记录 HTTP 状态、响应 JSON 和对应账号。建议先确认 `GET /healthz` 返回 `200`。

## 主流程

- [ ] A、B、C 均能登录，并通过 `GET /v1/users/me` 读取自己的资料。
- [ ] A 创建邀请码；B 可以预览并绑定；A、B 的 `GET /v1/couples/me` 返回同一个关系。
- [ ] A 申请一张 JPEG 上传授权，直传 COS 后发布带图记录；B 的时间线可见，照片 URL 可访问且有过期时间。
- [ ] A 分别发布纯文字和纯状态记录，B 的首页能看到最近五条及 A 的最近状态。
- [ ] A 连续申请九个上传授权、上传后按指定顺序发布；B 看到九张照片且顺序一致。
- [ ] 请求发布十个 `photoKeys`，确认返回 `400 TOO_MANY_PHOTOS` 且时间线没有新增记录。
- [ ] B 使用 A 未消费的对象键发布，确认返回 `400 INVALID_UPLOAD`。
- [ ] C 请求 A/B 的已知记录编号进行删除，确认不能删除；由于 C 没有该关系，应返回 `NO_ACTIVE_COUPLE`。
- [ ] B 删除 A 发布的记录，确认返回 `403 NOT_AUTHOR`；A 删除同一记录返回 `204`，随后双方时间线均不再出现。
- [ ] 翻页时准备多条相同或相近时间的记录，使用 `nextCursor` 继续请求，确认没有重复或遗漏。
- [ ] A 解除关系；随后 A、B 的时间线、首页、发布和删除均返回 `404 NO_ACTIVE_COUPLE`。

## 失败与恢复

- [ ] 使用过期 access token 返回 `401 UNAUTHORIZED`，refresh token 能换取新会话；退出后原 refresh token 不能继续使用。
- [ ] 上传不支持的格式或超过 10 MiB 的文件信息，确认返回 `INVALID_UPLOAD`。
- [ ] 上传授权过期或已被一次发布消费后再次使用，确认返回 `INVALID_UPLOAD`，且不会产生半条记录。
- [ ] 临时断网后重试照片上传，不重复创建记录；发布按钮在请求完成前保持禁用。
