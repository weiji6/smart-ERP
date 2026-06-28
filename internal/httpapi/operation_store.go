package httpapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"smarterp/internal/erp"
)

type operationRecordStore struct {
	mu      sync.RWMutex
	records map[string]map[int]erp.AnnualOperationRecord
	path    string
}

func newOperationRecordStore(path string) *operationRecordStore {
	store := &operationRecordStore{
		records: map[string]map[int]erp.AnnualOperationRecord{},
		path:    path,
	}
	if err := store.load(); err != nil {
		fmt.Printf("读取年度记录文件失败，已使用空记录: %v\n", err)
	}
	return store
}

func (s *operationRecordStore) Upsert(record erp.AnnualOperationRecord) (erp.AnnualOperationRecord, error) {
	erp.NormalizeAnnualRecord(&record)
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.records[record.Company] == nil {
		s.records[record.Company] = map[int]erp.AnnualOperationRecord{}
	}
	s.records[record.Company][record.Year] = record
	if err := s.saveLocked(); err != nil {
		return record, err
	}
	return record, nil
}

func (s *operationRecordStore) List(company string) ([]erp.AnnualOperationRecord, error) {
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
	return items, nil
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
	if err := s.saveLocked(); err != nil {
		return err
	}
	return nil
}

func (s *operationRecordStore) Companies() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	companies := make([]string, 0, len(s.records))
	for company := range s.records {
		companies = append(companies, company)
	}
	sort.Strings(companies)
	return companies, nil
}

func (s *operationRecordStore) load() error {
	if s.path == "" {
		return nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var records []erp.AnnualOperationRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}
	for _, record := range records {
		erp.NormalizeAnnualRecord(&record)
		if s.records[record.Company] == nil {
			s.records[record.Company] = map[int]erp.AnnualOperationRecord{}
		}
		s.records[record.Company][record.Year] = record
	}
	return nil
}

func (s *operationRecordStore) saveLocked() error {
	if s.path == "" {
		return nil
	}
	records := make([]erp.AnnualOperationRecord, 0)
	for _, recordsByYear := range s.records {
		for _, record := range recordsByYear {
			records = append(records, record)
		}
	}
	erp.SortAnnualRecords(records)
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化年度记录失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("创建年度记录目录失败: %w", err)
	}
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("写入年度记录文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("替换年度记录文件失败: %w", err)
	}
	return nil
}
