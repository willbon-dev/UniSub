# 配置说明

本文档按 [docs/config.example.yaml](config.example.yaml) 的结构，详细说明 UniSub 当前支持的配置项、适用场景和输出行为。

UniSub 现在支持三种目标平台：

- `V2rayN`
- `Happ`
- `Clash`

同时支持两种 source 类型：

- `manual`：直接在配置里写节点
- `remote`：从远程 URL 拉取订阅

## 顶层结构

```yaml
server:
subscriptions:
```

## server

### `server.listen`

- 含义：HTTP 服务监听地址
- 默认值：`127.0.0.1:8080`
- 示例：`0.0.0.0:8080`

### `server.read_timeout`

- 含义：读取客户端请求的超时时间
- 默认值：`10s`

### `server.write_timeout`

- 含义：写回订阅响应的超时时间
- 默认值：`30s`

### `server.shutdown_timeout`

- 含义：服务优雅关闭时等待中的请求收尾时间
- 默认值：`10s`

### `server.fetch_timeout`

- 含义：远程 source 的默认抓取超时时间
- 默认值：`20s`
- 说明：如果某个 `remote source` 自己写了 `timeout`，则优先使用 source 自己的值

### `server.max_response_bytes`

- 含义：远程 source 响应体的最大允许大小
- 默认值：`8388608`
- 说明：超过这个大小会直接报错，防止异常响应占满内存

## subscriptions

`subscriptions` 是统一订阅入口列表。每个元素对应一个可以通过 `/subscribe?secret=...` 访问的聚合订阅。

```yaml
subscriptions:
  - name: default-subscription
    secret: "123e4567-e89b-42d3-a456-426614174000"
    default_platform: V2rayN
    platform_options:
      happ: {}
      clash: {}
    sources: []
```

### `subscriptions[].name`

- 含义：订阅名称
- 作用：用于日志定位和内部区分

### `subscriptions[].secret`

- 含义：订阅访问密钥
- 要求：必须是 UUID，且在所有订阅之间唯一
- 使用方式：客户端通过 `GET /subscribe?secret=...` 访问对应订阅

### `subscriptions[].default_platform`

- 含义：默认输出平台
- 可选值：`V2rayN`、`Happ`、`Clash`
- 作用：当请求中没有显式传 `platform` 参数时，使用这个平台输出

示例：

```yaml
default_platform: Clash
```

## platform_options

`platform_options` 用于平台特定输出行为。

```yaml
platform_options:
  happ:
    routing: "happ://routing/onadd/..."
  clash:
    template: "./Self.ini"
```

### `platform_options.happ`

`Happ` 平台继续沿用原有实现：

- 支持 `routing`
- 支持各类 Happ 注释行
- 请求 `platform=Happ` 时，会在节点前插入 routing 和注释行

### `platform_options.clash`

#### `platform_options.clash.template`

- 含义：Clash 模板来源
- 类型：字符串
- 支持：
  - 本地文件路径，例如 `./Self.ini`
  - 远程 URL，例如 `https://example.com/Self.ini`

示例：

```yaml
platform_options:
  clash:
    template: "./Self.ini"
```

说明：

- 请求平台为 `Clash` 时必须能读到模板，否则会报错
- 当前实现按“最小 subconverter 兼容”处理模板
- 当前支持读取：
  - `custom_proxy_group=...`
  - `ruleset=...`
  - `clash_rule_base=...`
- UniSub 会将聚合后的 `proxies` 注入到最终生成的 Clash YAML 中

当前边界：

- 这不是完整的外部 `subconverter` 替代实现
- 不保证覆盖 ACL4SSR / subconverter 的全部高级语义
- 当前版本重点保证“能生成合法 Clash YAML 并把聚合节点注入进去”

## sources

每个 `source` 表示一个节点来源。

```yaml
sources:
  - name: source-name
    type: manual
    platforms: [Clash]
    style: clash_proxy
```

所有 source 都有以下通用字段。

### `subscriptions[].sources[].name`

- 含义：source 名称
- 作用：错误定位、日志输出

### `subscriptions[].sources[].type`

- 含义：source 类型
- 可选值：`manual`、`remote`

### `subscriptions[].sources[].platforms`

- 含义：该 source 适用的平台列表
- 当前是必填字段
- 可选值：`V2rayN`、`Happ`、`Clash`

示例：

```yaml
platforms: [V2rayN, Happ]
platforms: [Clash]
```

