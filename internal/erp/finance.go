package erp

import (
	"fmt"
	"sort"
)

type LoanCapacity struct {
	LastYearEquity Money `json:"lastYearEquity"`
	Limit          Money `json:"limit"`
	Used           Money `json:"used"`
	Available      Money `json:"available"`
}

type CashFlowForecast struct {
	Year          int          `json:"year"`
	Quarter       int          `json:"quarter"`
	PredictedCash Money        `json:"predictedCash"`
	Inflows       Money        `json:"inflows"`
	Outflows      Money        `json:"outflows"`
	NetFlow       Money        `json:"netFlow"`
	WarningLevel  WarningLevel `json:"warningLevel"`
	Suggestion    string       `json:"suggestion,omitempty"`
}

type ProfitStatement struct {
	SalesRevenue             Money `json:"salesRevenue"`
	DirectCost               Money `json:"directCost"`
	GrossProfit              Money `json:"grossProfit"`
	ComprehensiveExpense     Money `json:"comprehensiveExpense"`
	ProfitBeforeDepreciation Money `json:"profitBeforeDepreciation"`
	Depreciation             Money `json:"depreciation"`
	ProfitBeforeInterest     Money `json:"profitBeforeInterest"`
	FinancialExpense         Money `json:"financialExpense"`
	ProfitBeforeTax          Money `json:"profitBeforeTax"`
	IncomeTax                Money `json:"incomeTax"`
	NetProfit                Money `json:"netProfit"`
}

type BalanceSheet struct {
	Cash            Money `json:"cash"`
	Receivables     Money `json:"receivables"`
	Products        Money `json:"products"`
	Materials       Money `json:"materials"`
	CurrentAssets   Money `json:"currentAssets"`
	Factories       Money `json:"factories"`
	ProductionLines Money `json:"productionLines"`
	FixedAssets     Money `json:"fixedAssets"`
	TotalAssets     Money `json:"totalAssets"`
	LongLiability   Money `json:"longLiability"`
	ShortLiability  Money `json:"shortLiability"`
	TotalLiability  Money `json:"totalLiability"`
	OwnerEquity     Money `json:"ownerEquity"`
}

func CalcLoanCapacity(state CompanyState) LoanCapacity {
	limit := state.LastYearEquity * LoanLimitMultiple
	used := TotalLoanPrincipal(state)
	return LoanCapacity{
		LastYearEquity: state.LastYearEquity,
		Limit:          limit,
		Used:           used,
		Available:      maxMoney(0, limit-used),
	}
}

func TotalLoanPrincipal(state CompanyState) Money {
	var total Money
	for _, loan := range state.LongLoans {
		total += loan.Principal
	}
	for _, loan := range state.ShortLoans {
		total += loan.Principal
	}
	return total
}

func LongLoanAnnualInterest(state CompanyState) Money {
	var principal Money
	for _, loan := range state.LongLoans {
		principal += loan.Principal
	}
	return RoundPercent(principal, LongLoanAnnualRatePermille)
}

func ShortLoanInterest(principal Money) Money {
	return RoundPercent(principal, ShortLoanAnnualRatePermille)
}

func DiscountFee(amount Money, paymentPeriod int) Money {
	if paymentPeriod <= 2 {
		return CeilPercent(amount, DiscountRateShortPermille)
	}
	return CeilPercent(amount, DiscountRateLongPermille)
}

func RawMaterialAuctionCash(qty int) Money {
	if qty <= 0 {
		return 0
	}
	return FloorPercent(Money(qty), 800)
}

func ProductAuctionCash(product Product, qty int) Money {
	if qty <= 0 {
		return 0
	}
	rule, ok := ProductRules[product]
	if !ok {
		return 0
	}
	return Money(qty) * rule.DirectCost
}

