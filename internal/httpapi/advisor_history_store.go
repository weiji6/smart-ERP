package httpapi

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"smarterp/internal/erp"
)

type advisorHistoryStore struct {
	mu      sync.RWMutex
	nextID  int64
	records map[string][]erp.AdvisorQARecord
}

func newAdvisorHistoryStore() *advisorHistoryStore {
	return &advisorHistoryStore{records: map[string][]erp.AdvisorQARecord{}}
}

func (s *advisorHistoryStore) Add(record erp.AdvisorQARecord) (erp.AdvisorQARecord, error) {
	if record.Company == "" {
		record.Company = "默认企业"
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	record.ID = fmt.Sprintf("qa-%d", s.nextID)
	if record.CreatedAt == "" {
		record.CreatedAt = time.Now().Format(time.RFC3339)
	}
	s.records[record.Company] = append(s.records[record.Company], record)
	return record, nil
}

func (s *advisorHistoryStore) List(company string, limit int) ([]erp.AdvisorQARecord, error) {
	if company == "" {
		company = "默认企业"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]erp.AdvisorQARecord(nil), s.records[company]...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *advisorHistoryStore) Clear(company string) error {
	if company == "" {
		company = "默认企业"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, company)
	return nil
}
