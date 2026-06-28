# ERP沙盘智能决策系统技术方案与系统开发报告

## 1. 报告说明

### 1.1 报告目的

本文档基于当前已实现的 `SmartERP` 项目代码、两份 ERP 沙盘规则文档以及系统前端页面，重新修正原《ERP沙盘智能决策系统_技术方案.md》中的规划性内容，形成一份与当前可运行系统一致的专业开发报告。

原技术方案中包含 Go-Zero、Vue3、MySQL、Redis、Python FastAPI、XGBoost 等完整工程化设想。当前项目已经完成的是一个可运行、可演示、可继续扩展的单体版本：

- 后端采用 Go 标准库 `net/http` 实现 HTTP API。
- 前端采用原生 HTML/CSS/JavaScript，由 Go 服务直接托管。
- 规则、财务、生产、订单、市场、AI 顾问逻辑主要位于 `internal/erp`。
- 年度运营记录与 Q/A 历史建议记录支持可选 MySQL 持久化；未配置 MySQL 时使用内存存储。
- AI 调用采用 OpenAI 兼容 Chat Completions API，并支持本地配置文件。

因此，本报告不再把尚未落地的技术组件描述为已实现，而是明确区分：

- **当前已实现版本**：已经在本地 `http://localhost:8081/` 可运行的系统。
- **后续演进版本**：可迁移到 Go-Zero、Vue3、数据库和模型服务的方向。

### 1.2 项目当前状态

当前系统已经具备 ERP 沙盘智能决策支持的核心闭环：

1. 用户在页面录入企业状态、当前流程、经营决策和订单。
2. 后端根据 PDF 规则计算贷款额度、现金流、订单可行性、BOM、采购计划和规则风险。
3. 系统将规则文档、运营流程、年度经营历史、Q/A 历史和当前诊断结果组织为 AI Prompt。
4. AI 可用时返回专业决策建议；AI 不可用时返回规则引擎兜底建议。
5. 页面展示 AI 建议、规则风险、现金流预测、订单评分、采购计划和历史 Q/A。
6. 年度运营数据可保存，后续 AI 建议会参考历史经营表现。

## 2. 项目背景与需求分析

### 2.1 业务背景

ERP 沙盘模拟经营是一类以企业经营决策为核心的教学或竞赛活动。参赛团队通常需要在 5 至 7 个经营年度内持续做出广告投放、订单选择、产品研发、市场开拓、ISO 认证、原料采购、生产线建设、厂房规划、贷款融资和现金流管理等决策。

系统服务的核心目标不是替代参赛者，而是辅助参赛者快速完成以下工作：

- 规则查询与自动计算。
- 决策风险识别。
- 现金流压力预判。
- 订单利润和交付可行性判断。
- 广告与市场策略建议。
- 年度经营复盘。
- AI 问答式经营建议。

### 2.2 主要业务痛点

| 痛点 | 表现 | 系统解决方式 |
| --- | --- | --- |
| 规则复杂 | 贷款、贴现、原料提前期、广告排序、违约、税务等规则容易记错 | 将规则固化为 Go 常量、计算函数和运营流程节点 |
| 现金流敏感 | 一次广告、贷款、采购或建线失误可能导致现金断裂 | 提供未来 8 季度现金流预测和预警 |
| 决策耦合强 | 广告、订单、产能、采购、融资互相影响 | 在 AI 顾问中统一输入企业状态、决策、订单和历史记录 |
| 手工计算慢 | 比赛环境下计算时间紧张 | 页面提供表单、指标卡、表格和自动诊断 |
| AI 容易答偏 | 用户问市场时可能回答广告，或回答太空泛 | Prompt 明确问题优先级，并注入规则文档和当前诊断数据 |
| 历史经验难沉淀 | 每年经营结果没有结构化复盘 | 提供年度运营记录和历史 Q/A 建议记录 |

### 2.3 当前版本需求边界

当前版本定位为单机可运行的决策支持原型，重点覆盖核心规则和 AI 顾问闭环，不追求一次性完成完整竞赛平台。

已覆盖：

- 规则参数管理。
- 运营流程展示。
- 企业状态录入。
- 决策录入。
- 订单录入。
- 贷款额度计算。
- 现金流预测。
- BOM 与采购计划。
- 订单评分。
- 广告机会预测和预算分配。
- 策略对比模拟。
- 年度运营记录。
- Q/A 历史建议记录。
- AI 建议生成与规则兜底。
- AI 配置诊断。

