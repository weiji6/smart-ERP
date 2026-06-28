# 年度运营流程表

结论：每年使用同一套运营顺序，系统已将流程节点固化为 `OperationStep`，前端可通过 `GET /api/v1/operation-flow` 获取。

## 流程节点

| stepCode | 节点 | 阶段 | 是否人工输入 | 说明 |
| --- | --- | --- | --- | --- |
| `year_opening_cash` | 年初现金余额 | 年初 | 是 | 填写新年度开始现金余额 |
| `annual_planning` | 新年度规划会议 | 年初 | 否 | 确认年度广告、融资、研发、市场、产能和订单目标 |
| `ad_investment` | 广告投放 | 年初 | 是 | 输入广告费，影响订货会选单顺序和机会 |
| `order_selection` | 参加订货会选订单/登记订单 | 年初 | 是 | 按广告排名选单，登记订单、交期和账期 |
| `pay_tax` | 支付应付税 | 年初 | 否 | 系统自动扣除上年度应交所得税 |
| `pay_long_interest` | 支付长贷利息 | 年初 | 否 | 长贷按贷款总额 10% 四舍五入付息 |
| `update_long_loan` | 更新长期贷款/长期贷款还款 | 年初 | 否 | 到期长期贷款还本 |
| `apply_long_loan` | 申请长期贷款 | 年初 | 是 | 长短贷合计不超过上年权益 3 倍 |
| `quarter_opening_check` | 季初盘点 | 季初 | 是 | 产品下线、生产线完工自动更新，并填写余额 |
| `update_short_loan` | 更新短期贷款/短期贷款还本付息 | 季初 | 否 | 短贷到期一次还本付息 |
| `apply_short_loan` | 申请短期贷款 | 季初 | 是 | 先归还到期短贷，再在额度内申请新短贷 |
| `material_inbound` | 原材料入库/更新原料订单 | 季度经营 | 是 | 到期原料必须全额现金购买入库 |
| `place_material_order` | 下原料订单 | 季度经营 | 是 | R1/R2 提前 1 季，R3/R4 提前 2 季 |
| `factory_buy_or_rent` | 购买/租用厂房 | 季度经营 | 是 | 最多拥有 2 个厂房 |
| `production_update` | 更新生产/完工入库 | 季度经营 | 否 | 系统更新在制品、完工入库和产能状态 |
| `production_line_action` | 新建/在建/转产/变卖生产线 | 季度经营 | 是 | 新建、继续投资、转产或变卖生产线 |
| `emergency_purchase` | 紧急采购 | 随时 | 是 | 原料紧急采购按标准成本 2 倍 |
| `start_production` | 开始下一批生产 | 季度经营 | 是 | 选择生产线和产品开工 |
| `receivable_collection` | 更新应收款/应收款收现 | 季度经营 | 是 | 录入到期金额，系统更新应收账款 |
| `order_delivery` | 按订单交货 | 季度经营 | 是 | 订单必须在规定季或提前交货 |
| `product_rnd` | 产品研发投资 | 季度经营 | 是 | 可暂停，不允许超前或集中投入 |
| `factory_disposal` | 厂房出售/退租/租转买 | 季度经营 | 是 | 厂房出售形成 4 账期应收款 |
| `market_iso_investment` | 新市场开拓/ISO资格投资 | 第四季度 | 是 | 仅第四季度允许操作 |
| `pay_admin_fee` | 支付管理费/续租/检测产品开发情况 | 季末 | 否 | 系统自动扣除管理费、续租费并检测研发状态 |
| `sell_inventory` | 出售库存 | 随时 | 是 | 原料八折向下取整，成品按直接成本 |
| `factory_discount` | 厂房贴现 | 随时 | 是 | 厂房出售所得 4 账期应收款可紧急贴现 |
| `receivable_discount` | 应收款贴现 | 随时 | 是 | 1/2 账期 10%，3/4 账期 12.5%，费用向上取整 |
| `quarter_closing_cash` | 季末余额 | 季末 | 是 | 填写季末现金余额 |
| `pay_default_penalty` | 缴纳违约订单罚款 | 年末 | 否 | 违约订单按销售总额 20% 四舍五入扣罚 |
| `pay_maintenance` | 支付设备维护费 | 年末 | 否 | 建成生产线和转产中生产线交维护费 |
| `depreciation` | 计提折旧 | 年末 | 否 | 当年建成生产线当年不提折旧 |
| `market_iso_certificate` | 新市场/ISO资格换证 | 年末 | 否 | 系统自动更新已完成市场和 ISO 资格 |
| `closing` | 结账 | 年末 | 否 | 生成综合费用表、利润表和资产负债表 |

## AI 顾问请求字段

`POST /api/v1/advisor/decision`

- `state`: 当前企业状态。
- `stepCode`: 当前流程节点，例如 `ad_investment`。
- `decisions`: 用户输入的本轮决策。
- `orders`: 可选订单或已选订单，用于评分、BOM 和采购计划。
- `competitorAds`: 竞争对手广告投入，用于广告排名估算。
- `question`: 用户希望 AI 重点回答的问题。
- `useAI`: 是否调用 AI API，默认 `true`。
- `includePrompt`: 是否返回发送给 AI 的提示词，默认 `false`。

