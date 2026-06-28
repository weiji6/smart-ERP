package erp

import "fmt"

// Money 表示 ERP 沙盘中的金额，单位为 W（万元）。
type Money int64

type Product string

const (
	ProductP1 Product = "P1"
	ProductP2 Product = "P2"
	ProductP3 Product = "P3"
	ProductP4 Product = "P4"
)

type Material string

const (
	MaterialR1 Material = "R1"
	MaterialR2 Material = "R2"
	MaterialR3 Material = "R3"
	MaterialR4 Material = "R4"
)

type Market string

const (
	MarketLocal         Market = "本地"
	MarketRegional      Market = "区域"
	MarketDomestic      Market = "国内"
	MarketAsia          Market = "亚洲"
	MarketInternational Market = "国际"
)

type LineType string

const (
	LineManual  LineType = "手工线"
	LineAuto    LineType = "自动线"
	LineFlex    LineType = "柔性线"
	LineLeasing LineType = "租赁线"
)

type WarningLevel string

const (
	WarningSafe     WarningLevel = "safe"
	WarningWarning  WarningLevel = "warning"
	WarningDanger   WarningLevel = "danger"
	WarningCritical WarningLevel = "critical"
)

type ProductRule struct {
	Product          Product          `json:"product"`
	DevFeePerQuarter Money            `json:"devFeePerQuarter"`
	DevTotal         Money            `json:"devTotal"`
	DevQuarters      int              `json:"devQuarters"`
	ProcessingFee    Money            `json:"processingFee"`
	DirectCost       Money            `json:"directCost"`
	BOM              map[Material]int `json:"bom"`
}

type MaterialRule struct {
	Material Material `json:"material"`
	Price    Money    `json:"price"`
	LeadTime int      `json:"leadTime"`
}

type LineRule struct {
	Type               LineType `json:"type"`
	PurchaseCost       Money    `json:"purchaseCost"`
	InstallQuarters    int      `json:"installQuarters"`
	ProductionCycle    int      `json:"productionCycle"`
	SwitchCost         Money    `json:"switchCost"`
	SwitchQuarters     int      `json:"switchQuarters"`
	AnnualMaintenance  Money    `json:"annualMaintenance"`
	ResidualValue      Money    `json:"residualValue"`
	AnnualDepreciation Money    `json:"annualDepreciation"`
}

type FactoryType string

const (
	FactoryLarge FactoryType = "大厂房"
	FactorySmall FactoryType = "小厂房"
)

type FactoryRule struct {
	Type     FactoryType `json:"type"`
	BuyCost  Money       `json:"buyCost"`
	Rent     Money       `json:"rent"`
	SellCash Money       `json:"sellCash"`
	Capacity int         `json:"capacity"`
}

type FactoryState struct {
	ID                string      `json:"id"`
	Type              FactoryType `json:"type"`
	Ownership         string      `json:"ownership"`
	Rented            bool        `json:"rented"`
	Purchased         bool        `json:"purchased"`
	StartQuarterIndex int         `json:"startQuarterIndex,omitempty"`
}

type RnDProgress struct {
	Product  Product `json:"product"`
	Spent    Money   `json:"spent"`
	Quarters int     `json:"quarters"`
	Finished bool    `json:"finished"`
}

type MarketProgress struct {
	Market Market `json:"market"`
	Spent  Money  `json:"spent"`
	Years  int    `json:"years"`
	Opened bool   `json:"opened"`
}

type ISOProgress struct {
	Type     string `json:"type"`
	Spent    Money  `json:"spent"`
	Years    int    `json:"years"`
	Finished bool   `json:"finished"`
}

