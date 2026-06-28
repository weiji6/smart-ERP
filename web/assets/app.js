const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => Array.from(document.querySelectorAll(selector));
const API_BASE_URL = window.location.protocol === "file:" ? "http://localhost:8080" : "";
const RECORD_PAGE_TURN_OUT_MS = 620;
const RECORD_PAGE_TURN_IN_MS = 760;

const state = {
  flow: [],
  decisions: [
    { type: "ad", market: "本地", product: "P2", amount: 6, quantity: 0, note: "本地P2主投" },
    { type: "short_loan", amount: 20, quantity: 0, note: "保持现金垫" },
  ],
  orders: [
    { id: "A", year: 1, market: "本地", product: "P2", quantity: 2, totalPrice: 18, deliveryQuarter: 2, paymentPeriod: 1, isoRequired: "", deliveryTime: "Y1Q2", collectionTime: "Y1Q3" },
  ],
  factories: [
    { id: "F1", type: "大厂房", ownership: "租用", rented: true, purchased: false, startQuarterIndex: 1 },
  ],
  productionLines: [
    { id: "L1", factoryId: "F1", type: "自动线", product: "P2", built: true, status: "built", netValue: 15, isProducing: true, productionStartedQuarterIndex: 1, productionProgress: 1, productionTotal: 1, currentProductionProduct: "P2" },
  ],
  cashFlowPoints: [
    { quarterIndex: 1, openingCash: 60, inflows: 20, outflows: 7, closingCash: 73, note: "Y1Q1 短贷到账并完成广告、原料投入" },
    { quarterIndex: 2, openingCash: 73, inflows: 0, outflows: 1, closingCash: 72, note: "Y1Q2 常规管理费" },
    { quarterIndex: 3, openingCash: 72, inflows: 18, outflows: 1, closingCash: 89, note: "Y1Q3 订单回款" },
    { quarterIndex: 4, openingCash: 89, inflows: 0, outflows: 4, closingCash: 85, note: "Y1Q4 年末费用" },
  ],
  operationRecords: [],
  selectedRecordYear: 1,
  adviceHistory: [],
  historyAnalysis: null,
  lastAdvisor: null,
  aiStatus: null,
  dbStatus: null,
};

const productReady = {
  P1: { product: "P1", finished: true },
  P2: { product: "P2", finished: true },
};

const marketReady = {
  本地: { market: "本地", opened: true },
};

const factoryRules = {
  大厂房: { capacity: 6, rent: 5, buyCost: 40 },
  小厂房: { capacity: 4, rent: 3, buyCost: 30 },
};

const lineRules = {
  手工线: { cycle: 2, installQuarters: 0, maintenance: 1 },
  自动线: { cycle: 1, installQuarters: 3, maintenance: 2 },
  柔性线: { cycle: 1, installQuarters: 4, maintenance: 2 },
  租赁线: { cycle: 1, installQuarters: 0, maintenance: 6 },
};

const decisionFormConfig = {
  ad: {
    fields: ["market", "product", "amount", "note"],
    defaults: { market: "本地", product: "P2", amount: 6, quantity: 0, note: "广告投放" },
  },
  short_loan: {
    fields: ["amount", "note"],
    defaults: { amount: 20, quantity: 0, note: "短期贷款" },
  },
  long_loan: {
    fields: ["amount", "note"],
    defaults: { amount: 40, quantity: 0, note: "长期贷款" },
  },
  market_iso: {
    fields: ["market", "isoType", "amount", "note"],
    defaults: { market: "国内", isoType: "", amount: 1, quantity: 0, note: "市场开拓或ISO认证" },
  },
  production_line: {
    fields: ["lineType", "product", "factoryType", "amount", "quantity", "note"],
    defaults: { lineType: "自动线", product: "P2", factoryType: "", amount: 15, quantity: 1, note: "生产线建设或调整" },
  },
  material_order: {
    fields: ["material", "quantity", "amount", "note"],
    defaults: { material: "R1", amount: 0, quantity: 3, note: "按提前期下原料订单" },
  },
};

function currentCompanyState() {
  const longLoanBalance = numberValue("#longLoanInput");
  const shortLoanBalance = numberValue("#shortLoanInput");
  const decisionCashPlans = buildDecisionCashPlans();
  return {
    name: $("#recordCompany")?.value.trim() || "默认企业",
    year: numberValue("#yearInput"),
    quarter: numberValue("#quarterInput"),
    cash: numberValue("#cashInput"),
    equity: numberValue("#equityInput"),
    lastYearEquity: numberValue("#lastEquityInput"),
    longLoans: longLoanBalance > 0 ? [{ id: "L-current", type: "long", principal: longLoanBalance }] : [],
    shortLoans: shortLoanBalance > 0 ? [{ id: "S-current", type: "short", principal: shortLoanBalance }] : [],
    rnd: productReady,
    markets: marketReady,
    materialInventory: { R1: 0, R2: 0, R3: 0, R4: 0 },
    productInventory: { P1: 0, P2: 0, P3: 0, P4: 0 },
    factories: state.factories.map((factory) => factory.type),
    factoryStates: state.factories,
    productionLines: state.productionLines,
    plannedIncomes: decisionCashPlans.plannedIncomes,
    plannedExpenses: decisionCashPlans.plannedExpenses,
    receivables: decisionCashPlans.receivables,
  };
}

function buildDecisionCashPlans() {
  const plannedIncomes = [];
  const plannedExpenses = [];
  state.decisions.forEach((item) => {
    const quarterIndex = quarterIndexFromYearQuarter(item.year || numberValue("#yearInput"), item.quarter || numberValue("#quarterInput"));
    const amount = Number(item.amount || 0);
    if (amount <= 0) return;
    if (item.type === "short_loan" || item.type === "long_loan") {
      plannedIncomes.push({ quarterIndex, category: decisionName(item.type), amount });
    } else {
      plannedExpenses.push({ quarterIndex, category: decisionName(item.type), amount });
    }
  });
  return { plannedIncomes, plannedExpenses, receivables: [] };
}

function numberValue(selector) {
  const value = Number($(selector).value);
  return Number.isFinite(value) ? value : 0;
}

function calcLoanAvailable(company) {
  const usedLong = (company.longLoans || []).reduce((sum, item) => sum + Number(item.principal || 0), 0);
  const usedShort = (company.shortLoans || []).reduce((sum, item) => sum + Number(item.principal || 0), 0);
  return Math.max(0, Number(company.lastYearEquity || 0) * 3 - usedLong - usedShort - calcPlannedLoanAmount());
}

function calcPlannedLoanAmount() {
  return state.decisions
    .filter((item) => item.type === "long_loan" || item.type === "short_loan")
    .reduce((sum, item) => sum + Number(item.amount || 0), 0);
}

function renderCompanyMetrics() {
  const company = currentCompanyState();
  $("#metricCash").textContent = money(company.cash);
  $("#metricEquity").textContent = money(company.equity);
  $("#metricLoan").textContent = money(calcLoanAvailable(company));
}

async function api(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `HTTP ${response.status}`);
  }
  return response.json();
}

async function checkHealth() {
  try {
    await api("/health");
    $("#apiDot").className = "status-dot ok";
    await Promise.all([loadAIStatus(), loadDBStatus()]);
    renderSystemStatus();
  } catch (error) {
    $("#apiDot").className = "status-dot bad";
    $("#apiStatus").textContent = "后端离线";
  }
}

async function loadAIStatus() {
  try {
    const data = await api("/api/v1/ai/status");
    state.aiStatus = data;
  } catch (error) {
    state.aiStatus = { error: "AI状态未知" };
  }
}

