let rawGraph;
let cy;
let currentFilter = 'topic';
let selectedNodeID = '';
let currentPath = '';
let currentTitle = '项目总览';

const els = {
  loadState: document.getElementById('loadState'),
  leftState: document.getElementById('leftState'),
  warningCount: document.getElementById('warningCount'),
  warningCountCard: document.getElementById('warningCountCard'),
  projectName: document.getElementById('projectName'),
  topicCount: document.getElementById('topicCount'),
  workflowCount: document.getElementById('workflowCount'),
  edgeCount: document.getElementById('edgeCount'),
  builtAt: document.getElementById('builtAt'),
  hintTitle: document.getElementById('hintTitle'),
  hintText: document.getElementById('hintText'),
  detailEyebrow: document.getElementById('detailEyebrow'),
  detailTitle: document.getElementById('detailTitle'),
  detailType: document.getElementById('detailType'),
  detailSummary: document.getElementById('detailSummary'),
  detailStatus: document.getElementById('detailStatus'),
  detailConfidence: document.getElementById('detailConfidence'),
  detailIcon: document.getElementById('detailIcon'),
  relationTitle1: document.getElementById('relationTitle1'),
  relationTitle2: document.getElementById('relationTitle2'),
  relationTitle3: document.getElementById('relationTitle3'),
  relationList1: document.getElementById('relationList1'),
  relationList2: document.getElementById('relationList2'),
  relationList3: document.getElementById('relationList3'),
  warningList: document.getElementById('warningList'),
  fallback: document.getElementById('cyFallback'),
  fallbackText: document.getElementById('fallbackText')
};

const markdownSamples = {
  default: '# 文档预览\n\n无法直接读取当前 Markdown 文件。\n\n请确认 `zatools devwiki graph` 正在 DevWiki 根目录下运行，并且节点 path 指向 `wiki/` 下的 Markdown 文件。\n\n```text\nzatools devwiki graph --force\n```'
};

const typeName = { topic: 'Topic', workflow: 'Workflow' };
const typeStyle = {
  topic: { color: '#1769de', bg: '#eaf3ff', iconBg: 'linear-gradient(145deg, #2f7df6, #0e63e8)', dot: 'cap', icon: '□' },
  workflow: { color: '#d76f00', bg: '#fff3e3', iconBg: 'linear-gradient(145deg, #ff9d28, #f07b00)', dot: 'flow', icon: '⌁' }
};

function loadGraph() {
  showFallback('正在加载图谱数据...');
  fetch('/graph.json')
    .then((res) => {
      if (!res.ok) throw new Error('graph.json 加载失败');
      return res.json();
    })
    .then((data) => {
      rawGraph = data;
      els.loadState.textContent = '图谱已加载';
      els.leftState.textContent = '已加载';
      updateSummary(data);
      initCytoscape(toCytoscapeElements(data));
      clearSelection();
    })
    .catch((err) => {
      els.loadState.textContent = '加载失败';
      showFallback(err.message + '。请重新执行 zatools devwiki graph --force。');
      renderLoadError(err);
    });
}

function toCytoscapeElements(data) {
  const positions = computeLayeredPositions(data.nodes, data.edges);
  const nodes = data.nodes.map((node) => ({
    data: {
      id: node.id,
      title: node.title || node.slug,
      slug: node.slug,
      type: node.type,
      summary: node.summary || '',
      path: node.path,
      confidence: node.confidence || '-',
      status: node.status || '-',
      nodeLabel: node.title || node.slug,
      icon: typeStyle[node.type] ? typeStyle[node.type].icon : '⌬'
    },
    position: positions[node.id]
  }));
  const edges = data.edges.map((edge) => ({
    data: {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      relation: edge.type,
      label: edge.label,
      sources: edge.sources || []
    }
  }));
  return { nodes, edges };
}

