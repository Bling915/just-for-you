# 小程序开发指南

## 环境准备

需要 Node.js 20+、Go、MySQL 8、微信开发者工具。仓库根目录执行 `npm install`；微信开发者工具导入仓库根目录后，选择“工具 → 构建 npm”。`project.config.json` 已指定 `miniprogram/` 为源码根目录。

复制 `project.private.config.json.example` 为本机的 `project.private.config.json`，在开发者工具里选择自己的 AppID。不要把真实 AppID 或密钥写入受版本控制的文件。

## 启动

```bash
docker compose up -d
make migrate-up
make build-server
./server/bin/api
```

开发默认 API 地址为 `http://127.0.0.1:8080`。真机必须使用已备案的 HTTPS 域名，并在微信公众平台配置 request 合法域名；COS 上传域名也必须加入 request 合法域名。照片使用 `wx.request` PUT，不使用 `uploadFile`。

## 验证

```bash
make test
make verify
```

`make verify` 会运行小程序测试、Go 全量测试、三个业务域 race 测试和 `go vet`。工具编译报错时，先检查是否执行了“构建 npm”、AppID 是否有效、后端地址是否可访问。登录失败可检查开发者工具网络面板，但不要打印微信 code、Token 或签名上传 URL。

## 调试建议

- 两个微信账号分别在真机登录，使用一次性六位邀请码绑定。
- 慢网或 COS 失败时，发布页保留草稿和成功项，只重新上传失败照片。
- 修改本地 API 地址时只改本机私有配置；上线前恢复合法域名校验。