async function loadDBStatus() {
  try {
    const data = await api("/api/v1/db/status");
    state.dbStatus = data;
  } catch (error) {
    state.dbStatus = { error: "数据库状态未知", storageMode: "unknown" };
  }
}

function renderSystemStatus() {
  const ai = state.aiStatus || {};
  const db = state.dbStatus || {};
  const parts = ["后端在线"];
  const titleParts = [];

  if (ai.error) {
    parts.push(ai.error);
  } else if (!ai.enabled) {
    parts.push("AI未配置");
  } else if (ai.warning) {
    parts.push("AI配置警告");
    titleParts.push(ai.warning);
  } else {
    parts.push("AI已配置");
  }

  if (db.error) {
    parts.push("存储状态未知");
    titleParts.push(db.error);
  } else if (db.storageMode === "mysql") {
    parts.push("存储MySQL");
  } else {
    parts.push("存储内存");
  }

  if (ai.model) {
    titleParts.push(`AI模型：${ai.model}`);
  }
  if (ai.keyPreview) {
    titleParts.push(`${ai.keySource || "API Key"} ${ai.keyPreview}`);
  }
  if (ai.chatCompletionsURL || ai.baseURL) {
    titleParts.push(ai.chatCompletionsURL || ai.baseURL);
  }
  if (db.configPath) {
    titleParts.push(`DB配置：${db.configPath}${db.configLoaded ? " 已加载" : " 未加载"}`);
  }

  $("#apiStatus").textContent = parts.join(" · ");
  $("#apiStatus").title = titleParts.join("\n");
}

async function loadFlow() {
  const data = await api("/api/v1/operation-flow");
  state.flow = data.steps || [];
  renderFlow();
  renderStepSelect();
}

async function loadOperationRecords() {
  const company = encodeURIComponent($("#recordCompany")?.value.trim() || "默认企业");
  const data = await api(`/api/v1/operation-records?company=${company}`);
  state.operationRecords = data.records || [];
  state.historyAnalysis = data.analysis || null;
  renderOperationRecords();
}

async function loadAdvisorHistory() {
  const company = encodeURIComponent($("#recordCompany")?.value.trim() || "默认企业");
  const data = await api(`/api/v1/advisor/history?company=${company}&limit=20`);
  state.adviceHistory = data.records || [];
  renderQAHistory();
}

function renderStepSelect() {
  const select = $("#stepSelect");
  select.innerHTML = state.flow.map((step) => (
    `<option value="${escapeHtml(step.code)}">${escapeHtml(step.name)}</option>`
  )).join("");
  select.value = "ad_investment";
  updateStepSummary();
}

function updateStepSummary() {
  const step = state.flow.find((item) => item.code === $("#stepSelect").value);
  $("#stepSummary").textContent = step ? `${step.stage} · ${step.name}` : "未选择流程";
}

function renderFlow() {
  $("#flowCount").textContent = `${state.flow.length} 个节点`;
  $("#flowList").innerHTML = state.flow.map((step) => {
    const q4 = step.onlyQ4 ? `<span class="tag warn">仅Q4</span>` : "";
    return `
      <article class="flow-item">
        <div class="flow-title">
          <span>${escapeHtml(step.name)}</span>
          ${q4 || `<span class="tag">${escapeHtml(step.stage)}</span>`}
        </div>
        <p>${escapeHtml(step.description || "")}</p>
      </article>
    `;
  }).join("");
}

function renderDecisions() {
  const rows = $("#decisionRows");
  if (!state.decisions.length) {
    rows.innerHTML = emptyRow(6);
    return;
  }
  rows.innerHTML = state.decisions.map((item, index) => `
    <tr>
      <td>${escapeHtml(decisionName(item.type))}</td>
      <td>${escapeHtml(decisionTarget(item))}</td>
      <td>${money(item.amount)}</td>
      <td>${item.quantity || "-"}</td>
      <td>${escapeHtml(item.note || "-")}</td>
      <td><button class="row-action" type="button" data-remove-decision="${index}">删除</button></td>
    </tr>
  `).join("");
}

function renderOrders() {
  const rows = $("#orderRows");
  if (!state.orders.length) {
    rows.innerHTML = emptyRow(12);
    return;
  }
  rows.innerHTML = state.orders.map((item, index) => `
    <tr>
      <td>${escapeHtml(item.id)}</td>
      <td>Y${item.year || numberValue("#yearInput") || 1}</td>
      <td>${escapeHtml(item.market)}</td>
      <td>${escapeHtml(item.product)}</td>
      <td>${item.quantity}</td>
      <td>${money(item.totalPrice)}</td>
      <td>Q${item.deliveryQuarter || "-"}</td>
      <td>${paymentPeriodLabel(item.paymentPeriod)}</td>
      <td>${escapeHtml(item.isoRequired || "-")}</td>
      <td>${escapeHtml(item.deliveryTime || deliveryTimeLabel(item))}</td>
      <td>${escapeHtml(item.collectionTime || collectionTimeLabel(item))}</td>
      <td><button class="row-action" type="button" data-remove-order="${index}">删除</button></td>
    </tr>
  `).join("");
}

function renderCashFlowQuarterSelect() {
  const select = $("#cashFlowQuarter");
  if (!select) return;
  const start = quarterIndexFromYearQuarter(numberValue("#yearInput"), numberValue("#quarterInput"));
  select.innerHTML = Array.from({ length: 28 }, (_, index) => {
    const quarterIndex = start + index;
    return `<option value="${quarterIndex}">${quarterLabel(quarterIndex)}</option>`;
  }).join("");
}

function renderCashFlowPoints() {
  const rows = $("#cashFlowPointRows");
  if (!rows) return;
  if (!state.cashFlowPoints.length) {
    rows.innerHTML = emptyRow(7);
  } else {
    const items = sortedCashFlowPoints();
    rows.innerHTML = items.map((item) => {
      const index = state.cashFlowPoints.indexOf(item);
      return `
      <tr>
        <td>${quarterLabel(item.quarterIndex)}</td>
        <td>${money(item.openingCash)}</td>
        <td>${money(item.inflows)}</td>
        <td>${money(item.outflows)}</td>
        <td>${money(item.closingCash)}</td>
        <td>${escapeHtml(item.note || "-")}</td>
        <td><button class="row-action" type="button" data-remove-cash-flow="${index}">删除</button></td>
      </tr>
    `;
    }).join("");
  }
  updateCashFlowSummary();
}

function renderProductionStatus() {
  renderFactoryOptions();
  renderFactoryBoard();
  renderProductionBoard();
  renderProductionLineRows();
}

function renderFactoryOptions() {
  const options = state.factories.length
    ? state.factories.map((factory) => `<option value="${escapeHtml(factory.id)}">${escapeHtml(factory.id)} · ${escapeHtml(factory.type)}</option>`).join("")
    : '<option value="">暂无厂房</option>';
  if ($("#lineFactoryInput")) $("#lineFactoryInput").innerHTML = options;
}

function renderFactoryBoard() {
  const board = $("#factoryBoard");
  if (!board) return;
  if (!state.factories.length) {
    board.innerHTML = '<div class="empty qa-empty">暂无厂房，请先添加租用或购买的厂房。</div>';
    return;
  }
  board.innerHTML = state.factories.map((factory, index) => {
    const used = state.productionLines.filter((line) => line.factoryId === factory.id && line.status !== "sold").length;
    const rule = factoryRules[factory.type] || { capacity: 0 };
    const utilization = rule.capacity > 0 ? Math.min(100, Math.round(used / rule.capacity * 100)) : 0;
    return `
      <article class="factory-card">
        <div class="factory-card-head">
          <div>
            <strong>${escapeHtml(factory.id)}</strong>
            <span>${escapeHtml(factory.type)} · ${escapeHtml(factory.ownership || "-")}</span>
          </div>
          <button class="row-action" type="button" data-remove-factory="${index}">删除</button>
        </div>
        <div class="capacity-line">
          <span>容量 ${used}/${rule.capacity || 0} 条</span>
          <span>${utilization}%</span>
        </div>
        <div class="progress-track"><span style="width:${utilization}%"></span></div>
        <div class="factory-meta">租金 ${money(rule.rent || 0)} / 买价 ${money(rule.buyCost || 0)} · ${quarterLabel(factory.startQuarterIndex || 1)}</div>
      </article>
    `;
  }).join("");
}