function computeLayeredPositions(nodes, edges) {
  const byType = { topic: [], workflow: [] };
  nodes.forEach((node) => {
    if (!byType[node.type]) byType[node.type] = [];
    byType[node.type].push(node);
  });
  Object.keys(byType).forEach((type) => byType[type].sort((a, b) => a.slug.localeCompare(b.slug)));
  const columns = { topic: 220, workflow: 620 };
  const positions = {};
  Object.entries(byType).forEach(([type, list]) => {
    const step = Math.max(92, Math.min(150, 520 / Math.max(1, list.length)));
    const start = 130 + Math.max(0, (6 - list.length) * 18);
    list.forEach((node, index) => {
      positions[node.id] = { x: columns[type] || 430, y: start + index * step };
    });
  });
  const relatedOffsets = new Map();
  edges.filter((edge) => edge.type === 'related').forEach((edge, index) => {
    relatedOffsets.set(edge.source + edge.target, (index % 5) * 12);
  });
  return positions;
}

function initCytoscape(elements) {
  if (!window.cytoscape) {
    showFallback('Cytoscape.js 未加载成功。请确认静态资源 assets/cytoscape.min.js 存在。');
    return;
  }
  hideFallback();
  cy = cytoscape({
    container: document.getElementById('cy'),
    elements,
    minZoom: 0.45,
    maxZoom: 2.2,
    wheelSensitivity: 0.18,
    boxSelectionEnabled: false,
    layout: forceLayoutOptions(false),
    style: graphStyle()
  });
  cy.on('tap', 'node', (event) => selectNode(event.target));
  cy.on('tap', (event) => {
    if (event.target === cy) clearSelection();
  });
  cy.ready(() => {
    const firstTopic = cy.nodes('[type = "topic"]')[0];
    const firstNode = firstTopic || cy.nodes()[0];
    applyFilterAndSearch();
    if (firstNode) selectNode(firstNode);
  });
}

function graphStyle() {
  return [
    { selector: 'node', style: { 'width': 126, 'height': 62, 'shape': 'round-rectangle', 'background-color': '#ffffff', 'background-fill': 'linear-gradient', 'background-gradient-direction': 'to-bottom-right', 'background-gradient-stop-colors': '#ffffff #eaf3ff', 'background-gradient-stop-positions': '0% 100%', 'border-width': 1.8, 'border-color': '#7db6ff', 'label': 'data(nodeLabel)', 'font-family': 'Inter, PingFang SC, Microsoft YaHei, sans-serif', 'font-size': 11.5, 'font-weight': 760, 'line-height': 1.25, 'color': '#1f2937', 'text-valign': 'center', 'text-halign': 'center', 'text-wrap': 'wrap', 'text-max-width': 112, 'text-outline-width': 0, 'shadow-blur': 20, 'shadow-opacity': 0.18, 'shadow-color': '#7db6ff', 'shadow-offset-x': 0, 'shadow-offset-y': 8, 'overlay-opacity': 0, 'transition-property': 'opacity, border-width, border-color, background-color, width, height, shadow-opacity', 'transition-duration': 180 } },
    { selector: 'node[type = "topic"]', style: { 'background-gradient-stop-colors': '#ffffff #eaf3ff', 'border-color': '#60a5fa', 'color': '#1d4ed8', 'shadow-color': '#60a5fa' } },
    { selector: 'node[type = "workflow"]', style: { 'background-gradient-stop-colors': '#ffffff #fff0d9', 'border-color': '#fb923c', 'color': '#b95b00', 'shadow-color': '#fb923c' } },
    { selector: 'edge', style: { 'width': 1.6, 'curve-style': 'bezier', 'line-color': '#9aa7b8', 'target-arrow-color': '#9aa7b8', 'target-arrow-shape': 'triangle', 'arrow-scale': 0.9, 'opacity': 0.78, 'overlay-opacity': 0, 'transition-property': 'opacity, width, line-color, target-arrow-color', 'transition-duration': 180 } },
    { selector: 'edge[relation = "implemented_by"]', style: { 'line-color': '#49c98b', 'target-arrow-color': '#49c98b', 'width': 1.8 } },
    { selector: 'edge[relation = "related"]', style: { 'line-style': 'dashed', 'line-dash-pattern': [6, 6], 'line-color': '#a4afbf', 'target-arrow-shape': 'none', 'opacity': 0.62 } },
    { selector: '.selected', style: { 'width': 150, 'height': 74, 'border-width': 3.4, 'border-color': '#16b86f', 'background-gradient-stop-colors': '#ffffff #dcfce7', 'shadow-blur': 30, 'shadow-opacity': 0.34, 'shadow-color': '#16b86f', 'z-index': 10 } },
    { selector: '.neighbor', style: { 'border-width': 3, 'opacity': 1, 'z-index': 8 } },
    { selector: '.second-hop', style: { 'opacity': 0.82, 'border-style': 'dashed' } },
    { selector: '.search-match', style: { 'border-width': 3.2, 'border-color': '#7c3aed', 'shadow-blur': 26, 'shadow-opacity': 0.30, 'shadow-color': '#7c3aed', 'z-index': 9 } },
    { selector: '.search-related', style: { 'border-width': 2.4, 'border-color': '#94a3b8', 'border-style': 'dashed', 'opacity': 0.92, 'z-index': 7 } },
    { selector: '.search-edge', style: { 'width': 2.2, 'opacity': 0.86 } },
    { selector: '.faded', style: { 'opacity': 0.18 } },
    { selector: '.hidden-by-type, .hidden-by-search', style: { 'display': 'none' } }
  ];
}

