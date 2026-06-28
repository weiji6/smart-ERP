package erp

import (
	"fmt"
	"sort"
	"strings"
)

type OperationStep struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Stage       string `json:"stage"`
	CashFlow    string `json:"cashFlow"`
	ManualInput bool   `json:"manualInput"`
	OnlyQ4      bool   `json:"onlyQ4"`
	Description string `json:"description"`
}

type DecisionInput struct {
	Type          string         `json:"type"`
	Year          int            `json:"year"`
	Quarter       int            `json:"quarter"`
	Amount        Money          `json:"amount,omitempty"`
	Market        Market         `json:"market,omitempty"`
	Product       Product        `json:"product,omitempty"`
	Material      Material       `json:"material,omitempty"`
	Quantity      int            `json:"quantity,omitempty"`
	LineType      LineType       `json:"lineType,omitempty"`
	FactoryType   FactoryType    `json:"factoryType,omitempty"`
	OrderID       string         `json:"orderId,omitempty"`
	PaymentPeriod int            `json:"paymentPeriod,omitempty"`
	Extra         map[string]any `json:"extra,omitempty"`
	Note          string         `json:"note,omitempty"`
}

type AdvisorContext struct {
	State            CompanyState            `json:"state"`
	StepCode         string                  `json:"stepCode"`
	Decisions        []DecisionInput         `json:"decisions"`
	Orders           []Order                 `json:"orders"`
	CompetitorAds    []Money                 `json:"competitorAds"`
	OperationRecords []AnnualOperationRecord `json:"operationRecords"`
	AdviceHistory    []AdvisorQARecord       `json:"adviceHistory"`
	Question         string                  `json:"question"`
}

type AdvisorDiagnostics struct {
	CurrentStep        OperationStep            `json:"currentStep"`
	NextSteps          []OperationStep          `json:"nextSteps"`
	LoanCapacity       LoanCapacity             `json:"loanCapacity"`
	CashFlow           []CashFlowForecast       `json:"cashFlow"`
	HistoryAnalysis    OperationHistoryAnalysis `json:"historyAnalysis"`
	AdPlan             AdPlanRecommendation     `json:"adPlan"`
	MarketPlan         MarketPlanRecommendation `json:"marketPlan"`
	OrderScores        []OrderScore             `json:"orderScores,omitempty"`
	MaterialNeeds      []MaterialRequirement    `json:"materialNeeds,omitempty"`
	PurchasePlan       []PurchasePlanItem       `json:"purchasePlan,omitempty"`
	RiskFindings       []string                 `json:"riskFindings"`
	RuleReminders      []string                 `json:"ruleReminders"`
	RuleRecommendation string                   `json:"ruleRecommendation"`
}

