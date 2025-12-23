## Epay Bot

轻量级易支付订单通知 Telegram 机器人，旨在提供稳定、高性能、低资源占用的订单通知。

## 功能特性

*   **无侵入性**：无需修改易支付，无需服务端权限，直接与易支付进行交互。
*   **多商户支持**：每个 Telegram 用户独立配置商户信息。
*   **实时通知**：自动轮询并推送新的支付成功订单和结算记录。
*   **智能轮询**：多次请求失败会自动调整轮询间隔，节省资源。
*   **便捷管理**：通过 Telegram 按钮菜单进行商户配置、查询订单和开关通知。

## Docker快速开始
将机器人的 `API Token` 替换到变量中
```
docker run -d \
  --name epay-bot \
  --restart always \
  -v $(pwd)/data:/app/data \
  -e TELEGRAM_BOT_TOKEN=your_token_here \
  ghcr.io/sky22333/epay-bot
```

### 使用方法

在 Telegram 中向机器人发送 `/start` 开始使用。

*   **配置商户**：点击“设置商户信息”，按提示输入易支付域名、商户ID和密钥。
*   **查询数据**：配置完成后，可查询最近订单和结算记录。
*   **开启通知**：点击“开启自动通知”以接收实时推送。

## 目录结构

*   `bot/`: 机器人核心逻辑与交互处理
*   `db/`: 数据库操作层
*   `model/`: 数据结构定义
*   `service/`: 易支付 API 客户端与轮询服务
*   `main.go`: 程序入口


#### UA请求头
```
EpayBot-Client/1.0 (Monitoring Orders & Settlements)
```
