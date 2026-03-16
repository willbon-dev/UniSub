# UniSub

UniSub 是一个用 Go 编写的统一订阅聚合服务。程序启动时读取 YAML 配置，通过 `GET /subscribe?secret=...` 返回聚合后的代理订阅内容，支持手动节点、远程订阅、节点名称正则过滤、平台定制输出，以及 systemd/nginx/GitHub Release 部署。

当前支持的远程订阅类型：

- `base64_lines`: HTTP 响应体整体为 Base64，解码后得到多行代理链接。

## 功能

- 启动时加载 YAML 配置，按 `secret` 暴露统一订阅
- 支持手动节点和远程订阅源
- 远程订阅支持独立刷新间隔、正则 include/exclude 过滤
- 支持中文节点名称过滤
- 支持 `vmess/vless/ss/trojan/...` 原样透传
- 支持 `V2rayN` 和 `Happ` 平台输出
- 支持 `refresh=1` 强制刷新该统一订阅下所有远程源

## 快速开始

```bash
go run ./cmd/unisub -config ./docs/config.example.yaml
```

本地直接使用二进制启动：

```bash
go build -o ./bin/unisub ./cmd/unisub
./bin/unisub -config ./docs/config.example.yaml
```

如果你已经有编译好的二进制，也可以直接：

```bash
./unisub -config ./docs/config.example.yaml
```

请求示例：

```bash
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000'
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000&platform=Happ'
curl 'http://127.0.0.1:8080/subscribe?secret=123e4567-e89b-42d3-a456-426614174000&refresh=1'
```

## 配置

完整模板见 [docs/config.example.yaml](docs/config.example.yaml)。

关键字段：

- `subscriptions[].secret`: 访问统一订阅时使用的 UUID
- `subscriptions[].default_platform`: 默认返回平台，可为 `V2rayN` 或 `Happ`
- `subscriptions[].platform_options.happ.routing`: Happ 平台首行附加 routing
- `sources[].type`: `manual` 或 `remote`
- `sources[].prefix`: 为该 source 下所有节点名称追加统一前缀；为空时不处理
- `sources[].remote_type`: 远程订阅格式，当前仅支持 `base64_lines`
- `sources[].include_patterns` / `exclude_patterns`: 基于节点名称的 Go regexp 过滤

## 安装与运行

脚本安装时，默认配置文件保存到 `/etc/unisub/config.yaml`，程序安装到 `/opt/unisub/bin/unisub`。

首次安装：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash
```

更新：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash
```

修改配置后重启服务：

```bash
sudo systemctl restart unisub
sudo systemctl status unisub
```

## 部署文档

- [Ubuntu 24.04 服务安装](docs/deploy-ubuntu-24.04.md)
- [Nginx TLS 与反向代理](docs/nginx.md)
- [发布流程与 GitHub Release](docs/release.md)

## 发布产物

打 tag 后 GitHub Actions 会自动构建并发布多平台二进制压缩包和校验和文件。

## Attribution

本项目由 Codex 编写。