func OperationFlow() []OperationStep {
	return []OperationStep{
		{Code: "year_opening_cash", Name: "年初现金余额", Stage: "年初", CashFlow: "余额", ManualInput: true, Description: "填写新年度开始现金余额。"},
		{Code: "annual_planning", Name: "新年度规划会议", Stage: "年初", CashFlow: "无现金流动", Description: "确认年度广告、融资、研发、市场、产能和订单目标。"},
		{Code: "ad_investment", Name: "广告投放", Stage: "年初", CashFlow: "现金流出", ManualInput: true, Description: "输入广告费并确认，影响订货会选单顺序和机会。"},
		{Code: "order_selection", Name: "参加订货会选订单/登记订单", Stage: "年初", CashFlow: "无即时现金流动", ManualInput: true, Description: "按广告排名选单，登记订单、交期和账期。"},
		{Code: "pay_tax", Name: "支付应付税", Stage: "年初", CashFlow: "系统自动", Description: "系统自动扣除上年度应交所得税。"},
		{Code: "pay_long_interest", Name: "支付长贷利息", Stage: "年初", CashFlow: "系统自动", Description: "长期贷款按贷款总额 10% 四舍五入付息。"},
		{Code: "update_long_loan", Name: "更新长期贷款/长期贷款还款", Stage: "年初", CashFlow: "系统自动", Description: "到期长期贷款还本，不允许提前还款。"},
		{Code: "apply_long_loan", Name: "申请长期贷款", Stage: "年初", CashFlow: "现金流入", ManualInput: true, Description: "输入贷款数额，长短贷合计不超过上年权益 3 倍。"},
		{Code: "quarter_opening_check", Name: "季初盘点", Stage: "季初", CashFlow: "自动更新", ManualInput: true, Description: "产品下线、生产线完工自动更新，并填写余额。"},
		{Code: "update_short_loan", Name: "更新短期贷款/短期贷款还本付息", Stage: "季初", CashFlow: "系统自动", Description: "短贷到期一次还本付息。"},
		{Code: "apply_short_loan", Name: "申请短期贷款", Stage: "季初", CashFlow: "现金流入", ManualInput: true, Description: "先归还到期短贷，再在额度内申请新短贷。"},
		{Code: "material_inbound", Name: "原材料入库/更新原料订单", Stage: "季度经营", CashFlow: "现金流出", ManualInput: true, Description: "到期原料必须全额现金购买入库。"},
		{Code: "place_material_order", Name: "下原料订单", Stage: "季度经营", CashFlow: "无即时现金流动", ManualInput: true, Description: "R1/R2 提前 1 季，R3/R4 提前 2 季下单。"},
		{Code: "factory_buy_or_rent", Name: "购买/租用厂房", Stage: "季度经营", CashFlow: "现金流出", ManualInput: true, Description: "选择厂房并确认，最多拥有 2 个厂房。"},
		{Code: "production_update", Name: "更新生产/完工入库", Stage: "季度经营", CashFlow: "系统自动", Description: "系统更新在制品、完工入库和产能状态。"},
		{Code: "production_line_action", Name: "新建/在建/转产/变卖生产线", Stage: "季度经营", CashFlow: "现金流出或流入", ManualInput: true, Description: "新建、继续投资、转产或变卖生产线。"},
		{Code: "emergency_purchase", Name: "紧急采购", Stage: "随时", CashFlow: "现金流出", ManualInput: true, Description: "原料紧急采购按标准成本 2 倍，多付部分记入损失。"},
		{Code: "start_production", Name: "开始下一批生产", Stage: "季度经营", CashFlow: "现金流出", ManualInput: true, Description: "选择生产线和产品开工，确认加工和原料需求。"},
		{Code: "receivable_collection", Name: "更新应收款/应收款收现", Stage: "季度经营", CashFlow: "现金流入", ManualInput: true, Description: "录入到期金额，系统更新应收账款。"},
		{Code: "order_delivery", Name: "按订单交货", Stage: "季度经营", CashFlow: "形成应收或收入", ManualInput: true, Description: "选择订单交货，订单必须在规定季或提前交货。"},
		{Code: "product_rnd", Name: "产品研发投资", Stage: "季度经营", CashFlow: "现金流出", ManualInput: true, Description: "研发可暂停，不允许超前或集中投入。"},
		{Code: "factory_disposal", Name: "厂房出售/退租/租转买", Stage: "季度经营", CashFlow: "现金流入或流出", ManualInput: true, Description: "厂房出售形成 4 账期应收款，必要时可贴现。"},
		{Code: "market_iso_investment", Name: "新市场开拓/ISO资格投资", Stage: "第四季度", CashFlow: "现金流出", ManualInput: true, OnlyQ4: true, Description: "仅第四季度允许操作，允许暂停但不允许加速。"},
		{Code: "pay_admin_fee", Name: "支付管理费/续租/检测产品开发情况", Stage: "季末", CashFlow: "系统自动", Description: "系统自动扣除管理费、续租费并检测研发状态。"},
		{Code: "sell_inventory", Name: "出售库存", Stage: "随时", CashFlow: "现金流入", ManualInput: true, Description: "原料按八折向下取整，成品按直接成本。"},
		{Code: "factory_discount", Name: "厂房贴现", Stage: "随时", CashFlow: "现金流入", ManualInput: true, Description: "厂房出售所得 4 账期应收款可紧急贴现。"},
		{Code: "receivable_discount", Name: "应收款贴现", Stage: "随时", CashFlow: "现金流入", ManualInput: true, Description: "1/2 账期 10%，3/4 账期 12.5%，贴现费向上取整。"},
		{Code: "quarter_closing_cash", Name: "季末余额", Stage: "季末", CashFlow: "余额", ManualInput: true, Description: "填写季末现金余额。"},
		{Code: "pay_default_penalty", Name: "缴纳违约订单罚款", Stage: "年末", CashFlow: "系统自动", Description: "违约订单按销售总额 20% 四舍五入扣罚。"},
		{Code: "pay_maintenance", Name: "支付设备维护费", Stage: "年末", CashFlow: "系统自动", Description: "建成生产线和转产中生产线需交维护费。"},
		{Code: "depreciation", Name: "计提折旧", Stage: "年末", CashFlow: "无现金流动", Description: "当年建成生产线当年不提折旧，净值到残值后停止折旧。"},
		{Code: "market_iso_certificate", Name: "新市场/ISO资格换证", Stage: "年末", CashFlow: "系统自动", Description: "系统自动更新已完成市场和 ISO 资格。"},
		{Code: "closing", Name: "结账", Stage: "年末", CashFlow: "报表结算", Description: "生成综合费用表、利润表和资产负债表。"},
	}
}

