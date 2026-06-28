package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"smarterp/internal/erp"
)

const defaultOperationRecordPath = "data/operation_records.json"

type Server struct {
	mux              *http.ServeMux
	operationRecords operationRecordRepository
	advisorHistory   advisorHistoryRepository
	dbStatus         DBStatus
}

func NewServer() http.Handler {
	operationRecords := newOperationRecordStore(defaultOperationRecordPath)
	s := &Server{
		mux:              http.NewServeMux(),
		operationRecords: operationRecords,
		advisorHistory:   newAdvisorHistoryStore(),
		dbStatus: DBStatus{
			Enabled:     false,
			StorageMode: "file",
			ConfigPath:  defaultOperationRecordPath,
		},
	}
	s.configureStorage()
	s.routes()
	return s
}

func (s *Server) configureStorage() {
	cfg, enabled := LoadDBConfigFromEnv()
	s.dbStatus.ConfigPath = cfg.ConfigPath
	s.dbStatus.ConfigLoaded = cfg.ConfigLoaded
	s.dbStatus.Driver = cfg.Driver
	if !enabled {
		s.dbStatus.ConfigPath = defaultOperationRecordPath
		s.dbStatus.ConfigLoaded = true
		return
	}

	db, err := OpenDB(cfg)
	if err != nil {
		s.dbStatus.Error = err.Error()
		log.Printf("MySQL 数据库连接失败，已回退到本地文件存储: %v", err)
		return
	}
	if cfg.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := migrateMySQL(ctx, db); err != nil {
			s.dbStatus.Error = err.Error()
			_ = db.Close()
			log.Printf("MySQL 数据库建表失败，已回退到本地文件存储: %v", err)
			return
		}
	}

	mysqlOperationRecords := newMySQLOperationRecordStore(db)
	if err := copyOperationRecords(s.operationRecords, mysqlOperationRecords); err != nil {
		s.dbStatus.Error = err.Error()
		_ = db.Close()
		log.Printf("年度记录迁移到 MySQL 失败，已回退到本地文件存储: %v", err)
		return
	}

	s.operationRecords = mysqlOperationRecords
	s.advisorHistory = newMySQLAdvisorHistoryStore(db)
	s.dbStatus.Enabled = true
	s.dbStatus.StorageMode = "mysql"
}

func copyOperationRecords(src, dst operationRecordRepository) error {
	companies, err := src.Companies()
	if err != nil {
		return err
	}
	for _, company := range companies {
		records, err := src.List(company)
		if err != nil {
			return err
		}
		for _, record := range records {
			if _, err := dst.Upsert(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /api/v1/rules", s.rules)
	s.mux.HandleFunc("POST /api/v1/finance/cashflow", s.cashflow)
	s.mux.HandleFunc("POST /api/v1/finance/loan-capacity", s.loanCapacity)
	s.mux.HandleFunc("POST /api/v1/finance/balance-sheet", s.balanceSheet)
	s.mux.HandleFunc("POST /api/v1/production/bom", s.bom)
	s.mux.HandleFunc("POST /api/v1/production/purchase-plan", s.purchasePlan)
	s.mux.HandleFunc("POST /api/v1/production/capacity", s.capacity)
	s.mux.HandleFunc("POST /api/v1/order/score", s.orderScore)
	s.mux.HandleFunc("POST /api/v1/market/predict", s.marketPredict)
	s.mux.HandleFunc("POST /api/v1/market/ad-optimize", s.adOptimize)
	s.mux.HandleFunc("POST /api/v1/strategy/compare", s.strategyCompare)
	s.mux.HandleFunc("GET /api/v1/operation-flow", s.operationFlow)
	s.mux.HandleFunc("GET /api/v1/operation-records", s.listOperationRecords)
	s.mux.HandleFunc("POST /api/v1/operation-records", s.saveOperationRecord)
	s.mux.HandleFunc("DELETE /api/v1/operation-records", s.deleteOperationRecord)
	s.mux.HandleFunc("GET /api/v1/advisor/history", s.listAdvisorHistory)
	s.mux.HandleFunc("DELETE /api/v1/advisor/history", s.clearAdvisorHistory)
	s.mux.HandleFunc("POST /api/v1/advisor/decision", s.decisionAdvisor)
	s.mux.HandleFunc("GET /api/v1/ai/status", s.aiStatus)
	s.mux.HandleFunc("POST /api/v1/ai/check", s.aiCheck)
	s.mux.HandleFunc("GET /api/v1/db/status", s.dbStatusHandler)
	s.mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))
	s.mux.HandleFunc("GET /", s.index)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/index.html")
}

func (s *Server) rules(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"products":  erp.ProductRules,
		"materials": erp.MaterialRules,
		"lines":     erp.LineRules,
		"factories": erp.FactoryRules,
		"markets":   erp.MarketOpenRules,
		"iso":       erp.ISORules,
	})
}