作用：

- 请求某个平台时，只会聚合 `platforms` 包含该平台的 source
- 不同平台之间不会再无差别混用 source

### `subscriptions[].sources[].style`

- 含义：当前 source 的解析风格
- 当前是必填字段

支持的 style：

- `link_line`
- `link_lines_base64`
- `clash_proxy`
- `clash_proxies_yaml`

详细说明见下文“style 一览”。

### `subscriptions[].sources[].prefix`

- 含义：给该 source 下的节点名称加统一前缀
- 作用：区分来源

示例：

```yaml
prefix: "[JP] "
```

### `subscriptions[].sources[].include_patterns`

- 含义：节点名白名单正则
- 作用：只有名称匹配这些正则的节点才会保留

### `subscriptions[].sources[].exclude_patterns`

- 含义：节点名黑名单正则
- 作用：名称匹配这些正则的节点会被过滤掉

说明：

- `include_patterns` / `exclude_patterns` 都基于节点名生效
- 对链接节点读取显示名
- 对 Clash 节点读取 `name`

## style 一览

### `link_line`

- 适用 source 类型：`manual`
- 内容格式：单条代理链接
- 适用平台：`V2rayN`、`Happ`
- 常见内容：
  - `vmess://...`
  - `vless://...`
  - `ss://...`
  - `trojan://...`

注意：

- `link_line` 不能与 `Clash` 平台组合
- 如果 `platforms` 里包含 `Clash`，配置校验会报错

### `link_lines_base64`

- 适用 source 类型：`remote`
- 内容格式：整个 HTTP 响应为 Base64，解码后得到多行代理链接
- 适用平台：`V2rayN`、`Happ`

说明：

- 这就是传统机场订阅常见的返回方式
- 当前旧配置中的 `remote_type: base64_lines` 会自动兼容映射到这个 style

### `clash_proxy`

- 适用 source 类型：`manual`
- 内容格式：单条 Clash 风格代理对象
- 适用平台：`Clash`

说明：

- `entries` 中可以直接写字符串形式的对象
- 也可以直接写原生 YAML 对象
- 当前实现会同时兼容这两种写法

### `clash_proxies_yaml`

- 适用 source 类型：`remote`
- 内容格式：远程返回 Clash YAML，并从其中读取 `proxies:`
- 适用平台：`Clash`

说明：

- 如果上游默认返回的是 Base64 链接订阅，而不是 Clash YAML，那么这个 style 会报解析错误
- 某些上游需要特定 `User-Agent` 才会切换到 Clash YAML，这时需要使用 `request_headers`

## manual source

### `manual + link_line`

示例：

```yaml
- name: manual-links
  type: manual
  style: link_line
  platforms: [V2rayN, Happ]
  prefix: "[Manual] "
  entries:
    - "vmess://REPLACE_WITH_YOUR_NODE"
    - "vless://REPLACE_WITH_YOUR_NODE"
    - "ss://REPLACE_WITH_YOUR_NODE"
```

说明：

- `entries` 为必填
- 每一项都是一条代理链接
- 会参与前缀改名和正则过滤

### `manual + clash_proxy`

字符串写法示例：

```yaml
- name: manual-clash-proxy
  type: manual
  style: clash_proxy
  platforms: [Clash]
  entries:
    - "{ name: 'Example Clash Node', type: vmess, server: demo.example.com, port: 443 }"
```

原生 YAML 对象写法示例：

```yaml
- name: manual-clash-proxy
  type: manual
  style: clash_proxy
  platforms: [Clash]
  entries:
    - { name: "Example Clash Node", type: vmess, server: demo.example.com, port: 443 }
```

说明：

- 两种写法当前都支持
- 会按 `name` 做过滤和前缀处理
- 会按代理内容去重

## remote source

### `remote + link_lines_base64`

示例：

```yaml
- name: upstream-base64-lines
  type: remote
  style: link_lines_base64
  platforms: [V2rayN, Happ]
  url: "https://example.com/api/v1/client/subscribe?token=REPLACE_ME"
  refresh_interval: 30m
  timeout: 15s
```

字段说明：

- `url`：必填，远程订阅地址
- `refresh_interval`：必填，缓存刷新间隔
- `timeout`：可选，当前 source 的独立请求超时

### `remote + clash_proxies_yaml`

示例：