func BuildAdvisorDiagnostics(ctx AdvisorContext) AdvisorDiagnostics {
	state := ctx.State
	projectedState := ApplyDecisionsToState(ctx.State, ctx.Decisions)
	currentStep := FindOperationStep(ctx.StepCode)
	nextSteps := NextOperationSteps(ctx.StepCode, 5)
	cashFlow := PredictCashFlow(projectedState, 8)
	historyAnalysis := AnalyzeOperationHistory(ctx.OperationRecords)
	orderScores := ScoreOrders(state, ctx.Orders)
	materialNeeds := CalcMaterialRequirements(ctx.Orders)
	purchasePlan := GeneratePurchasePlan(ctx.Orders, state.MaterialInventory, state.CurrentQuarterIndex())
	adPlan := BuildAdPlanRecommendation(ctx, cashFlow, orderScores)
	marketPlan := BuildMarketPlanRecommendation(ctx, cashFlow)

	riskFindings := advisorRiskFindings(state, ctx.Decisions, cashFlow, orderScores)
	riskFindings = append(riskFindings, historyAnalysis.Findings...)
	reminders := advisorRuleReminders(state, currentStep, ctx.Decisions)
	recommendation := BuildRuleAdvisorAnswer(ctx, AdvisorDiagnostics{
		CurrentStep:     currentStep,
		NextSteps:       nextSteps,
		LoanCapacity:    CalcLoanCapacity(projectedState),
		CashFlow:        cashFlow,
		HistoryAnalysis: historyAnalysis,
		AdPlan:          adPlan,
		MarketPlan:      marketPlan,
		OrderScores:     orderScores,
		MaterialNeeds:   materialNeeds,
		PurchasePlan:    purchasePlan,
		RiskFindings:    riskFindings,
		RuleReminders:   reminders,
	})

	return AdvisorDiagnostics{
		CurrentStep:        currentStep,
		NextSteps:          nextSteps,
		LoanCapacity:       CalcLoanCapacity(projectedState),
		CashFlow:           cashFlow,
		HistoryAnalysis:    historyAnalysis,
		AdPlan:             adPlan,
		MarketPlan:         marketPlan,
		OrderScores:        orderScores,
		MaterialNeeds:      materialNeeds,
		PurchasePlan:       purchasePlan,
		RiskFindings:       riskFindings,
		RuleReminders:      reminders,
		RuleRecommendation: recommendation,
	}
}

func ApplyDecisionsToState(state CompanyState, decisions []DecisionInput) CompanyState {
	if len(decisions) == 0 {
		return state
	}

	state.PlannedExpenses = append([]PlannedExpense(nil), state.PlannedExpenses...)
	state.PlannedIncomes = append([]PlannedIncome(nil), state.PlannedIncomes...)
	state.LongLoans = append([]Loan(nil), state.LongLoans...)
	state.ShortLoans = append([]Loan(nil), state.ShortLoans...)

	for i, decision := range decisions {
		index := decisionQuarterIndex(state, decision)
		switch decision.Type {
		case "ad":
			addPlannedExpense(&state, index, "广告费", decision.Amount)
		case "market_iso":
			addPlannedExpense(&state, index, "市场/ISO投资", decision.Amount)
		case "production_line":
			addPlannedExpense(&state, index, "生产线投资", decision.Amount)
		case "material_order":
			amount := decision.Amount
			if amount == 0 && decision.Quantity > 0 {
				amount = materialOrderCost(decision.Material, decision.Quantity)
			}
			addPlannedExpense(&state, index, "原料采购", amount)
		case "short_loan":
			if decision.Amount <= 0 {
				continue
			}
			state.ShortLoans = append(state.ShortLoans, Loan{
				ID:                fmt.Sprintf("decision-short-%d", i+1),
				Type:              "short",
				Principal:         decision.Amount,
				StartQuarterIndex: index,
				DueQuarterIndex:   index + 4,
			})
			addPlannedIncome(&state, index, "申请短期贷款", decision.Amount)
		case "long_loan":
			if decision.Amount <= 0 {
				continue
			}
			state.LongLoans = append(state.LongLoans, Loan{
				ID:                fmt.Sprintf("decision-long-%d", i+1),
				Type:              "long",
				Principal:         decision.Amount,
				StartQuarterIndex: index,
				DueQuarterIndex:   index + 20,
			})
			addPlannedIncome(&state, index, "申请长期贷款", decision.Amount)
		}
	}
	return state
}

func decisionQuarterIndex(state CompanyState, decision DecisionInput) int {
	year := decision.Year
	quarter := decision.Quarter
	if year == 0 {
		year = state.Year
	}
	if quarter == 0 {
		quarter = state.Quarter
	}
	return QuarterIndex(year, quarter)
}

func addPlannedExpense(state *CompanyState, index int, category string, amount Money) {
	if amount <= 0 {
		return
	}
	state.PlannedExpenses = append(state.PlannedExpenses, PlannedExpense{
		QuarterIndex: index,
		Category:     category,
		Amount:       amount,
	})
}

func addPlannedIncome(state *CompanyState, index int, category string, amount Money) {
	if amount <= 0 {
		return
	}
	state.PlannedIncomes = append(state.PlannedIncomes, PlannedIncome{
		QuarterIndex: index,
		Category:     category,
		Amount:       amount,
	})
}

func materialOrderCost(material Material, quantity int) Money {
	if quantity <= 0 {
		return 0
	}
	rule, ok := MaterialRules[material]
	if !ok {
		return 0
	}
	return Money(quantity) * rule.Price
}

func FindOperationStep(code string) OperationStep {
	flow := OperationFlow()
	for _, step := range flow {
		if step.Code == code {
			return step
		}
	}
	if code == "" {
		return flow[0]
	}
	return OperationStep{Code: code, Name: "自定义决策步骤", Stage: "自定义", Description: "未匹配到运营流程表中的标准步骤。"}
}