func (s *Server) operationFlow(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"steps":            erp.OperationFlow(),
		"years":            erp.AnnualOperationYears(erp.CompetitionYears),
		"reportTemplate":   erp.AnnualReportTemplate(),
		"competitionYears": erp.CompetitionYears,
	})
}

func (s *Server) listOperationRecords(w http.ResponseWriter, r *http.Request) {
	company := r.URL.Query().Get("company")
	records, err := s.operationRecords.List(company)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	companies, err := s.operationRecords.Companies()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"records":   records,
		"analysis":  erp.AnalyzeOperationHistory(records),
		"companies": companies,
	})
}

func (s *Server) saveOperationRecord(w http.ResponseWriter, r *http.Request) {
	var record erp.AnnualOperationRecord
	if !decode(w, r, &record) {
		return
	}
	if record.Year <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("year 必须大于 0"))
		return
	}

	saved, err := s.operationRecords.Upsert(record)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	records, err := s.operationRecords.List(saved.Company)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"record":   saved,
		"records":  records,
		"analysis": erp.AnalyzeOperationHistory(records),
	})
}

func (s *Server) deleteOperationRecord(w http.ResponseWriter, r *http.Request) {
	company := r.URL.Query().Get("company")
	year := queryInt(r, "year", 0)
	if err := s.operationRecords.Delete(company, year); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	records, err := s.operationRecords.List(company)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"records":  records,
		"analysis": erp.AnalyzeOperationHistory(records),
	})
}

func (s *Server) listAdvisorHistory(w http.ResponseWriter, r *http.Request) {
	company := r.URL.Query().Get("company")
	limit := queryInt(r, "limit", 20)
	records, err := s.advisorHistory.List(company, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"records": records,
	})
}

func (s *Server) clearAdvisorHistory(w http.ResponseWriter, r *http.Request) {
	company := r.URL.Query().Get("company")
	if err := s.advisorHistory.Clear(company); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": []erp.AdvisorQARecord{}})
}

func (s *Server) aiStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, NewAIClientFromEnv().Status())
}

func (s *Server) aiCheck(w http.ResponseWriter, r *http.Request) {
	aiClient := NewAIClientFromEnv()
	status := aiClient.Status()
	if !status.Enabled {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     false,
			"status": status,
			"error":  "未配置 AI_API_KEY 或 OPENAI_API_KEY，当前会使用规则兜底",
		})
		return
	}

	sample, err := aiClient.Ask(r.Context(), "请只回复 ok，用于验证 AI 接口连通性。")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     false,
			"status": status,
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"status": status,
		"sample": sample,
	})
}

func (s *Server) dbStatusHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.dbStatus)
}

type stateRequest struct {
	State erp.CompanyState `json:"state"`
}

type cashflowRequest struct {
	State    erp.CompanyState `json:"state"`
	Quarters int              `json:"quarters"`
}

func (s *Server) cashflow(w http.ResponseWriter, r *http.Request) {
	var req cashflowRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	if req.Quarters == 0 {
		req.Quarters = 20
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"currentCash": req.State.Cash,
		"forecasts":   erp.PredictCashFlow(req.State, req.Quarters),
	})
}

func (s *Server) loanCapacity(w http.ResponseWriter, r *http.Request) {
	var req stateRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	writeJSON(w, http.StatusOK, erp.CalcLoanCapacity(req.State))
}

func (s *Server) balanceSheet(w http.ResponseWriter, r *http.Request) {
	var req stateRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	writeJSON(w, http.StatusOK, erp.CalcBalanceSheet(req.State))
}

type ordersRequest struct {
	State  erp.CompanyState `json:"state"`
	Orders []erp.Order      `json:"orders"`
}

func (s *Server) bom(w http.ResponseWriter, r *http.Request) {
	var req ordersRequest
	if !decode(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"requirements": erp.CalcMaterialRequirements(req.Orders),
	})
}

