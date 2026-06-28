# ERP 沙盘智能决策系统

结论：当前版本已实现一个可运行的 Go 后端核心决策系统，规则以 `1规则和财务报表.pdf` 为准，`ERP沙盘智能决策系统_技术方案.md` 作为架构和模块参考。

## 已实现能力

- ERP 沙盘规则参数表：产品研发、BOM、原料提前期、生产线、厂房、贷款、贴现、取整规则。
- 财务引擎：贷款额度、长短贷利息、贴现费、现金流预测、现金流预警、资产负债表。
- 生产计划：BOM 原料需求、采购计划、生产线产能测算。
- 订单决策：研发、市场、ISO、产能约束校验，订单综合评分。
- 市场与广告：广告选单机会、广告排名估算、广告预算分配。
- 策略模拟：内置均衡、重研发、重广告、重产能策略对比。
- AI 决策顾问：根据输入决策、运营流程节点、现金流、订单和采购计划调用 AI API 输出建议。
- 前端页面：提供决策顾问、运营流程、现金流、订单评分、采购计划视图。
- HTTP API：提供 `/api/v1` 风格接口，后续可迁移到 Go-Zero handler。

## 项目文档

- [系统开发报告与修正版技术方案](docs/system_development_report.md)
- [ERP 沙盘规则摘要](docs/rules.md)
- [年度运营流程表](docs/operation_flow.md)

## 运行

当前机器 Go 版本为 1.24.3，因此 `go.mod` 使用 `go 1.24` 以保证本地可测试。代码没有使用低版本专属写法，升级到 Go 1.26+ 不需要改业务逻辑。

```bash
env GOCACHE=/private/tmp/smarterp-go-cache go test ./...
go run ./cmd/server
```

服务默认监听：

```text
http://localhost:8080
```

前端页面：

```text
http://localhost:8080/
```

当前前端位于 `web/`，使用原生 HTML/CSS/JavaScript，不依赖 npm 构建。这样可以直接由 Go 服务托管，后续再迁移成 Vue3 组件。

健康检查：

```bash
curl http://localhost:8080/health
```

### 年度运营数据记录

系统支持记录每年的经营数据，AI 决策顾问会自动读取同企业的历史年度记录，并在后续建议中参考现金、权益、广告转化、利润、债务和违约等趋势。

```bash
curl -X POST http://localhost:8080/api/v1/operation-records \
  -H 'Content-Type: application/json' \
  -d '{
    "company": "默认企业",
    "year": 1,
    "openingCash": 60,
    "closingCash": 42,
    "openingEquity": 60,
    "closingEquity": 58,
    "salesRevenue": 32,
    "netProfit": -2,
    "adExpense": 6,
    "comprehensiveExpense": 12,
    "financialExpense": 4,
    "longLoanBalance": 40,
    "shortLoanBalance": 0,
    "keyEvents": "本地P2主投，自动线投产",
    "review": "销售规模偏小，下一年需要提高订单毛利并控制广告投入"
  }'
```

查询年度记录：

```bash
curl 'http://localhost:8080/api/v1/operation-records?company=默认企业'
```

删除某年记录：

```bash
curl -X DELETE 'http://localhost:8080/api/v1/operation-records?company=默认企业&year=1'
```

### AI API 配置

`/api/v1/advisor/decision` 支持 OpenAI 兼容的 Chat Completions API。没有配置 Key 时不会报错，会自动返回规则引擎兜底建议。

推荐使用本地配置文件，避免启动进程读不到当前 shell 的环境变量：

```bash
cp config/ai.example.json config/ai.local.json
```

然后编辑 `config/ai.local.json`：

```json
{
  "apiKey": "你的新 API Key",
  "baseURL": "https://openrouter.ai/api/v1",
  "model": "openai/gpt-4o-mini"
}
```

`config/ai.local.json` 已加入 `.gitignore`，不要提交真实 Key。程序启动后会自动读取这个文件。
开发环境下，AI 状态检查和生成建议会按请求重新读取配置文件；修改 `config/ai.local.json` 后，刷新页面或重新点击“生成建议”即可生效。

如果使用 OpenAI 官方 Key，可以改成：

```json
{
  "apiKey": "你的 OpenAI 官方 API Key",
  "baseURL": "https://api.openai.com/v1",
  "model": "gpt-4o-mini"
}
```

环境变量仍然可用，并且优先级高于配置文件，适合临时覆盖：

```bash
export AI_API_KEY="你的 API Key"
export AI_API_BASE_URL="https://api.openai.com/v1"
export AI_MODEL="gpt-4o-mini"
```

如果你的服务商不是标准 `/chat/completions` 路径，可以直接指定完整地址：

```bash
export AI_CHAT_COMPLETIONS_URL="https://example.com/v1/chat/completions"
```

检查当前运行中的服务实际读到了什么配置：

```bash
curl http://localhost:8080/api/v1/ai/status
curl -X POST http://localhost:8080/api/v1/ai/check
```

## API 示例

### 现金流预测

```bash
curl -X POST http://localhost:8080/api/v1/finance/cashflow \
  -H 'Content-Type: application/json' \
  -d '{
    "quarters": 4,
    "state": {
      "year": 1,
      "quarter": 1,
      "cash": 60,
      "lastYearEquity": 60,
      "equity": 60,
      "plannedExpenses": [
        {"quarterIndex": 1, "category": "广告费", "amount": 8},
        {"quarterIndex": 2, "category": "研发费", "amount": 2}
      ],
      "plannedIncomes": [
        {"quarterIndex": 3, "category": "回款", "amount": 30}
      ]
    }
  }'
```