暂未完整覆盖：

- 多用户登录与权限。
- 多用户级数据库模型和数据迁移管理。
- 完整自动结账。
- 完整厂房/生产线状态流转。
- 真实机器学习模型训练。
- 多企业竞赛实时对抗。
- WebSocket 实时协作。

## 3. 规则依据与系统实现口径

### 3.1 规则来源

系统规则主要来自以下两份文档：

- `1规则和财务报表.pdf`
- `2运营流程表.pdf`

项目中已经整理为 Markdown：

- `docs/rules.md`
- `docs/operation_flow.md`

AI Prompt 会读取这两份 Markdown 文档，将其作为回答约束的一部分。

### 3.2 已固化的核心规则

#### 3.2.1 融资规则

- 长贷和短贷合计额度不超过上年所有者权益的 3 倍。
- 长贷每年年初申请，每年年初付息，到期还本，年利率 10%，最多 5 年。
- 短贷每季度初申请，期限 1 年，到期一次还本付息，年利率 5%。
- 长短贷利息四舍五入。
- 长贷和短贷不允许提前还款。
- 本轮新增贷款决策会进入现金流预测和贷款额度占用。

实现位置：

- `internal/erp/finance.go`
- `internal/erp/advisor.go`

#### 3.2.2 贴现与拍卖规则

- 1、2 账期贴现率 10%。
- 3、4 账期贴现率 12.5%。
- 贴现费用向上取整。
- 原材料拍卖按 80% 折价并向下取整。
- 成品拍卖按产品直接成本计算。

实现位置：

- `internal/erp/finance.go`

#### 3.2.3 厂房与生产线规则

厂房：

| 厂房 | 买价 | 租金 | 售价 | 容量 |
| --- | ---: | ---: | ---: | ---: |
| 大厂房 | 40W | 5W/年 | 40W | 6 条 |
| 小厂房 | 30W | 3W/年 | 30W | 4 条 |

生产线：

| 生产线 | 购置费 | 安装周期 | 生产周期 | 转产费 | 转产周期 | 维修费 | 残值 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 手工线 | 5W | 0Q | 2Q | 0W | 0Q | 1W/年 | 1W |
| 自动线 | 15W | 3Q | 1Q | 2W | 1Q | 2W/年 | 3W |
| 柔性线 | 20W | 4Q | 1Q | 0W | 0Q | 2W/年 | 4W |
| 租赁线 | 0W | 0Q | 1Q | 2W | 1Q | 6W/年 | -6W |

实现位置：

- `internal/erp/rules.go`
- `internal/erp/production.go`
- `internal/erp/finance.go`

#### 3.2.4 产品研发与 BOM

| 产品 | 研发费 | 总额 | 周期 | 加工费 | 直接成本 | BOM |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| P1 | 1W/季 | 2W | 2 季 | 1W/个 | 2W/个 | R1 |
| P2 | 1W/季 | 3W | 3 季 | 1W/个 | 3W/个 | R2 + R3 |
| P3 | 1W/季 | 4W | 4 季 | 1W/个 | 4W/个 | R1 + R3 + R4 |
| P4 | 1W/季 | 6W | 6 季 | 1W/个 | 5W/个 | R2 + R3 + 2R4 |

实现位置：

- `internal/erp/rules.go`
- `internal/erp/production.go`
- `internal/erp/order.go`

#### 3.2.5 市场开拓与 ISO

- 市场开拓和 ISO 认证只允许在第四季度操作。
- ISO9000：1W/年，2 年，总 2W。
- ISO14000：2W/年，2 年，总 4W。

市场规则：

| 市场 | 每年开拓费 | 年限 | 总费用 |
| --- | ---: | ---: | ---: |
| 本地 | 1W | 1 年 | 1W |
| 区域 | 1W | 1 年 | 1W |
| 国内 | 1W | 2 年 | 2W |
| 亚洲 | 1W | 3 年 | 3W |
| 国际 | 1W | 4 年 | 4W |

实现位置：

- `internal/erp/rules.go`
- `internal/erp/advisor_market.go`
- `internal/erp/advisor.go`

#### 3.2.6 广告与选单规则

- 单个市场单个产品最低广告额为 1W。
- 投放 1W 理论获得 1 次选单机会。
- 此后每增加 2W 增加 1 次选单机会。
- 选单排序优先级：
  - 本市场本产品广告额。
  - 本市场广告总额。
  - 上年本市场销售排名。
  - 先投广告者。