function renderProductionBoard() {
  const board = $("#productionBoard");
  if (!board) return;
  if (!state.productionLines.length) {
    board.innerHTML = '<div class="empty qa-empty">暂无生产线。</div>';
    $("#productionSummary").textContent = "暂无生产线";
    return;
  }
  const built = state.productionLines.filter((line) => line.status === "built" && line.built).length;
  const producing = state.productionLines.filter((line) => line.isProducing && line.status === "built").length;
  const building = state.productionLines.filter((line) => line.status === "building").length;
  const switching = state.productionLines.filter((line) => line.status === "switching").length;
  $("#productionSummary").textContent = `厂房 ${state.factories.length} 个 · 生产线 ${state.productionLines.length} 条 · 已建成 ${built} 条 · 生产中 ${producing} 条`;
  board.innerHTML = state.productionLines.map((line) => {
    const total = Number(line.productionTotal || lineRules[line.type]?.cycle || 1);
    const progress = Math.max(0, Math.min(total, Number(line.productionProgress || 0)));
    const percent = total > 0 ? Math.round(progress / total * 100) : 0;
    const status = productionLineStatusName(line);
    const factory = state.factories.find((item) => item.id === line.factoryId);
    return `
      <article class="line-card ${line.status || "built"}">
        <div class="line-card-top">
          <div>
            <strong>${escapeHtml(line.id)}</strong>
            <span>${escapeHtml(line.type)} · ${escapeHtml(line.product || "-")}</span>
          </div>
          <span class="tag ${lineStatusTagClass(line)}">${escapeHtml(status)}</span>
        </div>
        <div class="line-meta">
          <span>${escapeHtml(factory ? `${factory.id} ${factory.type}` : "未分配厂房")}</span>
          <span>净值 ${money(line.netValue)}</span>
        </div>
        <div class="capacity-line">
          <span>${line.isProducing ? `生产 ${escapeHtml(line.currentProductionProduct || line.product || "-")}` : "空闲/未开工"}</span>
          <span>${progress}/${total} 期</span>
        </div>
        <div class="progress-track"><span style="width:${percent}%"></span></div>
      </article>
    `;
  }).join("");
}

function renderProductionLineRows() {
  const rows = $("#productionLineRows");
  if (!rows) return;
  if (!state.productionLines.length) {
    rows.innerHTML = emptyRow(8);
    return;
  }
  rows.innerHTML = state.productionLines.map((line, index) => `
    <tr>
      <td>${escapeHtml(line.id)}</td>
      <td>${escapeHtml(line.factoryId || "-")}</td>
      <td>${escapeHtml(line.type)}</td>
      <td>${escapeHtml(line.product || "-")}</td>
      <td>${escapeHtml(productionLineStatusName(line))}</td>
      <td>${line.isProducing ? `${line.productionProgress || 0}/${line.productionTotal || lineRules[line.type]?.cycle || 1}期` : "-"}</td>
      <td>${money(line.netValue)}</td>
      <td><button class="row-action" type="button" data-remove-line="${index}">删除</button></td>
    </tr>
  `).join("");
}

function renderOperationRecords() {
  const records = [...state.operationRecords].sort((a, b) => Number(a.year || 0) - Number(b.year || 0));
  if (!records.length) {
    $("#recordStorageSummary").textContent = "暂无已保存报表，填写当前页后点击保存年度。";
    $("#recordSavedList").innerHTML = "";
    $("#recordPrevBtn").disabled = true;
    $("#recordNextBtn").disabled = true;
    state.selectedRecordYear = numberValue("#recordYear") || 1;
    syncRecordMirrors();
    return;
  }

  const years = records.map((record) => Number(record.year || 0)).filter(Boolean);
  if (!years.includes(Number(state.selectedRecordYear))) {
    state.selectedRecordYear = years[years.length - 1];
  }

  const selectedIndex = years.indexOf(Number(state.selectedRecordYear));
  $("#recordStorageSummary").textContent = `已保存 ${records.length} 年报表，当前翻阅第 ${state.selectedRecordYear} 年。`;
  $("#recordSavedList").innerHTML = records.map((record) => {
    const year = Number(record.year || 0);
    const active = year === Number(state.selectedRecordYear) ? " active" : "";
    return `<button class="record-year-btn${active}" type="button" data-record-year="${year}">第${year}年</button>`;
  }).join("");
  $("#recordPrevBtn").disabled = selectedIndex <= 0;
  $("#recordNextBtn").disabled = selectedIndex >= years.length - 1;

  const selectedRecord = records.find((record) => Number(record.year) === Number(state.selectedRecordYear));
  if (selectedRecord) {
    fillRecordForm(selectedRecord);
  }
  syncRecordMirrors();
}

function fillRecordForm(record) {
  if (!record) return;
  const expense = record["综合费用表"] || {};
  const profit = record["利润表"] || {};
  const balance = record["资产负债表"] || {};

  $("#recordCompany").value = record.company || "默认企业";
  $("#recordYear").value = record.year || 1;

  setNumberValue("#recordManagementExpense", expense["管理费"]);
  setNumberValue("#recordAdExpense", expense["广告费"]);
  setNumberValue("#recordMaintenanceExpense", expense["设备维护费"]);
  setNumberValue("#recordSwitchExpense", expense["转产费"]);
  setNumberValue("#recordRentExpense", expense["租金"]);
  setNumberValue("#recordMarketExpense", expense["市场准入开拓"]);
  setNumberValue("#recordRnDExpense", expense["产品研发"]);
  setNumberValue("#recordISOExpense", expense["ISO 认证资格"]);
  setNumberValue("#recordInfoExpense", expense["信息费"]);
  setNumberValue("#recordOtherLoss", expense["其他"]);
  setNumberValue("#recordComprehensiveExpense", expense["合计"]);

  setNumberValue("#recordSalesRevenue", profit["销售收入"]);
  setNumberValue("#recordDirectCost", profit["直接成本"]);
  setNumberValue("#recordGrossProfit", profit["毛利"]);
  setNumberValue("#recordProfitComprehensiveExpense", profit["综合费用"]);
  setNumberValue("#recordProfitBeforeDepreciation", profit["折旧前利润"]);
  setNumberValue("#recordDepreciation", profit["折旧"]);
  setNumberValue("#recordProfitBeforeInterest", profit["支付利息前利润"]);
  setNumberValue("#recordFinancialExpense", profit["财务费用"]);
  setNumberValue("#recordProfitBeforeTax", profit["税前利润"]);
  setNumberValue("#recordIncomeTax", profit["所得税"]);
  setNumberValue("#recordNetProfit", profit["年度净利润"]);

  setNumberValue("#recordClosingCash", balance["现金"]);
  setNumberValue("#recordReceivables", balance["应收款"]);
  setNumberValue("#recordWorkInProgress", balance["在制品"]);
  setNumberValue("#recordFinishedGoods", balance["产成品"]);
  setNumberValue("#recordRawMaterials", balance["原材料"]);
  setNumberValue("#recordCurrentAssets", balance["流动资产合计"]);
  setNumberValue("#recordFactoryValue", balance["厂房"]);
  setNumberValue("#recordProductionLineValue", balance["生产线"]);
  setNumberValue("#recordConstructionInProgress", balance["在建工程"]);
  setNumberValue("#recordFixedAssets", balance["固定资产合计"]);
  setNumberValue("#recordTotalAssets", balance["资产总计"]);
  setNumberValue("#recordLongLoanBalance", balance["长期负债"]);
  setNumberValue("#recordShortLoanBalance", balance["短期负债"]);
  setNumberValue("#recordIncomeTaxPayable", balance["应交所得税"]);
  setNumberValue("#recordTotalLiability", balance["负债合计"]);
  setNumberValue("#recordShareCapital", balance["股东资本"]);
  setNumberValue("#recordRetainedEarnings", balance["利润留存"]);
  setNumberValue("#recordBalanceNetProfit", balance["年度净利"]);
  setNumberValue("#recordOwnerEquityTotal", balance["所有者权益合计"]);
  setNumberValue("#recordLiabilityEquityTotal", balance["负债和所有者权益总计"]);
}