func NextOperationSteps(code string, limit int) []OperationStep {
	flow := OperationFlow()
	if limit <= 0 {
		limit = 5
	}
	start := 0
	for i, step := range flow {
		if step.Code == code {
			start = i + 1
			break
		}
	}
	end := start + limit
	if end > len(flow) {
		end = len(flow)
	}
	if start >= len(flow) {
		return nil
	}
	return flow[start:end]
}

func AdvisorPrompt(ctx AdvisorContext, diagnostics AdvisorDiagnostics) string {
	var b strings.Builder
	b.WriteString("你是资深ERP沙盘模拟经营顾问，目标是在规则允许范围内最大化所有者权益，并避免现金流断裂。\n")
	b.WriteString("请严格依据沙盘规则、年度运营流程表、当前企业状态、历史年度运营记录和规则引擎诊断来评估用户决策。\n")
	b.WriteString("核心要求：必须正面回答用户问题；如果用户问题包含多个子问题，要逐条回答，不允许只给泛泛经营建议。\n")
	b.WriteString("问题意图优先级高于当前流程节点；当前流程节点只能作为上下文，不能把广告节点、订单节点等误当成用户真正想问的主题。\n")
	b.WriteString("输出要求：\n")
	b.WriteString("1. 第一段必须是【直接回答】，明确回应用户原问题，并说明建议执行、调整后执行或不建议执行。\n")
	b.WriteString("2. 再给【关键依据】：现金流、贷款额度、广告ROI、订单交付、产能/原料、研发/市场/ISO节奏、历史经营趋势。\n")
	b.WriteString("3. 再给【调整方案】：列出金额、季度、优先级和下一步流程动作。\n")
	b.WriteString("4. 最后给【仍需确认】：列出你无法从输入中确定、但会影响结论的数据。\n")
	b.WriteString("5. 内容要专业、具体、可执行，避免空泛口号；金额单位均为W（万元）。\n\n")
	b.WriteString(fmt.Sprintf("当前企业：第%d年Q%d，现金%dW，权益%dW，上年权益%dW。\n",
		ctx.State.Year, ctx.State.Quarter, ctx.State.Cash, ctx.State.Equity, ctx.State.LastYearEquity))
	b.WriteString(fmt.Sprintf("当前流程节点：%s（%s）。说明：%s\n", diagnostics.CurrentStep.Name, diagnostics.CurrentStep.Stage, diagnostics.CurrentStep.Description))
	b.WriteString(fmt.Sprintf("贷款额度：上限%dW，已用%dW，可用%dW。\n",
		diagnostics.LoanCapacity.Limit, diagnostics.LoanCapacity.Used, diagnostics.LoanCapacity.Available))
	b.WriteString("\n年度运营历史：\n")
	b.WriteString(FormatOperationHistory(ctx.OperationRecords, diagnostics.HistoryAnalysis))
	b.WriteString("\n历史建议记录：\n")
	b.WriteString(FormatAdvisorQAHistory(ctx.AdviceHistory))
	if diagnostics.AdPlan.Relevant {
		b.WriteString("\n广告专项测算：\n")
		b.WriteString(FormatAdPlan(diagnostics.AdPlan))
	}
	if diagnostics.MarketPlan.Relevant {
		b.WriteString("\n市场开拓专项测算：\n")
		b.WriteString(FormatMarketPlan(diagnostics.MarketPlan))
	}

	if len(ctx.Decisions) > 0 {
		b.WriteString("\n用户输入的决策：\n")
		for _, decision := range ctx.Decisions {
			b.WriteString(formatDecision(decision))
			b.WriteByte('\n')
		}
	}

	if len(diagnostics.RiskFindings) > 0 {
		b.WriteString("\n规则引擎发现的风险：\n")
		for _, item := range diagnostics.RiskFindings {
			b.WriteString("- " + item + "\n")
		}
	}
	if len(diagnostics.RuleReminders) > 0 {
		b.WriteString("\n流程和规则提醒：\n")
		for _, item := range diagnostics.RuleReminders {
			b.WriteString("- " + item + "\n")
		}
	}

	b.WriteString("\n未来8季度现金流预测：\n")
	for _, f := range diagnostics.CashFlow {
		b.WriteString(fmt.Sprintf("- Y%dQ%d：现金%dW，流入%dW，流出%dW，预警%s。\n",
			f.Year, f.Quarter, f.PredictedCash, f.Inflows, f.Outflows, f.WarningLevel))
	}

	if len(diagnostics.OrderScores) > 0 {
		b.WriteString("\n订单评分：\n")
		for _, score := range diagnostics.OrderScores {
			b.WriteString(fmt.Sprintf("- 订单%s：可行=%v，评分=%.2f，利润=%dW，原因=%s。\n",
				score.Order.ID, score.Feasible, score.Score, score.Profit, score.Reason))
		}
	}

	if len(diagnostics.PurchasePlan) > 0 {
		b.WriteString("\n建议采购计划：\n")
		for _, item := range diagnostics.PurchasePlan {
			y, q := YearQuarter(item.OrderQuarter)
			ny, nq := YearQuarter(item.NeedQuarter)
			b.WriteString(fmt.Sprintf("- %s 数量%d，Y%dQ%d下单，Y%dQ%d需要，成本%dW。\n",
				item.Material, item.Quantity, y, q, ny, nq, item.Cost))
		}
	}

	if ctx.Question != "" {
		b.WriteString("\n用户问题：" + ctx.Question + "\n")
	} else {
		b.WriteString("\n用户问题：未填写具体问题，请围绕当前流程和决策主动指出最重要的经营建议。\n")
	}

	b.WriteString("\n请按以下结构返回：\n")
	b.WriteString("一、直接回答用户问题：必须引用或复述用户问题，并直接给出答案。\n")
	b.WriteString("如果用户问广告怎么投，必须给出具体投放表：市场、产品、金额、预计选单机会、理由；不能只说“结合现金流和产能”。\n")
	b.WriteString("如果用户问是否开拓/开放新市场，必须给出具体市场开拓表：市场、是否建议开、何时投、每年费用、完成年份、优先级和理由；不能回答广告投放方案。\n")
	b.WriteString("二、关键依据：分别分析现金流、融资、广告/订单、产能/采购、研发/市场/ISO、历史数据启示。\n")
	b.WriteString("三、修改方案：给出可直接照做的步骤，包含金额、季度、动作和优先级。\n")
	b.WriteString("四、下一流程提醒：说明当前节点之后3-5个运营动作，以及必须提前准备的数据。\n")
	b.WriteString("五、仍需确认：列出缺失数据和人工确认点。")
	return b.String()
}

