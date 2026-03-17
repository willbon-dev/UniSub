# 配置说明

本文档按 `docs/config.example.yaml` 的结构，逐项说明 UniSub 的配置含义、作用和使用场景。

- 示例配置文件：[docs/config.example.yaml](config.example.yaml)
- Happ 应用管理文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li
- Happ 路由文档：https://www.happ.su/main/zh/dev-docs/routing
- Happ Provider ID 文档：https://www.happ.su/main/zh/dev-docs/provide-id
- Happ 路由编辑在线工具：https://utils.docs.rw/happ-rb

## 顶层结构

```yaml
server:
subscriptions:
```

### `server`

#### `server.listen`

- 含义：HTTP 服务监听地址。
- 作用：决定 UniSub 绑定的 IP 和端口。
- 示例：`127.0.0.1:8080`

#### `server.read_timeout`

- 含义：读取客户端请求的超时时间。
- 作用：防止慢连接长期占用服务端资源。
- 示例：`10s`

#### `server.write_timeout`

- 含义：向客户端写回订阅内容的超时时间。
- 作用：避免客户端接收过慢导致连接长期挂起。
- 示例：`30s`

#### `server.shutdown_timeout`

- 含义：服务优雅关闭时的最长等待时间。
- 作用：控制进程退出前等待中的请求收尾时间。
- 示例：`10s`

#### `server.fetch_timeout`

- 含义：远程订阅拉取的默认超时时间。
- 作用：为 `remote` 类型源提供统一的网络请求超时上限。
- 示例：`20s`

#### `server.max_response_bytes`

- 含义：远程订阅响应体的最大允许字节数。
- 作用：避免异常大响应拖垮内存或处理流程。
- 示例：`8388608`

### `subscriptions`

`subscriptions` 是统一订阅列表。每个元素对应一个可通过 `/subscribe?secret=...` 访问的聚合订阅。

#### `subscriptions[].name`

- 含义：该统一订阅的名称。
- 作用：用于内部区分配置项和日志定位。

#### `subscriptions[].secret`

- 含义：该统一订阅的访问密钥，必须是 UUID。
- 作用：客户端通过 `GET /subscribe?secret=...` 使用它访问对应订阅。

#### `subscriptions[].default_platform`

- 含义：默认输出平台。
- 作用：当请求中未显式传 `platform` 参数时，决定返回 `V2rayN` 还是 `Happ` 格式。
- 可选值：`V2rayN`、`Happ`

#### `subscriptions[].platform_options`

- 含义：平台定制输出配置。
- 作用：当前仅用于 `Happ` 平台注释行和路由行生成。

### `subscriptions[].sources`

每个 `source` 表示一个节点来源，支持手动录入和远程拉取两种方式。

#### `subscriptions[].sources[].name`

- 含义：源名称。
- 作用：用于日志和错误定位。

#### `subscriptions[].sources[].type`

- 含义：源类型。
- 作用：决定当前源是读取 `entries` 还是拉取 `url`。
- 可选值：`manual`、`remote`

#### `subscriptions[].sources[].prefix`

- 含义：节点名前缀。
- 作用：会附加到该源下所有节点的显示名称前，便于区分来源。
- 示例：`[JP] `

#### `subscriptions[].sources[].entries`

- 含义：手动节点列表。
- 作用：仅在 `type: manual` 时使用，支持直接填写 `vmess://`、`vless://`、`ss://`、`trojan://` 等链接。

#### `subscriptions[].sources[].remote_type`

- 含义：远程订阅解析类型。
- 作用：决定 UniSub 如何解释远程 HTTP 响应内容。
- 当前支持：`base64_lines`

#### `subscriptions[].sources[].url`

- 含义：远程订阅地址。
- 作用：仅在 `type: remote` 时使用，用于拉取上游订阅内容。

#### `subscriptions[].sources[].refresh_interval`

- 含义：远程订阅缓存刷新间隔。
- 作用：在间隔内复用缓存，超过间隔后重新拉取上游。
- 示例：`30m`

#### `subscriptions[].sources[].timeout`

- 含义：当前远程源的单独请求超时。
- 作用：覆盖 `server.fetch_timeout`，适合为个别慢源单独放宽或收紧超时。
- 示例：`15s`

#### `subscriptions[].sources[].include_patterns`

- 含义：节点名正则白名单。
- 作用：只有名称匹配这些 Go regexp 的节点会被保留。

#### `subscriptions[].sources[].exclude_patterns`

- 含义：节点名正则黑名单。
- 作用：名称匹配这些 Go regexp 的节点会被过滤掉。

## Happ 配置

`subscriptions[].platform_options.happ` 下的字段只在 `platform=Happ` 或默认平台为 `Happ` 时生效。UniSub 会把这些配置转换为 Happ 所需的首行 `routing` 或注释行。

