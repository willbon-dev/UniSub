# v0.4.1

`v0.4.1` 是一次小版本修复发布，重点完善了 Clash `rule-providers` 对纯文本规则集的兼容性，并补充了一份修正后的示例模板，方便排查和迁移。

## Highlights

- 修复 `ruleset` 远程规则生成的 `rule-providers` 缺少 `format: text` 的问题
- 将文本规则 provider 的缓存路径后缀调整为更贴切的 `.txt`
- 补充 `internal/service` 回归测试，覆盖文本规则 provider 输出
- 新增 `docs/Self.fixed.ini`，提供修正后的 Clash 模板参考

## Fixed

- 修复部分 Clash 客户端读取 `.list` 规则源时出现 `file must have a 'payload' field` 的兼容性问题
- 保持 `ACL4SSR` 一类纯文本规则集在 `RULE-SET` 模式下可正常工作

## Documentation

本版本同步更新了：

- `docs/release.md`
- `docs/Self.fixed.ini`

如果你此前使用的是基于 `ACL4SSR` / `subconverter` 风格模板的 `ruleset=...` 配置，升级到 `v0.4.1` 后，UniSub 生成的 Clash `rule-providers` 会显式标记为文本格式，兼容性会更好。
