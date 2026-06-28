package erp

import (
	"fmt"
	"sort"
	"strings"
)

type AdPlanRecommendation struct {
	Relevant       bool         `json:"relevant"`
	Budget         Money        `json:"budget"`
	SafeCashAfter  Money        `json:"safeCashAfter"`
	Items          []AdPlanItem `json:"items,omitempty"`
	Reason         string       `json:"reason,omitempty"`
	MissingContext []string     `json:"missingContext,omitempty"`
}

type AdPlanItem struct {
	Market           Market  `json:"market"`
	Product          Product `json:"product"`
	AdCost           Money   `json:"adCost"`
	SelectionChances int     `json:"selectionChances"`
	PredictedRanking int     `json:"predictedRanking"`
	OrderAmount      Money   `json:"orderAmount"`
	PredictedProfit  Money   `json:"predictedProfit"`
	ROI              float64 `json:"roi"`
	Reason           string  `json:"reason"`
}

func BuildAdPlanRecommendation(ctx AdvisorContext, cashFlow []CashFlowForecast, orderScores []OrderScore) AdPlanRecommendation {
	if !isAdQuestion(ctx.Question) && !(strings.TrimSpace(ctx.Question) == "" && ctx.StepCode == "ad_investment") {
		return AdPlanRecommendation{}
	}

	budget := adBudgetFromDecisions(ctx.Decisions)
	if budget <= 0 {
		budget = suggestedAdBudget(ctx.State)
	}
	if budget <= 0 {
		return AdPlanRecommendation{
			Relevant:       true,
			Reason:         "当前现金不足以形成有效广告预算，建议先融资或压缩其他支出。",
			MissingContext: []string{"广告总预算", "可选订单池", "竞争对手广告"},
		}
	}

	markets := availableMarkets(ctx.State, ctx.Orders)
	products := availableProducts(ctx.State, ctx.Orders)
	candidates := adPlanCandidates(markets, products, budget, ctx.CompetitorAds)
	items := pickAdPlanItems(candidates, budget, ctx.Orders, orderScores, ctx.CompetitorAds)
	used := Money(0)
	for _, item := range items {
		used += item.AdCost
	}
	missing := adMissingContext(ctx)
	reason := "优先保证至少1W获得选单机会，再把增量预算投向已研发、已开市场、预计利润和ROI更高的组合。"
	if len(ctx.Orders) > 0 {
		reason = "优先围绕当前可交付订单和高毛利产品投放，避免广告拿单后产能或原料跟不上。"
	}

	return AdPlanRecommendation{
		Relevant:       true,
		Budget:         used,
		SafeCashAfter:  ctx.State.Cash - used,
		Items:          items,
		Reason:         reason,
		MissingContext: missing,
	}
}

func FormatAdPlan(plan AdPlanRecommendation) string {
	if !plan.Relevant {
		return ""
	}
	if len(plan.Items) == 0 {
		return "广告专项建议：暂无可执行广告投放表。" + plan.Reason + "\n"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("广告专项建议：建议广告总预算 %dW，投后账面现金约 %dW。%s\n", plan.Budget, plan.SafeCashAfter, plan.Reason))
	b.WriteString("具体投放表：\n")
	for _, item := range plan.Items {
		b.WriteString(fmt.Sprintf("- %s/%s：投%dW，预计选单机会%d次，预计排名第%d，预计销售额%dW，预计利润%dW，ROI %.2f。理由：%s\n",
			item.Market, item.Product, item.AdCost, item.SelectionChances, item.PredictedRanking,
			item.OrderAmount, item.PredictedProfit, item.ROI, item.Reason))
	}
	if len(plan.MissingContext) > 0 {
		b.WriteString("需要补充后可进一步优化：")
		b.WriteString(strings.Join(plan.MissingContext, "、"))
		b.WriteString("。\n")
	}
	return b.String()
}

func isAdQuestion(question string) bool {
	question = strings.ToLower(question)
	return strings.Contains(question, "广告") || strings.Contains(question, "投放") || strings.Contains(question, "选单")
}

