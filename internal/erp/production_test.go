package erp

import "testing"

func TestMaterialRequirementsFollowPDFBOM(t *testing.T) {
	orders := []Order{
		{Product: ProductP2, Quantity: 10},
		{Product: ProductP4, Quantity: 5},
	}

	got := CalcMaterialRequirements(orders)
	want := map[Material]int{
		MaterialR1: 0,
		MaterialR2: 15,
		MaterialR3: 15,
		MaterialR4: 10,
	}
	for _, item := range got {
		if item.Quantity != want[item.Material] {
			t.Fatalf("%s 需求错误，want %d got %d", item.Material, want[item.Material], item.Quantity)
		}
		if item.Cost != Money(item.Quantity) {
			t.Fatalf("%s 原料单价应按 PDF 为 1W，got cost %d", item.Material, item.Cost)
		}
	}
}

func TestPurchasePlanRespectsLeadTimeAndInventory(t *testing.T) {
	orders := []Order{{Product: ProductP4, Quantity: 2, DeliveryQuarter: 3}}
	inventory := map[Material]int{MaterialR2: 1, MaterialR3: 0, MaterialR4: 1}

	got := GeneratePurchasePlan(orders, inventory, QuarterIndex(1, 1))
	if len(got) != 3 {
		t.Fatalf("采购计划数量错误: %+v", got)
	}

	byMaterial := map[Material]PurchasePlanItem{}
	for _, item := range got {
		byMaterial[item.Material] = item
	}
	if byMaterial[MaterialR2].Quantity != 1 || byMaterial[MaterialR2].OrderQuarter != QuarterIndex(1, 2) {
		t.Fatalf("R2 采购计划错误: %+v", byMaterial[MaterialR2])
	}
	if byMaterial[MaterialR4].Quantity != 3 || byMaterial[MaterialR4].OrderQuarter != QuarterIndex(1, 1) {
		t.Fatalf("R4 采购计划错误: %+v", byMaterial[MaterialR4])
	}
}

func TestCapacityCountsLineCycle(t *testing.T) {
	lines := []ProductionLine{
		{Type: LineManual, Product: ProductP1, Built: true},
		{Type: LineAuto, Product: ProductP2, Built: true},
	}

	got := CalcMaxCapacity(lines, QuarterIndex(1, 1), 4)
	total := map[Product]int{}
	for _, item := range got {
		total[item.Product] += item.Quantity
	}
	if total[ProductP1] != 2 {
		t.Fatalf("手工线一年应产 2 个，got %d", total[ProductP1])
	}
	if total[ProductP2] != 4 {
		t.Fatalf("自动线一年应产 4 个，got %d", total[ProductP2])
	}
}