function updateSummary(data) {
  const counts = { topic: 0, workflow: 0 };
  data.nodes.forEach((node) => counts[node.type]++);
  els.projectName.textContent = data.project && data.project.name ? data.project.name : '项目总览';
  els.topicCount.textContent = counts.topic;
  els.workflowCount.textContent = counts.workflow;
  els.edgeCount.textContent = data.edges.length;
  els.warningCount.textContent = (data.warnings || []).length;
  els.warningCountCard.textContent = (data.warnings || []).length;
  els.builtAt.textContent = formatTime(data.built_at);
  renderWarnings(data.warnings || []);
}

function renderWarnings(warnings) {
  if (!warnings.length) {
    els.warningList.innerHTML = '<div class="related-item muted">暂无 warning</div>';
    return;
  }
  els.warningList.innerHTML = warnings.map((warning, index) => '<div class="related-item"><span class="related-dot flow"></span><span class="related-title">' + escapeHTML(warning.message || ('warning.' + (index + 1))) + '</span></div>').join('');
}

function relatedItems(node, type) {
  const ids = new Set();
  node.connectedEdges().forEach((edge) => {
    const other = edge.source().id() === node.id() ? edge.target() : edge.source();
    if (!type || other.data('type') === type) ids.add(other.id());
  });
  return Array.from(ids).map((id) => cy.getElementById(id)).filter((item) => item && item.length);
}

function renderRelationList(el, items) {
  if (!items.length) {
    el.innerHTML = '<div class="related-item muted">暂无直接关系</div>';
    return;
  }
  el.innerHTML = items.map((item) => {
    const data = item.data();
    const style = typeStyle[data.type] || typeStyle.topic;
    return '<div class="related-item related-row"><button class="related-main related-button" type="button" data-node-id="' + escapeHTML(data.id) + '"><span class="related-dot ' + style.dot + '"></span><span class="related-title">' + escapeHTML(data.title || data.slug) + '</span></button><button class="preview-link related-preview" type="button" data-path="' + escapeHTML(data.path || '') + '" data-title="' + escapeHTML(data.title || data.slug) + '">预览</button></div>';
  }).join('');
  el.querySelectorAll('[data-node-id]').forEach((button) => {
    button.addEventListener('click', () => {
      const target = cy.getElementById(button.dataset.nodeId);
      if (target && target.length) selectNode(target);
    });
  });
}