func adBudgetFromDecisions(decisions []DecisionInput) Money {
	var budget Money
	for _, decision := range decisions {
		if decision.Type == "ad" {
			budget += decision.Amount
		}
	}
	return budget
}

func suggestedAdBudget(state CompanyState) Money {
	if state.Cash <= DefaultSafeCash {
		return MinSingleAd
	}
	budget := (state.Cash - DefaultSafeCash) / 3
	if budget < MinSingleAd {
		return MinSingleAd
	}
	if budget > 12 {
		return 12
	}
	return budget
}

func availableMarkets(state CompanyState, orders []Order) []Market {
	seen := map[Market]bool{}
	for _, order := range orders {
		if order.Market != "" {
			seen[order.Market] = true
		}
	}
	for market, progress := range state.Markets {
		if progress.Opened {
			seen[market] = true
		}
	}
	if len(seen) == 0 {
		seen[MarketLocal] = true
	}
	return sortedMarkets(seen)
}

func availableProducts(state CompanyState, orders []Order) []Product {
	seen := map[Product]bool{}
	for _, order := range orders {
		if order.Product != "" {
			seen[order.Product] = true
		}
	}
	for product, progress := range state.RnD {
		if progress.Finished {
			seen[product] = true
		}
	}
	if len(seen) == 0 {
		seen[ProductP1] = true
	}
	return sortedProducts(seen)
}

