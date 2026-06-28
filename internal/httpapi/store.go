package httpapi

import "smarterp/internal/erp"

type operationRecordRepository interface {
	Upsert(record erp.AnnualOperationRecord) (erp.AnnualOperationRecord, error)
	List(company string) ([]erp.AnnualOperationRecord, error)
	Delete(company string, year int) error
	Companies() ([]string, error)
}

type advisorHistoryRepository interface {
	Add(record erp.AdvisorQARecord) (erp.AdvisorQARecord, error)
	List(company string, limit int) ([]erp.AdvisorQARecord, error)
	Clear(company string) error
}

type DBStatus struct {
	Enabled      bool   `json:"enabled"`
	Driver       string `json:"driver,omitempty"`
	ConfigPath   string `json:"configPath,omitempty"`
	ConfigLoaded bool   `json:"configLoaded"`
	StorageMode  string `json:"storageMode"`
	Error        string `json:"error,omitempty"`
}
