package erp

import (
	"math"
	"sort"
)

type MarketPrediction struct {
	Market               Market  `json:"market"`
	Product              Product `json:"product"`
	AdCost               Money   `json:"adCost"`
	PredictedRanking     int     `json:"predictedRanking"`
	PredictedOrderAmount Money   `json:"predictedOrderAmount"`
	PredictedProfit      Money   `json:"predictedProfit"`
	ROI                  float64 `json:"roi"`
	SelectionChances     int     `json:"selectionChances"`
	Confidence           float64 `json:"confidence"`
}

type AdCandidate struct {
	Market      Market  `json:"market"`
	Product     Product `json:"product"`
	AdCost      Money   `json:"adCost"`
	OrderAmount Money   `json:"orderAmount"`
	Profit      Money   `json:"profit"`
	ROI         float64 `json:"roi"`
}

func PredictMarketOrder(market Market, product Product, adCost Money, competitorAds []Money) MarketPrediction {
	selectionChances := AdSelectionChances(adCost)
	ranking := PredictAdRanking(adCost, competitorAds)
	basePrice := EstimatedOrderPrice(market, product)
	amount := basePrice * Money(selectionChances)
	rule := ProductRules[product]
	profit := amount - Money(selectionChances)*rule.DirectCost
	roi := 0.0
	if adCost > 0 {
		roi = float64(profit) / float64(adCost)
	}
	return MarketPrediction{
		Market:               market,
		Product:              product,
		AdCost:               adCost,
		PredictedRanking:     ranking,
		PredictedOrderAmount: amount,
		PredictedProfit:      profit,
		ROI:                  math.Round(roi*100) / 100,
		SelectionChances:     selectionChances,
		Confidence:           0.62,
	}
}

func AdSelectionChances(adCost Money) int {
	if adCost < MinSingleAd {
		return 0
	}
	return 1 + int((adCost-MinSingleAd)/2)
}

func PredictAdRanking(adCost Money, competitorAds []Money) int {
	ranking := 1
	for _, ad := range competitorAds {
		if ad > adCost {
			ranking++
		}
	}
	return ranking
}

func OptimizeAdAllocation(markets []Market, products []Product, budget Money, competitorAds []Money) []AdCandidate {
	if budget <= 0 {
		return nil
	}
	var candidates []AdCandidate
	for _, market := range markets {
		for _, product := range products {
			for ad := Money(1); ad <= budget; ad++ {
				prediction := PredictMarketOrder(market, product, ad, competitorAds)
				candidates = append(candidates, AdCandidate{
					Market:      market,
					Product:     product,
					AdCost:      ad,
					OrderAmount: prediction.PredictedOrderAmount,
					Profit:      prediction.PredictedProfit,
					ROI:         prediction.ROI,
				})
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].ROI == candidates[j].ROI {
			return candidates[i].Profit > candidates[j].Profit
		}
		return candidates[i].ROI > candidates[j].ROI
	})

	var picked []AdCandidate
	used := Money(0)
	usedKey := map[string]bool{}
	for _, c := range candidates {
		key := string(c.Market) + ":" + string(c.Product)
		if usedKey[key] || used+c.AdCost > budget {
			continue
		}
		picked = append(picked, c)
		used += c.AdCost
		usedKey[key] = true
		if used == budget {
			break
		}
	}
	return picked
}

func EstimatedOrderPrice(market Market, product Product) Money {
	base := map[Product]Money{
		ProductP1: 6,
		ProductP2: 9,
		ProductP3: 12,
		ProductP4: 15,
	}[product]
	multiplier := map[Market]Money{
		MarketLocal:         100,
		MarketRegional:      110,
		MarketDomestic:      125,
		MarketAsia:          140,
		MarketInternational: 155,
	}[market]
	return Money(int64(base) * int64(multiplier) / 100)
}
