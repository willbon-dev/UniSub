# Ubuntu 24.04 部署

## 目录建议

- 二进制：`/opt/unisub/bin/unisub`
- 配置：`/etc/unisub/config.yaml`
- 日志：交给 `systemd journal`

## 手动安装

1. 创建服务用户：

```bash
sudo useradd --system --home /var/lib/unisub --shell /usr/sbin/nologin unisub
```

2. 创建目录：

```bash
sudo mkdir -p /opt/unisub/bin /etc/unisub
sudo chown -R root:root /opt/unisub /etc/unisub
```

3. 放置二进制和配置：

```bash
sudo install -m 0755 unisub /opt/unisub/bin/unisub
sudo install -m 0640 docs/config.example.yaml /etc/unisub/config.yaml
```

4. 创建 `systemd` 服务 `/etc/systemd/system/unisub.service`：

```ini
[Unit]
Description=UniSub unified subscription service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=unisub
Group=unisub
ExecStart=/opt/unisub/bin/unisub -config /etc/unisub/config.yaml
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/unisub
WorkingDirectory=/var/lib/unisub

[Install]
WantedBy=multi-user.target
```

5. 启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now unisub
sudo systemctl status unisub
```

## 本地二进制直接启动

不走 systemd 时，可以直接运行二进制：

```bash
go build -o ./bin/unisub ./cmd/unisub
./bin/unisub -config ./docs/config.example.yaml
```

如果已经下载 release 包并解压出 `unisub`，则直接：

```bash
./unisub -config /etc/unisub/config.yaml
```

## 修改配置后如何重启

脚本安装和手动安装都默认使用同一个配置路径：

```bash
/etc/unisub/config.yaml
```

修改完成后执行：

```bash
sudo systemctl restart unisub
sudo systemctl status unisub
```

## 升级

```bash
sudo systemctl stop unisub
sudo install -m 0755 unisub /opt/unisub/bin/unisub
sudo systemctl start unisub
```

## 查看日志

```bash
journalctl -u unisub -f
```

## 脚本化安装与维护

安装最新版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash
```

安装指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/install.sh | sudo bash -s -- --version v0.2.0
```

更新到最新版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash
```

更新到指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/update.sh | sudo bash -s -- --version v0.2.0
```

重装并保留配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/reinstall.sh | sudo bash
```

重装并清空配置与状态：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/reinstall.sh | sudo bash -s -- --purge-config
```

卸载但保留配置：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/uninstall.sh | sudo bash
```

卸载并彻底清理配置与状态：

```bash
curl -fsSL https://raw.githubusercontent.com/willbon-dev/UniSub/main/scripts/uninstall.sh | sudo bash -s -- --purge-config
```
