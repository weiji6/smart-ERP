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

	ComprehensiveExpenseReport ComprehensiveExpenseReport `json:"综合费用表"`
	ProfitReport               ProfitReport               `json:"利润表"`
	BalanceSheetReport         BalanceSheetReport         `json:"资产负债表"`

	OpeningCash              Money `json:"-"`
	ClosingCash              Money `json:"-"`
	OpeningEquity            Money `json:"-"`
	ClosingEquity            Money `json:"-"`
	SalesRevenue             Money `json:"-"`
	DirectCost               Money `json:"-"`
	GrossProfit              Money `json:"-"`
	NetProfit                Money `json:"-"`
	IncomeTax                Money `json:"-"`
	ProfitBeforeDepreciation Money `json:"-"`
	Depreciation             Money `json:"-"`
	ProfitBeforeInterest     Money `json:"-"`
	ProfitBeforeTax          Money `json:"-"`
	TotalAssets              Money `json:"-"`
	TotalLiability           Money `json:"-"`

	AdExpense            Money `json:"-"`
	ComprehensiveExpense Money `json:"-"`
	ManagementExpense    Money `json:"-"`
	RnDExpense           Money `json:"-"`
	MarketExpense        Money `json:"-"`
	ISOExpense           Money `json:"-"`
	MaintenanceExpense   Money `json:"-"`
	SwitchExpense        Money `json:"-"`
	RentExpense          Money `json:"-"`
	InfoExpense          Money `json:"-"`
	OtherLoss            Money `json:"-"`
	FinancialExpense     Money `json:"-"`
	DefaultPenalty       Money `json:"-"`

	LongLoanAdded    Money `json:"-"`
	ShortLoanAdded   Money `json:"-"`
	LongLoanBalance  Money `json:"-"`
	ShortLoanBalance Money `json:"-"`
	IncomeTaxPayable Money `json:"-"`
	ShareCapital     Money `json:"-"`
	RetainedEarnings Money `json:"-"`
	OwnerEquityTotal Money `json:"-"`

	Receivables            Money `json:"-"`
	WorkInProgress         Money `json:"-"`
	FinishedGoods          Money `json:"-"`
	RawMaterials           Money `json:"-"`
	CurrentAssets          Money `json:"-"`
	FactoryValue           Money `json:"-"`
	ProductionLineValue    Money `json:"-"`
	ConstructionInProgress Money `json:"-"`
	FixedAssets            Money `json:"-"`

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

type ComprehensiveExpenseReport struct {
	ManagementFee           Money `json:"管理费"`
	AdvertisingFee          Money `json:"广告费"`
	EquipmentMaintenanceFee Money `json:"设备维护费"`
	TransferFee             Money `json:"转产费"`
	Rent                    Money `json:"租金"`
	MarketAccessDevelopment Money `json:"市场准入开拓"`
	ProductDevelopment      Money `json:"产品研发"`
	ISOCertification        Money `json:"ISO 认证资格"`
	InformationFee          Money `json:"信息费"`
	Other                   Money `json:"其他"`
	Total                   Money `json:"合计"`
}

type ProfitReport struct {
	SalesRevenue             Money `json:"销售收入"`
	DirectCost               Money `json:"直接成本"`
	GrossProfit              Money `json:"毛利"`
	ComprehensiveExpense     Money `json:"综合费用"`
	ProfitBeforeDepreciation Money `json:"折旧前利润"`
	Depreciation             Money `json:"折旧"`
	ProfitBeforeInterest     Money `json:"支付利息前利润"`
	FinancialExpense         Money `json:"财务费用"`
	ProfitBeforeTax          Money `json:"税前利润"`
	IncomeTax                Money `json:"所得税"`
	AnnualNetProfit          Money `json:"年度净利润"`
}

