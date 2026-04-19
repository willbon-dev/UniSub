# v0.4.0

`v0.4.0` 重点为 UniSub 增加了完整的 Clash 聚合订阅能力，并补齐了新的配置模型与文档说明。

## Highlights

- 新增 `Clash` 平台输出
- 新增 `platform_options.clash.template`
- 新增 `sources[].platforms`
- 新增 `sources[].style`
- 支持 `manual` 手动 Clash 节点
- 支持 `remote` Clash YAML 订阅，并读取其中的 `proxies:`
- 支持 `request_headers`，适配需要特定 `User-Agent` 的上游
- 支持基于 `subconverter` 风格模板生成合法 Clash YAML

## Config Changes

- `default_platform` 现在支持 `Clash`
- `sources` 必须显式声明 `platforms`
- `sources` 必须显式声明 `style`
- 旧 `remote_type: base64_lines` 仍兼容，但推荐迁移为 `style: link_lines_base64`

## Clash Support

当前 Clash 能力包括：

- `manual + clash_proxy`
- `remote + clash_proxies_yaml`
- `request_headers` 自定义请求头
- `custom_proxy_group`
- `ruleset`
- `clash_rule_base`

当前实现按“最小 subconverter 兼容”处理模板，不尝试完整复刻外部 subconverter / ACL4SSR 全部高级语义。

## Documentation

本版本同步更新了：

- `README.md`
- `docs/config.example.yaml`
- `docs/config.md`

其中新增了：

- 字段级配置说明
- `style` 使用说明
- `manual` / `remote` 示例
- `request_headers` 说明
- 常见报错排查