实现位置：

- `internal/erp/market.go`
- `internal/erp/advisor_ad.go`

#### 3.2.7 原料与紧急采购

| 原料 | 单价 | 提前期 |
| --- | ---: | ---: |
| R1 | 1W/个 | 1 季 |
| R2 | 1W/个 | 1 季 |
| R3 | 1W/个 | 2 季 |
| R4 | 1W/个 | 2 季 |

- 紧急采购按标准成本 2 倍计算。
- 多付部分记入损失。

实现位置：

- `internal/erp/rules.go`
- `internal/erp/production.go`

#### 3.2.8 违约与所得税

- 违约金按违约订单销售总额的 20% 计算，四舍五入。
- 所得税率为 25%，四舍五入。
- PDF 原文包含 5 年补亏规则，即超过初始权益的部分才交税。

当前实现情况：

- 违约与所得税规则已经进入规则文档和 AI Prompt。
- `CalcProfitStatement` 当前实现为正税前利润按 25% 计税。
- 完整 5 年补亏自动结账逻辑尚未实现，属于后续财务闭环增强项。

## 4. 修正后的系统总体技术方案

### 4.1 当前技术架构

当前系统采用轻量级单体架构：

```text
浏览器页面
  |
  | HTTP / JSON
  v
Go net/http 服务
  |
  +-- 静态资源托管：web/index.html, web/assets
  +-- HTTP API：internal/httpapi
  +-- ERP 规则引擎：internal/erp
  +-- AI 客户端：OpenAI 兼容 Chat Completions
  +-- 存储层：可选 MySQL，未配置时回退内存存储
```

### 4.2 分层说明

| 层级 | 当前实现 | 主要职责 |
| --- | --- | --- |
| 前端展示层 | 原生 HTML/CSS/JavaScript | 页面展示、表单录入、API 调用、图表绘制、状态反馈 |
| HTTP 接口层 | Go `net/http` + `ServeMux` | 路由注册、请求解析、响应输出、静态资源托管 |
| 业务规则层 | `internal/erp` | ERP 规则计算、现金流、订单、生产、市场、策略、AI Prompt |
| 存储层 | MySQL 或内存 Map + Mutex | 年度运营记录、Q/A 历史记录 |
| AI 接入层 | OpenAI 兼容 Chat Completions | AI 建议生成、状态检查、连通性测试 |
| 文档规则层 | Markdown 规则文档 | AI Prompt 规则上下文 |

### 4.3 目录结构

```text
SmartERP
├── cmd/server
│   └── main.go                    # 服务入口
├── config
│   └── ai.example.json             # AI 配置示例
├── docs
│   ├── rules.md                    # ERP 规则摘要
│   ├── operation_flow.md           # 年度运营流程
│   └── system_development_report.md
├── internal
│   ├── erp                         # 核心业务规则与决策逻辑
│   └── httpapi                     # HTTP API 与 AI 客户端
├── web
│   ├── index.html                  # 单页入口
│   └── assets
│       ├── app.js                  # 前端交互逻辑
│       └── styles.css              # 页面样式
├── README.md
└── go.mod
```

### 4.4 当前架构与原技术方案差异

| 原方案内容 | 当前实际实现 | 修正说明 |
| --- | --- | --- |
| Go-Zero 微服务 | Go 标准库单体 HTTP 服务 | 当前版本优先实现核心业务闭环，后续可迁移到 Go-Zero |
| Vue3 + Element Plus | 原生 HTML/CSS/JavaScript | 降低构建复杂度，便于本地直接运行 |
| MySQL 持久化 | 已支持可选 MySQL | 配置 DSN 后自动建表并持久化年度记录和 Q/A 历史 |
| Redis 缓存 | 未使用 | 当前数据量小，不需要缓存层 |
| Python FastAPI + XGBoost | 未引入独立模型服务 | 当前使用规则模型和 AI 大模型，后续可扩展训练模型 |
| WebSocket 实时通信 | 未实现 | 当前交互为 HTTP 请求响应 |
| 完整竞赛平台 | 决策支持工具 | 当前系统定位为沙盘经营辅助决策系统 |

## 5. 核心模块设计与实现

### 5.1 规则参数模块

规则参数位于 `internal/erp/rules.go`，包括：

- 产品规则。
- 原料规则。
- 生产线规则。
- 厂房规则。
- 市场开拓规则。
- ISO 认证规则。
- 贷款、贴现、税率、广告最低额、安全现金垫等常量。