### 路由

#### `subscriptions[].platform_options.happ.routing`

- 含义：Happ 路由链接。
- 作用：作为返回订阅的首行输出，供 Happ 导入或自动启用路由配置。
- 输出形式：`happ://routing/...`
- 官方文档：https://www.happ.su/main/zh/dev-docs/routing
- 路由编辑器：https://utils.docs.rw/happ-rb

### 应用管理基础参数

#### `subscriptions[].platform_options.happ.profile_update_interval`

- 含义：订阅自动更新时间间隔，单位小时。
- 作用：在 Happ 中创建定时更新任务。
- 输出示例：`#profile-update-interval: 1`
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.profile_title`

- 含义：订阅显示名称。
- 作用：控制 Happ 中订阅条目的标题。
- 输出示例：`#profile-title: UniSub`
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_userinfo`

- 含义：订阅流量和到期信息。
- 作用：在 Happ 中显示上传、下载、总流量和到期时间。
- 输出示例：`#subscription-userinfo: upload=0; download=0; total=0; expire=1790951622`
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.support_url`

- 含义：支持页面链接。
- 作用：在订阅项旁显示支持入口按钮。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.profile_web_page_url`

- 含义：订阅主页链接。
- 作用：在订阅项旁显示站点入口按钮。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.announce`

- 含义：订阅公告文本。
- 作用：在 Happ 中展示订阅公告信息。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.routing_enable`

- 含义：是否允许全局路由功能。
- 作用：控制 Happ 中路由能力的启用或禁用。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.custom_tunnel_config`

- 含义：自定义隧道配置。
- 作用：向桌面版 Happ 的 sing-box 核心传入隧道配置。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

### Provider ID 相关参数

以下字段属于高级参数，通常需要配合 `provider_id` 使用。

#### `subscriptions[].platform_options.happ.provider_id`

- 含义：Provider ID。
- 作用：启用 Happ 的高级订阅管理与部分应用设置控制能力。
- 官方文档：https://www.happ.su/main/zh/dev-docs/provide-id

#### `subscriptions[].platform_options.happ.new_url`

- 含义：新的完整订阅地址。
- 作用：当旧订阅地址不可用时，自动替换为新的完整 URL。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.new_domain`

- 含义：新的订阅域名。
- 作用：只替换订阅地址的域名，保留其余路径和参数。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fallback_url`

- 含义：备用订阅地址。
- 作用：主地址不可访问、返回 300-599 或超时时自动回退。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

### 高级与应用设置参数

#### `subscriptions[].platform_options.happ.no_limit_enabled`

- 含义：为全部协议启用 No Limit 模式。
- 作用：提升 xray-core 内存限制，改善稳定性与性能。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.no_limit_xhttp_enabled`

- 含义：仅为 xhttp 启用 No Limit 模式。
- 作用：只对 xhttp 场景应用 No Limit 模式。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_always_hwid_enable`

- 含义：强制启用 HWID。
- 作用：防止用户在 Happ 中关闭 HWID 转发。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.notification_subs_expire`

- 含义：订阅到期通知开关。
- 作用：在到期前向用户发送续订提醒。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.hide_settings`

- 含义：隐藏服务器设置。
- 作用：禁止用户查看和编辑订阅中的服务器配置。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.server_address_resolve_enable`

- 含义：服务端地址预解析开关。
- 作用：在连接前由 Happ 先解析服务器域名。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.server_address_resolve_dns_domain`

- 含义：域名预解析使用的 DoH 域名。
- 作用：指定 Happ 进行解析时使用的 DNS 服务地址。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.server_address_resolve_dns_ip`

- 含义：域名预解析使用的 DNS IP。
- 作用：配合上面的 DNS 域名指定解析目标。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_autoconnect`

- 含义：启动应用时自动连接。
- 作用：打开 Happ 后自动连接服务器。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_autoconnect_type`

- 含义：自动连接策略。
- 作用：指定自动连接到哪一类服务器，例如最近使用节点。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_ping_onopen_enabled`

- 含义：打开应用时自动 Ping。
- 作用：在进入 Happ 时自动测试节点延迟。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_auto_update_enable`

- 含义：全局自动更新开关。
- 作用：统一控制 Happ 内所有订阅的自动更新能力。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fragmentation_enable`

- 含义：全局分片开关。
- 作用：统一控制 Happ 对订阅分片能力的启用状态。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fragmentation_packets`

- 含义：分片包类型配置。
- 作用：控制 Happ 分片参数中的 `packets`。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fragmentation_length`

- 含义：分片长度配置。
- 作用：控制 Happ 分片参数中的 `length`。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fragmentation_interval`