function setNumberValue(selector, value) {
  const input = $(selector);
  if (!input || value === undefined || value === null) return;
  input.value = Number(value || 0);
}

function turnRecordPage(direction) {
  const years = [...state.operationRecords]
    .map((record) => Number(record.year || 0))
    .filter(Boolean)
    .sort((a, b) => a - b);
  const index = years.indexOf(Number(state.selectedRecordYear));
  const nextYear = years[index + direction];
  if (!nextYear) return;
  animateRecordPageTurn(direction, () => {
    state.selectedRecordYear = nextYear;
    renderOperationRecords();
  });
}

function selectRecordYear(year) {
  const nextYear = Number(year);
  if (!nextYear || nextYear === Number(state.selectedRecordYear)) return;
  const direction = nextYear > Number(state.selectedRecordYear) ? 1 : -1;
  animateRecordPageTurn(direction, () => {
    state.selectedRecordYear = nextYear;
    renderOperationRecords();
  });
}

function animateRecordPageTurn(direction, updatePage) {
  const paper = $(".record-paper");
  if (!paper || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    updatePage();
    return;
  }

  const outClass = direction >= 0 ? "page-turn-next-out" : "page-turn-prev-out";
  const inClass = direction >= 0 ? "page-turn-next-in" : "page-turn-prev-in";
  paper.classList.remove("page-turn-next-out", "page-turn-prev-out", "page-turn-next-in", "page-turn-prev-in");
  void paper.offsetWidth;
  paper.classList.add(outClass);
  window.setTimeout(() => {
    updatePage();
    paper.classList.remove(outClass);
    paper.classList.add(inClass);
    window.setTimeout(() => {
      paper.classList.remove(inClass);
    }, RECORD_PAGE_TURN_IN_MS);
  }, RECORD_PAGE_TURN_OUT_MS);
}

function renderQAHistory() {
  const list = $("#qaHistoryList");
  if (!list) return;
  if (!state.adviceHistory.length) {
    list.innerHTML = '<div class="empty qa-empty">暂无建议记录</div>';
    return;
  }
  list.innerHTML = state.adviceHistory.map((item) => `
    <article class="qa-item">
      <div class="qa-meta">
        <span>Y${item.year || "-"}Q${item.quarter || "-"}</span>
        <span>${escapeHtml(item.stepName || "-")}</span>
        <span>${item.aiUsed ? "AI" : "规则"}</span>
        <span>${formatTime(item.createdAt)}</span>
      </div>
      <div class="qa-question">Q：${escapeHtml(item.question || "未填写具体问题")}</div>
      <div class="qa-answer">A：${escapeHtml(item.answer || "暂无回答")}</div>
    </article>
  `).join("");
}

function addDecision() {
  const type = $("#decisionType").value;
  const config = decisionFormConfig[type] || decisionFormConfig.ad;
  const decision = {
    type,
    year: numberValue("#yearInput"),
    quarter: numberValue("#quarterInput"),
  };

  if (config.fields.includes("market")) {
    decision.market = $("#decisionMarket").value;
  }
  if (config.fields.includes("product")) {
    decision.product = $("#decisionProduct").value;
  }
  if (config.fields.includes("material")) {
    decision.material = $("#decisionMaterial").value;
  }
  if (config.fields.includes("lineType")) {
    decision.lineType = $("#decisionLineType").value;
  }
  if (config.fields.includes("factoryType") && $("#decisionFactoryType").value) {
    decision.factoryType = $("#decisionFactoryType").value;
  }
  if (config.fields.includes("amount")) {
    decision.amount = numberValue("#decisionAmount");
  }
  if (config.fields.includes("quantity")) {
    decision.quantity = numberValue("#decisionQuantity");
  }
  if (config.fields.includes("isoType") && $("#decisionISOType").value) {
    decision.extra = { isoType: $("#decisionISOType").value };
  }
  const note = $("#decisionNote").value.trim();
  if (note) {
    decision.note = note;
  }

  state.decisions.push(decision);
  renderDecisions();
  renderCompanyMetrics();
}

function addOrder() {
  const order = {
    id: $("#orderId").value.trim() || `O${state.orders.length + 1}`,
    year: numberValue("#orderYear"),
    market: $("#orderMarket").value,
    product: $("#orderProduct").value,
    quantity: numberValue("#orderQuantity"),
    totalPrice: numberValue("#orderPrice"),
    deliveryQuarter: numberValue("#orderDelivery"),
    paymentPeriod: numberValue("#orderPaymentPeriod"),
    isoRequired: $("#orderISORequired").value,
    deliveryTime: $("#orderDeliveryTime").value.trim(),
    collectionTime: $("#orderCollectionTime").value.trim(),
  };
  state.orders.push({
    ...order,
    deliveryTime: order.deliveryTime || deliveryTimeLabel(order),
    collectionTime: order.collectionTime || collectionTimeLabel(order),
  });
  renderOrders();
  refreshCashFlow();
}

function addCashFlowPoint() {
  const point = {
    quarterIndex: numberValue("#cashFlowQuarter"),
    openingCash: numberValue("#cashFlowOpening"),
    inflows: numberValue("#cashFlowInflow"),
    outflows: numberValue("#cashFlowOutflow"),
    closingCash: numberValue("#cashFlowClosing"),
    note: $("#cashFlowNote").value.trim(),
  };
  const exists = state.cashFlowPoints.findIndex((item) => item.quarterIndex === point.quarterIndex);
  if (exists >= 0) {
    state.cashFlowPoints[exists] = point;
  } else {
    state.cashFlowPoints.push(point);
  }
  renderCashFlowPoints();
  refreshCashFlow();
  $("#cashFlowOpening").value = point.closingCash;
  $("#cashFlowInflow").value = 0;
  $("#cashFlowOutflow").value = 0;
  $("#cashFlowClosing").value = point.closingCash;
  $("#cashFlowNote").value = "";
}

function addFactory() {
  const id = $("#factoryIdInput").value.trim() || `F${state.factories.length + 1}`;
  const ownership = $("#factoryOwnershipInput").value;
  const factory = {
    id,
    type: $("#factoryTypeInput").value,
    ownership,
    rented: ownership === "租用",
    purchased: ownership === "购买",
    startQuarterIndex: numberValue("#factoryStartQuarterInput"),
  };
  const exists = state.factories.findIndex((item) => item.id === id);
  if (exists >= 0) {
    state.factories[exists] = factory;
  } else {
    state.factories.push(factory);
  }
  renderProductionStatus();
  $("#factoryIdInput").value = `F${state.factories.length + 1}`;
}