type BalanceSheetReport struct {
	Cash                    Money `json:"现金"`
	Receivable              Money `json:"应收款"`
	WorkInProgress          Money `json:"在制品"`
	FinishedGoods           Money `json:"产成品"`
	RawMaterials            Money `json:"原材料"`
	CurrentAssetsTotal      Money `json:"流动资产合计"`
	Factory                 Money `json:"厂房"`
	ProductionLine          Money `json:"生产线"`
	ConstructionInProgress  Money `json:"在建工程"`
	FixedAssetsTotal        Money `json:"固定资产合计"`
	AssetsTotal             Money `json:"资产总计"`
	LongTermLiability       Money `json:"长期负债"`
	ShortTermLiability      Money `json:"短期负债"`
	IncomeTaxPayable        Money `json:"应交所得税"`
	LiabilityTotal          Money `json:"负债合计"`
	ShareCapital            Money `json:"股东资本"`
	RetainedEarnings        Money `json:"利润留存"`
	AnnualNetProfit         Money `json:"年度净利"`
	OwnerEquityTotal        Money `json:"所有者权益合计"`
	LiabilityAndEquityTotal Money `json:"负债和所有者权益总计"`
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
	normalizeAnnualReportFields(record)
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

func normalizeAnnualReportFields(record *AnnualOperationRecord) {
	if record.ComprehensiveExpenseReport == (ComprehensiveExpenseReport{}) {
		record.ComprehensiveExpenseReport = ComprehensiveExpenseReport{
			ManagementFee:           record.ManagementExpense,
			AdvertisingFee:          record.AdExpense,
			EquipmentMaintenanceFee: record.MaintenanceExpense,
			TransferFee:             record.SwitchExpense,
			Rent:                    record.RentExpense,
			MarketAccessDevelopment: record.MarketExpense,
			ProductDevelopment:      record.RnDExpense,
			ISOCertification:        record.ISOExpense,
			InformationFee:          record.InfoExpense,
			Other:                   record.OtherLoss,
			Total:                   record.ComprehensiveExpense,
		}
	}
	if record.ProfitReport == (ProfitReport{}) {
		record.ProfitReport = ProfitReport{
			SalesRevenue:             record.SalesRevenue,
			DirectCost:               record.DirectCost,
			GrossProfit:              record.GrossProfit,
			ComprehensiveExpense:     record.ComprehensiveExpense,
			ProfitBeforeDepreciation: record.ProfitBeforeDepreciation,
			Depreciation:             record.Depreciation,
			ProfitBeforeInterest:     record.ProfitBeforeInterest,
			FinancialExpense:         record.FinancialExpense,
			ProfitBeforeTax:          record.ProfitBeforeTax,
			IncomeTax:                record.IncomeTax,
			AnnualNetProfit:          record.NetProfit,
		}
	}
	if record.BalanceSheetReport == (BalanceSheetReport{}) {
		ownerEquityTotal := record.OwnerEquityTotal
		if ownerEquityTotal == 0 {
			ownerEquityTotal = record.ClosingEquity
		}
		liabilityTotal := record.TotalLiability
		if liabilityTotal == 0 {
			liabilityTotal = record.LongLoanBalance + record.ShortLoanBalance
		}
		record.BalanceSheetReport = BalanceSheetReport{
			Cash:                    record.ClosingCash,
			Receivable:              record.Receivables,
			WorkInProgress:          record.WorkInProgress,
			FinishedGoods:           record.FinishedGoods,
			RawMaterials:            record.RawMaterials,
			CurrentAssetsTotal:      record.CurrentAssets,
			Factory:                 record.FactoryValue,
			ProductionLine:          record.ProductionLineValue,
			ConstructionInProgress:  record.ConstructionInProgress,
			FixedAssetsTotal:        record.FixedAssets,
			AssetsTotal:             record.TotalAssets,
			LongTermLiability:       record.LongLoanBalance,
			ShortTermLiability:      record.ShortLoanBalance,
			IncomeTaxPayable:        record.IncomeTaxPayable,
			LiabilityTotal:          liabilityTotal,
			ShareCapital:            record.ShareCapital,
			RetainedEarnings:        record.RetainedEarnings,
			AnnualNetProfit:         record.NetProfit,
			OwnerEquityTotal:        ownerEquityTotal,
			LiabilityAndEquityTotal: record.TotalAssets,
		}
	}

	expense := record.ComprehensiveExpenseReport
	record.ManagementExpense = expense.ManagementFee
	record.AdExpense = expense.AdvertisingFee
	record.MaintenanceExpense = expense.EquipmentMaintenanceFee
	record.SwitchExpense = expense.TransferFee
	record.RentExpense = expense.Rent
	record.MarketExpense = expense.MarketAccessDevelopment
	record.RnDExpense = expense.ProductDevelopment
	record.ISOExpense = expense.ISOCertification
	record.InfoExpense = expense.InformationFee
	record.OtherLoss = expense.Other
	record.ComprehensiveExpense = expense.Total

	profit := record.ProfitReport
	record.SalesRevenue = profit.SalesRevenue
	record.DirectCost = profit.DirectCost
	record.GrossProfit = profit.GrossProfit
	record.ProfitBeforeDepreciation = profit.ProfitBeforeDepreciation
	record.Depreciation = profit.Depreciation
	record.ProfitBeforeInterest = profit.ProfitBeforeInterest
	record.FinancialExpense = profit.FinancialExpense
	record.ProfitBeforeTax = profit.ProfitBeforeTax
	record.IncomeTax = profit.IncomeTax
	record.NetProfit = profit.AnnualNetProfit
	if record.ComprehensiveExpense == 0 {
		record.ComprehensiveExpense = profit.ComprehensiveExpense
	}
	if record.ComprehensiveExpenseReport.Total == 0 {
		record.ComprehensiveExpenseReport.Total = profit.ComprehensiveExpense
	}

	balance := record.BalanceSheetReport
	record.ClosingCash = balance.Cash
	record.Receivables = balance.Receivable
	record.WorkInProgress = balance.WorkInProgress
	record.FinishedGoods = balance.FinishedGoods
	record.RawMaterials = balance.RawMaterials
	record.CurrentAssets = balance.CurrentAssetsTotal
	record.FactoryValue = balance.Factory
	record.ProductionLineValue = balance.ProductionLine
	record.ConstructionInProgress = balance.ConstructionInProgress
	record.FixedAssets = balance.FixedAssetsTotal
	record.TotalAssets = balance.AssetsTotal
	record.LongLoanBalance = balance.LongTermLiability
	record.ShortLoanBalance = balance.ShortTermLiability
	record.IncomeTaxPayable = balance.IncomeTaxPayable
	record.TotalLiability = balance.LiabilityTotal
	record.ShareCapital = balance.ShareCapital
	record.RetainedEarnings = balance.RetainedEarnings
	record.OwnerEquityTotal = balance.OwnerEquityTotal
	if balance.OwnerEquityTotal != 0 {
		record.ClosingEquity = balance.OwnerEquityTotal
	}
	if record.NetProfit == 0 {
		record.NetProfit = balance.AnnualNetProfit
	}
	if record.ProfitReport.AnnualNetProfit == 0 {
		record.ProfitReport.AnnualNetProfit = balance.AnnualNetProfit
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
		b.WriteString(fmt.Sprintf("累计销售额 %dW，累计净利润 %dW，累计广告费 %dW，平均广告销售转化 %.2f。\n",
			analysis.TotalSalesRevenue, analysis.TotalNetProfit, analysis.TotalAdExpense, analysis.AverageAdROI))
	}
	for _, record := range items {
		b.WriteString(formatAnnualRecordDetail(record))
		if record.KeyEvents != "" {
			b.WriteString("关键事件：" + record.KeyEvents + "\n")
		}
		if record.Review != "" {
			b.WriteString("年度复盘：" + record.Review + "\n")
		}
	}
	if len(analysis.Findings) > 0 {
		b.WriteString("历史风险信号：\n")
		for _, finding := range analysis.Findings {
			b.WriteString("- " + finding + "\n")
		}
	}
	return b.String()
}

func formatAnnualRecordDetail(record AnnualOperationRecord) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nY%d 年度账本：\n", record.Year))
	b.WriteString(fmt.Sprintf("- 摘要：年末现金 %dW，所有者权益 %dW，销售收入 %dW，年度净利润 %dW，长贷 %dW，短贷 %dW。\n",
		record.ClosingCash, record.OwnerEquityTotal, record.SalesRevenue, record.NetProfit, record.LongLoanBalance, record.ShortLoanBalance))
	b.WriteString("- 综合费用表：")
	writeMoneyPairs(&b, []moneyPair{
		{"管理费", record.ComprehensiveExpenseReport.ManagementFee},
		{"广告费", record.ComprehensiveExpenseReport.AdvertisingFee},
		{"设备维护费", record.ComprehensiveExpenseReport.EquipmentMaintenanceFee},
		{"转产费", record.ComprehensiveExpenseReport.TransferFee},
		{"租金", record.ComprehensiveExpenseReport.Rent},
		{"市场准入开拓", record.ComprehensiveExpenseReport.MarketAccessDevelopment},
		{"产品研发", record.ComprehensiveExpenseReport.ProductDevelopment},
		{"ISO 认证资格", record.ComprehensiveExpenseReport.ISOCertification},
		{"信息费", record.ComprehensiveExpenseReport.InformationFee},
		{"其他", record.ComprehensiveExpenseReport.Other},
		{"合计", record.ComprehensiveExpenseReport.Total},
	})
	b.WriteByte('\n')
	b.WriteString("- 利润表：")
	writeMoneyPairs(&b, []moneyPair{
		{"销售收入", record.ProfitReport.SalesRevenue},
		{"直接成本", record.ProfitReport.DirectCost},
		{"毛利", record.ProfitReport.GrossProfit},
		{"综合费用", record.ProfitReport.ComprehensiveExpense},
		{"折旧前利润", record.ProfitReport.ProfitBeforeDepreciation},
		{"折旧", record.ProfitReport.Depreciation},
		{"支付利息前利润", record.ProfitReport.ProfitBeforeInterest},
		{"财务费用", record.ProfitReport.FinancialExpense},
		{"税前利润", record.ProfitReport.ProfitBeforeTax},
		{"所得税", record.ProfitReport.IncomeTax},
		{"年度净利润", record.ProfitReport.AnnualNetProfit},
	})
	b.WriteByte('\n')
	b.WriteString("- 资产负债表：")
	writeMoneyPairs(&b, []moneyPair{
		{"现金", record.BalanceSheetReport.Cash},
		{"应收款", record.BalanceSheetReport.Receivable},
		{"在制品", record.BalanceSheetReport.WorkInProgress},
		{"产成品", record.BalanceSheetReport.FinishedGoods},
		{"原材料", record.BalanceSheetReport.RawMaterials},
		{"流动资产合计", record.BalanceSheetReport.CurrentAssetsTotal},
		{"厂房", record.BalanceSheetReport.Factory},
		{"生产线", record.BalanceSheetReport.ProductionLine},
		{"在建工程", record.BalanceSheetReport.ConstructionInProgress},
		{"固定资产合计", record.BalanceSheetReport.FixedAssetsTotal},
		{"资产总计", record.BalanceSheetReport.AssetsTotal},
		{"长期负债", record.BalanceSheetReport.LongTermLiability},
		{"短期负债", record.BalanceSheetReport.ShortTermLiability},
		{"应交所得税", record.BalanceSheetReport.IncomeTaxPayable},
		{"负债合计", record.BalanceSheetReport.LiabilityTotal},
		{"股东资本", record.BalanceSheetReport.ShareCapital},
		{"利润留存", record.BalanceSheetReport.RetainedEarnings},
		{"年度净利", record.BalanceSheetReport.AnnualNetProfit},
		{"所有者权益合计", record.BalanceSheetReport.OwnerEquityTotal},
		{"负债和所有者权益总计", record.BalanceSheetReport.LiabilityAndEquityTotal},
	})
	b.WriteByte('\n')
	if len(record.Decisions) > 0 {
		b.WriteString("- 本年决策：\n")
		for _, decision := range record.Decisions {
			b.WriteString("  ")
			b.WriteString(formatDecision(decision))
			b.WriteByte('\n')
		}
	}
	if len(record.Orders) > 0 {
		b.WriteString("- 本年订单：\n")
		for _, order := range record.Orders {
			b.WriteString(fmt.Sprintf("  - 订单=%s，市场=%s，产品=%s，数量=%d，总价=%dW，交付=Q%d，账期=%d。\n",
				order.ID, order.Market, order.Product, order.Quantity, order.TotalPrice, order.DeliveryQuarter, order.PaymentPeriod))
		}
	}
	return b.String()
}

type moneyPair struct {
	name  string
	value Money
}

func writeMoneyPairs(b *strings.Builder, pairs []moneyPair) {
	for i, pair := range pairs {
		if i > 0 {
			b.WriteString("；")
		}
		b.WriteString(fmt.Sprintf("%s=%dW", pair.name, pair.value))
	}
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