function updateDetail(node) {
  const data = node.data();
  const style = typeStyle[data.type] || typeStyle.topic;
  selectedNodeID = data.id;
  currentPath = data.path || '';
  currentTitle = data.title || data.slug || '文档预览';
  els.detailEyebrow.textContent = '当前选中';
  els.detailTitle.textContent = data.title;
  els.detailType.textContent = typeName[data.type] || data.type;
  els.detailSummary.textContent = data.summary || '无摘要';
  els.detailStatus.textContent = data.status || '-';
  els.detailConfidence.textContent = data.confidence || '-';
  els.detailIcon.textContent = data.icon || style.icon;
  els.detailIcon.style.background = style.iconBg;
  els.detailType.style.color = style.color;
  els.detailType.style.background = style.bg;
  els.hintTitle.textContent = '已选中节点';
  els.hintText.textContent = '当前选中“' + data.title + '”，相关关系已高亮，非相关节点已淡化。';

  if (data.type === 'topic') {
    const workflows = relatedItems(node, 'workflow');
    const topics = relatedItems(node, 'topic');
    els.relationTitle1.textContent = '实现 Workflow (' + workflows.length + ')';
    els.relationTitle2.textContent = '相关 Topic (' + topics.length + ')';
    els.relationTitle3.textContent = '相关 Workflow (' + workflows.length + ')';
    renderRelationList(els.relationList1, workflows);
    renderRelationList(els.relationList2, topics);
    renderRelationList(els.relationList3, workflows);
  } else {
    const topics = relatedItems(node, 'topic');
    const workflows = relatedItems(node, 'workflow');
    els.relationTitle1.textContent = '支撑 Topic (' + topics.length + ')';
    els.relationTitle2.textContent = '相关 Workflow (' + workflows.length + ')';
    els.relationTitle3.textContent = '上层 Topic (' + topics.length + ')';
    renderRelationList(els.relationList1, topics);
    renderRelationList(els.relationList2, workflows);
    renderRelationList(els.relationList3, topics);
  }
}

function collectTwoHop(node, type) {
  const found = new Map();
  const oneHopNodes = node.closedNeighborhood().nodes();
  oneHopNodes.forEach((item) => {
    item.closedNeighborhood().nodes().forEach((candidate) => {
      if (candidate.id() !== node.id() && candidate.data('type') === type) found.set(candidate.id(), candidate);
    });
  });
  return Array.from(found.values());
}

function clearSelection() {
  if (!rawGraph) return;
  selectedNodeID = '';
  currentPath = '';
  currentTitle = rawGraph.project.name || '项目总览';
  if (cy) {
    cy.elements().removeClass('selected faded neighbor second-hop hidden-by-search hidden-by-type');
    applyFilterAndSearch();
  }
  const counts = { topic: 0, workflow: 0 };
  rawGraph.nodes.forEach((node) => counts[node.type]++);
  els.detailEyebrow.textContent = '项目总览';
  els.detailTitle.textContent = rawGraph.project.name || '项目总览';
  els.detailType.textContent = 'Overview';
  els.detailSummary.textContent = '未选中节点时，右侧展示项目总览、构建时间、warning 数量和图谱基础统计。';
  els.detailStatus.textContent = '已加载';
  els.detailConfidence.textContent = '-';
  els.detailIcon.textContent = '⌬';
  els.detailIcon.style.background = 'linear-gradient(145deg, #7c62ff, #2d7df5)';
  els.detailType.style.color = '#4938d0';
  els.detailType.style.background = '#eeeaff';
  els.hintTitle.textContent = '未选中节点';
  els.hintText.textContent = '在图中选择一个节点，查看详细信息和相关文档。';
  els.relationTitle1.textContent = 'Topic 数量';
  els.relationTitle2.textContent = 'Workflow 数量';
  els.relationTitle3.textContent = 'Workflow 数量';
  els.relationList1.innerHTML = '<div class="related-item"><span class="related-dot cap"></span><span class="related-title">Topic ' + counts.topic + '</span></div>';
  els.relationList2.innerHTML = '<div class="related-item"><span class="related-dot flow"></span><span class="related-title">Workflow ' + counts.workflow + '</span></div>';
  els.relationList3.innerHTML = '<div class="related-item"><span class="related-dot flow"></span><span class="related-title">Workflow ' + counts.workflow + '</span></div>';
}

function selectNode(node) {
  cy.elements().removeClass('selected faded neighbor second-hop');
  node.addClass('selected');
  const oneHop = node.closedNeighborhood();
  const twoHop = oneHop.closedNeighborhood();
  cy.elements().difference(twoHop).addClass('faded');
  oneHop.nodes().difference(node).addClass('neighbor');
  twoHop.nodes().difference(oneHop.nodes()).addClass('second-hop');
  updateDetail(node);
}

