package erp

import (
	"math"
	"sort"
)

type OrderScore struct {
	Order         Order   `json:"order"`
	Score         float64 `json:"score"`
	Feasible      bool    `json:"feasible"`
	Reason        string  `json:"reason,omitempty"`
	Profit        Money   `json:"profit"`
	ProfitMargin  float64 `json:"profitMargin"`
	PaymentScore  float64 `json:"paymentScore"`
	DeliveryScore float64 `json:"deliveryScore"`
	CashFlowScore float64 `json:"cashFlowScore"`
	CapacityScore float64 `json:"capacityScore"`
}

func ScoreOrders(state CompanyState, orders []Order) []OrderScore {
	scores := make([]OrderScore, 0, len(orders))
	for _, order := range orders {
		scores = append(scores, ScoreOrder(state, order))
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	return scores
}

func ScoreOrder(state CompanyState, order Order) OrderScore {
	if order.Quantity <= 0 || order.TotalPrice <= 0 {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: "订单数量和金额必须大于 0"}
	}
	if !isProductReady(state, order.Product) {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: "产品研发未完成，不能承接该订单"}
	}
	if !isMarketOpened(state, order.Market) {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: "市场尚未开拓完成，不能承接该订单"}
	}
	if order.ISORequired != "" && !isISOReady(state, order.ISORequired) {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: "ISO 认证未完成，不能承接该订单"}
	}
	if reason := factoryCapacityIssue(state); reason != "" {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: reason}
	}

	rule := ProductRules[order.Product]
	directCost := Money(order.Quantity) * rule.DirectCost
	profit := order.TotalPrice - directCost
	margin := float64(profit) / float64(order.TotalPrice)
	deliveryState := state
	deliveryState.ProductionLines = usableProductionLines(state)
	deliveryScore, feasible := deliveryFeasibilityScore(deliveryState, order)
	if !feasible {
		return OrderScore{Order: order, Score: -1, Feasible: false, Reason: "现有产能无法在交期前完成"}
	}
	paymentScore := math.Max(0, 100-float64(order.PaymentPeriod)*20)
	capacityScore := deliveryScore
	cashScore := cashFlowImpactScore(order, directCost)
	score := margin*100*0.30 + deliveryScore*0.25 + paymentScore*0.20 + capacityScore*0.15 + cashScore*0.10

	return OrderScore{
		Order:         order,
		Score:         math.Round(score*100) / 100,
		Feasible:      true,
		Profit:        profit,
		ProfitMargin:  math.Round(margin*10000) / 100,
		PaymentScore:  paymentScore,
		DeliveryScore: deliveryScore,
		CashFlowScore: cashScore,
		CapacityScore: capacityScore,
	}
}

func factoryCapacityIssue(state CompanyState) string {
	if len(state.ProductionLines) == 0 {
		return "没有可用生产线，无法承接需要生产的订单"
	}
	if len(state.FactoryStates) == 0 && len(state.Factories) == 0 {
		return ""
	}
	capacity := totalFactoryCapacity(state)
	if capacity <= 0 {
		return "没有可用厂房，无法承载生产线"
	}
	if len(state.ProductionLines) > capacity {
		return "生产线数量超过厂房容量，请先处理厂房容量或生产线布局"
	}
	if len(state.FactoryStates) == 0 {
		return ""
	}
	validFactories := activeFactoryIDs(state)
	for _, line := range state.ProductionLines {
		if line.FactoryID == "" {
			return "存在未分配厂房的生产线，请先补充生产线所在厂房"
		}
		if !validFactories[line.FactoryID] {
			return "存在挂靠在不可用厂房的生产线，请先修正厂房状态"
		}
	}
	return ""
}

func usableProductionLines(state CompanyState) []ProductionLine {
	if len(state.FactoryStates) == 0 {
		return state.ProductionLines
	}
	validFactories := activeFactoryIDs(state)
	lines := make([]ProductionLine, 0, len(state.ProductionLines))
	for _, line := range state.ProductionLines {
		if line.FactoryID == "" || !validFactories[line.FactoryID] {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func totalFactoryCapacity(state CompanyState) int {
	if len(state.FactoryStates) > 0 {
		total := 0
		for _, factory := range state.FactoryStates {
			if !isFactoryActive(factory) {
				continue
			}
			total += FactoryRules[factory.Type].Capacity
		}
		return total
	}
	total := 0
	for _, factoryType := range state.Factories {
		total += FactoryRules[factoryType].Capacity
	}
	return total
}

func activeFactoryIDs(state CompanyState) map[string]bool {
	ids := make(map[string]bool, len(state.FactoryStates))
	for _, factory := range state.FactoryStates {
		if isFactoryActive(factory) {
			ids[factory.ID] = true
		}
	}
	return ids
}

func isFactoryActive(factory FactoryState) bool {
	if factory.Type == "" {
		return false
	}
	if factory.Purchased || factory.Rented {
		return true
	}
	switch factory.Ownership {
	case "购买", "租用", "purchased", "rented":
		return true
	default:
		return false
	}
}

func deliveryFeasibilityScore(state CompanyState, order Order) (float64, bool) {
	start := state.CurrentQuarterIndex()
	deadline := QuarterIndex(state.Year, order.DeliveryQuarter)
	if deadline < start {
		deadline = QuarterIndex(state.Year+1, order.DeliveryQuarter)
	}
	quarters := deadline - start + 1
	if quarters <= 0 {
		return 0, false
	}
	capacity := 0
	for _, item := range CalcMaxCapacity(state.ProductionLines, start, quarters) {
		if item.Product == order.Product {
			capacity += item.Quantity
		}
	}
	capacity += state.ProductInventory[order.Product]
	if capacity < order.Quantity {
		return 0, false
	}
	slack := capacity - order.Quantity
	return math.Min(100, 60+float64(slack)*10), true
}

func cashFlowImpactScore(order Order, directCost Money) float64 {
	if order.TotalPrice <= 0 {
		return 0
	}
	profit := order.TotalPrice - directCost
	score := 50 + float64(profit)*2 - float64(order.PaymentPeriod)*10
	return math.Max(0, math.Min(100, score))
}

func isProductReady(state CompanyState, product Product) bool {
	progress, ok := state.RnD[product]
	return ok && progress.Finished
}

func isMarketOpened(state CompanyState, market Market) bool {
	progress, ok := state.Markets[market]
	return ok && progress.Opened
}

func isISOReady(state CompanyState, iso string) bool {
	progress, ok := state.ISO[iso]
	return ok && progress.Finished
}
