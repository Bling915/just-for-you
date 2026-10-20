# 小事成双对

一个为情侣设计的原生微信小程序：记录文字、状态、位置与最多九张照片，在两个人的共同时间线上保存日常。

项目包含 Go DDD 后端、MySQL 数据库与原生微信小程序前端。快速开始、配置方法和常见问题见 [小程序开发指南](docs/miniprogram-development.md)，代码职责见 [小程序架构说明](docs/miniprogram-architecture.md) 和 [后端架构说明](docs/backend-architecture.md)。

```bash
npm install
docker compose up -d
make migrate-up
make test
make build-server
```

第二阶段已包含情侣邀请与单人测试、状态互动、可编辑动态、纪念日、共同待办、愿望清单、日历、回忆地图、月度回顾及订阅通知设置。数据库升级请依次执行：

```bash
cd server
go run ./cmd/migrate -direction up -version 1 # 仅全新数据库
go run ./cmd/migrate -direction up -version 2
```

服务器除原有变量外，可选配置 `WECHAT_TEMPLATE_MOMENT_ID`、`WECHAT_TEMPLATE_PRESENCE_ID`、`WECHAT_TEMPLATE_NUDGE_ID`。未配置模板时通知入口会显示不可用，但不会影响动态、状态和互动写入。

双账号验收、单人模式、位置权限与订阅消息限制见 [一二阶段验收清单](docs/testing/phase-one-two-manual-checklist.md)。

请勿提交真实 AppID、Token、COS 密钥、`.env` 或 `project.private.config.json`。
