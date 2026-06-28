package erp

import (
	"fmt"
	"sort"
	"strings"
)

// AnnualOperationRecord 记录单个经营年度的关键运营数据，金额单位均为 W（万元）。
type AnnualOperationRecord struct {
	Company string `json:"company"`
	Year    int    `json:"year"`

	OpeningCash    Money `json:"openingCash"`
	ClosingCash    Money `json:"closingCash"`
	OpeningEquity  Money `json:"openingEquity"`
	ClosingEquity  Money `json:"closingEquity"`
	SalesRevenue   Money `json:"salesRevenue"`
	NetProfit      Money `json:"netProfit"`
	IncomeTax      Money `json:"incomeTax"`
	TotalAssets    Money `json:"totalAssets"`
	TotalLiability Money `json:"totalLiability"`

	AdExpense            Money `json:"adExpense"`
	ComprehensiveExpense Money `json:"comprehensiveExpense"`
	RnDExpense           Money `json:"rndExpense"`
	MarketExpense        Money `json:"marketExpense"`
	ISOExpense           Money `json:"isoExpense"`
	MaintenanceExpense   Money `json:"maintenanceExpense"`
	FinancialExpense     Money `json:"financialExpense"`
	DefaultPenalty       Money `json:"defaultPenalty"`

	LongLoanAdded    Money `json:"longLoanAdded"`
	ShortLoanAdded   Money `json:"shortLoanAdded"`
	LongLoanBalance  Money `json:"longLoanBalance"`
	ShortLoanBalance Money `json:"shortLoanBalance"`

	Orders            []Order                   `json:"orders,omitempty"`
	Decisions         []DecisionInput           `json:"decisions,omitempty"`
	ProductInventory  map[Product]int           `json:"productInventory,omitempty"`
	MaterialInventory map[Material]int          `json:"materialInventory,omitempty"`
	RnD               map[Product]RnDProgress   `json:"rnd,omitempty"`
	Markets           map[Market]MarketProgress `json:"markets,omitempty"`
	ISO               map[string]ISOProgress    `json:"iso,omitempty"`
	ProductionLines   []ProductionLine          `json:"productionLines,omitempty"`

	KeyEvents string `json:"keyEvents,omitempty"`
	Review    string `json:"review,omitempty"`
}

type OperationHistoryAnalysis struct {
	Years             int      `json:"years"`
	LatestYear        int      `json:"latestYear"`
	TotalSalesRevenue Money    `json:"totalSalesRevenue"`
	TotalNetProfit    Money    `json:"totalNetProfit"`
	TotalAdExpense    Money    `json:"totalAdExpense"`
	AverageAdROI      float64  `json:"averageAdROI"`
	EquityTrend       string   `json:"equityTrend"`
	CashTrend         string   `json:"cashTrend"`
	DebtPressure      string   `json:"debtPressure"`
	Findings          []string `json:"findings"`
}

func NormalizeAnnualRecord(record *AnnualOperationRecord) {
	if record.Company == "" {
		record.Company = "默认企业"
	}
	if record.ProductInventory == nil {
		record.ProductInventory = map[Product]int{}
	}
	if record.MaterialInventory == nil {
		record.MaterialInventory = map[Material]int{}
	}
	if record.RnD == nil {
		record.RnD = map[Product]RnDProgress{}
	}
	if record.Markets == nil {
		record.Markets = map[Market]MarketProgress{}
	}
	if record.ISO == nil {
		record.ISO = map[string]ISOProgress{}
	}
}

func SortAnnualRecords(records []AnnualOperationRecord) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].Company == records[j].Company {
			return records[i].Year < records[j].Year
		}
		return records[i].Company < records[j].Company
	})
}