```yaml
- name: upstream-clash-yaml
  type: remote
  style: clash_proxies_yaml
  platforms: [Clash]
  url: "https://example.com/clash-subscription.yaml"
  request_headers:
    User-Agent: "clash-verge"
  refresh_interval: 30m
  timeout: 15s
```

字段说明：

- `url`：必填
- `refresh_interval`：必填
- `timeout`：可选
- `request_headers`：可选，自定义请求头

#### `subscriptions[].sources[].request_headers`

- 含义：远程请求时附带的 HTTP 请求头
- 典型用途：某些上游只有带特定 `User-Agent` 才会返回 Clash YAML

示例：

```yaml
request_headers:
  User-Agent: "clash-verge"
```

常见场景：

- 默认访问上游返回 Base64 链接订阅
- 带 `User-Agent: clash-verge` 后返回 `proxies:` 开头的 Clash YAML
- 这种情况下必须配置 `request_headers`，否则 `clash_proxies_yaml` 会解析失败

## 兼容说明

### `remote_type`

旧配置中的：

```yaml
remote_type: base64_lines
```

仍然兼容。

兼容规则：

- 如果 `style` 已写，则以 `style` 为准
- 如果 `style` 没写，但 `remote_type: base64_lines` 已写，则自动映射为：

```yaml
style: link_lines_base64
```

新配置建议直接使用 `style`，不要再把 `remote_type` 作为主入口。

## 输出规则

## `V2rayN`

- 只聚合 `platforms` 包含 `V2rayN` 的 source
- 返回纯链接行列表
- 不注入 Happ routing / 注释
- 不生成 Clash YAML

## `Happ`

- 只聚合 `platforms` 包含 `Happ` 的 source
- 返回纯文本订阅
- 节点前会插入：
  - `routing`
  - Happ 注释行

## `Clash`

- 只聚合 `platforms` 包含 `Clash` 的 source
- 当前只接受 Clash 风格 source
- 返回 `application/yaml` 的 Clash 配置

最终输出通常包含：

- `proxies`
- `proxy-groups`
- `rules`
- `rule-providers`

## 过滤、前缀、去重

### 过滤

- `include_patterns` / `exclude_patterns` 都基于节点名
- 先判断是否命中 `include_patterns`
- 再判断是否命中 `exclude_patterns`

### 前缀

- 如果设置了 `prefix`，会修改节点显示名
- 对链接节点修改链接中的显示名
- 对 Clash 节点修改 `name`

### 去重

- 链接节点按最终字符串去重
- Clash 节点按代理关键内容去重，而不是只看名字

## 常见错误排查

### 1. `style "clash_proxies_yaml" requires Clash platform`

原因：

- 你用了 `clash_proxies_yaml`
- 但 `platforms` 里没有 `Clash`

修正：

```yaml
platforms: [Clash]
```

### 2. `decode source "...": parse clash proxies yaml ... cannot unmarshal !!str "dm1lc3M..." ...`

原因：

- 你把上游配置成了 `clash_proxies_yaml`
- 但上游实际返回的是 Base64 链接订阅，不是 Clash YAML

常见修正方式：

1. 如果上游本来就是普通订阅，改成：

```yaml
style: link_lines_base64
platforms: [V2rayN, Happ]
```

2. 如果上游支持通过 `User-Agent` 切换为 Clash YAML，则加：

```yaml
request_headers:
  User-Agent: "clash-verge"
```

### 3. `requires platform_options.clash.template for Clash output`

原因：

- 你请求了 `platform=Clash`
- 但没有配置 `platform_options.clash.template`

修正：

```yaml
platform_options:
  clash:
    template: "./Self.ini"
```

### 4. `entries must not be empty for manual source`

原因：

- `manual source` 没有写 `entries`

### 5. `remote_type "..." is not supported`

原因：

- 旧配置里写了当前版本不认识的 `remote_type`

修正：

- 优先改用 `style`
- 如果你是传统 Base64 订阅，使用：

```yaml
style: link_lines_base64
```

## 推荐实践

- 给不同来源加 `prefix`，便于在客户端里区分节点来源
- 只在确实需要的平台上声明 `platforms`，不要把所有 source 都无脑挂到所有平台
- 新配置统一使用 `style`，不要继续依赖 `remote_type`
- 上游 Clash 订阅如果需要指定客户端身份，优先使用 `request_headers.User-Agent`
- 把你自己的真实 URL、token、UUID 留在私有配置里，不要写进 `docs/config.example.yaml`