func AdvisorAIPrompt(ctx AdvisorContext, diagnostics AdvisorDiagnostics) string {
	var b strings.Builder
	b.WriteString("你是资深ERP沙盘模拟经营顾问，目标是在规则允许范围内最大化所有者权益，并避免现金流断裂。\n")
	b.WriteString("你必须独立回答用户问题。规则引擎只提供事实、风险和数据，不提供最终答案；不要照抄任何固定模板或兜底话术。\n")
	b.WriteString("问题意图优先级高于当前流程节点；当前流程节点只能作为上下文，不能把广告节点、订单节点等误当成用户真正想问的主题。\n")
	b.WriteString("如果用户问广告，请你自己根据现金、市场、产品、订单、竞争广告和产能给出具体投放方案。\n")
	b.WriteString("如果用户问市场开拓，请你自己根据现金、研发、市场周期、后续订单价值和经营年限给出是否开、开哪个、何时开。\n")
	b.WriteString("输出必须先直接回答用户问题，再给依据、方案和需要补充的数据；金额单位均为W（万元）。\n\n")
	if docs := PromptRuleDocuments(); docs != "" {
		b.WriteString("以下是原始规则文档和运营流程文档上下文，回答必须优先遵守这些规则：\n")
		b.WriteString(docs)
		b.WriteString("\n\n")
	}

	b.WriteString(fmt.Sprintf("当前企业：第%d年Q%d，现金%dW，权益%dW，上年权益%dW。\n",
		ctx.State.Year, ctx.State.Quarter, ctx.State.Cash, ctx.State.Equity, ctx.State.LastYearEquity))
	b.WriteString(fmt.Sprintf("当前流程节点：%s（%s）。说明：%s\n", diagnostics.CurrentStep.Name, diagnostics.CurrentStep.Stage, diagnostics.CurrentStep.Description))
	b.WriteString(fmt.Sprintf("贷款额度：上限%dW，已用%dW，可用%dW。\n",
		diagnostics.LoanCapacity.Limit, diagnostics.LoanCapacity.Used, diagnostics.LoanCapacity.Available))

	b.WriteString("\n年度运营历史：\n")
	b.WriteString(FormatOperationHistory(ctx.OperationRecords, diagnostics.HistoryAnalysis))
	b.WriteString("\n历史建议记录：\n")
	b.WriteString(FormatAdvisorQAHistory(ctx.AdviceHistory))

	if len(ctx.Decisions) > 0 {
		b.WriteString("\n用户输入的决策：\n")
		for _, decision := range ctx.Decisions {
			b.WriteString(formatDecision(decision))
			b.WriteByte('\n')
		}
	}

	if len(ctx.CompetitorAds) > 0 {
		b.WriteString("\n竞争对手广告参考：")
		for i, ad := range ctx.CompetitorAds {
			if i > 0 {
				b.WriteString("，")
			}
			b.WriteString(fmt.Sprintf("%dW", ad))
		}
		b.WriteByte('\n')
	}

	if len(diagnostics.RiskFindings) > 0 {
		b.WriteString("\n规则引擎发现的风险：\n")
		for _, item := range diagnostics.RiskFindings {
			b.WriteString("- " + item + "\n")
		}
	}
	if len(diagnostics.RuleReminders) > 0 {
		b.WriteString("\n流程和规则提醒：\n")
		for _, item := range diagnostics.RuleReminders {
			b.WriteString("- " + item + "\n")
		}
	}

	b.WriteString("\n未来8季度现金流预测：\n")
	for _, f := range diagnostics.CashFlow {
		b.WriteString(fmt.Sprintf("- Y%dQ%d：现金%dW，流入%dW，流出%dW，预警%s。\n",
			f.Year, f.Quarter, f.PredictedCash, f.Inflows, f.Outflows, f.WarningLevel))
	}

	if len(diagnostics.OrderScores) > 0 {
		b.WriteString("\n订单评分：\n")
		for _, score := range diagnostics.OrderScores {
			b.WriteString(fmt.Sprintf("- 订单%s：市场=%s，产品=%s，数量=%d，总价=%dW，可行=%v，评分=%.2f，利润=%dW，原因=%s。\n",
				score.Order.ID, score.Order.Market, score.Order.Product, score.Order.Quantity, score.Order.TotalPrice,
				score.Feasible, score.Score, score.Profit, score.Reason))
		}
	}

	if len(diagnostics.PurchasePlan) > 0 {
		b.WriteString("\n规则引擎倒推的采购需求：\n")
		for _, item := range diagnostics.PurchasePlan {
			y, q := YearQuarter(item.OrderQuarter)
			ny, nq := YearQuarter(item.NeedQuarter)
			b.WriteString(fmt.Sprintf("- %s 数量%d，Y%dQ%d下单，Y%dQ%d需要，标准成本%dW，紧急采购成本%dW。\n",
				item.Material, item.Quantity, y, q, ny, nq, item.Cost, item.EmergencyCost))
		}
	}

	if ctx.Question != "" {
		b.WriteString("\n用户问题：" + ctx.Question + "\n")
	} else {
		b.WriteString("\n用户问题：未填写具体问题，请围绕当前流程和决策主动指出最重要的经营建议。\n")
	}

	b.WriteString("\n请按以下结构返回：\n")
	b.WriteString("一、直接回答用户问题：必须引用或复述用户问题，并给出明确答案。\n")
	b.WriteString("二、为什么：只列和用户问题直接相关的依据，不要被当前流程节点带偏。\n")
	b.WriteString("三、具体方案：给出可执行动作、金额、季度和优先级。若问题涉及广告或市场，必须给具体表格。\n")
	b.WriteString("四、风险和补充数据：列出仍需用户补充的数据，以及这些数据会如何改变结论。")
	return b.String()
}

