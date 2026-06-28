package erp

import (
	"fmt"
	"strings"
)

type MarketPlanRecommendation struct {
	Relevant       bool             `json:"relevant"`
	Items          []MarketPlanItem `json:"items,omitempty"`
	Reason         string           `json:"reason,omitempty"`
	MissingContext []string         `json:"missingContext,omitempty"`
}

type MarketPlanItem struct {
	Market        Market `json:"market"`
	Action        string `json:"action"`
	StartYear     int    `json:"startYear"`
	StartQuarter  int    `json:"startQuarter"`
	FeeThisYear   Money  `json:"feeThisYear"`
	TotalFee      Money  `json:"totalFee"`
	YearsRequired int    `json:"yearsRequired"`
	ReadyYear     int    `json:"readyYear"`
	Priority      int    `json:"priority"`
	Reason        string `json:"reason"`
}

func BuildMarketPlanRecommendation(ctx AdvisorContext, cashFlow []CashFlowForecast) MarketPlanRecommendation {
	if !isMarketQuestion(ctx.Question) {
		return MarketPlanRecommendation{}
	}

	items := make([]MarketPlanItem, 0, 3)
	for _, market := range []Market{MarketRegional, MarketDomestic, MarketAsia, MarketInternational} {
		progress := ctx.State.Markets[market]
		rule := MarketOpenRules[market]
		if progress.Opened {
			continue
		}
		remainingYears := rule.Years - progress.Years
		if remainingYears <= 0 {
			remainingYears = 1
		}
		startYear, startQuarter := nextMarketInvestmentQuarter(ctx.State.Year, ctx.State.Quarter)
		readyYear := startYear + remainingYears
		items = append(items, MarketPlanItem{
			Market:        market,
			Action:        marketAction(progress),
			StartYear:     startYear,
			StartQuarter:  startQuarter,
			FeeThisYear:   rule.FeePerYear,
			TotalFee:      rule.TotalFee,
			YearsRequired: remainingYears,
			ReadyYear:     readyYear,
			Priority:      marketPriority(ctx.State, market, readyYear),
			Reason:        marketReason(ctx.State, market, readyYear),
		})
	}

	sortMarketPlanItems(items)
	limit := 3
	if len(items) < limit {
		limit = len(items)
	}
	items = items[:limit]

	reason := "市场开拓只允许第四季度操作，建议按投入周期和产品成熟度排序，优先开周期短、能承接当前产品订单的市场。"
	if ctx.State.Quarter != 4 {
		reason = fmt.Sprintf("当前是Y%dQ%d，不能立即投资新市场；建议先确定目标市场，在Y%dQ4执行开拓投入。",
			ctx.State.Year, ctx.State.Quarter, ctx.State.Year)
	}

	return MarketPlanRecommendation{
		Relevant:       true,
		Items:          items,
		Reason:         reason,
		MissingContext: marketMissingContext(ctx),
	}
}

func FormatMarketPlan(plan MarketPlanRecommendation) string {
	if !plan.Relevant {
		return ""
	}
	if len(plan.Items) == 0 {
		return "市场开拓专项建议：当前已无明显待开拓市场，建议把重点放在订单利润、广告效率和产能交付。\n"
	}

	var b strings.Builder
	b.WriteString("市场开拓专项建议：" + plan.Reason + "\n")
	b.WriteString("具体开拓表：\n")
	for _, item := range plan.Items {
		b.WriteString(fmt.Sprintf("- 优先级%d：%s，动作=%s，Y%dQ%d投入%dW，累计总费用%dW，剩余周期%d年，预计Y%d可换证/可参与订单。理由：%s\n",
			item.Priority, item.Market, item.Action, item.StartYear, item.StartQuarter, item.FeeThisYear,
			item.TotalFee, item.YearsRequired, item.ReadyYear, item.Reason))
	}
	if len(plan.MissingContext) > 0 {
		b.WriteString("需要补充后可进一步判断：")
		b.WriteString(strings.Join(plan.MissingContext, "、"))
		b.WriteString("。\n")
	}
	return b.String()
}

func isMarketQuestion(question string) bool {
	question = strings.ToLower(question)
	return strings.Contains(question, "市场") ||
		strings.Contains(question, "开拓") ||
		strings.Contains(question, "开放") ||
		strings.Contains(question, "开不") ||
		strings.Contains(question, "区域") ||
		strings.Contains(question, "国内") ||
		strings.Contains(question, "亚洲") ||
		strings.Contains(question, "国际")
}

func nextMarketInvestmentQuarter(year, quarter int) (int, int) {
	if year <= 0 {
		year = 1
	}
	if quarter <= 4 {
		return year, 4
	}
	return year + 1, 4
}

func marketAction(progress MarketProgress) string {
	if progress.Years > 0 || progress.Spent > 0 {
		return "继续开拓"
	}
	return "新开拓"
}

func marketPriority(state CompanyState, market Market, readyYear int) int {
	switch market {
	case MarketRegional:
		return 1
	case MarketDomestic:
		if hasFinishedProduct(state, ProductP2) || hasFinishedProduct(state, ProductP3) {
			return 2
		}
		return 3
	case MarketAsia:
		if readyYear <= state.Year+2 && hasFinishedProduct(state, ProductP3) {
			return 3
		}
		return 4
	case MarketInternational:
		return 5
	default:
		return 9
	}
}

func marketReason(state CompanyState, market Market, readyYear int) string {
	switch market {
	case MarketRegional:
		return "周期短、费用低，适合作为本地市场之后的第一扩张方向。"
	case MarketDomestic:
		if hasFinishedProduct(state, ProductP2) || hasFinishedProduct(state, ProductP3) {
			return "国内市场周期2年，适合在P2/P3研发完成或接近完成时提前布局。"
		}
		return "国内市场价值高，但要等P2/P3研发和产能跟上，否则开了也难变现。"
	case MarketAsia:
		return fmt.Sprintf("亚洲市场周期较长，预计Y%d才可用，适合现金宽裕且P3/P4路线明确时提前开。", readyYear)
	case MarketInternational:
		return "国际市场周期最长，前期现金和产能压力大，通常不建议早期开，除非比赛策略明确走高端市场。"
	default:
		return "需要结合订单池和产品研发状态判断。"
	}
}

func hasFinishedProduct(state CompanyState, product Product) bool {
	progress, ok := state.RnD[product]
	return ok && progress.Finished
}

func marketMissingContext(ctx AdvisorContext) []string {
	var missing []string
	if len(ctx.Orders) == 0 {
		missing = append(missing, "后续年份订单池/市场需求")
	}
	if len(ctx.State.RnD) == 0 {
		missing = append(missing, "产品研发进度")
	}
	if len(ctx.State.ProductionLines) == 0 {
		missing = append(missing, "未来产能规划")
	}
	return missing
}

func sortMarketPlanItems(items []MarketPlanItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Priority < items[i].Priority {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
