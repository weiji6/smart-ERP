package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"smarterp/internal/erp"
)

type mysqlOperationRecordStore struct {
	db *sql.DB
}

func newMySQLOperationRecordStore(db *sql.DB) *mysqlOperationRecordStore {
	return &mysqlOperationRecordStore{db: db}
}

func (s *mysqlOperationRecordStore) Upsert(record erp.AnnualOperationRecord) (erp.AnnualOperationRecord, error) {
	erp.NormalizeAnnualRecord(&record)

	comprehensiveExpenseReportJSON, err := marshalJSON(record.ComprehensiveExpenseReport)
	if err != nil {
		return record, err
	}
	profitReportJSON, err := marshalJSON(record.ProfitReport)
	if err != nil {
		return record, err
	}
	balanceSheetReportJSON, err := marshalJSON(record.BalanceSheetReport)
	if err != nil {
		return record, err
	}
	ordersJSON, err := marshalJSON(record.Orders)
	if err != nil {
		return record, err
	}
	decisionsJSON, err := marshalJSON(record.Decisions)
	if err != nil {
		return record, err
	}
	productInventoryJSON, err := marshalJSON(record.ProductInventory)
	if err != nil {
		return record, err
	}
	materialInventoryJSON, err := marshalJSON(record.MaterialInventory)
	if err != nil {
		return record, err
	}
	rndJSON, err := marshalJSON(record.RnD)
	if err != nil {
		return record, err
	}
	marketsJSON, err := marshalJSON(record.Markets)
	if err != nil {
		return record, err
	}
	isoJSON, err := marshalJSON(record.ISO)
	if err != nil {
		return record, err
	}
	productionLinesJSON, err := marshalJSON(record.ProductionLines)
	if err != nil {
		return record, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.db.ExecContext(ctx, `INSERT INTO annual_operation_records (
		company, year, comprehensive_expense_report_json, profit_report_json, balance_sheet_report_json,
		opening_cash, closing_cash, opening_equity, closing_equity,
		sales_revenue, direct_cost, gross_profit, net_profit, income_tax,
		profit_before_depreciation, depreciation, profit_before_interest, profit_before_tax,
		total_assets, total_liability,
		ad_expense, comprehensive_expense, management_expense, rnd_expense, market_expense, iso_expense,
		maintenance_expense, switch_expense, rent_expense, info_expense, other_loss,
		financial_expense, default_penalty,
		long_loan_added, short_loan_added, long_loan_balance, short_loan_balance,
		income_tax_payable, share_capital, retained_earnings, owner_equity_total,
		receivables, work_in_progress, finished_goods, raw_materials, current_assets,
		factory_value, production_line_value, construction_in_progress, fixed_assets,
		orders_json, decisions_json, product_inventory_json, material_inventory_json,
		rnd_json, markets_json, iso_json, production_lines_json, key_events, review
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		comprehensive_expense_report_json = VALUES(comprehensive_expense_report_json),
		profit_report_json = VALUES(profit_report_json),
		balance_sheet_report_json = VALUES(balance_sheet_report_json),
		opening_cash = VALUES(opening_cash),
		closing_cash = VALUES(closing_cash),
		opening_equity = VALUES(opening_equity),
		closing_equity = VALUES(closing_equity),
		sales_revenue = VALUES(sales_revenue),
		direct_cost = VALUES(direct_cost),
		gross_profit = VALUES(gross_profit),
		net_profit = VALUES(net_profit),
		income_tax = VALUES(income_tax),
		profit_before_depreciation = VALUES(profit_before_depreciation),
		depreciation = VALUES(depreciation),
		profit_before_interest = VALUES(profit_before_interest),
		profit_before_tax = VALUES(profit_before_tax),
		total_assets = VALUES(total_assets),
		total_liability = VALUES(total_liability),
		ad_expense = VALUES(ad_expense),
		comprehensive_expense = VALUES(comprehensive_expense),
		management_expense = VALUES(management_expense),
		rnd_expense = VALUES(rnd_expense),
		market_expense = VALUES(market_expense),
		iso_expense = VALUES(iso_expense),
		maintenance_expense = VALUES(maintenance_expense),
		switch_expense = VALUES(switch_expense),
		rent_expense = VALUES(rent_expense),
		info_expense = VALUES(info_expense),
		other_loss = VALUES(other_loss),
		financial_expense = VALUES(financial_expense),
		default_penalty = VALUES(default_penalty),
		long_loan_added = VALUES(long_loan_added),
		short_loan_added = VALUES(short_loan_added),
		long_loan_balance = VALUES(long_loan_balance),
		short_loan_balance = VALUES(short_loan_balance),
		income_tax_payable = VALUES(income_tax_payable),
		share_capital = VALUES(share_capital),
		retained_earnings = VALUES(retained_earnings),
		owner_equity_total = VALUES(owner_equity_total),
		receivables = VALUES(receivables),
		work_in_progress = VALUES(work_in_progress),
		finished_goods = VALUES(finished_goods),
		raw_materials = VALUES(raw_materials),
		current_assets = VALUES(current_assets),
		factory_value = VALUES(factory_value),
		production_line_value = VALUES(production_line_value),
		construction_in_progress = VALUES(construction_in_progress),
		fixed_assets = VALUES(fixed_assets),
		orders_json = VALUES(orders_json),
		decisions_json = VALUES(decisions_json),
		product_inventory_json = VALUES(product_inventory_json),
		material_inventory_json = VALUES(material_inventory_json),
		rnd_json = VALUES(rnd_json),
		markets_json = VALUES(markets_json),
		iso_json = VALUES(iso_json),
		production_lines_json = VALUES(production_lines_json),
		key_events = VALUES(key_events),
		review = VALUES(review)`,
		record.Company, record.Year, comprehensiveExpenseReportJSON, profitReportJSON, balanceSheetReportJSON,
		record.OpeningCash, record.ClosingCash, record.OpeningEquity, record.ClosingEquity,
		record.SalesRevenue, record.DirectCost, record.GrossProfit, record.NetProfit, record.IncomeTax,
		record.ProfitBeforeDepreciation, record.Depreciation, record.ProfitBeforeInterest, record.ProfitBeforeTax,
		record.TotalAssets, record.TotalLiability,
		record.AdExpense, record.ComprehensiveExpense, record.ManagementExpense, record.RnDExpense, record.MarketExpense, record.ISOExpense,
		record.MaintenanceExpense, record.SwitchExpense, record.RentExpense, record.InfoExpense, record.OtherLoss,
		record.FinancialExpense, record.DefaultPenalty,
		record.LongLoanAdded, record.ShortLoanAdded, record.LongLoanBalance, record.ShortLoanBalance,
		record.IncomeTaxPayable, record.ShareCapital, record.RetainedEarnings, record.OwnerEquityTotal,
		record.Receivables, record.WorkInProgress, record.FinishedGoods, record.RawMaterials, record.CurrentAssets,
		record.FactoryValue, record.ProductionLineValue, record.ConstructionInProgress, record.FixedAssets,
		ordersJSON, decisionsJSON, productInventoryJSON, materialInventoryJSON,
		rndJSON, marketsJSON, isoJSON, productionLinesJSON, record.KeyEvents, record.Review)
	if err != nil {
		return record, fmt.Errorf("保存年度运营记录失败: %w", err)
	}
	return record, nil
}

func (s *mysqlOperationRecordStore) List(company string) ([]erp.AnnualOperationRecord, error) {
	if company == "" {
		company = "默认企业"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT
		company, year, comprehensive_expense_report_json, profit_report_json, balance_sheet_report_json,
		opening_cash, closing_cash, opening_equity, closing_equity,
		sales_revenue, direct_cost, gross_profit, net_profit, income_tax,
		profit_before_depreciation, depreciation, profit_before_interest, profit_before_tax,
		total_assets, total_liability,
		ad_expense, comprehensive_expense, management_expense, rnd_expense, market_expense, iso_expense,
		maintenance_expense, switch_expense, rent_expense, info_expense, other_loss,
		financial_expense, default_penalty,
		long_loan_added, short_loan_added, long_loan_balance, short_loan_balance,
		income_tax_payable, share_capital, retained_earnings, owner_equity_total,
		receivables, work_in_progress, finished_goods, raw_materials, current_assets,
		factory_value, production_line_value, construction_in_progress, fixed_assets,
		orders_json, decisions_json, product_inventory_json, material_inventory_json,
		rnd_json, markets_json, iso_json, production_lines_json, key_events, review
		FROM annual_operation_records WHERE company = ? ORDER BY year ASC`, company)
	if err != nil {
		return nil, fmt.Errorf("查询年度运营记录失败: %w", err)
	}
	defer rows.Close()

	records := make([]erp.AnnualOperationRecord, 0)
	for rows.Next() {
		record, err := scanAnnualOperationRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取年度运营记录失败: %w", err)
	}
	return records, nil
}

func (s *mysqlOperationRecordStore) Delete(company string, year int) error {
	if company == "" {
		company = "默认企业"
	}
	if year <= 0 {
		return fmt.Errorf("year 必须大于 0")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.db.ExecContext(ctx, `DELETE FROM annual_operation_records WHERE company = ? AND year = ?`, company, year); err != nil {
		return fmt.Errorf("删除年度运营记录失败: %w", err)
	}
	return nil
}

func (s *mysqlOperationRecordStore) Companies() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT company FROM annual_operation_records ORDER BY company ASC`)
	if err != nil {
		return nil, fmt.Errorf("查询企业列表失败: %w", err)
	}
	defer rows.Close()

	companies := make([]string, 0)
	for rows.Next() {
		var company string
		if err := rows.Scan(&company); err != nil {
			return nil, fmt.Errorf("读取企业列表失败: %w", err)
		}
		companies = append(companies, company)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取企业列表失败: %w", err)
	}
	sort.Strings(companies)
	return companies, nil
}

type annualRecordScanner interface {
	Scan(dest ...any) error
}

func scanAnnualOperationRecord(scanner annualRecordScanner) (erp.AnnualOperationRecord, error) {
	var record erp.AnnualOperationRecord
	var comprehensiveExpenseReportJSON, profitReportJSON, balanceSheetReportJSON sql.NullString
	var ordersJSON, decisionsJSON, productInventoryJSON, materialInventoryJSON sql.NullString
	var rndJSON, marketsJSON, isoJSON, productionLinesJSON sql.NullString
	var keyEvents, review sql.NullString

	err := scanner.Scan(
		&record.Company, &record.Year, &comprehensiveExpenseReportJSON, &profitReportJSON, &balanceSheetReportJSON,
		&record.OpeningCash, &record.ClosingCash, &record.OpeningEquity, &record.ClosingEquity,
		&record.SalesRevenue, &record.DirectCost, &record.GrossProfit, &record.NetProfit, &record.IncomeTax,
		&record.ProfitBeforeDepreciation, &record.Depreciation, &record.ProfitBeforeInterest, &record.ProfitBeforeTax,
		&record.TotalAssets, &record.TotalLiability,
		&record.AdExpense, &record.ComprehensiveExpense, &record.ManagementExpense, &record.RnDExpense, &record.MarketExpense, &record.ISOExpense,
		&record.MaintenanceExpense, &record.SwitchExpense, &record.RentExpense, &record.InfoExpense, &record.OtherLoss,
		&record.FinancialExpense, &record.DefaultPenalty,
		&record.LongLoanAdded, &record.ShortLoanAdded, &record.LongLoanBalance, &record.ShortLoanBalance,
		&record.IncomeTaxPayable, &record.ShareCapital, &record.RetainedEarnings, &record.OwnerEquityTotal,
		&record.Receivables, &record.WorkInProgress, &record.FinishedGoods, &record.RawMaterials, &record.CurrentAssets,
		&record.FactoryValue, &record.ProductionLineValue, &record.ConstructionInProgress, &record.FixedAssets,
		&ordersJSON, &decisionsJSON, &productInventoryJSON, &materialInventoryJSON,
		&rndJSON, &marketsJSON, &isoJSON, &productionLinesJSON, &keyEvents, &review)
	if err != nil {
		return record, fmt.Errorf("扫描年度运营记录失败: %w", err)
	}

	record.KeyEvents = nullStringValue(keyEvents)
	record.Review = nullStringValue(review)
	_ = unmarshalNullableJSON(comprehensiveExpenseReportJSON, &record.ComprehensiveExpenseReport)
	_ = unmarshalNullableJSON(profitReportJSON, &record.ProfitReport)
	_ = unmarshalNullableJSON(balanceSheetReportJSON, &record.BalanceSheetReport)
	_ = unmarshalNullableJSON(ordersJSON, &record.Orders)
	_ = unmarshalNullableJSON(decisionsJSON, &record.Decisions)
	_ = unmarshalNullableJSON(productInventoryJSON, &record.ProductInventory)
	_ = unmarshalNullableJSON(materialInventoryJSON, &record.MaterialInventory)
	_ = unmarshalNullableJSON(rndJSON, &record.RnD)
	_ = unmarshalNullableJSON(marketsJSON, &record.Markets)
	_ = unmarshalNullableJSON(isoJSON, &record.ISO)
	_ = unmarshalNullableJSON(productionLinesJSON, &record.ProductionLines)
	erp.NormalizeAnnualRecord(&record)
	return record, nil
}

type mysqlAdvisorHistoryStore struct {
	db *sql.DB
}

func newMySQLAdvisorHistoryStore(db *sql.DB) *mysqlAdvisorHistoryStore {
	return &mysqlAdvisorHistoryStore{db: db}
}

func (s *mysqlAdvisorHistoryStore) Add(record erp.AdvisorQARecord) (erp.AdvisorQARecord, error) {
	if record.Company == "" {
		record.Company = "默认企业"
	}
	createdAt := time.Now()
	if record.CreatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, record.CreatedAt); err == nil {
			createdAt = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `INSERT INTO advisor_qa_records (
		company, year, quarter, step_code, step_name, question, answer, mode, ai_used, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.Company, record.Year, record.Quarter, record.StepCode, record.StepName,
		record.Question, record.Answer, record.Mode, record.AIUsed, createdAt)
	if err != nil {
		return record, fmt.Errorf("保存历史建议失败: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return record, fmt.Errorf("读取历史建议 ID 失败: %w", err)
	}
	record.ID = fmt.Sprintf("qa-%d", id)
	record.CreatedAt = createdAt.Format(time.RFC3339)
	return record, nil
}

func (s *mysqlAdvisorHistoryStore) List(company string, limit int) ([]erp.AdvisorQARecord, error) {
	if company == "" {
		company = "默认企业"
	}
	if limit <= 0 {
		limit = 20
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT id, company, year, quarter, step_code, step_name,
		question, answer, mode, ai_used, created_at
		FROM advisor_qa_records WHERE company = ? ORDER BY created_at DESC, id DESC LIMIT ?`, company, limit)
	if err != nil {
		return nil, fmt.Errorf("查询历史建议失败: %w", err)
	}
	defer rows.Close()

	records := make([]erp.AdvisorQARecord, 0)
	for rows.Next() {
		var id int64
		var createdAt time.Time
		var record erp.AdvisorQARecord
		if err := rows.Scan(&id, &record.Company, &record.Year, &record.Quarter, &record.StepCode, &record.StepName,
			&record.Question, &record.Answer, &record.Mode, &record.AIUsed, &createdAt); err != nil {
			return nil, fmt.Errorf("读取历史建议失败: %w", err)
		}
		record.ID = fmt.Sprintf("qa-%d", id)
		record.CreatedAt = createdAt.Format(time.RFC3339)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取历史建议失败: %w", err)
	}
	return records, nil
}

func (s *mysqlAdvisorHistoryStore) Clear(company string) error {
	if company == "" {
		company = "默认企业"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.db.ExecContext(ctx, `DELETE FROM advisor_qa_records WHERE company = ?`, company); err != nil {
		return fmt.Errorf("清空历史建议失败: %w", err)
	}
	return nil
}

func marshalJSON(value any) (sql.NullString, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	return sql.NullString{String: string(data), Valid: true}, nil
}

func unmarshalNullableJSON(value sql.NullString, dst any) error {
	if !value.Valid || value.String == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(value.String), dst); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}
	return nil
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
