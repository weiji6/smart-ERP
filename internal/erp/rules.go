package erp

var ProductRules = map[Product]ProductRule{
	ProductP1: {
		Product: ProductP1, DevFeePerQuarter: 1, DevTotal: 2, DevQuarters: 2,
		ProcessingFee: 1, DirectCost: 2, BOM: map[Material]int{MaterialR1: 1},
	},
	ProductP2: {
		Product: ProductP2, DevFeePerQuarter: 1, DevTotal: 3, DevQuarters: 3,
		ProcessingFee: 1, DirectCost: 3, BOM: map[Material]int{MaterialR2: 1, MaterialR3: 1},
	},
	ProductP3: {
		Product: ProductP3, DevFeePerQuarter: 1, DevTotal: 4, DevQuarters: 4,
		ProcessingFee: 1, DirectCost: 4, BOM: map[Material]int{MaterialR1: 1, MaterialR3: 1, MaterialR4: 1},
	},
	ProductP4: {
		Product: ProductP4, DevFeePerQuarter: 1, DevTotal: 6, DevQuarters: 6,
		ProcessingFee: 1, DirectCost: 5, BOM: map[Material]int{MaterialR2: 1, MaterialR3: 1, MaterialR4: 2},
	},
}

var MaterialRules = map[Material]MaterialRule{
	MaterialR1: {Material: MaterialR1, Price: 1, LeadTime: 1},
	MaterialR2: {Material: MaterialR2, Price: 1, LeadTime: 1},
	MaterialR3: {Material: MaterialR3, Price: 1, LeadTime: 2},
	MaterialR4: {Material: MaterialR4, Price: 1, LeadTime: 2},
}

var LineRules = map[LineType]LineRule{
	LineManual: {
		Type: LineManual, PurchaseCost: 5, InstallQuarters: 0, ProductionCycle: 2,
		SwitchCost: 0, SwitchQuarters: 0, AnnualMaintenance: 1, ResidualValue: 1, AnnualDepreciation: 1,
	},
	LineAuto: {
		Type: LineAuto, PurchaseCost: 15, InstallQuarters: 3, ProductionCycle: 1,
		SwitchCost: 2, SwitchQuarters: 1, AnnualMaintenance: 2, ResidualValue: 3, AnnualDepreciation: 3,
	},
	LineFlex: {
		Type: LineFlex, PurchaseCost: 20, InstallQuarters: 4, ProductionCycle: 1,
		SwitchCost: 0, SwitchQuarters: 0, AnnualMaintenance: 2, ResidualValue: 4, AnnualDepreciation: 4,
	},
	LineLeasing: {
		Type: LineLeasing, PurchaseCost: 0, InstallQuarters: 0, ProductionCycle: 1,
		SwitchCost: 2, SwitchQuarters: 1, AnnualMaintenance: 6, ResidualValue: -6, AnnualDepreciation: 0,
	},
}

var FactoryRules = map[FactoryType]FactoryRule{
	FactoryLarge: {Type: FactoryLarge, BuyCost: 40, Rent: 5, SellCash: 40, Capacity: 6},
	FactorySmall: {Type: FactorySmall, BuyCost: 30, Rent: 3, SellCash: 30, Capacity: 4},
}

var MarketOpenRules = map[Market]struct {
	FeePerYear Money `json:"feePerYear"`
	Years      int   `json:"years"`
	TotalFee   Money `json:"totalFee"`
}{
	MarketLocal:         {FeePerYear: 1, Years: 1, TotalFee: 1},
	MarketRegional:      {FeePerYear: 1, Years: 1, TotalFee: 1},
	MarketDomestic:      {FeePerYear: 1, Years: 2, TotalFee: 2},
	MarketAsia:          {FeePerYear: 1, Years: 3, TotalFee: 3},
	MarketInternational: {FeePerYear: 1, Years: 4, TotalFee: 4},
}

var ISORules = map[string]struct {
	FeePerYear Money `json:"feePerYear"`
	Years      int   `json:"years"`
	TotalFee   Money `json:"totalFee"`
}{
	"ISO9000":  {FeePerYear: 1, Years: 2, TotalFee: 2},
	"ISO14000": {FeePerYear: 2, Years: 2, TotalFee: 4},
}

const (
	InitialCash                 Money = 60
	LoanLimitMultiple                 = 3
	LongLoanAnnualRatePermille        = 100
	ShortLoanAnnualRatePermille       = 50
	DiscountRateShortPermille         = 100
	DiscountRateLongPermille          = 125
	IncomeTaxRatePermille             = 250
	DefaultAdminFee             Money = 1
	DefaultInfoFee              Money = 1
	MinSingleAd                 Money = 1
	DefaultSafeCash             Money = 20
)
