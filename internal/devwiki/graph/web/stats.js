const WORDCLOUD_COLORS = ["#2f75ff", "#31bd73", "#7c4dff", "#ff8a18", "#0ea5e9", "#64748b"];

let dashboardData = null;
let filterText = "";

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function formatTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { hour12: false });
}

function formatShortTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleTimeString("zh-CN", { hour12: false });
}

function formatNumber(value) {
  return new Intl.NumberFormat("zh-CN").format(value ?? 0);
}

function kindLabel(kind) {
  switch (kind) {
    case "topic": return "主题";
    case "workflow": return "流程";
    case "index": return "索引";
    case "glossary": return "术语";
    default: return kind || "-";
  }
}

function eventTypeLabel(event) {
  if (event.endpoint === "read") {
    return event.view || "-";
  }
  return kindLabel(event.kind);
}

function kindTag(kind) {
  switch (kind) {
    case "topic": return "blue";
    case "workflow": return "green";
    case "index": return "purple";
    default: return "gray";
  }
}

function matchesFilter(text) {
  if (!filterText) return true;
  return String(text || "").toLowerCase().includes(filterText);
}

function iconSvg(type) {
  const common = 'width="11" height="11" viewBox="0 0 24 24" fill="none"';
  const stroke = 'stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"';
  if (type === "search") {
    return `<svg ${common}><path d="m21 21-4.3-4.3M10.8 18a7.2 7.2 0 1 1 0-14.4 7.2 7.2 0 0 1 0 14.4Z" ${stroke}/></svg>`;
  }
  return `<svg ${common}><path d="M14 3v4a2 2 0 0 0 2 2h4" ${stroke}/><path d="M7 21h10a3 3 0 0 0 3-3V8.5L14.5 3H7a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3Z" ${stroke}/></svg>`;
}

function summaryIcon(type) {
  const common = 'width="18" height="18" viewBox="0 0 24 24" fill="none"';
  const stroke = 'stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"';
  if (type === "search") {
    return `<svg ${common}><path d="m21 21-4.3-4.3M10.8 18a7.2 7.2 0 1 1 0-14.4 7.2 7.2 0 0 1 0 14.4Z" ${stroke}/></svg>`;
  }
  if (type === "trend") {
    return `<svg ${common}><path d="M4 17h16M6 15l4-4 3 3 5-7" ${stroke}/></svg>`;
  }
  if (type === "doc") {
    return `<svg ${common}><path d="M14 3v4a2 2 0 0 0 2 2h4" ${stroke}/><path d="M7 21h10a3 3 0 0 0 3-3V8.5L14.5 3H7a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3Z" ${stroke}/></svg>`;
  }
  return `<svg ${common}><path d="M12 9v4m0 4h.01M10.3 4.7 2.8 18a2 2 0 0 0 1.7 3h15a2 2 0 0 0 1.7-3L13.7 4.7a2 2 0 0 0-3.4 0Z" ${stroke}/></svg>`;
}

function eventQueryText(event) {
  if (event.endpoint === "read") {
    return event.slug || "-";
  }
  return (event.queries || []).join(" / ") || "-";
}

function renderSummary(data) {
  const summary = [
    { label: "今日 API", value: formatNumber(data.today_api_count), color: "blue", icon: "search" },
    { label: "累计 API", value: formatNumber(data.total_api_count), color: "green", icon: "trend" },
    { label: "已统计文档", value: formatNumber(data.tracked_document_count), color: "purple", icon: "doc" },
    { label: "空结果", value: formatNumber(data.today_empty_search_count), color: "orange", icon: "alert" },
  ];
  document.getElementById("summaryList").innerHTML = summary.map((item) => `
    <div class="summary-row">
      <div class="icon-circle icon-${item.color}">${summaryIcon(item.icon)}</div>
      <div class="summary-label">${item.label}</div>
      <div class="summary-value ${item.color}">${item.value}</div>
    </div>
  `).join("");
}

