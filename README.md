# UniSub

UniSub 是一个用 Go 编写的统一订阅聚合服务。程序启动后读取 YAML 配置，通过 `GET /subscribe?secret=...` 输出聚合后的代理订阅内容。

当前支持：

- 手动节点与远程订阅源
- `V2rayN`、`Happ`、`Clash` 三个平台输出
- source 级平台筛选
- `manual` / `remote` 共用 `style` 配置
- 手动 Clash 节点与远程 Clash YAML 订阅
- 基于 `subconverter` 风格模板的 Clash YAML 生成

## Quick Start

直接运行：

```bash
go run ./cmd/unisub -config ./docs/config.example.yaml
```

或先构建再运行：

```bash
go build -o ./bin/unisub ./cmd/unisub
./bin/unisub -config ./docs/config.example.yaml
```

## Scripts

如果你是部署到 Linux 服务器，也可以直接使用仓库自带脚本：

首次安装：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash
```

更新：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash
```

重装：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/reinstall.sh | sudo bash
```

卸载：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/uninstall.sh | sudo bash
```

这些脚本会自动处理：

- 下载或更新二进制
- 写入 systemd 服务
- 安装示例配置
- 启动或重启 UniSub

更详细的脚本使用方式见 [docs/deploy-ubuntu-24.04.md](docs/deploy-ubuntu-24.04.md) 和 [docs/release.md](docs/release.md)。

## API

```bash
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000'
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000&platform=Happ'
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000&platform=Clash'
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000&platform=Clash&refresh=1'
```

## Config Highlights

- `subscriptions[].default_platform` 支持 `V2rayN`、`Happ`、`Clash`
- `subscriptions[].platform_options.clash.template` 支持本地路径或远程 URL
- `subscriptions[].sources[].platforms` 为必填，显式声明 source 适用平台
- `subscriptions[].sources[].style` 为必填，决定 source 的解析方式
- `subscriptions[].sources[].request_headers` 可用于指定上游请求头，例如 `User-Agent: clash-verge`
- 旧 `remote_type: base64_lines` 仍兼容，但推荐迁移到 `style: link_lines_base64`

当前支持的 style：

- `link_line`: `manual` 单条代理链接
- `link_lines_base64`: `remote` Base64 多行链接订阅
- `clash_proxy`: `manual` 单条 Clash 代理对象
- `clash_proxies_yaml`: `remote` Clash YAML，其中读取 `proxies:`

说明：

- `clash_proxy` 的 `entries` 既支持字符串形式，也支持原生 YAML 对象形式
- `clash_proxies_yaml` 适合上游直接返回 Clash YAML 的场景
- 如果上游默认返回 Base64 链接订阅，但带特定 `User-Agent` 会切换到 Clash YAML，就需要配 `request_headers`

完整示例见 [docs/config.example.yaml](docs/config.example.yaml)。

## Clash Template Notes

Clash 输出使用 `platform_options.clash.template` 指向模板。模板支持：

- 本地文件路径，例如 `./Self.ini`
- 远程 URL，例如 `https://example.com/Self.ini`

当前实现按“最小 subconverter 兼容”处理模板：

- 支持读取 `custom_proxy_group=...`
- 支持读取 `ruleset=...`
- 支持读取 `clash_rule_base=...`
- UniSub 负责把聚合后的 `proxies` 注入最终 Clash YAML

当前版本不尝试完整复刻外部 `subconverter` / `ACL4SSR` 的全部高级能力。

## Docs

- 配置详细说明：[docs/config.md](docs/config.md)
- 示例配置：[docs/config.example.yaml](docs/config.example.yaml)

`docs/config.md` 包含：

- 每个配置字段的含义和限制
- `style` 的适用场景
- `manual` / `remote` 的完整示例
- `request_headers` 用法
- Clash 模板边界
- 常见报错排查
