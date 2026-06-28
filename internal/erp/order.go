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

	rule := ProductRules[order.Product]
	directCost := Money(order.Quantity) * rule.DirectCost
	profit := order.TotalPrice - directCost
	margin := float64(profit) / float64(order.TotalPrice)
	deliveryScore, feasible := deliveryFeasibilityScore(state, order)
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