func AnalyzeOperationHistory(records []AnnualOperationRecord) OperationHistoryAnalysis {
	if len(records) == 0 {
		return OperationHistoryAnalysis{}
	}

	items := make([]AnnualOperationRecord, 0, len(records))
	for _, record := range records {
		NormalizeAnnualRecord(&record)
		items = append(items, record)
	}
	SortAnnualRecords(items)

	var totalRevenue, totalProfit, totalAd, maxLiability Money
	findings := make([]string, 0, 6)
	for _, record := range items {
		liability := annualRecordLiability(record)
		totalRevenue += record.SalesRevenue
		totalProfit += record.NetProfit
		totalAd += record.AdExpense
		if liability > maxLiability {
			maxLiability = liability
		}
		if record.ClosingCash <= 10 {
			findings = append(findings, fmt.Sprintf("Y%d 年末现金 %dW，安全垫不足。", record.Year, record.ClosingCash))
		}
		if record.NetProfit < 0 {
			findings = append(findings, fmt.Sprintf("Y%d 净利润 %dW，需复盘毛利、广告和综合费用。", record.Year, record.NetProfit))
		}
		if record.DefaultPenalty > 0 {
			findings = append(findings, fmt.Sprintf("Y%d 发生违约罚款 %dW，后续接单必须先校验产能和原料提前期。", record.Year, record.DefaultPenalty))
		}
		if record.AdExpense > 0 && record.SalesRevenue < record.AdExpense*3 {
			findings = append(findings, fmt.Sprintf("Y%d 广告投入 %dW、销售额 %dW，广告转化偏低。", record.Year, record.AdExpense, record.SalesRevenue))
		}
	}

	first := items[0]
	last := items[len(items)-1]
	analysis := OperationHistoryAnalysis{
		Years:             len(items),
		LatestYear:        last.Year,
		TotalSalesRevenue: totalRevenue,
		TotalNetProfit:    totalProfit,
		TotalAdExpense:    totalAd,
		EquityTrend:       trendLabel(first.ClosingEquity, last.ClosingEquity),
		CashTrend:         trendLabel(first.ClosingCash, last.ClosingCash),
		DebtPressure:      debtPressureLabel(annualRecordLiability(last), last.ClosingEquity, maxLiability),
		Findings:          dedupeStrings(findings),
	}
	if totalAd > 0 {
		analysis.AverageAdROI = float64(totalRevenue) / float64(totalAd)
	}
	return analysis
}

func FormatOperationHistory(records []AnnualOperationRecord, analysis OperationHistoryAnalysis) string {
	if len(records) == 0 {
		return "暂无年度运营记录。\n"
	}

	items := make([]AnnualOperationRecord, 0, len(records))
	for _, record := range records {
		NormalizeAnnualRecord(&record)
		items = append(items, record)
	}
	SortAnnualRecords(items)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("已记录 %d 个经营年度，最近年份 Y%d；权益趋势：%s；现金趋势：%s；债务压力：%s。\n",
		analysis.Years, analysis.LatestYear, analysis.EquityTrend, analysis.CashTrend, analysis.DebtPressure))
	if analysis.TotalAdExpense > 0 {
		b.WriteString(fmt.Sprintf("累计销售额 %dW，累计净利润 %dW，累计广告 %dW，平均广告销售转化 %.2f。\n",
			analysis.TotalSalesRevenue, analysis.TotalNetProfit, analysis.TotalAdExpense, analysis.AverageAdROI))
	}
	for _, record := range items {
		b.WriteString(fmt.Sprintf("- Y%d：期初现金%dW，年末现金%dW，年末权益%dW，销售%dW，净利%dW，广告%dW，综合费%dW，长贷余额%dW，短贷余额%dW。",
			record.Year, record.OpeningCash, record.ClosingCash, record.ClosingEquity, record.SalesRevenue, record.NetProfit,
			record.AdExpense, record.ComprehensiveExpense, record.LongLoanBalance, record.ShortLoanBalance))
		if len(record.Orders) > 0 {
			b.WriteString(fmt.Sprintf("订单%d张。", len(record.Orders)))
		}
		if record.KeyEvents != "" {
			b.WriteString("关键事件：" + record.KeyEvents + "。")
		}
		if record.Review != "" {
			b.WriteString("复盘：" + record.Review + "。")
		}
		b.WriteByte('\n')
	}
	if len(analysis.Findings) > 0 {
		b.WriteString("历史风险信号：\n")
		for _, finding := range analysis.Findings {
			b.WriteString("- " + finding + "\n")
		}
	}
	return b.String()
}

func trendLabel(first, last Money) string {
	switch {
	case last > first:
		return "改善"
	case last < first:
		return "下滑"
	default:
		return "持平"
	}
}

func debtPressureLabel(latestLiability, latestEquity, maxLiability Money) string {
	if latestLiability <= 0 && maxLiability <= 0 {
		return "无明显债务压力"
	}
	if latestEquity <= 0 {
		return "极高"
	}
	ratio := float64(latestLiability) / float64(latestEquity)
	switch {
	case ratio >= 2:
		return "高"
	case ratio >= 1:
		return "中"
	default:
		return "低"
	}
}

func annualRecordLiability(record AnnualOperationRecord) Money {
	if record.TotalLiability > 0 {
		return record.TotalLiability
	}
	return record.LongLoanBalance + record.ShortLoanBalance
}