func DefaultCompanyState() CompanyState {
	return CompanyState{
		Name:              "默认企业",
		Year:              1,
		Quarter:           1,
		Cash:              InitialCash,
		LastYearEquity:    InitialCash,
		Equity:            InitialCash,
		InitialEquity:     InitialCash,
		ProductInventory:  map[Product]int{},
		MaterialInventory: map[Material]int{},
		RnD:               map[Product]RnDProgress{},
		Markets:           map[Market]MarketProgress{},
		ISO:               map[string]ISOProgress{},
		AdminFee:          DefaultAdminFee,
		InfoFee:           DefaultInfoFee,
	}
}

func PredictCashFlow(state CompanyState, quarters int) []CashFlowForecast {
	if quarters <= 0 {
		return nil
	}
	current := state.Cash
	start := state.CurrentQuarterIndex()
	forecasts := make([]CashFlowForecast, 0, quarters)

	for i := 0; i < quarters; i++ {
		index := start + i
		year, quarter := YearQuarter(index)
		inflows := plannedIncome(state, index) + collectReceivables(state, index)
		outflows := plannedExpense(state, index) + recurringQuarterOutflow(state, index) + loanOutflow(state, index)
		net := inflows - outflows
		current += net

		level := CashWarningLevel(current)
		forecasts = append(forecasts, CashFlowForecast{
			Year:          year,
			Quarter:       quarter,
			PredictedCash: current,
			Inflows:       inflows,
			Outflows:      outflows,
			NetFlow:       net,
			WarningLevel:  level,
			Suggestion:    CashSuggestion(year, quarter, current, level),
		})
	}
	return forecasts
}

func CashWarningLevel(cash Money) WarningLevel {
	switch {
	case cash <= 0:
		return WarningCritical
	case cash <= 10:
		return WarningDanger
	case cash <= 30:
		return WarningWarning
	default:
		return WarningSafe
	}
}

func CashSuggestion(year, quarter int, cash Money, level WarningLevel) string {
	switch level {
	case WarningWarning:
		return fmt.Sprintf("Y%dQ%d 现金预计为 %dW，建议压缩广告、研发或市场开拓支出。", year, quarter, cash)
	case WarningDanger:
		need := DefaultSafeCash - cash
		return fmt.Sprintf("Y%dQ%d 现金预计仅 %dW，建议至少补充短贷 %dW 并复核订单回款。", year, quarter, cash, maxMoney(10, need))
	case WarningCritical:
		need := DefaultSafeCash - cash
		return fmt.Sprintf("Y%dQ%d 现金预计为 %dW，存在破产风险，必须融资、贴现或削减支出，建议补足 %dW。", year, quarter, cash, need)
	default:
		return ""
	}
}

func recurringQuarterOutflow(state CompanyState, index int) Money {
	_, quarter := YearQuarter(index)
	out := state.AdminFee

	if quarter == 4 {
		out += state.InfoFee
		for _, f := range state.Factories {
			out += FactoryRules[f].Rent
		}
		for _, line := range state.ProductionLines {
			if line.Built {
				out += LineRules[line.Type].AnnualMaintenance
			}
		}
	}
	return out
}

func loanOutflow(state CompanyState, index int) Money {
	_, quarter := YearQuarter(index)
	var out Money
	if quarter == 1 {
		out += LongLoanAnnualInterest(state)
	}
	for _, loan := range state.LongLoans {
		if loan.DueQuarterIndex == index {
			out += loan.Principal
		}
	}
	for _, loan := range state.ShortLoans {
		if loan.DueQuarterIndex == index {
			out += loan.Principal + ShortLoanInterest(loan.Principal)
		}
	}
	return out
}

func collectReceivables(state CompanyState, index int) Money {
	var total Money
	for _, r := range state.Receivables {
		if r.DueQuarterIndex == index {
			total += r.Amount
		}
	}
	return total
}

func plannedExpense(state CompanyState, index int) Money {
	var total Money
	for _, e := range state.PlannedExpenses {
		if e.QuarterIndex == index {
			total += e.Amount
		}
	}
	return total
}