设计原则：

- 金额统一使用 `Money int64` 表示，单位为 W，避免浮点误差。
- 规则表直接使用 Go map 和 struct 表达，便于测试和引用。
- 规则数据和计算逻辑分离，保持可维护性。

### 5.2 财务引擎模块

实现位置：

- `internal/erp/finance.go`

主要能力：

- 贷款额度计算。
- 长贷年利息计算。
- 短贷到期本息计算。
- 贴现费用计算。
- 原料拍卖和成品拍卖金额计算。
- 未来季度现金流预测。
- 现金流预警级别判断。
- 年度折旧计算。
- 利润表和资产负债表计算。

关键设计：

- `CalcLoanCapacity` 使用 `上年权益 * 3 - 已有长短贷本金`。
- `PredictCashFlow` 汇总计划收入、计划支出、应收账款回款、周期性支出、贷款还款等因素。
- 当前 AI 顾问会先通过 `ApplyDecisionsToState` 将页面本轮决策转成计划收支，再进入现金流预测。

### 5.3 决策输入映射模块

实现位置：

- `internal/erp/advisor.go`

核心函数：

```go
ApplyDecisionsToState(state CompanyState, decisions []DecisionInput) CompanyState
```

作用：

- 将用户页面输入的决策转换为财务预测可使用的企业状态。
- 避免页面看起来录入了决策，但后端现金流预测没有引用的问题。

当前映射关系：

| 决策类型 | 现金流影响 |
| --- | --- |
| `ad` | 当季广告费支出 |
| `market_iso` | 当季市场/ISO 投资支出 |
| `production_line` | 当季生产线投资支出 |
| `material_order` | 当季原料采购支出 |
| `short_loan` | 当季贷款流入，4 个季度后还本付息 |
| `long_loan` | 当季贷款流入，按长期贷款本金进入额度与利息计算 |

### 5.4 生产与采购模块

实现位置：

- `internal/erp/production.go`

主要能力：

- 根据订单计算 BOM 原料需求。
- 根据现有原料库存和提前期生成采购计划。
- 计算未来季度产能。
- 计算生产线折旧。

当前页面在“订单与采购”视图展示：

- 订单录入。
- 订单评分。
- 原料采购计划。

### 5.5 订单评分模块

实现位置：

- `internal/erp/order.go`

订单评分考虑：

- 产品是否研发完成。
- 市场是否开拓完成。
- ISO 是否满足订单要求。
- 订单利润。
- 回款账期。
- 交货季度。
- 产能约束。
- 现金流影响。

输出：

- 是否可行。
- 综合评分。
- 预计利润。
- 利润率。
- 不可行原因。

### 5.6 市场与广告模块

实现位置：

- `internal/erp/market.go`
- `internal/erp/advisor_ad.go`

主要能力：

- 根据广告投入和竞争对手广告估计广告排名。
- 根据广告金额估算理论选单机会。
- 根据市场、产品和预算生成广告投放建议。
- 当用户问题明显与广告相关时，规则兜底可生成具体广告投放表。

### 5.7 市场开拓建议模块

实现位置：

- `internal/erp/advisor_market.go`

主要能力：

- 判断用户问题是否与市场开拓相关。
- 根据市场开拓周期、费用、产品成熟度、经营年限和现金状况给出开拓优先级。
- 当用户询问是否开放新市场时，避免错误返回广告方案。

### 5.8 年度运营记录模块

实现位置：

- `internal/erp/operation_record.go`
- `internal/httpapi/operation_store.go`

主要能力：

- 记录年度现金、权益、销售收入、净利润、广告费、综合费用、研发、市场、ISO、财务费用、贷款余额、违约罚款等数据。
- 分析年度经营趋势。
- 识别历史风险信号，例如现金安全垫不足、广告转化偏低、净利润为负、违约罚款等。
- AI 顾问会读取同企业历史记录，并在后续建议中引用。

当前存储方式：

- 默认未配置数据库时使用内存 Map。
- 配置 `config/db.local.json` 或 `MYSQL_DSN` 后使用 MySQL。
- MySQL 表：`annual_operation_records`。
- 内存模式使用 `sync.RWMutex` 保证并发读写安全。

### 5.9 Q/A 历史建议模块

实现位置：

- `internal/erp/advisor_history.go`
- `internal/httpapi/advisor_history_store.go`

主要能力：