func advisorRiskFindings(state CompanyState, decisions []DecisionInput, cashFlow []CashFlowForecast, orderScores []OrderScore) []string {
	var findings []string
	for _, f := range cashFlow {
		if f.WarningLevel == WarningDanger || f.WarningLevel == WarningCritical {
			findings = append(findings, fmt.Sprintf("Y%dQ%d 现金预测为 %dW，达到 %s 预警。", f.Year, f.Quarter, f.PredictedCash, f.WarningLevel))
			break
		}
	}
	for _, score := range orderScores {
		if !score.Feasible {
			findings = append(findings, fmt.Sprintf("订单 %s 不可行：%s。", score.Order.ID, score.Reason))
		}
	}

	loanCapacity := CalcLoanCapacity(state)
	var plannedLoan Money
	for _, decision := range decisions {
		switch decision.Type {
		case "long_loan", "short_loan":
			plannedLoan += decision.Amount
			if decision.Amount > 0 && decision.Amount < 10 {
				findings = append(findings, "长期/短期贷款必须为大于等于 10W 的整数。")
			}
		case "market_iso":
			if decision.Quarter != 0 && decision.Quarter != 4 {
				findings = append(findings, "市场开拓和 ISO 投资只能在第四季度操作。")
			}
		case "ad":
			if decision.Amount > 0 && decision.Amount < MinSingleAd {
				findings = append(findings, "单个广告投入低于最小广告额 1W，无法获得有效选单机会。")
			}
		}
	}
	if plannedLoan > loanCapacity.Available {
		findings = append(findings, fmt.Sprintf("计划贷款合计 %dW 超过当前可用贷款额度 %dW。", plannedLoan, loanCapacity.Available))
	}
	return dedupeStrings(findings)
}

func advisorRuleReminders(state CompanyState, step OperationStep, decisions []DecisionInput) []string {
	var reminders []string
	if step.OnlyQ4 && state.Quarter != 4 {
		reminders = append(reminders, "当前不是第四季度，该流程节点按规则不可执行。")
	}
	if step.Code == "place_material_order" {
		reminders = append(reminders, "R1/R2 提前 1 季下单，R3/R4 提前 2 季下单；没有下订单的原料不能采购入库。")
	}
	if step.Code == "order_delivery" {
		reminders = append(reminders, "订单必须在规定季或提前交货，延期将按订单总额 20% 罚款并收回订单。")
	}
	if step.Code == "receivable_discount" {
		reminders = append(reminders, "应收款贴现需按季度分开贴现，1/2 账期 10%，3/4 账期 12.5%，费用向上取整。")
	}
	for _, decision := range decisions {
		if decision.Type == "production_line" && len(state.Factories) == 0 {
			reminders = append(reminders, "新建生产线前必须先拥有或租用厂房，且厂房有空位。")
		}
	}
	return dedupeStrings(reminders)
}

