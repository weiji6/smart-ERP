package erp

import (
	"math"
	"sort"
)

type StrategyParams struct {
	Name           string    `json:"name"`
	AdBudget       Money     `json:"adBudget"`
	RnDPriority    []Product `json:"rndPriority"`
	MarketPriority []Market  `json:"marketPriority"`
	LoanMode       string    `json:"loanMode"`
	RiskFactor     float64   `json:"riskFactor"`
}

type StrategyResult struct {
	StrategyName   string  `json:"strategyName"`
	MeanEquity     Money   `json:"meanEquity"`
	MedianEquity   Money   `json:"medianEquity"`
	MinEquity      Money   `json:"minEquity"`
	MaxEquity      Money   `json:"maxEquity"`
	BankruptcyRate float64 `json:"bankruptcyRate"`
	RiskLevel      string  `json:"riskLevel"`
	Suggestion     string  `json:"suggestion"`
}

type StrategyComparison struct {
	Results        []StrategyResult `json:"results"`
	BestStrategy   string           `json:"bestStrategy"`
	Recommendation string           `json:"recommendation"`
}

func DefaultStrategies() []StrategyParams {
	return []StrategyParams{
		{Name: "均衡发展策略", AdBudget: 10, RnDPriority: []Product{ProductP2, ProductP3}, MarketPriority: []Market{MarketRegional, MarketDomestic}, LoanMode: "balanced", RiskFactor: 0.08},
		{Name: "重研发策略", AdBudget: 8, RnDPriority: []Product{ProductP2, ProductP3, ProductP4}, MarketPriority: []Market{MarketDomestic, MarketAsia}, LoanMode: "balanced", RiskFactor: 0.12},
		{Name: "重广告策略", AdBudget: 15, RnDPriority: []Product{ProductP2}, MarketPriority: []Market{MarketLocal, MarketRegional}, LoanMode: "conservative", RiskFactor: 0.18},
		{Name: "重产能策略", AdBudget: 10, RnDPriority: []Product{ProductP2, ProductP3}, MarketPriority: []Market{MarketDomestic}, LoanMode: "aggressive", RiskFactor: 0.22},
	}
}

func CompareStrategies(state CompanyState, strategies []StrategyParams, simulations int) StrategyComparison {
	if len(strategies) == 0 {
		strategies = DefaultStrategies()
	}
	if simulations <= 0 {
		simulations = 100
	}
	results := make([]StrategyResult, 0, len(strategies))
	for _, strategy := range strategies {
		results = append(results, SimulateStrategy(state, strategy, simulations))
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].BankruptcyRate == results[j].BankruptcyRate {
			return results[i].MeanEquity > results[j].MeanEquity
		}
		return results[i].BankruptcyRate < results[j].BankruptcyRate
	})
	best := ""
	recommendation := "暂无可用策略。"
	if len(results) > 0 {
		best = results[0].StrategyName
		recommendation = "推荐 " + best + "：在当前现金与权益约束下，收益和破产风险的组合最优。"
	}
	return StrategyComparison{Results: results, BestStrategy: best, Recommendation: recommendation}
}

func SimulateStrategy(state CompanyState, strategy StrategyParams, simulations int) StrategyResult {
	equities := make([]Money, 0, simulations)
	bankrupt := 0
	base := state.Equity
	if base == 0 {
		base = CalcBalanceSheet(state).OwnerEquity
	}
	for i := 0; i < simulations; i++ {
		noise := deterministicNoise(i, strategy.RiskFactor)
		gain := Money(math.Round(float64(strategy.AdBudget)*2.2 + float64(len(strategy.RnDPriority))*8 + float64(len(strategy.MarketPriority))*5))
		riskCost := Money(math.Round(float64(base) * strategy.RiskFactor * noise))
		loanBoost := loanModeBoost(strategy.LoanMode)
		equity := base + gain + loanBoost - riskCost
		if equity <= 0 || state.Cash-riskCost <= -10 {
			bankrupt++
			continue
		}
		equities = append(equities, equity)
	}
	if len(equities) == 0 {
		return StrategyResult{
			StrategyName: strategy.Name, BankruptcyRate: 1, RiskLevel: "高风险",
			Suggestion: "该策略现金压力过大，不建议执行。",
		}
	}
	sort.Slice(equities, func(i, j int) bool { return equities[i] < equities[j] })
	mean := meanMoney(equities)
	rate := float64(bankrupt) / float64(simulations)
	return StrategyResult{
		StrategyName:   strategy.Name,
		MeanEquity:     mean,
		MedianEquity:   equities[len(equities)/2],
		MinEquity:      equities[0],
		MaxEquity:      equities[len(equities)-1],
		BankruptcyRate: math.Round(rate*10000) / 100,
		RiskLevel:      riskLevel(rate),
		Suggestion:     strategySuggestion(strategy, rate),
	}
}

func deterministicNoise(i int, risk float64) float64 {
	v := float64((i*37)%100) / 100
	return 0.6 + v + risk
}

func loanModeBoost(mode string) Money {
	switch mode {
	case "aggressive":
		return 20
	case "balanced":
		return 10
	case "conservative":
		return 3
	default:
		return 0
	}
}

func meanMoney(items []Money) Money {
	var sum Money
	for _, item := range items {
		sum += item
	}
	return Money(math.Round(float64(sum) / float64(len(items))))
}

func riskLevel(rate float64) string {
	switch {
	case rate >= 0.15:
		return "高风险"
	case rate >= 0.05:
		return "中风险"
	default:
		return "稳健"
	}
}

func strategySuggestion(strategy StrategyParams, bankruptcyRate float64) string {
	if bankruptcyRate >= 0.15 {
		return "破产概率偏高，建议降低广告或产能扩张节奏，并保留至少 20W 现金垫。"
	}
	if len(strategy.RnDPriority) >= 3 {
		return "研发投入较重，适合中后期高毛利路线，注意不要与市场开拓集中挤占现金。"
	}
	return "策略风险可控，建议结合订单毛利和回款账期动态微调广告预算。"
}