function renderDocuments(documents) {
  const rows = (documents || []).filter((doc) =>
    matchesFilter([doc.slug, doc.kind, kindLabel(doc.kind)].join(" "))
  );
  const body = document.getElementById("documentsBody");
  if (rows.length === 0) {
    body.innerHTML = '<tr><td colspan="4" class="empty-cell">暂无数据</td></tr>';
    return;
  }
  body.innerHTML = rows.map((doc, index) => `
    <tr>
      <td class="rank">${index + 1}</td>
      <td title="${escapeHTML(doc.slug)}">${escapeHTML(doc.slug)}</td>
      <td><span class="tag tag-${kindTag(doc.kind)}">${escapeHTML(kindLabel(doc.kind))}</span></td>
      <td class="num">${formatNumber(doc.read_count)}</td>
    </tr>
  `).join("");
}

function renderEvents(events) {
  const rows = [...(events || [])].reverse().filter((event) =>
    matchesFilter([
      event.endpoint,
      event.kind,
      eventQueryText(event),
      ...(event.queries || []),
      event.slug,
    ].join(" "))
  );
  const body = document.getElementById("eventsBody");
  if (rows.length === 0) {
    body.innerHTML = '<tr><td colspan="6" class="empty-cell">暂无数据</td></tr>';
    return;
  }
  body.innerHTML = rows.map((event) => {
    const isSearch = event.endpoint === "search";
    const typeLabel = isSearch ? "搜索" : "阅读";
    const typeColor = isSearch ? "blue" : "green";
    const iconType = isSearch ? "search" : "doc";
    const detailType = eventTypeLabel(event);
    const detailTag = isSearch ? kindTag(event.kind) : "gray";
    const resultCell = isSearch ? formatNumber(event.result_count ?? 0) : "-";
    const emptyTag = isSearch
      ? `<span class="tag ${event.empty ? "tag-red" : "tag-green"}">${event.empty ? "是" : "否"}</span>`
      : `<span class="tag tag-gray">-</span>`;
    return `
      <tr>
        <td>${formatShortTime(event.ts)}</td>
        <td>
          <span class="event-type">
            <span class="type-dot icon-${typeColor}">${iconSvg(iconType)}</span>
            <span class="${typeColor}">${typeLabel}</span>
          </span>
        </td>
        <td><span class="tag tag-${detailTag}">${escapeHTML(detailType)}</span></td>
        <td><span class="query-cell"><span class="query-pill">${escapeHTML(eventQueryText(event))}</span></span></td>
        <td class="num">${resultCell}</td>
        <td style="text-align:center">${emptyTag}</td>
      </tr>
    `;
  }).join("");
}

function keywordFontSize(weight, maxWeight, width) {
  const normalized = Math.max(weight || 1, 1) / Math.max(maxWeight || 1, 1);
  const maxFont = Math.max(28, Math.min(width * 0.16, 52));
  const minFont = Math.max(16, maxFont * 0.42);
  return minFont + normalized * (maxFont - minFont);
}

function renderKeywordFallback(keywords) {
  const shell = document.querySelector(".keyword-cloud-shell");
  let fallback = document.getElementById("keywordFallback");
  if (!fallback) {
    fallback = document.createElement("div");
    fallback.id = "keywordFallback";
    fallback.className = "keyword-fallback";
    shell.appendChild(fallback);
  }
  fallback.innerHTML = (keywords || []).map((item) => {
    const weight = Math.max(item.weight || 1, 1);
    return `<span class="keyword-tag" style="--kw-weight:${weight}">${escapeHTML(item.text)}</span>`;
  }).join("");
  shell.classList.add("use-fallback");
}

function paintKeywordCloud(canvas, shell, keywords) {
  const width = Math.max(shell.clientWidth, 280);
  const height = Math.max(shell.clientHeight, 210);
  canvas.width = width;
  canvas.height = height;
  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;

  const weights = (keywords || []).map((item) => Math.max(item.weight || 1, 1));
  const maxWeight = Math.max(...weights, 1);
  const list = (keywords || []).map((item) => [item.text, Math.max(item.weight || 1, 1)]);

  if (typeof WordCloud !== "function" || (typeof WordCloud.isSupported === "function" && !WordCloud.isSupported)) {
    return false;
  }

  WordCloud(canvas, {
    list,
    gridSize: Math.round(Math.max(6, 10 * width / 1024)),
    weightFactor(size) {
      return keywordFontSize(size, maxWeight, width);
    },
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
    fontWeight: "700",
    color() {
      return WORDCLOUD_COLORS[Math.floor(Math.random() * WORDCLOUD_COLORS.length)];
    },
    rotateRatio: 0.2,
    rotationSteps: 2,
    backgroundColor: "transparent",
    shrinkToFit: true,
    drawOutOfBound: false,
    minSize: 12,
  });
  return true;
}