func plannedIncome(state CompanyState, index int) Money {
	var total Money
	for _, income := range state.PlannedIncomes {
		if income.QuarterIndex == index {
			total += income.Amount
		}
	}
	return total
}

func CalcAnnualDepreciation(state CompanyState, year int) Money {
	var total Money
	yearEnd := QuarterIndex(year, 4)
	for _, line := range state.ProductionLines {
		rule := LineRules[line.Type]
		if !line.Built || line.Type == LineLeasing {
			continue
		}
		builtYear, _ := YearQuarter(line.BuiltAt)
		if builtYear >= year || line.BuiltAt > yearEnd {
			continue
		}
		if line.NetValue <= rule.ResidualValue {
			continue
		}
		total += minMoney(rule.AnnualDepreciation, line.NetValue-rule.ResidualValue)
	}
	return total
}

func CalcProfitStatement(orders []Order, comprehensiveExpense, depreciation, financialExpense Money) ProfitStatement {
	var revenue, directCost Money
	for _, order := range orders {
		revenue += order.TotalPrice
		if rule, ok := ProductRules[order.Product]; ok {
			directCost += Money(order.Quantity) * rule.DirectCost
		}
	}
	gross := revenue - directCost
	beforeDep := gross - comprehensiveExpense
	beforeInterest := beforeDep - depreciation
	beforeTax := beforeInterest - financialExpense
	tax := Money(0)
	if beforeTax > 0 {
		tax = RoundPercent(beforeTax, IncomeTaxRatePermille)
	}
	return ProfitStatement{
		SalesRevenue:             revenue,
		DirectCost:               directCost,
		GrossProfit:              gross,
		ComprehensiveExpense:     comprehensiveExpense,
		ProfitBeforeDepreciation: beforeDep,
		Depreciation:             depreciation,
		ProfitBeforeInterest:     beforeInterest,
		FinancialExpense:         financialExpense,
		ProfitBeforeTax:          beforeTax,
		IncomeTax:                tax,
		NetProfit:                beforeTax - tax,
	}
}

func CalcBalanceSheet(state CompanyState) BalanceSheet {
	receivables := Money(0)
	for _, r := range state.Receivables {
		receivables += r.Amount
	}
	products := Money(0)
	for p, qty := range state.ProductInventory {
		products += Money(qty) * ProductRules[p].DirectCost
	}
	materials := Money(0)
	for m, qty := range state.MaterialInventory {
		materials += Money(qty) * MaterialRules[m].Price
	}
	factories := Money(0)
	for _, f := range state.Factories {
		factories += FactoryRules[f].BuyCost
	}
	lines := Money(0)
	for _, line := range state.ProductionLines {
		lines += maxMoney(0, line.NetValue)
	}
	currentAssets := state.Cash + receivables + products + materials
	fixedAssets := factories + lines
	longLiability := Money(0)
	for _, loan := range state.LongLoans {
		longLiability += loan.Principal
	}
	shortLiability := Money(0)
	for _, loan := range state.ShortLoans {
		shortLiability += loan.Principal
	}
	totalAssets := currentAssets + fixedAssets
	totalLiability := longLiability + shortLiability
	return BalanceSheet{
		Cash:            state.Cash,
		Receivables:     receivables,
		Products:        products,
		Materials:       materials,
		CurrentAssets:   currentAssets,
		Factories:       factories,
		ProductionLines: lines,
		FixedAssets:     fixedAssets,
		TotalAssets:     totalAssets,
		LongLiability:   longLiability,
		ShortLiability:  shortLiability,
		TotalLiability:  totalLiability,
		OwnerEquity:     totalAssets - totalLiability,
	}
}

func SortForecasts(items []CashFlowForecast) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Year == items[j].Year {
			return items[i].Quarter < items[j].Quarter
		}
		return items[i].Year < items[j].Year
	})
}