func (s *Server) purchasePlan(w http.ResponseWriter, r *http.Request) {
	var req ordersRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	writeJSON(w, http.StatusOK, map[string]any{
		"plan": erp.GeneratePurchasePlan(req.Orders, req.State.MaterialInventory, req.State.CurrentQuarterIndex()),
	})
}

type capacityRequest struct {
	Lines    []erp.ProductionLine `json:"lines"`
	Year     int                  `json:"year"`
	Quarter  int                  `json:"quarter"`
	Quarters int                  `json:"quarters"`
}

func (s *Server) capacity(w http.ResponseWriter, r *http.Request) {
	var req capacityRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Year == 0 {
		req.Year = 1
	}
	if req.Quarter == 0 {
		req.Quarter = 1
	}
	if req.Quarters == 0 {
		req.Quarters = 8
	}
	if err := erp.ValidateYearQuarter(req.Year, req.Quarter); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"capacity": erp.CalcMaxCapacity(req.Lines, erp.QuarterIndex(req.Year, req.Quarter), req.Quarters),
	})
}

func (s *Server) orderScore(w http.ResponseWriter, r *http.Request) {
	var req ordersRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	writeJSON(w, http.StatusOK, map[string]any{
		"scores": erp.ScoreOrders(req.State, req.Orders),
	})
}

type marketPredictRequest struct {
	Market        erp.Market  `json:"market"`
	Product       erp.Product `json:"product"`
	AdCost        erp.Money   `json:"adCost"`
	CompetitorAds []erp.Money `json:"competitorAds"`
}

func (s *Server) marketPredict(w http.ResponseWriter, r *http.Request) {
	var req marketPredictRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Market == "" || req.Product == "" {
		writeError(w, http.StatusBadRequest, errors.New("market 和 product 不能为空"))
		return
	}
	writeJSON(w, http.StatusOK, erp.PredictMarketOrder(req.Market, req.Product, req.AdCost, req.CompetitorAds))
}

type adOptimizeRequest struct {
	Markets       []erp.Market  `json:"markets"`
	Products      []erp.Product `json:"products"`
	Budget        erp.Money     `json:"budget"`
	CompetitorAds []erp.Money   `json:"competitorAds"`
}

func (s *Server) adOptimize(w http.ResponseWriter, r *http.Request) {
	var req adOptimizeRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Budget <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("budget 必须大于 0"))
		return
	}
	if len(req.Markets) == 0 {
		req.Markets = []erp.Market{erp.MarketLocal, erp.MarketRegional, erp.MarketDomestic}
	}
	if len(req.Products) == 0 {
		req.Products = []erp.Product{erp.ProductP1, erp.ProductP2, erp.ProductP3}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"allocation": erp.OptimizeAdAllocation(req.Markets, req.Products, req.Budget, req.CompetitorAds),
	})
}

type strategyCompareRequest struct {
	State       erp.CompanyState     `json:"state"`
	Strategies  []erp.StrategyParams `json:"strategies"`
	Simulations int                  `json:"simulations"`
}

func (s *Server) strategyCompare(w http.ResponseWriter, r *http.Request) {
	var req strategyCompareRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)
	writeJSON(w, http.StatusOK, erp.CompareStrategies(req.State, req.Strategies, req.Simulations))
}

type advisorRequest struct {
	State            erp.CompanyState            `json:"state"`
	StepCode         string                      `json:"stepCode"`
	Decisions        []erp.DecisionInput         `json:"decisions"`
	Orders           []erp.Order                 `json:"orders"`
	CompetitorAds    []erp.Money                 `json:"competitorAds"`
	OperationRecords []erp.AnnualOperationRecord `json:"operationRecords"`
	AdviceHistory    []erp.AdvisorQARecord       `json:"adviceHistory"`
	Question         string                      `json:"question"`
	UseAI            *bool                       `json:"useAI"`
	IncludePrompt    bool                        `json:"includePrompt"`
}

type advisorResponse struct {
	Mode        string                 `json:"mode"`
	AIEnabled   bool                   `json:"aiEnabled"`
	AIUsed      bool                   `json:"aiUsed"`
	AIError     string                 `json:"aiError,omitempty"`
	Advice      string                 `json:"advice"`
	Diagnostics erp.AdvisorDiagnostics `json:"diagnostics"`
	Prompt      string                 `json:"prompt,omitempty"`
	QARecord    erp.AdvisorQARecord    `json:"qaRecord"`
	QAHistory   []erp.AdvisorQARecord  `json:"qaHistory"`
}