- 每次调用 AI 顾问后保存一条 Q/A 记录。
- 前端展示历史建议记录。
- 后续 AI Prompt 会读取最近的历史建议，避免重复建议，并保持策略连续性。

当前存储方式：

- 默认未配置数据库时使用内存 Map。
- 配置 MySQL 后写入 `advisor_qa_records`。
- 支持按企业查询、限制条数、清空历史。

### 5.10 AI 顾问模块

实现位置：

- `internal/erp/advisor.go`
- `internal/erp/prompt_docs.go`
- `internal/httpapi/ai_client.go`

AI 顾问处理流程：

1. 前端提交企业状态、当前流程、决策、订单、竞争广告、年度记录、历史 Q/A 和用户问题。
2. 后端构造 `AdvisorContext`。
3. 规则引擎生成 `AdvisorDiagnostics`。
4. 如果 AI 配置可用，构造 AI Prompt：
   - 系统角色。
   - 原始规则文档。
   - 运营流程文档。
   - 当前企业状态。
   - 贷款额度。
   - 年度运营历史。
   - 历史 Q/A。
   - 当前决策。
   - 竞争对手广告。
   - 风险诊断。
   - 现金流预测。
   - 订单评分。
   - 采购需求。
   - 用户问题。
5. 调用 OpenAI 兼容 Chat Completions API。
6. AI 成功时返回 AI 回答。
7. AI 失败或未配置时返回规则兜底建议。
8. 保存本次 Q/A 到历史记录。

AI Prompt 约束：

- 必须先直接回答用户问题。
- 用户问题优先级高于当前流程节点。
- 当前流程节点只作为上下文，不能把问题带偏。
- 广告问题必须给具体投放表。
- 市场问题必须回答是否开、开哪个、何时开。
- 输出必须包含依据、方案和补充数据。

### 5.11 AI 配置与诊断模块

实现位置：

- `internal/httpapi/ai_client.go`

配置来源：

1. 默认读取 `config/ai.local.json`。
2. 环境变量可覆盖配置文件。

支持字段：

```json
{
  "apiKey": "你的 API Key",
  "baseURL": "https://openrouter.ai/api/v1",
  "model": "openai/gpt-4o-mini",
  "chatCompletionsURL": ""
}
```

诊断接口：

- `GET /api/v1/ai/status`
- `POST /api/v1/ai/check`

安全设计：

- 不返回完整 API Key。
- 只返回脱敏后的 Key 预览。
- 当 OpenRouter Key 配到 OpenAI 官方 Base URL 时给出配置警告。
- `config/ai.local.json` 被 `.gitignore` 忽略，避免泄露密钥。

## 6. HTTP API 设计

### 6.1 基础接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 健康检查 |
| GET | `/api/v1/rules` | 获取规则参数 |
| GET | `/api/v1/operation-flow` | 获取年度运营流程 |

### 6.2 财务接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/finance/cashflow` | 现金流预测 |
| POST | `/api/v1/finance/loan-capacity` | 贷款额度计算 |
| POST | `/api/v1/finance/balance-sheet` | 资产负债表计算 |

### 6.3 生产与订单接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/production/bom` | BOM 原料需求计算 |
| POST | `/api/v1/production/purchase-plan` | 原料采购计划 |
| POST | `/api/v1/production/capacity` | 产能测算 |
| POST | `/api/v1/order/score` | 订单评分 |

### 6.4 市场与策略接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/market/predict` | 市场订单机会预测 |
| POST | `/api/v1/market/ad-optimize` | 广告预算分配 |
| POST | `/api/v1/strategy/compare` | 多策略对比 |

### 6.5 年度记录接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/operation-records` | 查询年度运营记录 |
| POST | `/api/v1/operation-records` | 保存年度运营记录 |
| DELETE | `/api/v1/operation-records` | 删除年度运营记录 |

### 6.6 AI 顾问接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/advisor/decision` | 生成 AI/规则决策建议 |
| GET | `/api/v1/advisor/history` | 查询历史建议 |
| DELETE | `/api/v1/advisor/history` | 清空历史建议 |
| GET | `/api/v1/ai/status` | 查看 AI 配置状态 |
| POST | `/api/v1/ai/check` | 测试 AI 连通性 |

## 7. 前端页面设计

### 7.1 页面定位

前端不是营销型展示页，而是面向 ERP 沙盘经营过程的操作型工具界面。设计原则是：

- 第一屏展示关键经营指标。
- 当前后端和 AI 状态一打开即可看到。
- 表单按经营决策语义组织。
- AI 建议、规则风险和历史记录放在同一工作流中。
- 移动端不出现横向溢出。