func advisorRuleRecommendation(step OperationStep, risks []string, reminders []string) string {
	if len(risks) > 0 {
		return "不建议直接执行当前方案，需先处理现金流、额度或交付可行性风险。"
	}
	if step.Code == "ad_investment" {
		return "可执行前请结合订单毛利和产能，避免广告拿到订单后无法交付。"
	}
	if step.Code == "place_material_order" {
		return "优先按已接订单倒推原料提前期，避免紧急采购造成损失。"
	}
	if len(reminders) > 0 {
		return "当前方案可继续评估，但必须满足流程节点限制。"
	}
	return "当前方案未发现硬规则冲突，可结合 AI 建议进一步优化收益和风险。"
}

func BuildRuleAdvisorAnswer(ctx AdvisorContext, diagnostics AdvisorDiagnostics) string {
	question := strings.TrimSpace(ctx.Question)
	if question == "" {
		question = "当前方案是否建议执行"
	}

	var b strings.Builder
	b.WriteString("一、直接回答用户问题\n")
	if diagnostics.MarketPlan.Relevant && len(diagnostics.MarketPlan.Items) > 0 {
		b.WriteString(fmt.Sprintf("针对“%s”：%s。", question, summarizeMarketPlan(diagnostics.MarketPlan)))
		if ctx.State.Quarter != 4 {
			b.WriteString(" 注意市场开拓只能在第四季度投入，现在先做准备，到Q4再执行。")
		}
	} else if diagnostics.AdPlan.Relevant && len(diagnostics.AdPlan.Items) > 0 {
		b.WriteString(fmt.Sprintf("针对“%s”：建议按%s执行。", question, summarizeAdPlan(diagnostics.AdPlan)))
		if len(diagnostics.RiskFindings) > 0 {
			b.WriteString(" 但当前存在规则或交付风险，投放前要先处理风险项。")
		}
	} else if len(diagnostics.RiskFindings) > 0 {
		b.WriteString(fmt.Sprintf("针对“%s”：不建议直接执行当前方案，至少要先处理规则风险和交付可行性问题。", question))
	} else if len(diagnostics.RuleReminders) > 0 {
		b.WriteString(fmt.Sprintf("针对“%s”：可以继续评估，但必须先满足当前流程节点限制。", question))
	} else {
		b.WriteString(fmt.Sprintf("针对“%s”：当前输入没有发现硬规则冲突，可以执行，但仍建议按现金流、产能和订单毛利再做一次校验。", question))
	}
	b.WriteString("\n\n二、关键依据\n")
	if diagnostics.MarketPlan.Relevant {
		b.WriteString("- 问题主题：市场开拓/市场开放，以下判断优先围绕新市场是否值得开、何时开、开哪个市场。\n")
	} else {
		b.WriteString(fmt.Sprintf("- 当前节点：%s（%s），%s\n", diagnostics.CurrentStep.Name, diagnostics.CurrentStep.Stage, diagnostics.CurrentStep.Description))
	}
	b.WriteString(fmt.Sprintf("- 融资空间：上年权益%dW，贷款上限%dW，已用%dW，可用%dW。\n",
		diagnostics.LoanCapacity.LastYearEquity, diagnostics.LoanCapacity.Limit, diagnostics.LoanCapacity.Used, diagnostics.LoanCapacity.Available))

	if minForecast, ok := lowestCashForecast(diagnostics.CashFlow); ok {
		b.WriteString(fmt.Sprintf("- 现金流：未来8季度最低现金预计出现在Y%dQ%d，为%dW，预警级别%s。\n",
			minForecast.Year, minForecast.Quarter, minForecast.PredictedCash, minForecast.WarningLevel))
	}
	if diagnostics.HistoryAnalysis.Years > 0 {
		b.WriteString(fmt.Sprintf("- 历史经营：已记录%d年，累计销售%dW，累计净利%dW，债务压力%s。\n",
			diagnostics.HistoryAnalysis.Years, diagnostics.HistoryAnalysis.TotalSalesRevenue,
			diagnostics.HistoryAnalysis.TotalNetProfit, diagnostics.HistoryAnalysis.DebtPressure))
	}
	for _, score := range diagnostics.OrderScores {
		if !score.Feasible {
			b.WriteString(fmt.Sprintf("- 订单%s不可行：%s。\n", score.Order.ID, score.Reason))
		}
	}
	for _, risk := range diagnostics.RiskFindings {
		b.WriteString("- 风险：" + risk + "\n")
	}
	for _, reminder := range diagnostics.RuleReminders {
		b.WriteString("- 提醒：" + reminder + "\n")
	}

	b.WriteString("\n三、修改方案\n")
	if diagnostics.MarketPlan.Relevant {
		b.WriteString(FormatMarketPlan(diagnostics.MarketPlan))
	} else if diagnostics.AdPlan.Relevant {
		b.WriteString(FormatAdPlan(diagnostics.AdPlan))
	}
	if len(diagnostics.RiskFindings) > 0 {
		b.WriteString("- 优先处理不可行订单、现金预警或贷款额度问题，再决定是否执行当前决策。\n")
	} else if !diagnostics.MarketPlan.Relevant && !diagnostics.AdPlan.Relevant {
		b.WriteString("- 当前方案可作为基线方案，但广告、贷款、研发和采购金额建议保留至少30W现金安全垫。\n")
	}
	if len(diagnostics.PurchasePlan) > 0 && !diagnostics.MarketPlan.Relevant && !diagnostics.AdPlan.Relevant {
		for _, item := range diagnostics.PurchasePlan {
			y, q := YearQuarter(item.OrderQuarter)
			ny, nq := YearQuarter(item.NeedQuarter)
			b.WriteString(fmt.Sprintf("- %s建议Y%dQ%d下单%d个，Y%dQ%d需要，标准成本%dW；如果错过提前期，紧急采购成本约%dW。\n",
				item.Material, y, q, item.Quantity, ny, nq, item.Cost, item.EmergencyCost))
		}
	}
	if diagnostics.AdPlan.Relevant {
		b.WriteString("- 广告投放要绑定订单利润和交付能力：没有产能把握的产品不要重投，避免拿单后违约。\n")
	}

	b.WriteString("\n四、下一流程提醒\n")
	if diagnostics.MarketPlan.Relevant {
		b.WriteString("- 当前不是第四季度时，先记录目标市场、预算和产品路线，不要提前投入。\n")
		b.WriteString(fmt.Sprintf("- 到Y%dQ4执行“新市场开拓/ISO资格投资”，按优先级投入市场开拓费。\n", ctx.State.Year))
		b.WriteString("- 同步检查产品研发和产能是否能承接新市场订单，否则市场开了也难转化为权益。\n")
	} else {
		for _, step := range diagnostics.NextSteps {
			b.WriteString(fmt.Sprintf("- %s：%s\n", step.Name, step.Description))
		}
	}

	b.WriteString("\n五、仍需确认\n")
	b.WriteString("- 请补充竞争对手广告、可选订单池、现有生产线状态和原料在途订单；这些数据会直接影响选单、交付和采购判断。")
	return b.String()
}

