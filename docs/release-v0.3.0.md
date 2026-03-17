# UniSub v0.3.0 Release Notes

发布日期：`2026-03-17`

## 概览

`v0.3.0` 重点补齐了 Happ 平台的配置能力和配置文档体系。

- 支持更多 Happ 订阅注释字段输出
- 新增独立配置文档，逐项说明配置含义、作用和使用方式
- 为 Happ 配置项补充官方文档入口和路由编辑工具链接
- 移除 `subscription_lines` 原始注入方式，统一改为结构化配置

## 主要变更

### 1. Happ 配置能力增强

UniSub 现在可以从 `subscriptions[].platform_options.happ` 读取并输出更多 Happ 配置项，包括：

- 应用管理基础字段，如 `profile_update_interval`、`profile_title`、`subscription_userinfo`、`support_url`、`profile_web_page_url`、`announce`
- 路由与网络相关字段，如 `routing_enable`、`custom_tunnel_config`、`proxy_enable`、`tun_enable`、`tun_mode`、`tun_type`、`exclude_routes`
- 自动化与行为控制字段，如 `subscription_autoconnect`、`subscription_autoconnect_type`、`subscription_ping_onopen_enabled`、`subscription_auto_update_enable`、`subscription_auto_update_open_enable`、`app_auto_start`
- 高级与性能字段，如 `fragmentation_*`、`noises_*`、`mux_*`、`ping_type`、`check_url_via_proxy`、`change_user_agent`
- Provider ID 相关字段，如 `provider_id`、`new_url`、`new_domain`、`fallback_url`

当请求平台为 `Happ` 时，UniSub 会按 Happ 订阅规范输出：

- 首行 `routing`
- 后续注释行，例如 `#profile-update-interval: 1`
- 最后的去重节点列表

### 2. 配置文档独立拆分

新增独立配置文档：

- [docs/config.md](config.md)

文档内容包括：

- `server`、`subscriptions`、`sources` 的逐项说明
- `platform_options.happ` 每个已支持字段的含义和作用
- 每个 Happ 配置项对应的官方文档链接
- Happ 路由文档和在线路由编辑工具入口

相关入口文档：

- Happ 应用管理文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li
- Happ 路由文档：https://www.happ.su/main/zh/dev-docs/routing
- Happ Provider ID 文档：https://www.happ.su/main/zh/dev-docs/provide-id
- Happ 路由编辑器：https://utils.docs.rw/happ-rb

## 验证

发布前已执行：

```bash
env GOCACHE=/tmp/gocache GOPATH=/tmp/gopath go test ./...
```

测试结果：通过。