func (s *Server) decisionAdvisor(w http.ResponseWriter, r *http.Request) {
	var req advisorRequest
	if !decode(w, r, &req) {
		return
	}
	normalizeState(&req.State)

	useAI := true
	if req.UseAI != nil {
		useAI = *req.UseAI
	}
	operationRecords := req.OperationRecords
	if len(operationRecords) == 0 {
		records, err := s.operationRecords.List(req.State.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		operationRecords = records
	}
	adviceHistory := req.AdviceHistory
	if len(adviceHistory) == 0 {
		records, err := s.advisorHistory.List(req.State.Name, 5)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		adviceHistory = records
	}
	ctx := erp.AdvisorContext{
		State:            req.State,
		StepCode:         req.StepCode,
		Decisions:        req.Decisions,
		Orders:           req.Orders,
		CompetitorAds:    req.CompetitorAds,
		OperationRecords: operationRecords,
		AdviceHistory:    adviceHistory,
		Question:         req.Question,
	}
	diagnostics := erp.BuildAdvisorDiagnostics(ctx)
	prompt := erp.AdvisorPrompt(ctx, diagnostics)
	aiClient := NewAIClientFromEnv()
	if useAI && aiClient.Enabled() {
		prompt = erp.AdvisorAIPrompt(ctx, diagnostics)
	}

	resp := advisorResponse{
		Mode:        "rule",
		AIEnabled:   aiClient.Enabled(),
		AIUsed:      false,
		Advice:      diagnostics.RuleRecommendation,
		Diagnostics: diagnostics,
	}
	if req.IncludePrompt {
		resp.Prompt = prompt
	}

	if useAI && aiClient.Enabled() {
		advice, err := aiClient.Ask(r.Context(), prompt)
		if err != nil {
			resp.AIError = err.Error()
			resp.Mode = "rule_fallback"
		} else {
			resp.Mode = "ai"
			resp.AIUsed = true
			resp.Advice = advice
		}
	}
	if useAI && !aiClient.Enabled() {
		resp.AIError = "未配置 AI_API_KEY 或 OPENAI_API_KEY，已返回规则引擎建议"
		resp.Mode = "rule_fallback"
	}

	qaRecord, err := s.advisorHistory.Add(erp.AdvisorQARecord{
		Company:  req.State.Name,
		Year:     req.State.Year,
		Quarter:  req.State.Quarter,
		StepCode: diagnostics.CurrentStep.Code,
		StepName: diagnostics.CurrentStep.Name,
		Question: req.Question,
		Answer:   resp.Advice,
		Mode:     resp.Mode,
		AIUsed:   resp.AIUsed,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	resp.QARecord = qaRecord
	qaHistory, err := s.advisorHistory.List(req.State.Name, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	resp.QAHistory = qaHistory

	writeJSON(w, http.StatusOK, resp)
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func normalizeState(state *erp.CompanyState) {
	def := erp.DefaultCompanyState()
	if state.Name == "" {
		state.Name = def.Name
	}
	if state.Year == 0 {
		state.Year = def.Year
	}
	if state.Quarter == 0 {
		state.Quarter = def.Quarter
	}
	if state.Cash == 0 {
		state.Cash = def.Cash
	}
	if state.LastYearEquity == 0 {
		state.LastYearEquity = def.LastYearEquity
	}
	if state.Equity == 0 {
		state.Equity = def.Equity
	}
	if state.InitialEquity == 0 {
		state.InitialEquity = def.InitialEquity
	}
	if state.AdminFee == 0 {
		state.AdminFee = def.AdminFee
	}
	if state.InfoFee == 0 {
		state.InfoFee = def.InfoFee
	}
	if state.ProductInventory == nil {
		state.ProductInventory = map[erp.Product]int{}
	}
	if state.MaterialInventory == nil {
		state.MaterialInventory = map[erp.Material]int{}
	}
	if state.RnD == nil {
		state.RnD = map[erp.Product]erp.RnDProgress{}
	}
	if state.Markets == nil {
		state.Markets = map[erp.Market]erp.MarketProgress{}
	}
	if state.ISO == nil {
		state.ISO = map[string]erp.ISOProgress{}
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