function addProductionLine() {
  const type = $("#lineTypeInput").value;
  const status = $("#lineStatusInput").value;
  const producing = $("#lineProducingInput").value === "true";
  const rule = lineRules[type] || { cycle: 1 };
  const line = {
    id: $("#lineIdInput").value.trim() || `L${state.productionLines.length + 1}`,
    factoryId: $("#lineFactoryInput").value,
    type,
    product: $("#lineProductInput").value,
    built: status === "built",
    status,
    builtAtQuarterIndex: numberValue("#lineBuiltAtInput"),
    buildStartedQuarterIndex: numberValue("#lineBuiltAtInput") - Number(rule.installQuarters || 0),
    netValue: numberValue("#lineNetValueInput"),
    isProducing: producing,
    productionStartedQuarterIndex: numberValue("#lineBuiltAtInput"),
    productionProgress: numberValue("#lineProgressInput"),
    productionTotal: numberValue("#lineTotalInput") || Number(rule.cycle || 1),
    currentProductionProduct: $("#lineProductInput").value,
  };
  const exists = state.productionLines.findIndex((item) => item.id === line.id);
  if (exists >= 0) {
    state.productionLines[exists] = line;
  } else {
    state.productionLines.push(line);
  }
  renderProductionStatus();
  $("#lineIdInput").value = `L${state.productionLines.length + 1}`;
}

function sortedCashFlowPoints() {
  return [...state.cashFlowPoints].sort((a, b) => Number(a.quarterIndex || 0) - Number(b.quarterIndex || 0));
}

function cashWarningFromClosingCash(value) {
  const cash = Number(value || 0);
  if (cash <= 0) return "critical";
  if (cash <= 10) return "danger";
  if (cash <= 30) return "warning";
  return "safe";
}

function cashTrendItems() {
  return sortedCashFlowPoints().map((item) => ({
    year: Math.floor((Number(item.quarterIndex || 1) - 1) / 4) + 1,
    quarter: ((Number(item.quarterIndex || 1) - 1) % 4) + 1,
    quarterIndex: Number(item.quarterIndex || 1),
    openingCash: Number(item.openingCash || 0),
    inflows: Number(item.inflows || 0),
    outflows: Number(item.outflows || 0),
    netFlow: Number(item.inflows || 0) - Number(item.outflows || 0),
    closingCash: Number(item.closingCash || 0),
    warningLevel: cashWarningFromClosingCash(item.closingCash),
    note: item.note || "",
  }));
}

function updateCashFlowSummary() {
  const items = cashTrendItems();
  if (!$("#cashFlowSummary")) return;
  if (!items.length) {
    $("#cashFlowSummary").textContent = "录入季度现金后生成走势图";
    return;
  }
  const minItem = items.reduce((min, item) => item.closingCash < min.closingCash ? item : min, items[0]);
  const latest = items[items.length - 1];
  $("#cashFlowSummary").textContent = `已录入 ${items.length} 个季度；最新现金 ${money(latest.closingCash)}，最低现金 ${quarterLabel(minItem.quarterIndex)} ${money(minItem.closingCash)}。`;
}

function syncCashClosingFromFlow() {
  const opening = numberValue("#cashFlowOpening");
  const inflow = numberValue("#cashFlowInflow");
  const outflow = numberValue("#cashFlowOutflow");
  $("#cashFlowClosing").value = opening + inflow - outflow;
}

function prefillCashFlowFormFromLatest() {
  const items = sortedCashFlowPoints();
  if (!items.length) return;
  const latest = items[items.length - 1];
  const nextQuarter = Number(latest.quarterIndex || 1) + 1;
  if ($("#cashFlowQuarter")) {
    const option = Array.from($("#cashFlowQuarter").options).find((item) => Number(item.value) === nextQuarter);
    if (option) $("#cashFlowQuarter").value = nextQuarter;
  }
  $("#cashFlowOpening").value = latest.closingCash || 0;
  $("#cashFlowInflow").value = 0;
  $("#cashFlowOutflow").value = 0;
  $("#cashFlowClosing").value = latest.closingCash || 0;
}

function renderProductionQuarterSelects() {
  const start = quarterIndexFromYearQuarter(numberValue("#yearInput"), numberValue("#quarterInput"));
  const options = Array.from({ length: 28 }, (_, index) => {
    const quarterIndex = start + index;
    return `<option value="${quarterIndex}">${quarterLabel(quarterIndex)}</option>`;
  }).join("");
  if ($("#factoryStartQuarterInput")) $("#factoryStartQuarterInput").innerHTML = options;
  if ($("#lineBuiltAtInput")) $("#lineBuiltAtInput").innerHTML = options;
}

function productionLineStatusName(line) {
  if (line.status === "building") return "在建";
  if (line.status === "switching") return "转产中";
  if (line.status === "sold") return "已出售";
  if (line.isProducing) return "生产中";
  return "已建成";
}

function lineStatusTagClass(line) {
  if (line.status === "built" && line.isProducing) return "ok";
  if (line.status === "built") return "";
  if (line.status === "building" || line.status === "switching") return "warn";
  return "danger";
}

async function saveOperationRecord() {
  $("#saveRecordBtn").disabled = true;
  $("#saveRecordBtn").textContent = "保存中";
  try {
    syncRecordMirrors();
    const record = {
      company: $("#recordCompany").value.trim() || "默认企业",
      year: numberValue("#recordYear"),
      综合费用表: {
        管理费: numberValue("#recordManagementExpense"),
        广告费: numberValue("#recordAdExpense"),
        设备维护费: numberValue("#recordMaintenanceExpense"),
        转产费: numberValue("#recordSwitchExpense"),
        租金: numberValue("#recordRentExpense"),
        市场准入开拓: numberValue("#recordMarketExpense"),
        产品研发: numberValue("#recordRnDExpense"),
        "ISO 认证资格": numberValue("#recordISOExpense"),
        信息费: numberValue("#recordInfoExpense"),
        其他: numberValue("#recordOtherLoss"),
        合计: numberValue("#recordComprehensiveExpense"),
      },
      利润表: {
        销售收入: numberValue("#recordSalesRevenue"),
        直接成本: numberValue("#recordDirectCost"),
        毛利: numberValue("#recordGrossProfit"),
        综合费用: numberValue("#recordProfitComprehensiveExpense"),
        折旧前利润: numberValue("#recordProfitBeforeDepreciation"),
        折旧: numberValue("#recordDepreciation"),
        支付利息前利润: numberValue("#recordProfitBeforeInterest"),
        财务费用: numberValue("#recordFinancialExpense"),
        税前利润: numberValue("#recordProfitBeforeTax"),
        所得税: numberValue("#recordIncomeTax"),
        年度净利润: numberValue("#recordNetProfit"),
      },
      资产负债表: {
        现金: numberValue("#recordClosingCash"),
        应收款: numberValue("#recordReceivables"),
        在制品: numberValue("#recordWorkInProgress"),
        产成品: numberValue("#recordFinishedGoods"),
        原材料: numberValue("#recordRawMaterials"),
        流动资产合计: numberValue("#recordCurrentAssets"),
        厂房: numberValue("#recordFactoryValue"),
        生产线: numberValue("#recordProductionLineValue"),
        在建工程: numberValue("#recordConstructionInProgress"),
        固定资产合计: numberValue("#recordFixedAssets"),
        资产总计: numberValue("#recordTotalAssets"),
        长期负债: numberValue("#recordLongLoanBalance"),
        短期负债: numberValue("#recordShortLoanBalance"),
        应交所得税: numberValue("#recordIncomeTaxPayable"),
        负债合计: numberValue("#recordTotalLiability"),
        股东资本: numberValue("#recordShareCapital"),
        利润留存: numberValue("#recordRetainedEarnings"),
        年度净利: numberValue("#recordBalanceNetProfit"),
        所有者权益合计: numberValue("#recordOwnerEquityTotal"),
        负债和所有者权益总计: numberValue("#recordLiabilityEquityTotal"),
      },
    };
    const data = await api("/api/v1/operation-records", {
      method: "POST",
      body: JSON.stringify(record),
    });
    state.operationRecords = data.records || [];
    state.historyAnalysis = data.analysis || null;
    state.selectedRecordYear = record.year;
    renderOperationRecords();
  } catch (error) {
    window.alert(`保存失败：${error.message}`);
  } finally {
    $("#saveRecordBtn").disabled = false;
    $("#saveRecordBtn").textContent = "保存年度";
  }
}