### BOM 与采购计划

```bash
curl -X POST http://localhost:8080/api/v1/production/purchase-plan \
  -H 'Content-Type: application/json' \
  -d '{
    "state": {
      "year": 1,
      "quarter": 1,
      "materialInventory": {"R2": 1, "R4": 1}
    },
    "orders": [
      {"id": "O1", "market": "本地", "product": "P4", "quantity": 2, "totalPrice": 30, "deliveryQuarter": 3}
    ]
  }'
```

### 订单评分

```bash
curl -X POST http://localhost:8080/api/v1/order/score \
  -H 'Content-Type: application/json' \
  -d '{
    "state": {
      "year": 1,
      "quarter": 1,
      "cash": 60,
      "rnd": {"P2": {"product": "P2", "finished": true}},
      "markets": {"本地": {"market": "本地", "opened": true}},
      "productionLines": [
        {"id": "L1", "type": "自动线", "product": "P2", "built": true}
      ]
    },
    "orders": [
      {"id": "A", "market": "本地", "product": "P2", "quantity": 1, "totalPrice": 10, "deliveryQuarter": 1, "paymentPeriod": 0},
      {"id": "B", "market": "本地", "product": "P2", "quantity": 2, "totalPrice": 15, "deliveryQuarter": 1, "paymentPeriod": 2}
    ]
  }'
```

### 策略对比

```bash
curl -X POST http://localhost:8080/api/v1/strategy/compare \
  -H 'Content-Type: application/json' \
  -d '{
    "simulations": 100,
    "state": {"year": 1, "quarter": 1, "cash": 60, "equity": 60, "lastYearEquity": 60}
  }'
```

### AI 决策顾问

`stepCode` 来自运营流程表，完整列表可通过 `GET /api/v1/operation-flow` 获取。

```bash
curl -X POST http://localhost:8080/api/v1/advisor/decision \
  -H 'Content-Type: application/json' \
  -d '{
    "stepCode": "ad_investment",
    "question": "我这样投广告和接单是否稳妥？",
    "includePrompt": false,
    "state": {
      "year": 2,
      "quarter": 1,
      "cash": 45,
      "equity": 58,
      "lastYearEquity": 58,
      "rnd": {"P2": {"product": "P2", "finished": true}},
      "markets": {"本地": {"market": "本地", "opened": true}},
      "productionLines": [
        {"id": "L1", "type": "自动线", "product": "P2", "built": true}
      ]
    },
    "decisions": [
      {"type": "ad", "market": "本地", "product": "P2", "amount": 6},
      {"type": "short_loan", "amount": 20}
    ],
    "orders": [
      {"id": "A", "market": "本地", "product": "P2", "quantity": 2, "totalPrice": 18, "deliveryQuarter": 2, "paymentPeriod": 1}
    ],
    "competitorAds": [5, 4, 3]
  }'
```

返回中的关键字段：

- `mode`: `ai` 表示已使用 AI，`rule_fallback` 表示 AI 未配置或调用失败后返回规则建议，`rule` 表示仅规则模式。
- `advice`: 最终建议文本。
- `diagnostics.currentStep`: 当前运营流程节点。
- `diagnostics.cashFlow`: 未来 8 季度现金流预测。
- `diagnostics.orderScores`: 订单评分和不可行原因。
- `diagnostics.purchasePlan`: 按订单倒推的原料采购建议。
- `aiError`: AI 未配置或调用失败时的原因。

## 规则来源与修正

技术方案中有部分示例参数和 PDF 规则不一致，本项目以 PDF 为准：

- 原料价格：PDF 中 R1/R2/R3/R4 均为 `1W/个`，不是技术方案示例中的 `1/2/3/4W`。
- 柔性线安装周期：PDF 为 `4Q`，不是技术方案示例中的 `2Q`。
- 生产线维护费：PDF 为年维护费，手工线 `1W/年`，自动线/柔性线 `2W/年`，租赁线 `6W/年`。
- 贷款额度：长贷和短贷合计不超过上年权益的 `3 倍`，不是长贷、短贷分别计算额度。
- 短贷规则：短贷年限为 1 年，到期一次还本付息，利息四舍五入。
- 贴现规则：1、2 账期贴现率 `10%`，3、4 账期 `12.5%`，费用向上取整。

## 当前边界

- 当前版本先实现核心决策引擎和 HTTP API，没有引入数据库、Redis、Go-Zero 代码生成和 Vue 前端。
- 市场预测目前是规则兜底估算，不是 XGBoost 模型；等有历史广告和订单数据后，可以接 Python AI 服务替换预测函数。
- AI 决策顾问当前通过环境变量调用外部 AI API，尚未做对话历史持久化和建议版本管理。
- 策略模拟是轻量确定性模拟，用于比较策略倾向，不等价于完整比赛回放。

## 后续建议

1. 接入 MySQL，持久化企业状态、订单、贷款、市场与对手数据。
2. 用 Go-Zero 生成 API 层，把当前 `internal/erp` 作为业务核心复用。
3. 补 Vue3 前端，优先做数据录入、现金流、订单评分、采购计划、策略对比 5 个页面。
4. 累积广告投放和选单结果后，接 Python FastAPI + XGBoost 做市场预测。