type ProductionLine struct {
	ID                       string   `json:"id"`
	FactoryID                string   `json:"factoryId,omitempty"`
	Type                     LineType `json:"type"`
	Product                  Product  `json:"product"`
	Built                    bool     `json:"built"`
	Status                   string   `json:"status,omitempty"`
	BuildStarted             int      `json:"buildStartedQuarterIndex"`
	BuiltAt                  int      `json:"builtAtQuarterIndex"`
	NetValue                 Money    `json:"netValue"`
	IsProducing              bool     `json:"isProducing,omitempty"`
	ProductionStarted        int      `json:"productionStartedQuarterIndex,omitempty"`
	ProductionProgress       int      `json:"productionProgress,omitempty"`
	ProductionTotal          int      `json:"productionTotal,omitempty"`
	CurrentProductionProduct Product  `json:"currentProductionProduct,omitempty"`
}

type Loan struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Principal         Money  `json:"principal"`
	StartQuarterIndex int    `json:"startQuarterIndex"`
	DueQuarterIndex   int    `json:"dueQuarterIndex"`
}

type Receivable struct {
	Amount          Money `json:"amount"`
	DueQuarterIndex int   `json:"dueQuarterIndex"`
}

type PlannedExpense struct {
	QuarterIndex int    `json:"quarterIndex"`
	Category     string `json:"category"`
	Amount       Money  `json:"amount"`
}

type PlannedIncome struct {
	QuarterIndex int    `json:"quarterIndex"`
	Category     string `json:"category"`
	Amount       Money  `json:"amount"`
}

type CompanyState struct {
	Name              string                    `json:"name"`
	Year              int                       `json:"year"`
	Quarter           int                       `json:"quarter"`
	Cash              Money                     `json:"cash"`
	LastYearEquity    Money                     `json:"lastYearEquity"`
	Equity            Money                     `json:"equity"`
	InitialEquity     Money                     `json:"initialEquity"`
	LongLoans         []Loan                    `json:"longLoans"`
	ShortLoans        []Loan                    `json:"shortLoans"`
	Receivables       []Receivable              `json:"receivables"`
	Factories         []FactoryType             `json:"factories"`
	FactoryStates     []FactoryState            `json:"factoryStates,omitempty"`
	ProductionLines   []ProductionLine          `json:"productionLines"`
	ProductInventory  map[Product]int           `json:"productInventory"`
	MaterialInventory map[Material]int          `json:"materialInventory"`
	RnD               map[Product]RnDProgress   `json:"rnd"`
	Markets           map[Market]MarketProgress `json:"markets"`
	ISO               map[string]ISOProgress    `json:"iso"`
	PlannedExpenses   []PlannedExpense          `json:"plannedExpenses"`
	PlannedIncomes    []PlannedIncome           `json:"plannedIncomes"`
	AdminFee          Money                     `json:"adminFee"`
	InfoFee           Money                     `json:"infoFee"`
}

type Order struct {
	ID              string  `json:"id"`
	Year            int     `json:"year,omitempty"`
	Market          Market  `json:"market"`
	Product         Product `json:"product"`
	Quantity        int     `json:"quantity"`
	TotalPrice      Money   `json:"totalPrice"`
	DeliveryQuarter int     `json:"deliveryQuarter"`
	PaymentPeriod   int     `json:"paymentPeriod"`
	ISORequired     string  `json:"isoRequired,omitempty"`
	DeliveryTime    string  `json:"deliveryTime,omitempty"`
	CollectionTime  string  `json:"collectionTime,omitempty"`
}

func (s CompanyState) CurrentQuarterIndex() int {
	return QuarterIndex(s.Year, s.Quarter)
}

func QuarterIndex(year, quarter int) int {
	return (year-1)*4 + quarter
}

func YearQuarter(index int) (int, int) {
	if index <= 0 {
		return 1, 1
	}
	return (index-1)/4 + 1, (index-1)%4 + 1
}

func ValidateYearQuarter(year, quarter int) error {
	if year <= 0 {
		return fmt.Errorf("年份必须大于 0")
	}
	if quarter < 1 || quarter > 4 {
		return fmt.Errorf("季度必须在 1 到 4 之间")
	}
	return nil
}