function bindRecordMirror(sourceSelector, targetSelector) {
  const source = $(sourceSelector);
  const target = $(targetSelector);
  if (!source || !target) return;
  source.addEventListener("input", () => {
    target.value = source.value;
  });
}

function syncRecordMirrors() {
  const syncPairs = [
    ["#recordComprehensiveExpense", "#recordProfitComprehensiveExpense"],
    ["#recordNetProfit", "#recordBalanceNetProfit"],
    ["#recordTotalAssets", "#recordLiabilityEquityTotal"],
  ];
  syncPairs.forEach(([sourceSelector, targetSelector]) => {
    const source = $(sourceSelector);
    const target = $(targetSelector);
    if (source && target) {
      target.value = source.value;
    }
  });
}

async function deleteOperationRecord(year) {
  const company = encodeURIComponent($("#recordCompany").value.trim() || "默认企业");
  const data = await api(`/api/v1/operation-records?company=${company}&year=${year}`, { method: "DELETE" });
  state.operationRecords = data.records || [];
  state.historyAnalysis = data.analysis || null;
  renderOperationRecords();
}

async function clearAdvisorHistory() {
  const company = encodeURIComponent($("#recordCompany").value.trim() || "默认企业");
  const data = await api(`/api/v1/advisor/history?company=${company}`, { method: "DELETE" });
  state.adviceHistory = data.records || [];
  renderQAHistory();
}