### 7.2 主要视图

| 视图 | 功能 |
| --- | --- |
| 决策顾问 | 企业状态、输入决策、AI 建议、规则风险、历史 Q/A |
| 年度记录 | 录入年度经营数据、查看年度复盘和趋势 |
| 运营流程 | 展示完整年度运营流程节点 |
| 现金流 | 展示未来 8 季度现金流图表和明细 |
| 订单与采购 | 订单录入、订单评分、采购计划 |

### 7.3 已优化的 UI 细节

- 后端在线和 AI 配置状态移到顶部标题栏。
- 指标卡显示现金、权益、可贷额度和顾问模式。
- 决策输入根据类型动态切换字段。
- AI 建议显示本次调用状态：
  - AI 成功。
  - AI 未配置。
  - AI 调用失败后规则兜底。
- Q/A 历史以问答格式展示。
- 移动端侧边栏改为横向导航，避免文字挤压。
- 表格、面板、按钮和输入框统一为后台工具风格。

### 7.4 前端状态与后端数据一致性

当前页面已修正以下一致性问题：

- 企业状态中的长贷余额和短贷余额会传给后端。
- 可贷额度会实时扣除已有贷款和本轮贷款决策。
- 本轮广告、贷款、市场/ISO、生产线、原料订单会进入后端现金流预测。
- AI 建议使用后端返回的 `mode`、`aiUsed`、`aiError` 显示真实调用状态。

## 8. 数据模型设计

### 8.1 核心 Go 数据模型

主要类型位于 `internal/erp/types.go`：

- `Money`
- `Product`
- `Material`
- `Market`
- `LineType`
- `FactoryType`
- `RnDProgress`
- `MarketProgress`
- `ISOProgress`
- `ProductionLine`
- `Loan`
- `Receivable`
- `PlannedExpense`
- `PlannedIncome`
- `CompanyState`
- `Order`

### 8.2 决策输入模型

```go
type DecisionInput struct {
    Type          string
    Year          int
    Quarter       int
    Amount        Money
    Market        Market
    Product       Product
    Material      Material
    Quantity      int
    LineType      LineType
    FactoryType   FactoryType
    OrderID       string
    PaymentPeriod int
    Extra         map[string]any
    Note          string
}
```

### 8.3 AI 顾问上下文模型

```go
type AdvisorContext struct {
    State            CompanyState
    StepCode         string
    Decisions        []DecisionInput
    Orders           []Order
    CompetitorAds    []Money
    OperationRecords []AnnualOperationRecord
    AdviceHistory    []AdvisorQARecord
    Question         string
}
```

### 8.4 诊断输出模型

`AdvisorDiagnostics` 包含：

- 当前流程节点。
- 后续流程节点。
- 贷款额度。
- 现金流预测。
- 年度历史分析。
- 广告专项测算。
- 市场开拓专项测算。
- 订单评分。
- 原料需求。
- 采购计划。
- 风险发现。
- 规则提醒。
- 规则兜底建议。

## 9. AI Prompt 设计

### 9.1 Prompt 设计目标

系统 Prompt 的目标不是让 AI 泛泛聊天，而是让 AI 成为受规则约束的 ERP 沙盘经营顾问。

核心要求：

- 回答必须直接对应用户问题。
- 不能把当前流程节点误当成用户问题。
- 必须优先遵守 PDF 规则。
- 必须引用当前企业状态、现金流、贷款额度、订单、产能、采购和历史经营数据。
- 广告和市场问题必须给具体方案。
- 输出结构稳定，便于用户阅读。

### 9.2 Prompt 输入内容

AI Prompt 包含：

1. 角色定义。
2. 回答原则。
3. 原始规则文档。
4. 运营流程文档。
5. 当前企业状态。
6. 当前流程节点。
7. 贷款额度。
8. 年度运营历史。
9. 历史 Q/A 建议。
10. 用户输入的决策。
11. 竞争对手广告。
12. 规则风险。
13. 流程提醒。
14. 现金流预测。
15. 订单评分。
16. 采购计划。
17. 用户问题。
18. 输出格式要求。

### 9.3 AI 与规则兜底分离

当前系统特别区分：

- `AdvisorAIPrompt`：给 AI 使用，只提供事实、规则和诊断，不塞入硬编码的最终答案。
- `AdvisorPrompt` / `BuildRuleAdvisorAnswer`：AI 不可用时的规则兜底。

