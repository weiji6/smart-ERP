package erp

import "sort"

type MaterialRequirement struct {
	Material Material `json:"material"`
	Quantity int      `json:"quantity"`
	Cost     Money    `json:"cost"`
}

type PurchasePlanItem struct {
	Material      Material `json:"material"`
	Quantity      int      `json:"quantity"`
	Cost          Money    `json:"cost"`
	OrderQuarter  int      `json:"orderQuarterIndex"`
	NeedQuarter   int      `json:"needQuarterIndex"`
	LeadTime      int      `json:"leadTime"`
	EmergencyCost Money    `json:"emergencyCost"`
}

type CapacityItem struct {
	Product      Product `json:"product"`
	QuarterIndex int     `json:"quarterIndex"`
	Quantity     int     `json:"quantity"`
}

func CalcMaterialRequirements(orders []Order) []MaterialRequirement {
	total := map[Material]int{
		MaterialR1: 0,
		MaterialR2: 0,
		MaterialR3: 0,
		MaterialR4: 0,
	}
	for _, order := range orders {
		rule, ok := ProductRules[order.Product]
		if !ok {
			continue
		}
		for material, qty := range rule.BOM {
			total[material] += qty * order.Quantity
		}
	}

	items := make([]MaterialRequirement, 0, len(total))
	for material, qty := range total {
		rule := MaterialRules[material]
		items = append(items, MaterialRequirement{
			Material: material,
			Quantity: qty,
			Cost:     Money(qty) * rule.Price,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Material < items[j].Material })
	return items
}

func GeneratePurchasePlan(orders []Order, inventory map[Material]int, currentQuarterIndex int) []PurchasePlanItem {
	needByMaterialAndQuarter := map[Material]map[int]int{}
	for _, order := range orders {
		rule, ok := ProductRules[order.Product]
		if !ok {
			continue
		}
		needQuarter := currentQuarterIndex + maxInt(0, order.DeliveryQuarter-currentQuarter(currentQuarterIndex))
		for material, qty := range rule.BOM {
			if _, ok := needByMaterialAndQuarter[material]; !ok {
				needByMaterialAndQuarter[material] = map[int]int{}
			}
			needByMaterialAndQuarter[material][needQuarter] += qty * order.Quantity
		}
	}

	var plan []PurchasePlanItem
	for material, byQuarter := range needByMaterialAndQuarter {
		quarters := make([]int, 0, len(byQuarter))
		for q := range byQuarter {
			quarters = append(quarters, q)
		}
		sort.Ints(quarters)
		stock := inventory[material]
		rule := MaterialRules[material]
		for _, needQuarter := range quarters {
			needed := byQuarter[needQuarter]
			if stock >= needed {
				stock -= needed
				continue
			}
			deficit := needed - stock
			stock = 0
			orderQuarter := needQuarter - rule.LeadTime
			item := PurchasePlanItem{
				Material:      material,
				Quantity:      deficit,
				Cost:          Money(deficit) * rule.Price,
				OrderQuarter:  orderQuarter,
				NeedQuarter:   needQuarter,
				LeadTime:      rule.LeadTime,
				EmergencyCost: Money(deficit) * rule.Price * 2,
			}
			if orderQuarter < currentQuarterIndex {
				item.OrderQuarter = currentQuarterIndex
			}
			plan = append(plan, item)
		}
	}
	sort.Slice(plan, func(i, j int) bool {
		if plan[i].OrderQuarter == plan[j].OrderQuarter {
			return plan[i].Material < plan[j].Material
		}
		return plan[i].OrderQuarter < plan[j].OrderQuarter
	})
	return plan
}

func CalcMaxCapacity(lines []ProductionLine, startQuarterIndex, quarters int) []CapacityItem {
	if quarters <= 0 {
		return nil
	}
	capacity := map[Product]map[int]int{}
	for _, line := range lines {
		rule, ok := LineRules[line.Type]
		if !ok || rule.ProductionCycle <= 0 {
			continue
		}
		firstAvailable := startQuarterIndex
		if !line.Built {
			firstAvailable = line.BuildStarted + rule.InstallQuarters + 1
		}
		for q := startQuarterIndex; q < startQuarterIndex+quarters; q++ {
			if q < firstAvailable {
				continue
			}
			if (q-firstAvailable+1)%rule.ProductionCycle != 0 {
				continue
			}
			if _, ok := capacity[line.Product]; !ok {
				capacity[line.Product] = map[int]int{}
			}
			capacity[line.Product][q]++
		}
	}

	var items []CapacityItem
	for product, byQuarter := range capacity {
		for quarter, qty := range byQuarter {
			items = append(items, CapacityItem{Product: product, QuarterIndex: quarter, Quantity: qty})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].QuarterIndex == items[j].QuarterIndex {
			return items[i].Product < items[j].Product
		}
		return items[i].QuarterIndex < items[j].QuarterIndex
	})
	return items
}

func currentQuarter(index int) int {
	_, q := YearQuarter(index)
	return q
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
