# 小事成双对

一个为情侣设计的原生微信小程序：记录文字、状态、位置与最多九张照片，在两个人的共同时间线上保存日常。

项目包含 Go DDD 后端、MySQL 数据库与原生微信小程序前端。快速开始、配置方法和常见问题见 [小程序开发指南](docs/miniprogram-development.md)，代码职责见 [小程序架构说明](docs/miniprogram-architecture.md)。

```bash
npm install
docker compose up -d
make migrate-up
make test
make build-server
```

请勿提交真实 AppID、Token、COS 密钥、`.env` 或 `project.private.config.json`。
