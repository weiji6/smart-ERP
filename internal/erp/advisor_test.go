package erp

import (
	"strings"
	"testing"
)

func TestOperationFlowContainsPDFSteps(t *testing.T) {
	flow := OperationFlow()
	if len(flow) < 30 {
		t.Fatalf("运营流程节点数量过少: %d", len(flow))
	}
	if flow[2].Code != "ad_investment" {
		t.Fatalf("第三个节点应为广告投放，got %+v", flow[2])
	}
	if !FindOperationStep("market_iso_investment").OnlyQ4 {
		t.Fatalf("市场/ISO 投资应标记为仅第四季度可操作")
	}
}

func TestAdvisorDiagnosticsFindsLoanAndQ4Risks(t *testing.T) {
	state := DefaultCompanyState()
	state.Year = 2
	state.Quarter = 2
	state.LastYearEquity = 40
	state.LongLoans = []Loan{{Principal: 100}}

	ctx := AdvisorContext{
		State:    state,
		StepCode: "market_iso_investment",
		Decisions: []DecisionInput{
			{Type: "long_loan", Amount: 30},
			{Type: "market_iso", Quarter: 2, Market: MarketDomestic},
		},
	}
	got := BuildAdvisorDiagnostics(ctx)
	joined := strings.Join(append(got.RiskFindings, got.RuleReminders...), "\n")
	if !strings.Contains(joined, "超过当前可用贷款额度") {
		t.Fatalf("应发现贷款额度风险: %s", joined)
	}
	if !strings.Contains(joined, "第四季度") {
		t.Fatalf("应提醒市场/ISO 仅第四季度操作: %s", joined)
	}
}

func TestAdvisorDiagnosticsUsesDecisionCashEffects(t *testing.T) {
	state := DefaultCompanyState()
	state.Cash = 60
	state.LastYearEquity = 60

	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		Decisions: []DecisionInput{
			{Type: "ad", Amount: 6},
			{Type: "short_loan", Amount: 20},
		},
	}

	got := BuildAdvisorDiagnostics(ctx)
	if got.LoanCapacity.Available != 160 {
		t.Fatalf("计划短贷应占用贷款额度，got %+v", got.LoanCapacity)
	}
	if len(got.CashFlow) == 0 {
		t.Fatal("应生成现金流预测")
	}
	if got.CashFlow[0].Inflows != 20 || got.CashFlow[0].Outflows != 7 || got.CashFlow[0].PredictedCash != 73 {
		t.Fatalf("现金流应包含贷款流入、广告费和管理费，got %+v", got.CashFlow[0])
	}
}

func TestAdvisorDiagnosticsFindsTotalPlannedLoanRisk(t *testing.T) {
	state := DefaultCompanyState()
	state.LastYearEquity = 40
	state.LongLoans = []Loan{{Principal: 50}}

	ctx := AdvisorContext{
		State:    state,
		StepCode: "apply_short_loan",
		Decisions: []DecisionInput{
			{Type: "long_loan", Amount: 40},
			{Type: "short_loan", Amount: 40},
		},
	}

	got := BuildAdvisorDiagnostics(ctx)
	joined := strings.Join(got.RiskFindings, "\n")
	if !strings.Contains(joined, "计划贷款合计 80W 超过当前可用贷款额度 70W") {
		t.Fatalf("应按计划贷款合计校验额度，got %s", joined)
	}
}

func TestAdvisorPromptIncludesDecisionAndCashFlow(t *testing.T) {
	state := DefaultCompanyState()
	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		Decisions: []DecisionInput{
			{Type: "ad", Amount: 5, Market: MarketLocal, Product: ProductP1},
		},
		Question: "广告是否过高？",
	}
	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := AdvisorPrompt(ctx, diagnostics)
	for _, want := range []string{"广告投放", "金额=5W", "未来8季度现金流预测", "广告是否过高"} {
		if !strings.Contains(got, want) {
			t.Fatalf("提示词缺少 %q:\n%s", want, got)
		}
	}
}

func TestAdvisorPromptIncludesOperationHistory(t *testing.T) {
	state := DefaultCompanyState()
	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		OperationRecords: []AnnualOperationRecord{
			{
				Company:         "默认企业",
				Year:            1,
				OpeningCash:     60,
				ClosingCash:     8,
				OpeningEquity:   60,
				ClosingEquity:   52,
				SalesRevenue:    20,
				NetProfit:       -8,
				AdExpense:       10,
				TotalLiability:  40,
				LongLoanBalance: 40,
				KeyEvents:       "广告投入过高",
				Review:          "第二年要先保现金",
			},
		},
	}

	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := AdvisorPrompt(ctx, diagnostics)
	for _, want := range []string{"年度运营历史", "Y1", "广告投入过高", "第二年要先保现金", "历史风险信号"} {
		if !strings.Contains(got, want) {
			t.Fatalf("提示词缺少历史字段 %q:\n%s", want, got)
		}
	}
	if diagnostics.HistoryAnalysis.Years != 1 {
		t.Fatalf("应生成历史分析，got %+v", diagnostics.HistoryAnalysis)
	}
	if diagnostics.HistoryAnalysis.DebtPressure != "低" {
		t.Fatalf("应使用贷款余额兜底计算债务压力，got %+v", diagnostics.HistoryAnalysis)
	}
	if !strings.Contains(strings.Join(diagnostics.RiskFindings, "\n"), "安全垫不足") {
		t.Fatalf("应将历史风险写入规则诊断: %+v", diagnostics.RiskFindings)
	}
}