这样可以避免“AI 明明可用，但回答像写死模板”的问题。

## 10. 系统运行与配置

### 10.1 本地运行

```bash
env GOCACHE=/private/tmp/smarterp-go-cache go test ./...
env ADDR=:8081 GOCACHE=/private/tmp/smarterp-go-cache go run ./cmd/server
```

访问：

```text
http://localhost:8081/
```

健康检查：

```bash
curl http://localhost:8081/health
```

### 10.2 AI 配置

推荐使用本地配置文件：

```bash
cp config/ai.example.json config/ai.local.json
```

OpenRouter 示例：

```json
{
  "apiKey": "你的新 API Key",
  "baseURL": "https://openrouter.ai/api/v1",
  "model": "openai/gpt-4o-mini"
}
```

OpenAI 官方示例：

```json
{
  "apiKey": "你的 OpenAI 官方 API Key",
  "baseURL": "https://api.openai.com/v1",
  "model": "gpt-4o-mini"
}
```

检查配置：

```bash
curl http://localhost:8081/api/v1/ai/status
curl -X POST http://localhost:8081/api/v1/ai/check
```

### 10.3 安全注意事项

- 不要提交 `config/ai.local.json`。
- 不要在报告、README 或聊天记录中暴露完整 API Key。
- 前端只展示脱敏后的 Key 信息。
- 生产部署时建议使用环境变量或密钥管理服务注入配置。

## 11. 测试与验证

### 11.1 自动化测试

当前项目包含以下测试：

- 财务计算测试。
- 订单评分测试。
- 生产与采购测试。
- 市场广告测试。
- AI Prompt 与规则兜底测试。
- AI 配置读取测试。

执行命令：

```bash
env GOCACHE=/private/tmp/smarterp-go-cache go test ./...
```

### 11.2 已验证的关键场景

#### 场景 1：默认广告与短贷决策

输入：

- 初始现金 60W。
- 上年权益 60W。
- 广告 6W。
- 短贷 20W。

结果：

- 可贷额度：`60 * 3 - 20 = 160W`。
- Y1Q1 现金：`60 + 20 - 6 - 1 = 73W`。
- 现金流首行显示：流入 20W，流出 7W。

#### 场景 2：已有贷款与计划贷款合计超限

输入：

- 上年权益 40W。
- 已有长贷 50W。
- 计划长贷 40W。
- 计划短贷 40W。

规则：

```text
贷款上限 = 40 * 3 = 120W
当前可用 = 120 - 50 = 70W
计划贷款合计 = 80W
```

结果：

```text
计划贷款合计 80W 超过当前可用贷款额度 70W。
```

#### 场景 3：AI 配置状态

页面顶部展示：

- 后端在线。
- AI 已配置或未配置。
- 当前模型。

后端接口返回：

- `enabled`
- `configPath`
- `configLoaded`
- `keySource`
- `keyPreview`
- `baseURL`
- `chatCompletionsURL`
- `model`
- `warning`

#### 场景 4：移动端页面

已验证：

- 顶部状态第一屏可见。
- 无横向溢出。
- 指标卡正常展示。
- 现金流和贷款额度与桌面端一致。

## 12. 当前系统的不足

### 12.1 数据持久化能力仍需完善

当前系统已经支持可选 MySQL 持久化，能够保存年度运营记录和 Q/A 历史建议；未配置 MySQL 时仍使用内存存储。

改进方向：

- 增加数据库迁移版本管理。
- 继续扩展企业、订单、贷款、生产线、市场、ISO 等业务表。
- 增加多用户和多比赛场次隔离。
- 增加导入导出能力。

### 12.2 完整结账能力不足

当前系统具备现金流、贷款、利润表、资产负债表等基础计算，但尚未完成全自动年度结账流程。

后续应补充：

- 综合费用表。
- 完整利润表。
- 5 年补亏所得税计算。
- 自动更新权益。
- 自动更新贷款到期与还款。
- 自动更新市场/ISO换证。
- 自动更新折旧和维护费。

### 12.3 生产状态流转不完整

当前可计算产能和采购计划，但尚未完整模拟：

- 在建生产线季度推进。
- 转产状态。
- 在制品状态。
- 完工入库。
- 生产线所在厂房空间约束。

### 12.4 竞争对手建模较简化

当前竞争对手主要以广告投入数组参与广告排名估算，尚未建模：

