package erp

import "testing"

func TestScoreOrderRejectsUnreadyProduct(t *testing.T) {
	state := DefaultCompanyState()
	state.Markets[MarketLocal] = MarketProgress{Market: MarketLocal, Opened: true}

	got := ScoreOrder(state, Order{
		ID: "O1", Market: MarketLocal, Product: ProductP2, Quantity: 1,
		TotalPrice: 10, DeliveryQuarter: 2,
	})
	if got.Feasible {
		t.Fatalf("P2 未研发完成时不应可选")
	}
}

func TestScoreOrderAcceptsFeasibleOrder(t *testing.T) {
	state := DefaultCompanyState()
	state.RnD[ProductP2] = RnDProgress{Product: ProductP2, Finished: true}
	state.Markets[MarketLocal] = MarketProgress{Market: MarketLocal, Opened: true}
	state.ProductionLines = []ProductionLine{{Type: LineAuto, Product: ProductP2, Built: true}}

	got := ScoreOrder(state, Order{
		ID: "O1", Market: MarketLocal, Product: ProductP2, Quantity: 1,
		TotalPrice: 10, DeliveryQuarter: 1,
	})
	if !got.Feasible {
		t.Fatalf("订单应可行: %+v", got)
	}
	if got.Profit != 7 {
		t.Fatalf("P2 直接成本 3W，利润应为 7W，got %dW", got.Profit)
	}
}
