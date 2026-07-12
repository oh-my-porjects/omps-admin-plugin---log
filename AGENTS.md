# 审计日志公共模块

- V4 后台只使用 `admin-intents.yaml` 和 `api-docs/`；禁止恢复 `admin-web.yaml`、`AdminWebHint` 和 V1–V3 菜单/spec 协议。
- 审计列表字段必须来自真实接口，筛选参数也必须与 api-docs 已验证的 query 合同一致。
- 日志查询是只读能力；不得让后台页面臆造删除、重写或修复审计记录的动作。