- 竞争对手历史销售排名。
- 各市场各产品广告分布。
- 竞争对手产能和研发路线。
- 多队伍选单博弈。

### 12.5 AI 回答质量依赖外部模型

AI 建议质量受以下因素影响：

- 模型能力。
- API 连通性。
- Prompt 上下文长度。
- 用户输入数据完整度。

当前系统已经提供规则兜底，但如果要进入正式比赛辅助场景，仍建议增加更强的规则化方案生成能力。

## 13. 后续演进路线

### 13.1 第一阶段：完善数据库持久化

目标：

- 在现有 MySQL 支持基础上补齐完整业务表。
- 增加数据库迁移版本管理。
- 保留现有 API，不影响前端调用。

建议表：

- `companies`
- `annual_operation_records`
- `advisor_qa_records`
- `orders`
- `decisions`
- `loans`
- `production_lines`
- `market_progress`
- `iso_progress`

### 13.2 第二阶段：完整经营状态机

目标：

- 将运营流程从展示节点升级为可执行状态机。
- 用户按流程逐步执行决策。
- 系统自动推进季度和年度。

需要实现：

- 年初流程。
- 季初流程。
- 季度经营流程。
- 季末流程。
- 年末结账流程。

### 13.3 第三阶段：前端工程化

当前原生前端适合快速演示，但随着功能增加，应迁移到 Vue3 或 React。

建议：

- Vue3 + TypeScript。
- Pinia 管理状态。
- ECharts 展示财务图表。
- 表单 schema 化。
- API 类型自动生成。

### 13.4 第四阶段：Go-Zero 服务化

当系统需要多用户、多场次、多并发时，可迁移到 Go-Zero：

- API 网关。
- 业务服务拆分。
- 中间件统一日志、鉴权、限流。
- gRPC 内部服务。

建议拆分：

- 规则服务。
- 财务服务。
- 生产服务。
- 订单服务。
- AI 顾问服务。
- 复盘服务。

### 13.5 第五阶段：策略模型与仿真

在积累足够历史数据后，可引入模型：

- 广告投入与选单结果预测。
- 订单池价值预测。
- 多策略蒙特卡洛模拟。
- 竞争对手策略分类。
- 现金流风险概率预测。

注意：

当前没有足够真实比赛数据，直接引入 XGBoost 容易变成形式工程。建议先积累结构化历史记录，再训练模型。

## 14. 项目价值总结

当前系统已经从最初的技术设想落地为一个可运行的 ERP 沙盘智能决策工具，核心价值体现在：

1. **规则计算自动化**
   - 将 PDF 中复杂的规则转化为可执行代码。
   - 降低人工计算错误。

2. **现金流风险前置**
   - 能在执行决策前预测未来季度现金变化。
   - 支持提前调整广告、贷款、采购和产能计划。

3. **决策建议具体化**
   - 用户可以直接问“怎么投广告”“要不要开市场”“贷款是否超额”。
   - AI 会结合规则、现金流、订单、采购和历史记录回答。

4. **历史复盘结构化**
   - 年度经营数据和历史建议记录可被后续 AI 调用。
   - 系统具备从“单次计算器”升级为“持续经营顾问”的基础。

5. **工程结构可演进**
   - 当前单体架构足够轻量，便于本地演示。
   - 业务逻辑已经集中在 `internal/erp`，未来迁移数据库、Go-Zero 或 Vue3 时可以复用核心规则。

## 15. 结论

本项目当前版本已经完成 ERP 沙盘智能决策系统的核心功能闭环：

- 规则参数可查询。
- 运营流程可展示。
- 企业状态和决策可录入。
- 决策会进入现金流和贷款额度诊断。
- 订单可评分。
- 原料可倒推采购计划。
- 广告和市场可生成专项建议。
- AI 可读取规则文档、经营历史和当前诊断生成建议。
- AI 不可用时具备规则兜底。
- 年度数据和 Q/A 历史可记录。
- 前端页面可在桌面和移动端使用。

与原技术方案相比，当前版本更接近一个可交付的 MVP。系统已经补充可选 MySQL 持久化能力，后续应优先完善数据库模型、完整结账状态机和生产线状态流转，再考虑 Vue3、Go-Zero、Redis、XGBoost 等工程化与智能化扩展。

本报告建议将当前系统定位为：

```text
ERP 沙盘经营规则引擎 + AI 决策顾问 MVP
```

而不是完整竞赛平台。这样定位更准确，也更符合当前代码实现和后续可持续演进路径。
