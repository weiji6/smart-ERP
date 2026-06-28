package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

const defaultDBConfigPath = "config/db.local.json"

type DBConfig struct {
	Driver       string `json:"driver"`
	DSN          string `json:"dsn"`
	AutoMigrate  bool   `json:"autoMigrate"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxIdleConns int    `json:"maxIdleConns"`
	ConfigPath   string `json:"-"`
	ConfigLoaded bool   `json:"-"`
}

func LoadDBConfigFromEnv() (DBConfig, bool) {
	configPath := strings.TrimSpace(os.Getenv("DB_CONFIG_FILE"))
	if configPath == "" {
		configPath = defaultDBConfigPath
	}

	cfg := DBConfig{
		Driver:       "mysql",
		AutoMigrate:  true,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		ConfigPath:   configPath,
	}
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &cfg)
		cfg.ConfigPath = configPath
		cfg.ConfigLoaded = true
	}

	if v := strings.TrimSpace(os.Getenv("DB_DRIVER")); v != "" {
		cfg.Driver = v
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_DSN")); v != "" {
		cfg.DSN = v
	} else if v := strings.TrimSpace(os.Getenv("DB_DSN")); v != "" {
		cfg.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("DB_AUTO_MIGRATE")); v != "" {
		cfg.AutoMigrate = parseBool(v, cfg.AutoMigrate)
	}
	if v := strings.TrimSpace(os.Getenv("DB_MAX_OPEN_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxOpenConns = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("DB_MAX_IDLE_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.MaxIdleConns = n
		}
	}

	cfg.Driver = strings.TrimSpace(cfg.Driver)
	cfg.DSN = strings.TrimSpace(cfg.DSN)
	return cfg, cfg.DSN != ""
}

func OpenDB(cfg DBConfig) (*sql.DB, error) {
	dsn := cfg.DSN
	if cfg.Driver == "mysql" {
		if mysqlCfg, err := mysql.ParseDSN(cfg.DSN); err == nil {
			mysqlCfg.ParseTime = true
			if mysqlCfg.Loc == nil {
				mysqlCfg.Loc = time.Local
			}
			if mysqlCfg.Collation == "" {
				mysqlCfg.Collation = "utf8mb4_unicode_ci"
			}
			dsn = mysqlCfg.FormatDSN()
		}
	}

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns >= 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func migrateMySQL(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS annual_operation_records (
			company VARCHAR(191) NOT NULL,
			year INT NOT NULL,
			comprehensive_expense_report_json LONGTEXT NULL,
			profit_report_json LONGTEXT NULL,
			balance_sheet_report_json LONGTEXT NULL,
			opening_cash BIGINT NOT NULL DEFAULT 0,
			closing_cash BIGINT NOT NULL DEFAULT 0,
			opening_equity BIGINT NOT NULL DEFAULT 0,
			closing_equity BIGINT NOT NULL DEFAULT 0,
			sales_revenue BIGINT NOT NULL DEFAULT 0,
			direct_cost BIGINT NOT NULL DEFAULT 0,
			gross_profit BIGINT NOT NULL DEFAULT 0,
			net_profit BIGINT NOT NULL DEFAULT 0,
			income_tax BIGINT NOT NULL DEFAULT 0,
			profit_before_depreciation BIGINT NOT NULL DEFAULT 0,
			depreciation BIGINT NOT NULL DEFAULT 0,
			profit_before_interest BIGINT NOT NULL DEFAULT 0,
			profit_before_tax BIGINT NOT NULL DEFAULT 0,
			total_assets BIGINT NOT NULL DEFAULT 0,
			total_liability BIGINT NOT NULL DEFAULT 0,
			ad_expense BIGINT NOT NULL DEFAULT 0,
			comprehensive_expense BIGINT NOT NULL DEFAULT 0,
			management_expense BIGINT NOT NULL DEFAULT 0,
			rnd_expense BIGINT NOT NULL DEFAULT 0,
			market_expense BIGINT NOT NULL DEFAULT 0,
			iso_expense BIGINT NOT NULL DEFAULT 0,
			maintenance_expense BIGINT NOT NULL DEFAULT 0,
			switch_expense BIGINT NOT NULL DEFAULT 0,
			rent_expense BIGINT NOT NULL DEFAULT 0,
			info_expense BIGINT NOT NULL DEFAULT 0,
			other_loss BIGINT NOT NULL DEFAULT 0,
			financial_expense BIGINT NOT NULL DEFAULT 0,
			default_penalty BIGINT NOT NULL DEFAULT 0,
			long_loan_added BIGINT NOT NULL DEFAULT 0,
			short_loan_added BIGINT NOT NULL DEFAULT 0,
			long_loan_balance BIGINT NOT NULL DEFAULT 0,
			short_loan_balance BIGINT NOT NULL DEFAULT 0,
			income_tax_payable BIGINT NOT NULL DEFAULT 0,
			share_capital BIGINT NOT NULL DEFAULT 0,
			retained_earnings BIGINT NOT NULL DEFAULT 0,
			owner_equity_total BIGINT NOT NULL DEFAULT 0,
			receivables BIGINT NOT NULL DEFAULT 0,
			work_in_progress BIGINT NOT NULL DEFAULT 0,
			finished_goods BIGINT NOT NULL DEFAULT 0,
			raw_materials BIGINT NOT NULL DEFAULT 0,
			current_assets BIGINT NOT NULL DEFAULT 0,
			factory_value BIGINT NOT NULL DEFAULT 0,
			production_line_value BIGINT NOT NULL DEFAULT 0,
			construction_in_progress BIGINT NOT NULL DEFAULT 0,
			fixed_assets BIGINT NOT NULL DEFAULT 0,
			orders_json LONGTEXT NULL,
			decisions_json LONGTEXT NULL,
			product_inventory_json LONGTEXT NULL,
			material_inventory_json LONGTEXT NULL,
			rnd_json LONGTEXT NULL,
			markets_json LONGTEXT NULL,
			iso_json LONGTEXT NULL,
			production_lines_json LONGTEXT NULL,
			key_events TEXT NULL,
			review TEXT NULL,
			created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
			updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
			PRIMARY KEY (company, year)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS advisor_qa_records (
			id BIGINT NOT NULL AUTO_INCREMENT,
			company VARCHAR(191) NOT NULL,
			year INT NOT NULL DEFAULT 0,
			quarter INT NOT NULL DEFAULT 0,
			step_code VARCHAR(128) NOT NULL DEFAULT '',
			step_name VARCHAR(191) NOT NULL DEFAULT '',
			question LONGTEXT NULL,
			answer LONGTEXT NULL,
			mode VARCHAR(64) NOT NULL DEFAULT '',
			ai_used BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
			PRIMARY KEY (id),
			INDEX idx_advisor_company_created (company, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	columns := map[string]string{
		"comprehensive_expense_report_json": "LONGTEXT NULL",
		"profit_report_json":                "LONGTEXT NULL",
		"balance_sheet_report_json":         "LONGTEXT NULL",
		"direct_cost":                       "BIGINT NOT NULL DEFAULT 0",
		"gross_profit":                      "BIGINT NOT NULL DEFAULT 0",
		"profit_before_depreciation":        "BIGINT NOT NULL DEFAULT 0",
		"depreciation":                      "BIGINT NOT NULL DEFAULT 0",
		"profit_before_interest":            "BIGINT NOT NULL DEFAULT 0",
		"profit_before_tax":                 "BIGINT NOT NULL DEFAULT 0",
		"management_expense":                "BIGINT NOT NULL DEFAULT 0",
		"switch_expense":                    "BIGINT NOT NULL DEFAULT 0",
		"rent_expense":                      "BIGINT NOT NULL DEFAULT 0",
		"info_expense":                      "BIGINT NOT NULL DEFAULT 0",
		"other_loss":                        "BIGINT NOT NULL DEFAULT 0",
		"income_tax_payable":                "BIGINT NOT NULL DEFAULT 0",
		"share_capital":                     "BIGINT NOT NULL DEFAULT 0",
		"retained_earnings":                 "BIGINT NOT NULL DEFAULT 0",
		"owner_equity_total":                "BIGINT NOT NULL DEFAULT 0",
		"receivables":                       "BIGINT NOT NULL DEFAULT 0",
		"work_in_progress":                  "BIGINT NOT NULL DEFAULT 0",
		"finished_goods":                    "BIGINT NOT NULL DEFAULT 0",
		"raw_materials":                     "BIGINT NOT NULL DEFAULT 0",
		"current_assets":                    "BIGINT NOT NULL DEFAULT 0",
		"factory_value":                     "BIGINT NOT NULL DEFAULT 0",
		"production_line_value":             "BIGINT NOT NULL DEFAULT 0",
		"construction_in_progress":          "BIGINT NOT NULL DEFAULT 0",
		"fixed_assets":                      "BIGINT NOT NULL DEFAULT 0",
	}
	for name, definition := range columns {
		if err := addMySQLColumnIfMissing(ctx, db, "annual_operation_records", name, definition); err != nil {
			return err
		}
	}
	return nil
}

func addMySQLColumnIfMissing(ctx context.Context, db *sql.DB, tableName, columnName, definition string) error {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`, tableName, columnName).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", tableName, columnName, definition)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func parseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