- 含义：分片间隔配置。
- 作用：控制 Happ 分片参数中的 `interval`。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.fragmentation_maxsplit`

- 含义：最大分片数配置。
- 作用：控制 Happ 分片拆分上限。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.noises_enable`

- 含义：噪声开关。
- 作用：控制 Happ 的噪声功能是否启用。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.noises_type`

- 含义：噪声类型。
- 作用：指定 Happ 使用的噪声模式。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.noises_packet`

- 含义：噪声包配置。
- 作用：控制 Happ 噪声功能中的包参数。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.noises_delay`

- 含义：噪声延迟配置。
- 作用：控制 Happ 噪声发包延迟。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.noises_applyto`

- 含义：噪声应用范围。
- 作用：指定噪声对哪些连接或节点生效。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.ping_type`

- 含义：Ping 类型。
- 作用：设置 Happ 使用 `via Proxy`、`TCP` 或 `ICMP` 方式测试延迟。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.check_url_via_proxy`

- 含义：代理 Ping 使用的检测 URL。
- 作用：在 `ping_type` 为代理检测时指定检查地址。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.change_user_agent`

- 含义：订阅请求使用的 User-Agent。
- 作用：让 Happ 拉取订阅时使用指定请求头。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.app_auto_start`

- 含义：应用自动启动开关。
- 作用：控制 Happ 在系统启动后的自动启动行为。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscription_auto_update_open_enable`

- 含义：打开应用时检查自动更新。
- 作用：控制 Happ 在进入应用时触发订阅自动更新检查。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.per_app_proxy_mode`

- 含义：分应用代理模式。
- 作用：指定 Happ 的分应用代理工作模式。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.per_app_proxy_list`

- 含义：分应用代理列表。
- 作用：定义按应用分流时的目标应用集合。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sniffing_enable`

- 含义：嗅探开关。
- 作用：控制 Happ 的流量嗅探能力。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscriptions_collapse`

- 含义：订阅列表默认折叠。
- 作用：控制 Happ 中订阅分组的折叠展示状态。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.subscriptions_expand_now`

- 含义：立即展开订阅列表。
- 作用：控制 Happ 当前是否直接展开显示订阅项。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.ping_result`

- 含义：预设 Ping 结果文本。
- 作用：为 Happ 展示的延迟结果或状态文本提供初始值。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.mux_enable`

- 含义：MUX 开关。
- 作用：控制 Happ 是否启用连接复用。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.mux_tcp_connections`

- 含义：TCP MUX 连接数。
- 作用：设置 Happ 的 TCP 复用连接数量。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.mux_xudp_connections`

- 含义：XUDP MUX 连接数。
- 作用：设置 Happ 的 XUDP 复用连接数量。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.mux_quic`

- 含义：QUIC MUX 配置。
- 作用：控制 Happ 中与 QUIC 相关的复用行为。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.proxy_enable`

- 含义：代理总开关。
- 作用：控制 Happ 是否启用代理功能。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.tun_enable`

- 含义：TUN 开关。
- 作用：控制 Happ 是否启用 TUN 模式。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.tun_mode`

- 含义：TUN 模式类型。
- 作用：设置 Happ 的 TUN 工作模式。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.tun_type`

- 含义：TUN 内核类型。
- 作用：指定 Happ 使用的 TUN 实现。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.exclude_routes`

- 含义：排除路由列表。
- 作用：定义不经过 TUN 或代理处理的路由规则。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.color_profile`

- 含义：订阅配色方案。
- 作用：设置 Happ 中订阅卡片或条目的颜色主题。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_info_color`

- 含义：订阅信息区域颜色。
- 作用：控制附加订阅信息的显示颜色。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_info_text`

- 含义：订阅信息文本。
- 作用：在 Happ 中显示自定义附加说明文本。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_info_button_text`

- 含义：订阅信息按钮文本。
- 作用：定义附加信息区域按钮的显示文字。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_info_button_link`

- 含义：订阅信息按钮链接。
- 作用：定义附加信息区域按钮的跳转地址。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_expire`

- 含义：订阅到期开关。
- 作用：控制 Happ 中是否显示到期相关提示。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

#### `subscriptions[].platform_options.happ.sub_expire_button_link`

- 含义：到期按钮链接。
- 作用：为到期提示按钮设置跳转地址。
- 官方文档：https://www.happ.su/main/zh/dev-docs/ying-yong-guan-li

## 输出规则

### `V2rayN`

- 不注入 Happ 路由或注释行。
- 只输出去重后的节点链接。

### `Happ`

- 若配置了 `routing`，会先输出路由行。
- 其余 Happ 配置会按 Happ 订阅注释规则转换后插入到节点列表前。
- 布尔值会输出为 `1` 或 `0`。
- 字段名会自动转换为 Happ 所需的中横线格式，例如 `profile_update_interval` 会输出为 `#profile-update-interval: 1`。