func TestRuleAdvisorAnswerAnswersQuestion(t *testing.T) {
	state := DefaultCompanyState()
	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		Question: "第二年广告是否应该加到10W？",
		Orders: []Order{
			{ID: "B", Market: MarketLocal, Product: ProductP2, Quantity: 3, TotalPrice: 30, DeliveryQuarter: 2, PaymentPeriod: 1},
		},
	}

	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := diagnostics.RuleRecommendation
	for _, want := range []string{"直接回答用户问题", "第二年广告是否应该加到10W", "关键依据", "修改方案"} {
		if !strings.Contains(got, want) {
			t.Fatalf("规则建议缺少 %q:\n%s", want, got)
		}
	}
}

func TestRuleAdvisorAnswerIncludesConcreteAdPlan(t *testing.T) {
	state := DefaultCompanyState()
	state.Cash = 60
	state.Markets = map[Market]MarketProgress{
		MarketLocal:    {Market: MarketLocal, Opened: true},
		MarketRegional: {Market: MarketRegional, Opened: true},
	}
	state.RnD = map[Product]RnDProgress{
		ProductP1: {Product: ProductP1, Finished: true},
		ProductP2: {Product: ProductP2, Finished: true},
	}
	ctx := AdvisorContext{
		State:         state,
		StepCode:      "ad_investment",
		Question:      "你觉得我怎么投广告比较好？",
		CompetitorAds: []Money{5, 4, 3},
		Decisions: []DecisionInput{
			{Type: "ad", Market: MarketLocal, Product: ProductP2, Amount: 6},
		},
	}

	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := diagnostics.RuleRecommendation
	for _, want := range []string{"广告专项建议", "具体投放表", "投", "预计选单机会", "理由"} {
		if !strings.Contains(got, want) {
			t.Fatalf("广告建议缺少 %q:\n%s", want, got)
		}
	}
	if !diagnostics.AdPlan.Relevant || len(diagnostics.AdPlan.Items) == 0 {
		t.Fatalf("应生成广告专项方案: %+v", diagnostics.AdPlan)
	}
}

func TestMarketQuestionDoesNotReturnAdPlan(t *testing.T) {
	state := DefaultCompanyState()
	state.Markets = map[Market]MarketProgress{
		MarketLocal: {Market: MarketLocal, Opened: true},
	}
	state.RnD = map[Product]RnDProgress{
		ProductP1: {Product: ProductP1, Finished: true},
		ProductP2: {Product: ProductP2, Finished: true},
	}
	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		Question: "我现在要不要开放新市场？",
		Decisions: []DecisionInput{
			{Type: "ad", Market: MarketLocal, Product: ProductP2, Amount: 6},
		},
	}

	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := diagnostics.RuleRecommendation
	if diagnostics.AdPlan.Relevant {
		t.Fatalf("市场问题不应触发广告方案: %+v", diagnostics.AdPlan)
	}
	if !diagnostics.MarketPlan.Relevant {
		t.Fatalf("市场问题应触发市场方案")
	}
	for _, want := range []string{"市场开拓专项建议", "具体开拓表", "区域", "Y1Q4"} {
		if !strings.Contains(got, want) {
			t.Fatalf("市场建议缺少 %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "广告专项建议") {
		t.Fatalf("市场问题不应返回广告专项建议:\n%s", got)
	}
}

func TestAIPromptDoesNotEmbedFallbackPlans(t *testing.T) {
	state := DefaultCompanyState()
	state.Markets = map[Market]MarketProgress{
		MarketLocal: {Market: MarketLocal, Opened: true},
	}
	ctx := AdvisorContext{
		State:    state,
		StepCode: "ad_investment",
		Question: "我现在要不要开放新市场？",
	}

	diagnostics := BuildAdvisorDiagnostics(ctx)
	got := AdvisorAIPrompt(ctx, diagnostics)
	for _, forbidden := range []string{"广告专项建议", "市场开拓专项建议", "具体开拓表", "建议按这个顺序开拓市场"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("AI Prompt 不应包含兜底固定方案 %q:\n%s", forbidden, got)
		}
	}
	for _, want := range []string{"必须独立回答用户问题", "用户问题：我现在要不要开放新市场？", "未来8季度现金流预测"} {
		if !strings.Contains(got, want) {
			t.Fatalf("AI Prompt 缺少必要上下文 %q:\n%s", want, got)
		}
	}
	for _, want := range []string{"ERP 沙盘规则摘要", "年度运营流程表", "融资", "广告投放"} {
		if !strings.Contains(got, want) {
			t.Fatalf("AI Prompt 缺少规则文档内容 %q:\n%s", want, got)
		}
	}
}
