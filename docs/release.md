# 发布流程

## 当前版本

- 当前准备发布版本：`v0.3.0`
- 当前版本说明文档：[docs/release-v0.3.0.md](release-v0.3.0.md)

## 自动发布

推送语义化版本 tag，例如：

```bash
git tag v0.3.0
git push origin v0.3.0
```

GitHub Actions 会自动：

- 执行 `go test ./...`
- 构建 Linux/macOS/Windows 的 `amd64` 和 `arm64`
- 打包产物
- 生成 `checksums.txt`
- 创建对应 GitHub Release

## 手动发布建议

1. 确认工作区干净，测试通过。
2. 更新 `README`、示例配置和文档。
3. 创建 tag 并推送。
4. 在 GitHub Release 页面核对产物和说明。

## 一键安装脚本

默认安装最新版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash
```

安装指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash -s -- --version v0.3.0
```

脚本默认会：

- 从 GitHub Release 下载 Ubuntu 24.04 对应 Linux amd64/arm64 包
- 安装到 `/opt/unisub`
- 写入 `/etc/unisub/config.yaml`
- 创建 `unisub` 用户
- 注册并启动 `systemd` 服务

## 维护脚本

更新到最新版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash
```

更新到指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash -s -- --version v0.3.0
```

重装并保留配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/reinstall.sh | sudo bash
```

重装并清除配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/reinstall.sh | sudo bash -s -- --purge-config
```

卸载但保留配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/uninstall.sh | sudo bash
```

卸载并彻底清理配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/uninstall.sh | sudo bash -s -- --purge-config
```

说明：

- `install.sh` 适合首次安装
- `update.sh` 只更新二进制和 service 定义，保留配置
- `reinstall.sh` 用于异常恢复，可选保留或清空配置
- `uninstall.sh` 会移除服务和二进制，可选是否清理配置与运行状态
- 安装、更新、重装脚本结束时都会打印配置文件路径和重启命令