async function runAdvisor() {
  $("#runAdvisorBtn").disabled = true;
  $("#runAdvisorBtn").textContent = "分析中";
  $("#adviceBox").textContent = "正在计算规则诊断并请求建议。";
  try {
    const payload = {
      stepCode: $("#stepSelect").value,
      question: $("#questionInput").value.trim(),
      state: currentCompanyState(),
      decisions: state.decisions,
      orders: state.orders,
      competitorAds: [5, 4, 3],
      operationRecords: state.operationRecords,
      adviceHistory: state.adviceHistory.slice(0, 5),
    };
    const data = await api("/api/v1/advisor/decision", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    state.lastAdvisor = data;
    state.adviceHistory = data.qaHistory || state.adviceHistory;
    renderAdvisor(data);
    renderQAHistory();
    updateMetrics(data);
    switchView($(".nav-item.active")?.dataset.view || "advisor");
  } catch (error) {
    $("#adviceBox").textContent = `请求失败：${error.message}`;
  } finally {
    $("#runAdvisorBtn").disabled = false;
    $("#runAdvisorBtn").textContent = "生成建议";
  }
}

function renderAdvisor(data) {
  const usedAI = data.mode === "ai" || data.aiUsed;
  $("#advisorMode").textContent = usedAI ? "AI建议" : "规则建议";
  $("#metricMode").textContent = usedAI ? "AI" : "规则";
  $("#adviceBox").textContent = data.advice || "暂无建议";
  if (data.aiError) {
    $("#advisorMode").textContent = "规则兜底";
    $("#advisorMode").title = data.aiError;
    $("#aiRunStatus").className = "ai-run-status warn";
    $("#aiRunStatus").textContent = `本次调用：AI已配置，但调用失败，已用规则兜底。原因：${data.aiError}`;
  } else {
    $("#advisorMode").title = usedAI ? "本次回答来自 AI" : "";
    $("#aiRunStatus").className = `ai-run-status ${usedAI ? "ok" : ""}`;
    $("#aiRunStatus").textContent = usedAI
      ? `本次调用：AI成功，模式 ${data.mode || "ai"}。`
      : `本次调用：未使用AI，模式 ${data.mode || "rule"}。`;
  }

  const risks = data.diagnostics?.riskFindings || [];
  $("#riskList").innerHTML = risks.length
    ? risks.map((item) => `<li>${escapeHtml(item)}</li>`).join("")
    : "<li>未发现硬规则冲突</li>";

  refreshCashFlow();
  renderScores(data.diagnostics?.orderScores || []);
  renderPurchasePlan(data.diagnostics?.purchasePlan || []);
}

function updateMetrics(data) {
  const company = currentCompanyState();
  $("#metricCash").textContent = money(company.cash);
  $("#metricEquity").textContent = money(company.equity);
  const currentAvailable = data.diagnostics?.loanCapacity?.available;
  $("#metricLoan").textContent = money(currentAvailable === undefined ? calcLoanAvailable(company) : currentAvailable);
}

function renderCashFlow(items) {
  const rows = $("#cashRows");
  rows.innerHTML = items.length ? items.map((item) => `
    <tr>
      <td>Y${item.year}Q${item.quarter}</td>
      <td>${money(item.closingCash)}</td>
      <td>${money(item.inflows)}</td>
      <td>${money(item.outflows)}</td>
      <td>${money(item.netFlow)}</td>
      <td>${warningTag(item.warningLevel)}</td>
    </tr>
  `).join("") : emptyRow(6);
  drawCashChart(items);
}

function refreshCashFlow() {
  const rows = $("#cashRows");
  if (!rows) return;
  const items = cashTrendItems();
  renderCashFlow(items);
  renderCashFlowPoints();
}

function renderScores(items) {
  const rows = $("#scoreRows");
  rows.innerHTML = items.length ? items.map((item) => `
    <tr>
      <td>${escapeHtml(item.order.id)}</td>
      <td>${item.feasible ? '<span class="tag ok">可行</span>' : '<span class="tag danger">不可行</span>'}</td>
      <td>${item.score}</td>
      <td>${money(item.profit)}</td>
      <td>${escapeHtml(item.reason || "-")}</td>
    </tr>
  `).join("") : emptyRow(5);
}

function renderPurchasePlan(items) {
  const rows = $("#purchaseRows");
  rows.innerHTML = items.length ? items.map((item) => `
    <tr>
      <td>${escapeHtml(item.material)}</td>
      <td>${item.quantity}</td>
      <td>${quarterLabel(item.orderQuarter)}</td>
      <td>${quarterLabel(item.needQuarter)}</td>
      <td>${money(item.cost)}</td>
    </tr>
  `).join("") : emptyRow(5);
}

function drawCashChart(items) {
  const canvas = $("#cashChart");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  const rect = canvas.getBoundingClientRect();
  const ratio = window.devicePixelRatio || 1;
  canvas.width = Math.max(320, Math.floor(rect.width * ratio));
  canvas.height = Math.floor(320 * ratio);
  ctx.scale(ratio, ratio);

  const width = rect.width;
  const height = 320;
  ctx.clearRect(0, 0, width, height);
  ctx.fillStyle = "#fbfcfd";
  ctx.fillRect(0, 0, width, height);

  const padding = { left: 48, right: 18, top: 22, bottom: 42 };
  const values = items.map((item) => item.closingCash);
  const min = Math.min(0, ...values);
  const max = Math.max(60, ...values);
  const span = Math.max(1, max - min);
  const xStep = items.length > 1 ? (width - padding.left - padding.right) / (items.length - 1) : 0;

  ctx.strokeStyle = "#d9e0e5";
  ctx.lineWidth = 1;
  for (let i = 0; i < 5; i += 1) {
    const y = padding.top + ((height - padding.top - padding.bottom) / 4) * i;
    line(ctx, padding.left, y, width - padding.right, y);
  }

  const zeroY = yFor(0);
  ctx.strokeStyle = "#b91c1c";
  ctx.setLineDash([5, 5]);
  line(ctx, padding.left, zeroY, width - padding.right, zeroY);
  ctx.setLineDash([]);

  ctx.strokeStyle = "#0f766e";
  ctx.lineWidth = 3;
  ctx.beginPath();
  items.forEach((item, index) => {
    const x = padding.left + xStep * index;
    const y = yFor(item.closingCash);
    if (index === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.stroke();

  ctx.fillStyle = "#0f766e";
  items.forEach((item, index) => {
    const x = padding.left + xStep * index;
    const y = yFor(item.closingCash);
    ctx.beginPath();
    ctx.arc(x, y, 4, 0, Math.PI * 2);
    ctx.fill();
  });

  ctx.fillStyle = "#65717c";
  ctx.font = "12px sans-serif";
  items.forEach((item, index) => {
    const x = padding.left + xStep * index;
    ctx.fillText(`Y${item.year}Q${item.quarter}`, x - 18, height - 18);
  });
  ctx.fillText(`${max}W`, 8, padding.top + 4);
  ctx.fillText(`${min}W`, 8, height - padding.bottom);

  function yFor(value) {
    return padding.top + (max - value) / span * (height - padding.top - padding.bottom);
  }
}

function line(ctx, x1, y1, x2, y2) {
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.lineTo(x2, y2);
  ctx.stroke();
}

function switchView(viewName) {
  $$(".nav-item").forEach((btn) => btn.classList.toggle("active", btn.dataset.view === viewName));
  $$(".view").forEach((view) => view.classList.remove("active"));
  $(`#${viewName}View`)?.classList.add("active");
  if (viewName === "finance") {
    refreshCashFlow();
  }
}

function resetSample() {
  $("#yearInput").value = 1;
  $("#quarterInput").value = 1;
  $("#cashInput").value = 60;
  $("#equityInput").value = 60;
  $("#lastEquityInput").value = 60;
  $("#recordCompany").value = "默认企业";
  $("#recordYear").value = 1;
  $("#recordClosingCash").value = 42;
  $("#recordSalesRevenue").value = 32;
  $("#recordDirectCost").value = 12;
  $("#recordGrossProfit").value = 20;
  $("#recordProfitBeforeDepreciation").value = 4;
  $("#recordDepreciation").value = 3;
  $("#recordProfitBeforeInterest").value = 1;
  $("#recordProfitBeforeTax").value = -3;
  $("#recordNetProfit").value = -2;
  $("#recordTotalAssets").value = 98;
  $("#recordTotalLiability").value = 40;
  $("#recordOwnerEquityTotal").value = 58;
  $("#recordShareCapital").value = 60;
  $("#recordRetainedEarnings").value = 0;
  $("#recordManagementExpense").value = 4;
  $("#recordAdExpense").value = 6;
  $("#recordComprehensiveExpense").value = 16;
  $("#recordProfitComprehensiveExpense").value = 16;
  $("#recordRnDExpense").value = 2;
  $("#recordMarketExpense").value = 1;
  $("#recordISOExpense").value = 0;
  $("#recordMaintenanceExpense").value = 2;
  $("#recordSwitchExpense").value = 0;
  $("#recordRentExpense").value = 0;
  $("#recordInfoExpense").value = 1;
  $("#recordOtherLoss").value = 0;
  $("#recordFinancialExpense").value = 4;
  $("#recordLongLoanBalance").value = 40;
  $("#recordShortLoanBalance").value = 0;
  $("#longLoanInput").value = 0;
  $("#shortLoanInput").value = 0;
  $("#recordIncomeTax").value = 0;
  $("#recordIncomeTaxPayable").value = 0;
  $("#recordReceivables").value = 0;
  $("#recordWorkInProgress").value = 0;
  $("#recordFinishedGoods").value = 0;
  $("#recordRawMaterials").value = 0;
  $("#recordCurrentAssets").value = 42;
  $("#recordFactoryValue").value = 40;
  $("#recordProductionLineValue").value = 15;
  $("#recordConstructionInProgress").value = 0;
  $("#recordFixedAssets").value = 55;
  $("#recordBalanceNetProfit").value = -2;
  $("#recordLiabilityEquityTotal").value = 98;
  $("#orderYear").value = 1;
  $("#orderPaymentPeriod").value = 1;
  $("#orderISORequired").value = "";
  $("#orderDeliveryTime").value = "Y1Q2";
  $("#orderCollectionTime").value = "Y1Q3";
  $("#factoryIdInput").value = "F1";
  $("#factoryTypeInput").value = "大厂房";
  $("#factoryOwnershipInput").value = "租用";
  $("#lineIdInput").value = "L1";
  $("#lineTypeInput").value = "自动线";
  $("#lineProductInput").value = "P2";
  $("#lineStatusInput").value = "built";
  $("#lineNetValueInput").value = 15;
  $("#lineProducingInput").value = "true";
  $("#lineProgressInput").value = 1;
  $("#lineTotalInput").value = 1;
  $("#stepSelect").value = "ad_investment";
  $("#decisionType").value = "ad";
  updateDecisionFormFields();
  state.decisions = [
    { type: "ad", market: "本地", product: "P2", amount: 6, quantity: 0, note: "本地P2主投" },
    { type: "short_loan", amount: 20, quantity: 0, note: "保持现金垫" },
  ];
  state.orders = [
    { id: "A", year: 1, market: "本地", product: "P2", quantity: 2, totalPrice: 18, deliveryQuarter: 2, paymentPeriod: 1, isoRequired: "", deliveryTime: "Y1Q2", collectionTime: "Y1Q3" },
  ];
  state.factories = [
    { id: "F1", type: "大厂房", ownership: "租用", rented: true, purchased: false, startQuarterIndex: 1 },
  ];
  state.productionLines = [
    { id: "L1", factoryId: "F1", type: "自动线", product: "P2", built: true, status: "built", netValue: 15, isProducing: true, productionStartedQuarterIndex: 1, productionProgress: 1, productionTotal: 1, currentProductionProduct: "P2" },
  ];
  state.cashFlowPoints = [
    { quarterIndex: 1, openingCash: 60, inflows: 20, outflows: 7, closingCash: 73, note: "Y1Q1 短贷到账并完成广告、原料投入" },
    { quarterIndex: 2, openingCash: 73, inflows: 0, outflows: 1, closingCash: 72, note: "Y1Q2 常规管理费" },
    { quarterIndex: 3, openingCash: 72, inflows: 18, outflows: 1, closingCash: 89, note: "Y1Q3 订单回款" },
    { quarterIndex: 4, openingCash: 89, inflows: 0, outflows: 4, closingCash: 85, note: "Y1Q4 年末费用" },
  ];
  renderCashFlowQuarterSelect();
  syncRecordMirrors();
  renderProductionQuarterSelects();
  prefillCashFlowFormFromLatest();
  renderCashFlowPoints();
  renderProductionStatus();
  renderDecisions();
  renderOrders();
  renderOperationRecords();
  renderQAHistory();
  renderCompanyMetrics();
  updateStepSummary();
  refreshCashFlow();
  runAdvisor();
}

function money(value) {
  const n = Number(value || 0);
  return `${n}W`;
}

function quarterLabel(index) {
  const n = Number(index || 1);
  const year = Math.floor((n - 1) / 4) + 1;
  const quarter = ((n - 1) % 4) + 1;
  return `Y${year}Q${quarter}`;
}

function quarterIndexFromYearQuarter(year, quarter) {
  return (Number(year || 1) - 1) * 4 + Number(quarter || 1);
}

function paymentPeriodLabel(value) {
  const n = Number(value || 0);
  return n <= 0 ? "现金" : `${n}期`;
}

function deliveryTimeLabel(order) {
  return `Y${order.year || 1}Q${order.deliveryQuarter || 1}`;
}

function collectionTimeLabel(order) {
  const year = Number(order.year || 1);
  const delivery = Number(order.deliveryQuarter || 1);
  const period = Number(order.paymentPeriod || 0);
  return quarterLabel(quarterIndexFromYearQuarter(year, delivery) + period);
}

function warningTag(level) {
  const map = {
    safe: '<span class="tag ok">安全</span>',
    warning: '<span class="tag warn">关注</span>',
    danger: '<span class="tag danger">危险</span>',
    critical: '<span class="tag danger">紧急</span>',
  };
  return map[level] || `<span class="tag">${escapeHtml(level || "-")}</span>`;
}

function decisionName(type) {
  return {
    ad: "广告",
    short_loan: "短贷",
    long_loan: "长贷",
    market_iso: "市场/ISO",
    production_line: "生产线",
    material_order: "原料订单",
  }[type] || type;
}

function decisionTarget(item) {
  const parts = [];
  if (item.market) parts.push(item.market);
  if (item.product) parts.push(item.product);
  if (item.material) parts.push(item.material);
  if (item.lineType) parts.push(item.lineType);
  if (item.factoryType) parts.push(item.factoryType);
  if (item.extra?.isoType) parts.push(item.extra.isoType);
  return parts.length ? parts.join(" / ") : "-";
}

function updateDecisionFormFields() {
  const type = $("#decisionType").value;
  const config = decisionFormConfig[type] || decisionFormConfig.ad;
  const visible = new Set(config.fields);

  $$("[data-decision-field]").forEach((field) => {
    field.hidden = !visible.has(field.dataset.decisionField);
  });

  if (config.defaults.amount !== undefined) {
    $("#decisionAmount").value = config.defaults.amount;
  }
  if (config.defaults.quantity !== undefined) {
    $("#decisionQuantity").value = config.defaults.quantity;
  }
  if (config.defaults.market !== undefined) {
    $("#decisionMarket").value = config.defaults.market;
  }
  if (config.defaults.product !== undefined) {
    $("#decisionProduct").value = config.defaults.product;
  }
  if (config.defaults.material !== undefined) {
    $("#decisionMaterial").value = config.defaults.material;
  }
  if (config.defaults.lineType !== undefined) {
    $("#decisionLineType").value = config.defaults.lineType;
  }
  if (config.defaults.factoryType !== undefined) {
    $("#decisionFactoryType").value = config.defaults.factoryType;
  }
  if (config.defaults.isoType !== undefined) {
    $("#decisionISOType").value = config.defaults.isoType;
  }
  $("#decisionNote").value = config.defaults.note || "";

  if (type === "market_iso") {
    $("#decisionProduct").value = "";
  }
  if (type === "material_order") {
    $("#decisionMarket").value = "";
    $("#decisionProduct").value = "";
  }
}

function emptyRow(cols) {
  return `<tr><td colspan="${cols}" class="empty">暂无数据</td></tr>`;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function formatTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function bindEvents() {
  $$(".nav-item").forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.view));
  });
  $("#stepSelect").addEventListener("change", updateStepSummary);
  $("#decisionType").addEventListener("change", updateDecisionFormFields);
  $("#addDecisionBtn").addEventListener("click", addDecision);
  $("#addOrderBtn").addEventListener("click", addOrder);
  $("#addCashFlowPointBtn").addEventListener("click", addCashFlowPoint);
  $("#addFactoryBtn").addEventListener("click", addFactory);
  $("#addProductionLineBtn").addEventListener("click", addProductionLine);
  $("#refreshCashFlowBtn").addEventListener("click", refreshCashFlow);
  $("#saveRecordBtn").addEventListener("click", saveOperationRecord);
  $("#recordPrevBtn").addEventListener("click", () => turnRecordPage(-1));
  $("#recordNextBtn").addEventListener("click", () => turnRecordPage(1));
  $("#recordSavedList").addEventListener("click", (event) => {
    const year = event.target.dataset.recordYear;
    if (!year) return;
    selectRecordYear(year);
  });
  $("#clearQAHistoryBtn").addEventListener("click", clearAdvisorHistory);
  $("#runAdvisorBtn").addEventListener("click", runAdvisor);
  $("#resetBtn").addEventListener("click", resetSample);
  bindRecordMirror("#recordComprehensiveExpense", "#recordProfitComprehensiveExpense");
  bindRecordMirror("#recordProfitComprehensiveExpense", "#recordComprehensiveExpense");
  bindRecordMirror("#recordNetProfit", "#recordBalanceNetProfit");
  bindRecordMirror("#recordBalanceNetProfit", "#recordNetProfit");
  bindRecordMirror("#recordTotalAssets", "#recordLiabilityEquityTotal");
  bindRecordMirror("#recordLiabilityEquityTotal", "#recordTotalAssets");
  $("#decisionRows").addEventListener("click", (event) => {
    const index = event.target.dataset.removeDecision;
    if (index === undefined) return;
    state.decisions.splice(Number(index), 1);
    renderDecisions();
    renderCompanyMetrics();
    refreshCashFlow();
  });
  $("#orderRows").addEventListener("click", (event) => {
    const index = event.target.dataset.removeOrder;
    if (index === undefined) return;
    state.orders.splice(Number(index), 1);
    renderOrders();
    refreshCashFlow();
  });
  $("#cashFlowPointRows").addEventListener("click", (event) => {
    const index = event.target.dataset.removeCashFlow;
    if (index === undefined) return;
    state.cashFlowPoints.splice(Number(index), 1);
    renderCashFlowPoints();
    refreshCashFlow();
  });
  $("#factoryBoard").addEventListener("click", (event) => {
    const index = event.target.dataset.removeFactory;
    if (index === undefined) return;
    const [removed] = state.factories.splice(Number(index), 1);
    state.productionLines = state.productionLines.map((line) => line.factoryId === removed.id ? { ...line, factoryId: "" } : line);
    renderProductionStatus();
  });
  $("#productionLineRows").addEventListener("click", (event) => {
    const index = event.target.dataset.removeLine;
    if (index === undefined) return;
    state.productionLines.splice(Number(index), 1);
    renderProductionStatus();
  });
  $("#recordCompany").addEventListener("change", () => {
    loadOperationRecords();
    loadAdvisorHistory();
  });
  ["#yearInput", "#quarterInput", "#cashInput", "#equityInput", "#lastEquityInput", "#longLoanInput", "#shortLoanInput"].forEach((selector) => {
    $(selector).addEventListener("input", () => {
      renderCompanyMetrics();
      renderCashFlowQuarterSelect();
      renderProductionQuarterSelects();
      refreshCashFlow();
    });
  });
  ["#cashFlowOpening", "#cashFlowInflow", "#cashFlowOutflow"].forEach((selector) => {
    $(selector).addEventListener("input", syncCashClosingFromFlow);
  });
  window.addEventListener("resize", () => {
    refreshCashFlow();
  });
}

async function init() {
  bindEvents();
  updateDecisionFormFields();
  renderCashFlowQuarterSelect();
  renderProductionQuarterSelects();
  prefillCashFlowFormFromLatest();
  renderCashFlowPoints();
  renderProductionStatus();
  renderDecisions();
  renderOrders();
  refreshCashFlow();
  renderOperationRecords();
  renderQAHistory();
  renderCompanyMetrics();
  await checkHealth();
  try {
    await loadFlow();
    await loadOperationRecords();
    await loadAdvisorHistory();
    await runAdvisor();
  } catch (error) {
    $("#adviceBox").textContent = `初始化失败：${error.message}`;
  }
}

init();
