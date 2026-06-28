package erp

import "testing"

func TestPDFRoundingRules(t *testing.T) {
	if got := ShortLoanInterest(21); got != 1 {
		t.Fatalf("短贷利息应四舍五入为 1W，got %dW", got)
	}
	if got := RoundPercent(78, LongLoanAnnualRatePermille); got != 8 {
		t.Fatalf("长贷利息应四舍五入为 8W，got %dW", got)
	}
	if got := DiscountFee(26, 1); got != 3 {
		t.Fatalf("1 账期贴现费应向上取整为 3W，got %dW", got)
	}
	if got := DiscountFee(424, 2); got != 43 {
		t.Fatalf("2 账期贴现费应向上取整为 43W，got %dW", got)
	}
	if got := RawMaterialAuctionCash(2); got != 1 {
		t.Fatalf("2 个原料拍卖应向下取整为 1W，got %dW", got)
	}
}

func TestLoanCapacityUsesTotalLongAndShortLoans(t *testing.T) {
	state := DefaultCompanyState()
	state.LastYearEquity = 40
	state.LongLoans = []Loan{{Principal: 50}}

	got := CalcLoanCapacity(state)
	if got.Limit != 120 || got.Used != 50 || got.Available != 70 {
		t.Fatalf("贷款额度计算错误: %+v", got)
	}
}

func TestPredictCashFlowWarnsCritical(t *testing.T) {
	state := DefaultCompanyState()
	state.Cash = 5
	state.PlannedExpenses = []PlannedExpense{{QuarterIndex: QuarterIndex(1, 1), Category: "广告费", Amount: 10}}

	got := PredictCashFlow(state, 1)
	if len(got) != 1 {
		t.Fatalf("预测结果数量错误: %d", len(got))
	}
	if got[0].WarningLevel != WarningCritical {
		t.Fatalf("预警等级应为 critical，got %s", got[0].WarningLevel)
	}
}

func TestBalanceSheetCalculatesOwnerEquity(t *testing.T) {
	state := DefaultCompanyState()
	state.Cash = 30
	state.Receivables = []Receivable{{Amount: 20, DueQuarterIndex: QuarterIndex(1, 2)}}
	state.MaterialInventory = map[Material]int{MaterialR1: 3}
	state.ProductInventory = map[Product]int{ProductP2: 2}
	state.LongLoans = []Loan{{Principal: 10}}

	got := CalcBalanceSheet(state)
	if got.OwnerEquity != 49 {
		t.Fatalf("权益应为 49W，got %dW", got.OwnerEquity)
	}
}
