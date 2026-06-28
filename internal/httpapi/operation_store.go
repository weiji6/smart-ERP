package httpapi

import (
	"fmt"
	"sort"
	"sync"

	"smarterp/internal/erp"
)

type operationRecordStore struct {
	mu      sync.RWMutex
	records map[string]map[int]erp.AnnualOperationRecord
}

func newOperationRecordStore() *operationRecordStore {
	return &operationRecordStore{records: map[string]map[int]erp.AnnualOperationRecord{}}
}

func (s *operationRecordStore) Upsert(record erp.AnnualOperationRecord) erp.AnnualOperationRecord {
	erp.NormalizeAnnualRecord(&record)
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.records[record.Company] == nil {
		s.records[record.Company] = map[int]erp.AnnualOperationRecord{}
	}
	s.records[record.Company][record.Year] = record
	return record
}

func (s *operationRecordStore) List(company string) []erp.AnnualOperationRecord {
	if company == "" {
		company = "默认企业"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	recordsByYear := s.records[company]
	items := make([]erp.AnnualOperationRecord, 0, len(recordsByYear))
	for _, record := range recordsByYear {
		items = append(items, record)
	}
	erp.SortAnnualRecords(items)
	return items
}

func (s *operationRecordStore) Delete(company string, year int) error {
	if company == "" {
		company = "默认企业"
	}
	if year <= 0 {
		return fmt.Errorf("year 必须大于 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	recordsByYear := s.records[company]
	if recordsByYear == nil {
		return nil
	}
	delete(recordsByYear, year)
	if len(recordsByYear) == 0 {
		delete(s.records, company)
	}
	return nil
}

func (s *operationRecordStore) Companies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	companies := make([]string, 0, len(s.records))
	for company := range s.records {
		companies = append(companies, company)
	}
	sort.Strings(companies)
	return companies
}