function renderKeywords(keywords, available) {
  const shell = document.querySelector(".keyword-cloud-shell");
  const canvas = document.getElementById("keywordCanvas");
  const empty = document.getElementById("keywordCloudEmpty");
  const meta = document.getElementById("keywordsMeta");
  const filtered = (keywords || []).filter((item) => matchesFilter(item.text));

  shell.classList.remove("has-data", "use-fallback");
  const fallback = document.getElementById("keywordFallback");
  if (fallback) {
    fallback.innerHTML = "";
  }

  if (!available || filtered.length === 0) {
    empty.textContent = "暂无关键词数据，Server 会每小时自动刷新";
    meta.textContent = "Server 每小时自动从 queries 日志刷新";
    return;
  }

  shell.classList.add("has-data");
  meta.textContent = `基于 keywords.json，共 ${filtered.length} 个关键词`;

  requestAnimationFrame(() => {
    const painted = paintKeywordCloud(canvas, shell, filtered);
    if (!painted) {
      shell.classList.remove("has-data");
      renderKeywordFallback(filtered);
      meta.textContent = `共 ${filtered.length} 个关键词（词云不可用，已降级为标签展示）`;
    }
  });
}

function renderDashboard(data) {
  dashboardData = data;
  const projectLabel = data.project_name || data.project_slug || "-";

  document.getElementById("projectChip").textContent = data.project_slug || projectLabel;
  document.getElementById("sidebarProjectName").textContent = projectLabel;
  document.getElementById("heroTodaySearch").textContent = formatNumber(data.today_api_count);
  document.getElementById("heroTotalSearch").textContent = formatNumber(data.total_api_count);
  document.getElementById("metaUpdatedAt").textContent = formatTime(data.updated_at);
  document.getElementById("metaToday").textContent = data.today ? `今日 (${data.today})` : "今日";
  document.getElementById("asideLogFile").textContent = data.today_log_file || "queries-*.jsonl";

  const hasData = (data.today_api_count || 0) > 0
    || (data.total_api_count || 0) > 0
    || (data.tracked_document_count || 0) > 0
    || (data.today_events || []).length > 0;
  document.getElementById("metaStatus").textContent = hasData ? "正常" : "暂无数据";
  document.getElementById("metaStatus").style.color = hasData ? "#16a05d" : "#d93131";

  renderSummary(data);
  renderDocuments(data.top_documents);
  renderEvents(data.today_events);
  renderKeywords(data.keywords, data.keywords_available);
}

async function loadDashboard() {
  const loadState = document.getElementById("loadState");
  const overviewState = document.getElementById("overviewState");
  try {
    const response = await fetch("/api/stats/summary");
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    renderDashboard(data);
    loadState.textContent = "已加载";
    overviewState.innerHTML = '<i class="dot" style="width:7px;height:7px;box-shadow:none"></i>已同步';
    overviewState.classList.remove("error");
  } catch (error) {
    loadState.textContent = "加载失败";
    overviewState.innerHTML = '<i class="dot" style="width:7px;height:7px;box-shadow:none;background:#ef4444"></i>同步失败';
    overviewState.classList.add("error");
    console.error(error);
  }
}

document.getElementById("filterInput").addEventListener("input", (event) => {
  filterText = event.target.value.trim().toLowerCase();
  if (!dashboardData) return;
  renderDocuments(dashboardData.top_documents);
  renderEvents(dashboardData.today_events);
  renderKeywords(dashboardData.keywords, dashboardData.keywords_available);
});

document.getElementById("refreshBtn").addEventListener("click", () => {
  loadDashboard();
});

document.getElementById("openStatsDirBtn").addEventListener("click", () => {
  const logFile = dashboardData?.today_log_file || "queries-*.jsonl";
  window.alert(`统计数据目录：.devwiki/stats/\n今日日志：${logFile}`);
});

window.addEventListener("resize", () => {
  if (!dashboardData || !dashboardData.keywords_available) return;
  renderKeywords(dashboardData.keywords, dashboardData.keywords_available);
});

loadDashboard();
