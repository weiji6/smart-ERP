package erp

import (
	"fmt"
	"strings"
)

const CompetitionYears = 7

type AnnualOperationYear struct {
	Year        int                 `json:"year"`
	Quarters    []AnnualQuarterFlow `json:"quarters"`
	ClosingForm AnnualReportSpec    `json:"closingForm"`
}

type AnnualQuarterFlow struct {
	Quarter int             `json:"quarter"`
	Steps   []OperationStep `json:"steps"`
}

type AnnualReportSpec struct {
	ComprehensiveExpense []string `json:"comprehensiveExpense"`
	Profit               []string `json:"profit"`
	BalanceAssets        []string `json:"balanceAssets"`
	BalanceLiabilities   []string `json:"balanceLiabilities"`
	Formulas             []string `json:"formulas"`
}

func AnnualReportTemplate() AnnualReportSpec {
	return AnnualReportSpec{
		ComprehensiveExpense: []string{
			"管理费", "广告费", "设备维护费", "转产费", "租金", "市场准入开拓",
			"产品研发", "ISO 认证资格", "信息费", "其他", "合计",
		},
		Profit: []string{
			"销售收入", "直接成本", "毛利", "综合费用", "折旧前利润", "折旧",
			"支付利息前利润", "财务费用", "税前利润", "所得税", "年度净利润",
		},
		BalanceAssets: []string{
			"现金", "应收款", "在制品", "产成品", "原材料", "流动资产合计",
			"厂房", "生产线", "在建工程", "固定资产合计", "资产总计",
		},
		BalanceLiabilities: []string{
			"长期负债", "短期负债", "应交所得税", "负债合计", "股东资本",
			"利润留存", "年度净利", "所有者权益合计", "负债和所有者权益总计",
		},
		Formulas: []string{
			"销售收入=今年所卖订单总额",
			"直接成本=所卖产品总成本",
			"毛利=销售收入-直接成本",
			"综合费用=综合费用表合计",
			"折旧前利润=毛利-综合费用",
			"支付利息前利润=折旧前利润-折旧",
			"税前利润=支付利息前利润-财务费用",
			"所得税=税前利润*25%，按 5 年补亏规则，超过初始权益的部分才交税",
			"年度净利润=税前利润-所得税",
			"利润留存=去年利润留存+去年年度净利",
			"资产总计=流动资产合计+固定资产合计",
			"负债和所有者权益总计=负债合计+所有者权益合计",
		},
	}
}

func AnnualOperationYears(years int) []AnnualOperationYear {
	if years <= 0 {
		years = CompetitionYears
	}
	flow := OperationFlow()
	quarters := []AnnualQuarterFlow{
		{Quarter: 1, Steps: quarterSteps(flow, 1)},
		{Quarter: 2, Steps: quarterSteps(flow, 2)},
		{Quarter: 3, Steps: quarterSteps(flow, 3)},
		{Quarter: 4, Steps: quarterSteps(flow, 4)},
	}
	items := make([]AnnualOperationYear, 0, years)
	for year := 1; year <= years; year++ {
		items = append(items, AnnualOperationYear{
			Year:        year,
			Quarters:    quarters,
			ClosingForm: AnnualReportTemplate(),
		})
	}
	return items
}

func quarterSteps(flow []OperationStep, quarter int) []OperationStep {
	items := make([]OperationStep, 0, len(flow))
	for _, step := range flow {
		if includeStepInQuarter(step, quarter) {
			items = append(items, step)
		}
	}
	return items
}

func includeStepInQuarter(step OperationStep, quarter int) bool {
	if step.OnlyQ4 {
		return quarter == 4
	}
	switch step.Stage {
	case "年初":
		return quarter == 1
	case "第四季度", "年末":
		return quarter == 4
	default:
		return true
	}
}

func FormatAnnualFlowControl(year, quarter int) string {
	if year <= 0 {
		year = 1
	}
	if quarter <= 0 {
		quarter = 1
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("沙盘按年度推进：当前为 Y%dQ%d；每一年都必须从年初融资/广告/选单开始，逐季完成采购、生产、交付、收款，Q4 完成市场/ISO、折旧、违约罚款和结账。\n", year, quarter))
	b.WriteString("年度控制要求：\n")
	b.WriteString("- 所有决策必须带 year/quarter，不能把跨年度决策混在同一轮里判断。\n")
	b.WriteString("- 贷款额度按上年所有者权益*3 控制；长期贷款年初付息，短期贷款按季度到期还本付息。\n")
	b.WriteString("- 市场开拓和 ISO 认证只能在第四季度投资；产品研发按季度投入，不能集中加速。\n")
	b.WriteString("- 订单必须在当年规定季或提前交付，年末未完成订单按销售额 20% 罚款并记入综合费用表“其他”。\n")
	b.WriteString("- 每年结束必须生成综合费用表、利润表、资产负债表，下一年经营以这些报表作为期初依据。\n")
	return b.String()
}

func FormatAnnualReportTemplate() string {
	spec := AnnualReportTemplate()
	var b strings.Builder
	b.WriteString("年度结账必须按 PDF 原表填写三张表：\n")
	b.WriteString("- 综合费用表字段：" + strings.Join(spec.ComprehensiveExpense, "、") + "\n")
	b.WriteString("- 利润表字段：" + strings.Join(spec.Profit, "、") + "\n")
	b.WriteString("- 资产负债表资产端字段：" + strings.Join(spec.BalanceAssets, "、") + "\n")
	b.WriteString("- 资产负债表负债/权益端字段：" + strings.Join(spec.BalanceLiabilities, "、") + "\n")
	b.WriteString("关键公式：\n")
	for _, formula := range spec.Formulas {
		b.WriteString("- " + formula + "\n")
	}
	return b.String()
}

func FormatAnnualLedger(records []AnnualOperationRecord, currentDecisions []DecisionInput, state CompanyState) string {
	var b strings.Builder
	b.WriteString(FormatAnnualFlowControl(state.Year, state.Quarter))
	b.WriteByte('\n')
	b.WriteString(FormatAnnualReportTemplate())
	b.WriteByte('\n')
	b.WriteString(FormatOperationHistory(records, AnalyzeOperationHistory(records)))
	b.WriteByte('\n')
	b.WriteString("从第一年至当前输入的所有决策：\n")
	allDecisions := collectAnnualDecisions(records, currentDecisions)
	if len(allDecisions) == 0 {
		b.WriteString("暂无已登记决策。\n")
		return b.String()
	}
	for _, decision := range allDecisions {
		b.WriteString(formatDecision(decision))
		b.WriteByte('\n')
	}
	return b.String()
}

func collectAnnualDecisions(records []AnnualOperationRecord, currentDecisions []DecisionInput) []DecisionInput {
	items := make([]DecisionInput, 0, len(currentDecisions))
	normalized := make([]AnnualOperationRecord, 0, len(records))
	for _, record := range records {
		NormalizeAnnualRecord(&record)
		normalized = append(normalized, record)
	}
	SortAnnualRecords(normalized)
	for _, record := range normalized {
		for _, decision := range record.Decisions {
			if decision.Year == 0 {
				decision.Year = record.Year
			}
			items = append(items, decision)
		}
	}
	items = append(items, currentDecisions...)
	return items
}