function applyFilterAndSearch() {
  if (!cy) return;
  const kw = document.getElementById('searchInput').value.trim().toLowerCase();
  const searchState = collectSearchVisibleNodeIDs(kw);
  cy.elements().removeClass('hidden-by-search hidden-by-type search-match search-related search-edge');
  cy.nodes().forEach((node) => {
    const data = node.data();
    const typeHidden = data.type !== currentFilter;
    if (!kw && typeHidden) {
      node.addClass('hidden-by-type');
      return;
    }
    if (kw && !searchState.visible.has(node.id())) {
      node.addClass('hidden-by-search');
      return;
    }
    if (kw && searchState.matches.has(node.id())) {
      node.addClass('search-match');
      return;
    }
    if (kw && searchState.related.has(node.id())) {
      node.addClass('search-related');
    }
  });
  cy.edges().forEach((edge) => {
    const sourceHidden = edge.source().hasClass('hidden-by-type') || edge.source().hasClass('hidden-by-search');
    const targetHidden = edge.target().hasClass('hidden-by-type') || edge.target().hasClass('hidden-by-search');
    if (sourceHidden || targetHidden) edge.addClass('hidden-by-type');
    if (kw && !sourceHidden && !targetHidden) edge.addClass('search-edge');
  });
  updateSearchStateText(kw, searchState);
}

function collectSearchVisibleNodeIDs(kw) {
  const matches = new Set();
  const related = new Set();
  const visible = new Set();
  if (!kw) return { matches, related, visible };
  cy.nodes().forEach((node) => {
    const data = node.data();
    if (data.type === currentFilter && nodeMatchesSearch(node, kw)) {
      matches.add(node.id());
      visible.add(node.id());
    }
  });
  matches.forEach((id) => {
    const node = cy.getElementById(id);
    node.connectedEdges().forEach((edge) => {
      const other = edge.source().id() === id ? edge.target() : edge.source();
      visible.add(other.id());
      if (!matches.has(other.id())) related.add(other.id());
    });
  });
  return { matches, related, visible };
}

function nodeMatchesSearch(node, kw) {
  const data = node.data();
  const text = (data.title + ' ' + data.slug + ' ' + data.summary).toLowerCase();
  return text.includes(kw);
}

function updateSearchStateText(kw, searchState) {
  if (!kw) {
    els.leftState.textContent = '已加载';
    return;
  }
  els.leftState.textContent = '命中 ' + searchState.matches.size + ' / 关联 ' + searchState.related.size;
}

function forceLayoutOptions(animate) {
  return { name: 'cose', animate: Boolean(animate), animationDuration: 700, nodeRepulsion: 9000, idealEdgeLength: 120, edgeElasticity: 110, gravity: 0.25, fit: true, padding: 64 };
}

function showFallback(message) {
  els.fallbackText.textContent = message;
  els.fallback.style.display = 'grid';
}

function hideFallback() {
  els.fallback.style.display = 'none';
}

function showToast(text) {
  const toast = document.getElementById('toast');
  toast.textContent = text || '操作成功';
  toast.classList.add('show');
  setTimeout(() => toast.classList.remove('show'), 1400);
}

function formatTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function escapeHTML(value) {
  return String(value || '').replace(/[&<>"']/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch]));
}

async function loadMarkdownFromLocal(path) {
  const candidates = [path, path ? './' + path : ''].filter(Boolean);
  let lastError = null;
  for (const url of candidates) {
    try {
      const res = await fetch(url, { cache: 'no-cache' });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      return { text: await res.text(), source: 'local', url };
    } catch (err) {
      lastError = err;
    }
  }
  return {
    text: '# 文档预览\n\n无法直接读取当前 Markdown 文件。\n\n请确认 `zatools devwiki graph` 正在 DevWiki 根目录下运行，并且节点 path 指向 `wiki/` 下的 Markdown 文件。\n\n```text\nzatools devwiki graph --force\n```',
    source: 'fallback',
    url: path || '',
    error: lastError
  };
}

function renderMarkdownPreview(contentEl, markdown, afterRender) {
  contentEl.innerHTML = '';
  const previewEl = document.createElement('div');
  previewEl.className = 'markdown-preview-inner';
  contentEl.appendChild(previewEl);
  if (!window.Vditor || typeof Vditor.preview !== 'function') {
    previewEl.innerHTML = '<div class="markdown-loading">Vditor 未加载成功，请确认 vendor/vditor 静态资源存在。</div>';
    if (afterRender) afterRender();
    return;
  }
  Vditor.preview(previewEl, markdown || '', {
    mode: 'light',
    cdn: '/assets/vendor/vditor',
    markdown: {
      toc: true,
      mark: true,
      footnotes: true,
      autoSpace: true
    },
    hljs: {
      style: 'github'
    },
    after() {
      if (afterRender) afterRender();
    }
  });
}

async function openMarkdownPreview(path, title) {
  const modal = document.getElementById('markdownModal');
  const contentEl = document.getElementById('markdownContent');
  if (!modal || !contentEl) return;
  document.getElementById('markdownTitle').textContent = title || '文档预览';
  document.getElementById('markdownMeta').textContent = '正在读取当前 Markdown';
  contentEl.innerHTML = '<div class="markdown-loading">正在读取 Markdown 文件...</div>';
  modal.classList.add('show');
  modal.setAttribute('aria-hidden', 'false');
  const result = await loadMarkdownFromLocal(path);
  document.getElementById('markdownMeta').textContent = result.source === 'local' ? '本地 Markdown' : 'fallback Markdown 预览';
  renderMarkdownPreview(contentEl, result.text, () => {
    showToast(result.source === 'local' ? '已读取本地 Markdown' : '已显示 fallback Markdown 预览');
  });
}

function closeMarkdownPreview() {
  const modal = document.getElementById('markdownModal');
  if (!modal) return;
  modal.classList.remove('show');
  modal.setAttribute('aria-hidden', 'true');
}

function renderLoadError(err) {
  els.detailTitle.textContent = '加载失败';
  els.detailType.textContent = 'Error';
  els.detailSummary.textContent = err.message;
  els.detailStatus.textContent = '失败';
  els.relationList1.innerHTML = '<div class="related-item muted">请重新执行 zatools devwiki graph --force</div>';
  els.relationList2.innerHTML = '';
  els.relationList3.innerHTML = '';
}

document.getElementById('clearBtn').addEventListener('click', clearSelection);
document.getElementById('typeSelect').addEventListener('change', (event) => {
  currentFilter = event.target.value;
  applyFilterAndSearch();
  showToast('已切换到 ' + typeName[currentFilter]);
});
document.getElementById('searchInput').addEventListener('input', applyFilterAndSearch);
document.getElementById('fitBtnTop').addEventListener('click', () => cy && cy.fit(undefined, 64));
document.getElementById('zoomInBtn').addEventListener('click', () => cy && cy.zoom({ level: cy.zoom() + 0.15, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } }));
document.getElementById('zoomOutBtnTop').addEventListener('click', () => cy && cy.zoom({ level: cy.zoom() - 0.15, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } }));
document.getElementById('previewCurrentBtn').addEventListener('click', () => openMarkdownPreview(currentPath, currentTitle));
document.getElementById('markdownClose').addEventListener('click', closeMarkdownPreview);
document.getElementById('markdownBackdrop').addEventListener('click', closeMarkdownPreview);
document.addEventListener('click', (event) => {
  const button = event.target.closest('.related-preview');
  if (!button) return;
  openMarkdownPreview(button.dataset.path, button.dataset.title);
});
document.getElementById('warningBtn').addEventListener('click', () => {
  els.warningList.scrollIntoView({ block: 'nearest' });
  showToast('已定位 warning 列表');
});
document.getElementById('leftWarningBtn').addEventListener('click', () => {
  els.warningList.scrollIntoView({ block: 'nearest' });
  showToast('已定位 warning 列表');
});
window.addEventListener('resize', () => {
  if (!cy) return;
  cy.resize();
  cy.fit(undefined, 64);
});
window.addEventListener('keydown', (event) => {
  if (event.key === 'Escape') {
    closeMarkdownPreview();
  }
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault();
    document.getElementById('searchInput').focus();
  }
});

loadGraph();