func summarizeAdPlan(plan AdPlanRecommendation) string {
	parts := make([]string, 0, len(plan.Items))
	for _, item := range plan.Items {
		parts = append(parts, fmt.Sprintf("%s/%s投%dW", item.Market, item.Product, item.AdCost))
	}
	return fmt.Sprintf("总预算%dW（%s）", plan.Budget, strings.Join(parts, "，"))
}

func summarizeMarketPlan(plan MarketPlanRecommendation) string {
	parts := make([]string, 0, len(plan.Items))
	for _, item := range plan.Items {
		parts = append(parts, fmt.Sprintf("优先级%d开%s：Y%dQ%d投%dW，预计Y%d可用",
			item.Priority, item.Market, item.StartYear, item.StartQuarter, item.FeeThisYear, item.ReadyYear))
	}
	return "建议按这个顺序开拓市场：" + strings.Join(parts, "；")
}

func lowestCashForecast(items []CashFlowForecast) (CashFlowForecast, bool) {
	if len(items) == 0 {
		return CashFlowForecast{}, false
	}
	lowest := items[0]
	for _, item := range items[1:] {
		if item.PredictedCash < lowest.PredictedCash {
			lowest = item
		}
	}
	return lowest, true
}

func formatDecision(decision DecisionInput) string {
	parts := []string{fmt.Sprintf("- 类型=%s", decision.Type)}
	if decision.Year > 0 || decision.Quarter > 0 {
		parts = append(parts, fmt.Sprintf("时间=Y%dQ%d", decision.Year, decision.Quarter))
	}
	if decision.Amount != 0 {
		parts = append(parts, fmt.Sprintf("金额=%dW", decision.Amount))
	}
	if decision.Market != "" {
		parts = append(parts, fmt.Sprintf("市场=%s", decision.Market))
	}
	if decision.Product != "" {
		parts = append(parts, fmt.Sprintf("产品=%s", decision.Product))
	}
	if decision.Material != "" {
		parts = append(parts, fmt.Sprintf("原料=%s", decision.Material))
	}
	if decision.Quantity != 0 {
		parts = append(parts, fmt.Sprintf("数量=%d", decision.Quantity))
	}
	if decision.LineType != "" {
		parts = append(parts, fmt.Sprintf("生产线=%s", decision.LineType))
	}
	if decision.FactoryType != "" {
		parts = append(parts, fmt.Sprintf("厂房=%s", decision.FactoryType))
	}
	if decision.OrderID != "" {
		parts = append(parts, fmt.Sprintf("订单=%s", decision.OrderID))
	}
	if len(decision.Extra) > 0 {
		keys := make([]string, 0, len(decision.Extra))
		for key := range decision.Extra {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		extras := make([]string, 0, len(keys))
		for _, key := range keys {
			extras = append(extras, fmt.Sprintf("%s=%v", key, decision.Extra[key]))
		}
		parts = append(parts, "附加="+strings.Join(extras, "/"))
	}
	if decision.Note != "" {
		parts = append(parts, fmt.Sprintf("备注=%s", decision.Note))
	}
	return strings.Join(parts, "，")
}

func dedupeStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