func adPlanCandidates(markets []Market, products []Product, budget Money, competitorAds []Money) []AdPlanItem {
	var items []AdPlanItem
	for _, market := range markets {
		for _, product := range products {
			for ad := Money(1); ad <= budget; ad++ {
				prediction := PredictMarketOrder(market, product, ad, competitorAds)
				items = append(items, AdPlanItem{
					Market:           market,
					Product:          product,
					AdCost:           ad,
					SelectionChances: prediction.SelectionChances,
					PredictedRanking: prediction.PredictedRanking,
					OrderAmount:      prediction.PredictedOrderAmount,
					PredictedProfit:  prediction.PredictedProfit,
					ROI:              prediction.ROI,
					Reason:           "预计ROI较高且满足最低广告额。",
				})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ROI == items[j].ROI {
			if items[i].PredictedProfit == items[j].PredictedProfit {
				return items[i].AdCost < items[j].AdCost
			}
			return items[i].PredictedProfit > items[j].PredictedProfit
		}
		return items[i].ROI > items[j].ROI
	})
	return items
}

func pickAdPlanItems(candidates []AdPlanItem, budget Money, orders []Order, orderScores []OrderScore, competitorAds []Money) []AdPlanItem {
	usedKey := map[string]bool{}
	picked := make([]AdPlanItem, 0, 4)
	sort.SliceStable(candidates, func(i, j int) bool {
		leftPriority := adPlanPriority(candidates[i], orders, orderScores)
		rightPriority := adPlanPriority(candidates[j], orders, orderScores)
		if leftPriority == rightPriority {
			if candidates[i].ROI == candidates[j].ROI {
				return candidates[i].PredictedProfit > candidates[j].PredictedProfit
			}
			return candidates[i].ROI > candidates[j].ROI
		}
		return leftPriority > rightPriority
	})

	for _, candidate := range candidates {
		key := string(candidate.Market) + ":" + string(candidate.Product)
		if usedKey[key] || candidate.AdCost != MinSingleAd {
			continue
		}
		candidate.Reason = adItemReason(candidate, orders, orderScores)
		picked = append(picked, candidate)
		usedKey[key] = true
		if Money(len(picked)) >= budget || len(picked) >= 4 {
			break
		}
	}

	for used := totalAdPlanCost(picked); used < budget && len(picked) > 0; used = totalAdPlanCost(picked) {
		index := bestAdIncrementIndex(picked, budget-used, orders, orderScores, competitorAds)
		if index < 0 {
			break
		}
		picked[index].AdCost++
		prediction := PredictMarketOrder(picked[index].Market, picked[index].Product, picked[index].AdCost, competitorAds)
		picked[index].SelectionChances = prediction.SelectionChances
		picked[index].PredictedRanking = prediction.PredictedRanking
		picked[index].OrderAmount = prediction.PredictedOrderAmount
		picked[index].PredictedProfit = prediction.PredictedProfit
		picked[index].ROI = prediction.ROI
		picked[index].Reason = adItemReason(picked[index], orders, orderScores)
	}

	sort.Slice(picked, func(i, j int) bool {
		if picked[i].AdCost == picked[j].AdCost {
			return adPlanPriority(picked[i], orders, orderScores) > adPlanPriority(picked[j], orders, orderScores)
		}
		return picked[i].AdCost > picked[j].AdCost
	})
	return picked
}

func totalAdPlanCost(items []AdPlanItem) Money {
	var total Money
	for _, item := range items {
		total += item.AdCost
	}
	return total
}

func bestAdIncrementIndex(items []AdPlanItem, remaining Money, orders []Order, orderScores []OrderScore, competitorAds []Money) int {
	if remaining <= 0 {
		return -1
	}
	bestIndex := 0
	bestScore := -1.0
	for i, item := range items {
		nextAd := item.AdCost + 1
		prediction := PredictMarketOrder(item.Market, item.Product, nextAd, competitorAds)
		score := prediction.ROI + float64(prediction.PredictedProfit)/10 + adPlanPriority(item, orders, orderScores)
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex
}

func adPlanPriority(item AdPlanItem, orders []Order, orderScores []OrderScore) float64 {
	score := 0.0
	for _, orderScore := range orderScores {
		if orderScore.Order.Market == item.Market && orderScore.Order.Product == item.Product {
			if orderScore.Feasible {
				score += 20
			} else {
				score += 5
			}
			score += float64(orderScore.Profit) / 5
		}
	}
	for _, order := range orders {
		if order.Market == item.Market && order.Product == item.Product {
			score += 3
		}
	}
	return score
}

func adItemReason(item AdPlanItem, orders []Order, orderScores []OrderScore) string {
	for _, score := range orderScores {
		if score.Order.Market == item.Market && score.Order.Product == item.Product {
			if score.Feasible {
				return fmt.Sprintf("匹配当前可交付订单%s，利润约%dW，适合主投。", score.Order.ID, score.Profit)
			}
			return fmt.Sprintf("匹配订单%s但%s，投放前要先补产能或调整交期。", score.Order.ID, score.Reason)
		}
	}
	for _, order := range orders {
		if order.Market == item.Market && order.Product == item.Product {
			return fmt.Sprintf("匹配当前关注订单%s，但仍需复核产能和原料。", order.ID)
		}
	}
	return "该市场/产品预计毛利较高，可作为补充投放以争取额外选单机会。"
}

func adMissingContext(ctx AdvisorContext) []string {
	var missing []string
	if len(ctx.CompetitorAds) == 0 {
		missing = append(missing, "竞争对手广告")
	}
	if len(ctx.Orders) == 0 {
		missing = append(missing, "可选订单池")
	}
	if len(ctx.State.ProductionLines) == 0 {
		missing = append(missing, "现有生产线和产能")
	}
	if len(ctx.State.MaterialInventory) == 0 {
		missing = append(missing, "原料库存和在途订单")
	}
	return missing
}

func sortedMarkets(markets map[Market]bool) []Market {
	order := []Market{MarketLocal, MarketRegional, MarketDomestic, MarketAsia, MarketInternational}
	out := make([]Market, 0, len(markets))
	for _, market := range order {
		if markets[market] {
			out = append(out, market)
		}
	}
	return out
}

func sortedProducts(products map[Product]bool) []Product {
	order := []Product{ProductP1, ProductP2, ProductP3, ProductP4}
	out := make([]Product, 0, len(products))
	for _, product := range order {
		if products[product] {
			out = append(out, product)
		}
	}
	return out
}
