package ui

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Local EMS Dashboard</title>
<style>
  :root { --bg: #0f172a; --card: #1e293b; --border: #334155; --text: #e2e8f0; --muted: #94a3b8; --dim: #64748b;
    --solar: #facc15; --load: #f97316; --grid: #38bdf8; --ess: #22c55e; --danger: #ef4444; --warn: #eab308;
    --success: #22c55e; --select-bg: #1e293b; --select-text: #e2e8f0; }

  [data-theme="light"] { --bg: #f8fafc; --card: #ffffff; --border: #e2e8f0; --text: #1e293b; --muted: #64748b; --dim: #94a3b8;
    --solar: #ca8a04; --load: #ea580c; --grid: #2563eb; --ess: #16a34a; --danger: #dc2626; --warn: #ca8a04;
    --success: #16a34a; --select-bg: #ffffff; --select-text: #1e293b; }

  @media (prefers-color-scheme: light) {
    [data-theme="system"] { --bg: #f8fafc; --card: #ffffff; --border: #e2e8f0; --text: #1e293b; --muted: #64748b; --dim: #94a3b8;
      --solar: #ca8a04; --load: #ea580c; --grid: #2563eb; --ess: #16a34a; --danger: #dc2626; --warn: #ca8a04;
      --success: #16a34a; --select-bg: #ffffff; --select-text: #1e293b; }
  }

  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family: -apple-system,'Segoe UI',Roboto,sans-serif; background:var(--bg); color:var(--text); min-height:100vh; display:flex; flex-direction:column; }
  .main-content { flex:1; }
  .footer { text-align:center; padding:16px; font-size:12px; color:var(--dim); border-top:1px solid var(--border); }

  /* --- Header --- */
  .header { background:var(--card); padding:12px 24px; display:flex; justify-content:space-between; align-items:center; border-bottom:1px solid var(--border); }
  .header h1 { font-size:16px; font-weight:600; }
  .header-left { display:flex; align-items:center; }
  .header-right { display:flex; align-items:center; gap:16px; }
  .dot { width:8px; height:8px; border-radius:50%; display:inline-block; margin-right:4px; }
  .dot.on { background:var(--ess); box-shadow:0 0 6px var(--ess); }
  .dot.off { background:var(--danger); }
  .status-text { font-size:12px; color:var(--muted); }

  /* --- Sidebar --- */
  .hamburger { background:none; border:none; color:var(--text); font-size:22px; cursor:pointer; padding:4px 8px; line-height:1; margin-right:12px; }
  .sidebar { position:fixed; top:0; left:-280px; width:280px; height:100vh; background:var(--card); border-right:1px solid var(--border);
    z-index:1000; transition:left .3s ease; display:flex; flex-direction:column; overflow-y:auto; }
  .sidebar.open { left:0; }
  .sidebar-overlay { display:none; position:fixed; inset:0; background:rgba(0,0,0,.5); z-index:999; }
  .sidebar-overlay.open { display:block; }
  .sidebar-header { padding:20px 20px 16px; border-bottom:1px solid var(--border); }
  .sidebar-header .logo { font-size:18px; font-weight:600; }
  .sidebar-header .version { font-size:11px; color:var(--dim); margin-top:2px; }
  .sidebar-close { position:absolute; top:16px; right:16px; background:none; border:none; color:var(--muted); font-size:20px; cursor:pointer; }
  .sidebar-nav { flex:1; padding:12px 0; }
  .nav-section { padding:4px 20px; font-size:10px; color:var(--dim); text-transform:uppercase; letter-spacing:1px; margin-top:12px; }
  .nav-item { display:flex; align-items:center; gap:12px; padding:10px 20px; color:var(--muted); cursor:pointer; transition:all .15s; font-size:14px; text-decoration:none; }
  .nav-item:hover { background:var(--border); color:var(--text); }
  .nav-item.active { background:rgba(250,204,21,0.1); color:var(--solar); border-right:3px solid var(--solar); }
  .nav-item .nav-icon { font-size:18px; width:24px; text-align:center; }
  .sidebar-footer { padding:12px 20px; border-top:1px solid var(--border); }
  .sidebar-footer .nav-item { padding:10px 0; }
  .sidebar-footer .nav-item:hover { background:none; color:var(--danger); }

  /* --- Pages --- */
  .tab-content { display:none; }
  .tab-content.active { display:block; }

  /* --- Live Layout (OpenEMS-style 2 columns) --- */
  .live-layout { display:grid; grid-template-columns:1fr 1fr; gap:16px; padding:16px 24px; align-items:start; }
  @media(max-width:900px) { .live-layout { grid-template-columns:1fr; padding:12px 16px; } }

  /* Energy Monitor Card */
  .monitor-card { background:var(--card); border:1px solid var(--border); border-radius:12px; overflow:hidden; }
  .monitor-header { padding:12px 16px; border-bottom:1px solid var(--border); font-size:13px; font-weight:600; color:var(--muted); display:flex; align-items:center; gap:8px; }
  .monitor-body { padding:16px; }
  .flow-svg { width:100%; max-width:420px; margin:0 auto; display:block; }
  .soc-section { padding:8px 0 0; }
  .soc-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:6px; }
  .soc-header .lbl { font-size:12px; color:var(--muted); text-transform:uppercase; letter-spacing:.5px; }
  .soc-header .val { font-size:14px; font-weight:600; }
  .soc-track { background:var(--border); border-radius:6px; height:22px; overflow:hidden; }
  .soc-fill { height:100%; border-radius:6px; transition:width .5s ease, background .5s ease; min-width:2px; }

  /* Widget Cards */
  .widget-column { display:grid; grid-template-columns:1fr 1fr; gap:10px; align-content:start; }
  @media(max-width:600px) { .widget-column { grid-template-columns:1fr; } }
  .widget-card { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:14px 16px; }
  .widget-header { font-size:12px; font-weight:600; color:var(--muted); display:flex; align-items:center; gap:8px; margin-bottom:8px; }
  .widget-dot { width:8px; height:8px; border-radius:50%; }
  .widget-badge { margin-left:auto; padding:2px 8px; border-radius:4px; font-size:10px; font-weight:600; background:var(--border); color:var(--text); }
  .widget-body {}
  .widget-value { font-size:22px; font-weight:700; font-variant-numeric:tabular-nums; }
  .widget-sub { font-size:11px; color:var(--dim); margin-top:2px; }
  .widget-pct { font-size:28px; font-weight:700; font-variant-numeric:tabular-nums; }
  .widget-pct small { font-size:14px; color:var(--dim); }
  .widget-bar { background:var(--border); border-radius:4px; height:6px; margin-top:6px; overflow:hidden; }
  .widget-bar-fill { height:100%; border-radius:4px; transition:width .5s ease; }

  /* Constraints */
  .constraints-bar { grid-column:1/-1; padding:4px 0; display:flex; gap:8px; flex-wrap:wrap; align-items:center; }
  .constraints-bar .lbl { font-size:11px; color:var(--dim); text-transform:uppercase; }
  .tag { background:#7c3aed22; color:#a78bfa; padding:3px 10px; border-radius:4px; font-size:11px; font-weight:500; }
  .tag.clamped { background:#dc262622; color:#f87171; }

  /* --- Chart --- */
  .chart-wrap { background:var(--card); border:1px solid var(--border); border-radius:12px; margin:0 24px 24px; padding:16px 16px 12px; }
  .chart-wrap .title { font-size:12px; color:var(--muted); text-transform:uppercase; letter-spacing:.5px; margin-bottom:8px; }
  canvas { width:100%!important; height:220px!important; }

  /* --- History Tab --- */
  .history-controls { display:flex; align-items:center; gap:12px; padding:20px 24px 12px; }
  .history-controls label { font-size:13px; color:var(--muted); }
  .history-controls input[type=date] { background:var(--card); border:1px solid var(--border); color:var(--text); padding:8px 12px; border-radius:6px; font-size:14px; }
  .history-controls button { background:var(--solar); color:#000; border:none; padding:8px 20px; border-radius:6px; font-size:13px; font-weight:600; cursor:pointer; }
  .summary-grid { display:grid; grid-template-columns:repeat(4,1fr); gap:12px; padding:0 24px 20px; }
  @media(max-width:700px) { .summary-grid { grid-template-columns:repeat(2,1fr); } }
  .summary-card { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:14px 16px; }
  .summary-card .lbl { font-size:11px; color:var(--muted); text-transform:uppercase; }
  .summary-card .val { font-size:20px; font-weight:700; margin-top:4px; }
  .summary-card .unit { font-size:12px; color:var(--dim); }
  .history-chart-wrap { background:var(--card); border:1px solid var(--border); border-radius:12px; margin:0 24px 16px; padding:16px; }
  .history-chart-wrap .title { font-size:12px; color:var(--muted); text-transform:uppercase; letter-spacing:.5px; margin-bottom:8px; }
  .history-empty { text-align:center; padding:48px; color:var(--dim); font-size:14px; }

  /* --- Config / About pages --- */
  .config-page { padding:24px; }
  .config-page h2 { font-size:16px; margin-bottom:16px; }
  .config-section { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:16px 20px; margin-bottom:12px; }
  .config-section h3 { font-size:13px; color:var(--solar); margin-bottom:12px; text-transform:uppercase; letter-spacing:.5px; }
  .config-row { display:flex; justify-content:space-between; align-items:center; padding:6px 0; border-bottom:1px solid var(--border); font-size:13px; }
  .config-row:last-child { border-bottom:none; }
  .config-row .key { color:var(--muted); }
  .config-row .val { color:var(--text); font-weight:500; font-variant-numeric:tabular-nums; }
  .about-page { padding:24px; }
  .about-page h2 { font-size:16px; margin-bottom:16px; }
  .about-card { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:20px; margin-bottom:12px; }
  .about-card h3 { font-size:13px; color:var(--solar); margin-bottom:10px; text-transform:uppercase; }
  .about-row { display:flex; justify-content:space-between; padding:4px 0; font-size:13px; }
  .about-row .key { color:var(--muted); }
  .about-row .val { color:var(--text); }

  /* ========== DEVICE LIST & GRID ========== */
  .dev-list { margin-top:8px; font-size:11px; }
  .dev-row { display:flex; justify-content:space-between; align-items:center; padding:3px 0; border-top:1px solid var(--border); }
  .dev-row:first-child { border-top:none; }
  .dev-id { color:var(--muted); font-family:monospace; }
  .dev-val { color:var(--text); font-variant-numeric:tabular-nums; }
  .dev-dot { display:inline-block; width:6px; height:6px; border-radius:50%; margin-right:4px; }
  .dev-dot.on { background:var(--success); }
  .dev-dot.off { background:var(--danger); }
  .dev-grid { display:grid; grid-template-columns:repeat(auto-fill, minmax(140px,1fr)); gap:6px; }
  .dev-chip { background:var(--bg); border:1px solid var(--border); border-radius:6px; padding:6px 8px; font-size:11px; }
  .dev-chip .dev-name { font-weight:600; font-family:monospace; font-size:11px; }
  .dev-chip .dev-info { color:var(--muted); margin-top:2px; font-variant-numeric:tabular-nums; }
  .dev-chip.fault { border-color:var(--danger); }

  /* ========== RESPONSIVE ========== */
  @media(max-width:600px) {
    .header { padding:10px 14px; }
    .header h1 { font-size:14px; }
    .chart-wrap, .history-chart-wrap { margin:0 14px 14px; padding:12px; border-radius:8px; }
    canvas { height:160px!important; }
    .history-controls { flex-wrap:wrap; padding:14px 14px 10px; gap:8px; }
    .history-controls input[type=date] { flex:1; min-width:140px; }
    .history-controls button { flex:1; text-align:center; }
    .summary-grid { grid-template-columns:1fr 1fr; gap:8px; padding:0 14px 14px; }
    .config-page, .about-page { padding:16px; }
  }
</style>
</head>
<body>
<script>
// Apply saved theme immediately to prevent flash
(function(){
  var t = localStorage.getItem('ems-theme') || 'dark';
  document.documentElement.setAttribute('data-theme', t);
})();
</script>

<!-- Sidebar Overlay -->
<div class="sidebar-overlay" id="sidebarOverlay" onclick="closeSidebar()"></div>

<!-- Sidebar -->
<nav class="sidebar" id="sidebar">
  <button class="sidebar-close" onclick="closeSidebar()">✕</button>
  <div class="sidebar-header">
    <div class="logo">⚡ Local EMS</div>
    <div class="version">v2026.10.0 — 5MWh BESS</div>
  </div>
  <div class="sidebar-nav">
    <div class="nav-section">Monitoring</div>
    <a class="nav-item" href="/live" data-page="live">
      <span class="nav-icon">🔴</span> Live
    </a>
    <a class="nav-item" href="/history" data-page="history">
      <span class="nav-icon">📊</span> <span data-i18n="history">History</span>
    </a>
    <div class="nav-section" data-i18n="nav_system">System</div>
    <a class="nav-item" href="/settings/config" data-page="config">
      <span class="nav-icon">⚙️</span> <span data-i18n="config">Edge Config</span>
    </a>
    <a class="nav-item" href="/settings/health" data-page="health">
      <span class="nav-icon">💚</span> <span data-i18n="health">System Health</span>
    </a>
    <a class="nav-item" href="/settings/about" data-page="about">
      <span class="nav-icon">ℹ️</span> <span data-i18n="about">About</span>
    </a>
    <div class="nav-section" data-i18n="nav_account">Account</div>
    <a class="nav-item" href="/user" data-page="user">
      <span class="nav-icon">👤</span> <span data-i18n="user">User</span>
    </a>
  </div>
  <div class="sidebar-footer">
    <a href="/logout" class="nav-item">
      <span class="nav-icon">🚪</span> Sign Out
    </a>
  </div>
</nav>

<!-- Header -->
<div class="main-content">
<div class="header">
  <div class="header-left">
    <button class="hamburger" onclick="openSidebar()">☰</button>
    <h1 id="pageTitle">⚡ Live</h1>
  </div>
  <div class="header-right">
    <span class="status-text"><span class="dot off" id="dot"></span><span id="connStatus">Connecting...</span></span>
  </div>
</div>

<!-- ==================== LIVE PAGE ==================== -->
<div class="tab-content" id="tab-live">
  <div class="live-layout">

    <!-- LEFT: Energy Monitor -->
    <div class="monitor-card">
      <div class="monitor-header"><span>📊</span> Energy Monitor</div>
      <div class="monitor-body">
        <svg viewBox="0 0 400 320" class="flow-svg">
          <!-- Flow lines (behind nodes) -->
          <line x1="102" y1="60" x2="298" y2="60"  style="stroke:var(--border)" stroke-width="1.5" id="linePG"/>
          <line x1="340" y1="102" x2="340" y2="218" style="stroke:var(--border)" stroke-width="1.5" id="linePC"/>
          <line x1="60"  y1="102" x2="60"  y2="218" style="stroke:var(--border)" stroke-width="1.5" id="lineGS"/>
          <line x1="102" y1="260" x2="298" y2="260" style="stroke:var(--border)" stroke-width="1.5" id="lineSC"/>
          <line x1="102" y1="80"  x2="298" y2="240" style="stroke:var(--border)" stroke-width="1"   stroke-dasharray="4" opacity=".3"/>

          <!-- Grid (top-left) -->
          <circle cx="60" cy="60" r="44" style="fill:var(--card)" stroke="#38bdf8" stroke-width="2"/>
          <text x="60" y="45" text-anchor="middle" font-size="20">🏭</text>
          <text x="60" y="68" text-anchor="middle" style="fill:var(--grid)" font-size="14" font-weight="700" id="svGrid">—</text>
          <text x="60" y="84" text-anchor="middle" style="fill:var(--dim)" font-size="9" id="svGridDir">GRID</text>

          <!-- Production (top-right) -->
          <circle cx="340" cy="60" r="44" style="fill:var(--card)" stroke="#facc15" stroke-width="2"/>
          <text x="340" y="45" text-anchor="middle" font-size="20">☀️</text>
          <text x="340" y="68" text-anchor="middle" style="fill:var(--solar)" font-size="14" font-weight="700" id="svProd">—</text>
          <text x="340" y="84" text-anchor="middle" style="fill:var(--dim)" font-size="9">PRODUCTION</text>

          <!-- Consumption (bottom-right) -->
          <circle cx="340" cy="260" r="44" style="fill:var(--card)" stroke="#f97316" stroke-width="2"/>
          <text x="340" y="245" text-anchor="middle" font-size="20">🏢</text>
          <text x="340" y="268" text-anchor="middle" style="fill:var(--load)" font-size="14" font-weight="700" id="svCons">—</text>
          <text x="340" y="284" text-anchor="middle" style="fill:var(--dim)" font-size="9">CONSUMPTION</text>

          <!-- Storage (bottom-left) -->
          <circle cx="60" cy="260" r="44" style="fill:var(--card)" stroke="#22c55e" stroke-width="2"/>
          <text x="60" y="245" text-anchor="middle" font-size="20">🔋</text>
          <text x="60" y="268" text-anchor="middle" style="fill:var(--ess)" font-size="14" font-weight="700" id="svStor">—</text>
          <text x="60" y="284" text-anchor="middle" style="fill:var(--dim)" font-size="9" id="svStorDir">STORAGE</text>

          <!-- Center self-consumption -->
          <circle cx="200" cy="160" r="34" style="fill:var(--card);stroke:var(--border)" stroke-width="1.5"/>
          <text x="200" y="155" text-anchor="middle" style="fill:var(--solar)" font-size="18" font-weight="700" id="svSelf">—%</text>
          <text x="200" y="173" text-anchor="middle" style="fill:var(--dim)" font-size="8">SELF-USE</text>
        </svg>

        <div class="soc-section">
          <div class="soc-header">
            <span class="lbl">State of Charge</span>
            <span class="val" id="socVal">—%</span>
          </div>
          <div class="soc-track"><div class="soc-fill" id="socFill" style="width:0%"></div></div>
        </div>
      </div>
    </div>

    <!-- RIGHT: Widget Cards -->
    <div class="widget-column">
      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--grid)"></span> Grid <span class="widget-badge" id="wGridBadge">—</span></div>
        <div class="widget-body"><div class="widget-value" style="color:var(--grid)" id="wGridVal">—</div></div>
      </div>
      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--solar)"></span> Production</div>
        <div class="widget-body"><div class="widget-value" style="color:var(--solar)" id="wProdVal">—</div></div>
      </div>
      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--ess)"></span> Storage <span class="widget-badge" id="wStorBadge">—</span></div>
        <div class="widget-body">
          <div class="widget-value" style="color:var(--ess)" id="wStorVal">—</div>
          <div class="widget-sub"><span id="wStorSOC">—</span>% SOC · <span id="wStorTemp">—</span>°C</div>
          <div class="dev-list" id="bmsList"></div>
        </div>
      </div>
      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--load)"></span> Consumption</div>
        <div class="widget-body"><div class="widget-value" style="color:var(--load)" id="wConsVal">—</div></div>
      </div>

      <!-- EV Charger Widget -->
      <div class="widget-card" style="grid-column:1/-1;">
        <div class="widget-header"><span class="widget-dot" style="background:#06b6d4"></span> EV Chargers <span class="widget-badge" id="wEVBadge">—</span></div>
        <div class="widget-body">
          <div style="display:flex;gap:24px;align-items:baseline;margin-bottom:8px;">
            <div class="widget-value" style="color:#06b6d4" id="wEVPower">—</div>
            <div class="widget-sub" id="wEVInfo">—</div>
          </div>
          <div class="dev-list" id="evList"></div>
        </div>
      </div>

      <!-- Device Status (spans full width) -->
      <div class="widget-card" style="grid-column:1/-1;">
        <div class="widget-header"><span class="widget-dot" style="background:#a78bfa"></span> Devices <span class="widget-badge" id="wDevTotal">—</span></div>
        <div class="widget-body">
          <div class="dev-grid" id="devGrid"></div>
        </div>
      </div>

      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--solar)"></span> Self-consumption</div>
        <div class="widget-body">
          <div class="widget-pct"><span id="wSelfPct">—</span><small>%</small></div>
          <div class="widget-bar"><div class="widget-bar-fill" id="wSelfBar" style="width:0%;background:var(--solar)"></div></div>
        </div>
      </div>
      <div class="widget-card">
        <div class="widget-header"><span class="widget-dot" style="background:var(--ess)"></span> Autarchy</div>
        <div class="widget-body">
          <div class="widget-pct"><span id="wAutPct">—</span><small>%</small></div>
          <div class="widget-bar"><div class="widget-bar-fill" id="wAutBar" style="width:0%;background:var(--ess)"></div></div>
        </div>
      </div>
      <div class="constraints-bar">
        <span class="lbl" style="margin-right:4px;">Constraints:</span>
        <span id="cTags" style="color:var(--dim)">None</span>
      </div>
    </div>
  </div>

  <div class="chart-wrap">
    <div class="title">Power History — last 120 cycles</div>
    <canvas id="liveChart"></canvas>
  </div>
</div>

<!-- ==================== HISTORY PAGE ==================== -->
<div class="tab-content" id="tab-history">
  <div class="history-controls">
    <label for="histDate">Select date:</label>
    <input type="date" id="histDate">
    <button onclick="loadHistory()">Load</button>
  </div>
  <div class="summary-grid" id="summaryGrid" style="display:none;">
    <div class="summary-card"><div class="lbl">☀ Peak Solar</div><div class="val"><span id="hPeakSolar">—</span><span class="unit">W</span></div></div>
    <div class="summary-card"><div class="lbl">⚡ Peak Load</div><div class="val"><span id="hPeakLoad">—</span><span class="unit">W</span></div></div>
    <div class="summary-card"><div class="lbl">🔌 Peak Import</div><div class="val"><span id="hPeakImport">—</span><span class="unit">W</span></div></div>
    <div class="summary-card"><div class="lbl">📤 Peak Export</div><div class="val"><span id="hPeakExport">—</span><span class="unit">W</span></div></div>
    <div class="summary-card"><div class="lbl">🔋 SOC Range</div><div class="val"><span id="hSOCRange">—</span></div></div>
    <div class="summary-card"><div class="lbl">🔋 Avg SOC</div><div class="val"><span id="hAvgSOC">—</span><span class="unit">%</span></div></div>
    <div class="summary-card"><div class="lbl">🌡 Avg Temp</div><div class="val"><span id="hAvgTemp">—</span><span class="unit">°C</span></div></div>
    <div class="summary-card"><div class="lbl">🔄 Total Cycles</div><div class="val"><span id="hCycles">—</span></div></div>
  </div>
  <div class="history-chart-wrap" id="histChartWrap" style="display:none;">
    <div class="title">Power Profile — <span id="histChartDate"></span></div>
    <canvas id="histChart"></canvas>
  </div>
  <div class="history-chart-wrap" id="socChartWrap" style="display:none;">
    <div class="title">SOC Profile — <span id="socChartDate"></span></div>
    <canvas id="socChart"></canvas>
  </div>
  <div class="history-empty" id="histEmpty">Select a date and click <b>Load</b> to view historical data.</div>
</div>

<!-- ==================== EDGE CONFIG PAGE ==================== -->
<div class="tab-content" id="tab-config">
  <div class="config-page">
    <h2>⚙️ Edge Configuration</h2>
    <div class="config-section">
      <h3 data-i18n="controllers">Controllers (Priority Order)</h3>
      <div class="config-row"><span class="key">P1 — limit_discharge</span><span class="val">minSoc: 15% · forceSoc: 10%</span></div>
      <div class="config-row"><span class="key">P2 — sell_to_grid_limit</span><span class="val">maxSell: 0 W (zero-export)</span></div>
      <div class="config-row"><span class="key">P3 — peak_shaving</span><span class="val">threshold: 8000 W</span></div>
      <div class="config-row"><span class="key">P3 — time_of_use</span><span class="val">schedule: 22:00–06:00 cheap</span></div>
      <div class="config-row"><span class="key">P4 — balancing</span><span class="val">self-consumption</span></div>
    </div>
    <div class="config-section">
      <h3 data-i18n="battery_ess">Battery ESS</h3>
      <div class="config-row"><span class="key">Capacity</span><span class="val">5,000 Wh</span></div>
      <div class="config-row"><span class="key">Max Charge Rate</span><span class="val">2,500 W</span></div>
      <div class="config-row"><span class="key">Max Discharge Rate</span><span class="val">5,000 W</span></div>
      <div class="config-row"><span class="key">Min SOC</span><span class="val">15%</span></div>
      <div class="config-row"><span class="key">Force Charge SOC</span><span class="val">10%</span></div>
    </div>
    <div class="config-section">
      <h3 data-i18n="devices_modbus">Devices (Modbus TCP)</h3>
      <div class="config-row"><span class="key">Grid Meter</span><span class="val">Unit 1 · FC03</span></div>
      <div class="config-row"><span class="key">Solar Meter</span><span class="val">Unit 2 · FC03</span></div>
      <div class="config-row"><span class="key">BMS</span><span class="val">Unit 3 · FC03</span></div>
      <div class="config-row"><span class="key">PCS</span><span class="val">Unit 4 · FC03/FC16</span></div>
    </div>
    <div class="config-section">
      <h3 data-i18n="cycle">Cycle</h3>
      <div class="config-row"><span class="key">Interval</span><span class="val">1,000 ms</span></div>
      <div class="config-row"><span class="key">Resolver</span><span class="val">Constraint-based (strictest wins)</span></div>
      <div class="config-row"><span class="key">Scheduler</span><span class="val">FixedOrder</span></div>
    </div>
  </div>
</div>

<!-- ==================== SYSTEM HEALTH PAGE ==================== -->
<div class="tab-content" id="tab-health">
  <div class="config-page">
    <h2>💚 System Health</h2>
    <div class="config-section" id="healthSection"><h3>Loading...</h3></div>
  </div>
</div>

<!-- ==================== ABOUT PAGE ==================== -->
<div class="tab-content" id="tab-about">
  <div class="about-page">
    <h2>ℹ️ About</h2>
    <div class="about-card">
      <h3>Local EMS</h3>
      <div class="about-row"><span class="key">Version</span><span class="val">2026.10.0</span></div>
      <div class="about-row"><span class="key">Architecture</span><span class="val">Constraint-based Controller Chain</span></div>
      <div class="about-row"><span class="key">Runtime</span><span class="val">Go 1.27</span></div>
      <div class="about-row"><span class="key">Storage</span><span class="val">SQLite (local)</span></div>
      <div class="about-row"><span class="key">Protocol</span><span class="val">Modbus TCP (zero-dependency)</span></div>
    </div>
  </div>
</div>

<!-- ==================== USER PAGE ==================== -->
<div class="tab-content" id="tab-user">
  <div class="config-page">
    <h2>👤 User</h2>
    <div class="config-section">
      <h3 data-i18n="user_info">User Information</h3>
      <div class="config-row"><span class="key" data-i18n="username">Username</span><span class="val">admin</span></div>
      <div class="config-row"><span class="key" data-i18n="role">Role</span><span class="val">Owner</span></div>
      <div class="config-row"><span class="key" data-i18n="last_login">Last Login</span><span class="val" id="uLastLogin">—</span></div>
    </div>
    <div class="config-section">
      <h3 data-i18n="contact">Contact Details</h3>
      <div class="config-row"><span class="key" data-i18n="company">Company</span><span class="val">—</span></div>
      <div class="config-row"><span class="key" data-i18n="email">Email</span><span class="val">admin@local-ems.io</span></div>
      <div class="config-row"><span class="key" data-i18n="phone">Phone</span><span class="val">—</span></div>
      <div class="config-row"><span class="key" data-i18n="address">Address</span><span class="val">—</span></div>
    </div>
    <div class="config-section">
      <h3 data-i18n="general_settings">General Settings</h3>
      <div class="config-row"><span class="key" data-i18n="language">Language</span><span class="val">
        <select id="langSelect" onchange="setLang(this.value)" style="background:var(--select-bg);color:var(--select-text);border:1px solid var(--border);border-radius:6px;padding:4px 12px;font-size:13px;cursor:pointer;outline:none;">
          <option value="en">🇬🇧 English</option>
          <option value="vi">🇻🇳 Tiếng Việt</option>
        </select>
      </span></div>
      <div class="config-row"><span class="key" data-i18n="theme">Theme</span><span class="val">
        <select id="themeSelect" onchange="setTheme(this.value)" style="background:var(--select-bg);color:var(--select-text);border:1px solid var(--border);border-radius:6px;padding:4px 12px;font-size:13px;cursor:pointer;outline:none;">
          <option value="dark">🌙 Dark</option>
          <option value="light">☀️ Light</option>
          <option value="system">💻 System</option>
        </select>
      </span></div>
      <div class="config-row"><span class="key" data-i18n="session_timeout">Session Timeout</span><span class="val">24 hours</span></div>
      <div class="config-row"><span class="key" data-i18n="timezone">Timezone</span><span class="val" id="uTimezone">—</span></div>
    </div>
    <div style="padding-top:12px;">
      <a href="/logout" data-i18n="sign_out" style="display:inline-block;padding:10px 24px;background:var(--danger);color:#fff;border-radius:6px;font-size:13px;font-weight:600;text-decoration:none;">Sign Out</a>
    </div>
  </div>
</div>
</div><!-- /main-content -->

<footer class="footer">Local EMS | Version 2026.10.0</footer>

<script>
// ==================== ROUTING ====================
const ACTIVE_PAGE = /*ACTIVE_PAGE*/ || 'live';
const pageUrls = { live:'/live', history:'/history', config:'/settings/config', health:'/settings/health', about:'/settings/about', user:'/user' };
const pageTitles = { live:'⚡ Live', history:'📊 History', config:'⚙️ Edge Config', health:'💚 System Health', about:'ℹ️ About', user:'👤 User' };

function openSidebar() { document.getElementById('sidebar').classList.add('open'); document.getElementById('sidebarOverlay').classList.add('open'); }
function closeSidebar() { document.getElementById('sidebar').classList.remove('open'); document.getElementById('sidebarOverlay').classList.remove('open'); }

// Populate user page dynamic fields
document.getElementById('uTimezone').textContent = Intl.DateTimeFormat().resolvedOptions().timeZone;
document.getElementById('uLastLogin').textContent = new Date().toLocaleString();

// Theme management
function setTheme(theme) {
  localStorage.setItem('ems-theme', theme);
  document.documentElement.setAttribute('data-theme', theme);
}
// Init theme select dropdown
(function() {
  var saved = localStorage.getItem('ems-theme') || 'dark';
  var sel = document.getElementById('themeSelect');
  if (sel) sel.value = saved;
})();

// ===== i18n Language System =====
const i18n = {
  en: {
    // Nav
    live: 'Live', history: 'History', config: 'Edge Config', health: 'System Health', about: 'About', user: 'User',
    nav_system: 'System', nav_account: 'Account',
    // Page titles
    live_title: '📊 Live Dashboard', history_title: '📈 History', config_title: '⚙️ Edge Configuration',
    health_title: '💚 System Health', about_title: 'ℹ️ About', user_title: '👤 User',
    // Widgets
    grid: 'Grid', production: 'Production', storage: 'Storage', consumption: 'Consumption',
    ev_chargers: 'EV Chargers', devices: 'Devices', self_consumption: 'Self-Consumption', autarchy: 'Autarchy',
    // Config
    controllers: 'Controllers (Priority Order)', battery_ess: 'Battery ESS', devices_modbus: 'Devices (Modbus TCP)',
    cycle: 'Cycle', general_settings: 'General Settings', language: 'Language', theme: 'Theme',
    session_timeout: 'Session Timeout', timezone: 'Timezone',
    // User
    user_info: 'User Information', username: 'Username', role: 'Role', last_login: 'Last Login',
    contact: 'Contact Details', company: 'Company', email: 'Email', phone: 'Phone', address: 'Address',
    sign_out: 'Sign Out',
    // About
    version: 'Version', architecture: 'Architecture', runtime: 'Runtime', storage_label: 'Storage', protocol: 'Protocol',
    // Misc
    capacity: 'Capacity', max_charge: 'Max Charge Rate', max_discharge: 'Max Discharge Rate',
    min_soc: 'Min SOC', force_charge_soc: 'Force Charge SOC', interval: 'Interval',
    resolver: 'Resolver', scheduler: 'Scheduler',
  },
  vi: {
    // Nav
    live: 'Trực tiếp', history: 'Lịch sử', config: 'Cấu hình', health: 'Sức khỏe HT', about: 'Thông tin', user: 'Người dùng',
    nav_system: 'Hệ thống', nav_account: 'Tài khoản',
    // Page titles
    live_title: '📊 Bảng điều khiển', history_title: '📈 Lịch sử', config_title: '⚙️ Cấu hình hệ thống',
    health_title: '💚 Sức khỏe hệ thống', about_title: 'ℹ️ Thông tin', user_title: '👤 Người dùng',
    // Widgets
    grid: 'Lưới điện', production: 'Sản xuất', storage: 'Lưu trữ', consumption: 'Tiêu thụ',
    ev_chargers: 'Trạm sạc EV', devices: 'Thiết bị', self_consumption: 'Tự tiêu thụ', autarchy: 'Tự chủ NL',
    // Config
    controllers: 'Bộ điều khiển (Ưu tiên)', battery_ess: 'Pin ESS', devices_modbus: 'Thiết bị (Modbus TCP)',
    cycle: 'Chu kỳ', general_settings: 'Cài đặt chung', language: 'Ngôn ngữ', theme: 'Giao diện',
    session_timeout: 'Hết phiên', timezone: 'Múi giờ',
    // User
    user_info: 'Thông tin người dùng', username: 'Tên đăng nhập', role: 'Vai trò', last_login: 'Đăng nhập cuối',
    contact: 'Thông tin liên hệ', company: 'Công ty', email: 'Email', phone: 'Điện thoại', address: 'Địa chỉ',
    sign_out: 'Đăng xuất',
    // About
    version: 'Phiên bản', architecture: 'Kiến trúc', runtime: 'Nền tảng', storage_label: 'Lưu trữ', protocol: 'Giao thức',
    // Misc
    capacity: 'Dung lượng', max_charge: 'Sạc tối đa', max_discharge: 'Xả tối đa',
    min_soc: 'SOC tối thiểu', force_charge_soc: 'SOC ép sạc', interval: 'Chu kỳ',
    resolver: 'Bộ xử lý', scheduler: 'Bộ lập lịch',
  }
};

function setLang(lang) {
  localStorage.setItem('ems-lang', lang);
  applyLang(lang);
}

function applyLang(lang) {
  const dict = i18n[lang] || i18n.en;
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (dict[key]) el.textContent = dict[key];
  });
  // Update page titles map
  pageTitles.live = dict.live_title || pageTitles.live;
  pageTitles.history = dict.history_title || pageTitles.history;
  pageTitles.config = dict.config_title || pageTitles.config;
  pageTitles.health = dict.health_title || pageTitles.health;
  pageTitles.about = dict.about_title || pageTitles.about;
  pageTitles.user = dict.user_title || pageTitles.user;
  // Re-apply current page title
  const activePage = document.querySelector('.tab-content.active');
  if (activePage) {
    const pageId = activePage.id.replace('tab-','');
    document.getElementById('pageTitle').textContent = pageTitles[pageId] || pageId;
  }
}

// Init language
(function() {
  var saved = localStorage.getItem('ems-lang') || 'en';
  var sel = document.getElementById('langSelect');
  if (sel) sel.value = saved;
  if (saved !== 'en') applyLang(saved);
})();

function activatePage(page, pushHistory) {
  // Hide all pages, show target
  document.querySelectorAll('.tab-content').forEach(e => e.classList.remove('active'));
  const t = document.getElementById('tab-'+page);
  if (t) t.classList.add('active');
  // Update nav highlight
  document.querySelectorAll('.sidebar-nav .nav-item').forEach(e => {
    e.classList.toggle('active', e.getAttribute('data-page') === page);
  });
  // Update header title
  document.getElementById('pageTitle').textContent = pageTitles[page] || page;
  // Update browser URL
  if (pushHistory && pageUrls[page]) {
    history.pushState({page:page}, '', pageUrls[page]);
  }
  // Close sidebar on mobile
  closeSidebar();
  // Dynamic data
  if (page === 'health') loadHealthData();
}

// SPA navigation: intercept sidebar links
document.querySelectorAll('.sidebar-nav .nav-item[data-page]').forEach(link => {
  link.addEventListener('click', function(e) {
    e.preventDefault();
    const page = this.getAttribute('data-page');
    activatePage(page, true);
  });
});

// Browser back/forward
window.addEventListener('popstate', function(e) {
  if (e.state && e.state.page) activatePage(e.state.page, false);
});

// Activate page from server injection on load
activatePage(ACTIVE_PAGE, false);
// Set initial history state
history.replaceState({page: ACTIVE_PAGE}, '', pageUrls[ACTIVE_PAGE] || location.pathname);

async function loadHealthData() {
  try {
    const res = await fetch('/api/health'); const data = await res.json();
    const s = data.status, st = data.stats;
    document.getElementById('healthSection').innerHTML =
      '<h3>Status: '+(s.healthy?'✅ Healthy':'❌ Unhealthy')+'</h3>'+
      '<div class="config-row"><span class="key">Uptime</span><span class="val">'+s.uptime+'</span></div>'+
      '<div class="config-row"><span class="key">Message</span><span class="val">'+s.message+'</span></div>'+
      '<div class="config-row"><span class="key">Total Cycles</span><span class="val">'+(st.cycle_count||0).toLocaleString()+'</span></div>'+
      '<div class="config-row"><span class="key">Errors</span><span class="val">'+(st.error_count||0)+'</span></div>'+
      '<div class="config-row"><span class="key">Panics Recovered</span><span class="val">'+(st.panic_count||0)+'</span></div>'+
      '<div class="config-row"><span class="key">Last Cycle</span><span class="val">'+(st.last_cycle||'—')+'</span></div>';
  } catch(e) { document.getElementById('healthSection').innerHTML = '<h3>Error</h3><p>'+e.message+'</p>'; }
}

// ==================== LIVE ====================
const MAX=120, hist={s:[],l:[],g:[],e:[]};

function connectWS() {
  const ws = new WebSocket('ws://'+location.host+'/ws');
  ws.onopen = () => { document.getElementById('dot').className='dot on'; document.getElementById('connStatus').textContent='Live'; };
  ws.onclose = () => { document.getElementById('dot').className='dot off'; document.getElementById('connStatus').textContent='Reconnecting...'; setTimeout(connectWS,2000); };
  ws.onmessage = e => updateLive(JSON.parse(e.data));
}

function kw(w) { return Math.abs(w)>=1000 ? (Math.abs(w)/1000).toFixed(1)+' kW' : Math.abs(w)+' W'; }

function updateLive(d) {
  const solar = d.solar_w||0, grid = d.grid_w||0, load = d.load_w||0, ess = d.ess_w||0;

  // SVG flow diagram
  document.getElementById('svGrid').textContent = kw(grid);
  document.getElementById('svGridDir').textContent = grid >= 0 ? 'BUYING' : 'SELLING';
  document.getElementById('svProd').textContent = kw(solar);
  document.getElementById('svCons').textContent = kw(load);
  document.getElementById('svStor').textContent = kw(ess);
  document.getElementById('svStorDir').textContent = ess >= 0 ? 'DISCHARGE' : 'CHARGE';

  // Flow line colors (active when power > 100W)
  const inactiveStroke = getComputedStyle(document.documentElement).getPropertyValue('--border').trim();
  document.getElementById('linePG').style.stroke = solar > 100 ? '#facc15' : inactiveStroke;
  document.getElementById('linePC').style.stroke = solar > 100 ? '#facc15' : inactiveStroke;
  document.getElementById('lineGS').style.stroke = Math.abs(ess) > 100 ? '#22c55e' : inactiveStroke;
  document.getElementById('lineSC').style.stroke = Math.abs(ess) > 100 ? '#22c55e' : inactiveStroke;

  // Self-consumption %: how much of production is used locally
  const selfPct = solar > 0 ? Math.min(100, Math.round((1 - Math.max(0,-grid) / solar) * 100)) : 0;
  document.getElementById('svSelf').textContent = selfPct + '%';

  // Autarchy %: how much consumption comes from local sources
  const autPct = load > 0 ? Math.min(100, Math.round((1 - Math.max(0,grid) / load) * 100)) : 0;

  // Widget cards
  document.getElementById('wGridVal').textContent = kw(grid);
  document.getElementById('wGridBadge').textContent = grid >= 0 ? 'Buy' : 'Sell';
  document.getElementById('wGridBadge').style.background = grid >= 0 ? '#dc262633' : '#22c55e33';
  document.getElementById('wGridBadge').style.color = grid >= 0 ? '#f87171' : '#22c55e';

  document.getElementById('wProdVal').textContent = kw(solar);

  document.getElementById('wStorVal').textContent = kw(ess);
  document.getElementById('wStorBadge').textContent = ess >= 0 ? 'Discharge' : 'Charge';
  document.getElementById('wStorBadge').style.background = ess >= 0 ? '#f9731633' : '#22c55e33';
  document.getElementById('wStorBadge').style.color = ess >= 0 ? '#f97316' : '#22c55e';
  document.getElementById('wStorSOC').textContent = (d.soc||0).toFixed(1);
  document.getElementById('wStorTemp').textContent = (d.temp_c||0).toFixed(1);

  document.getElementById('wConsVal').textContent = kw(load);

  document.getElementById('wSelfPct').textContent = selfPct;
  document.getElementById('wSelfBar').style.width = selfPct + '%';

  document.getElementById('wAutPct').textContent = autPct;
  document.getElementById('wAutBar').style.width = autPct + '%';

  // SOC
  const soc = (d.soc||0).toFixed(1);
  document.getElementById('socVal').textContent = soc+'%';
  const fill = document.getElementById('socFill');
  fill.style.width = soc+'%';
  fill.style.background = d.soc>50?'var(--ess)':d.soc>20?'var(--warn)':'var(--danger)';

  // Constraints
  const ct = document.getElementById('cTags');
  if (d.constraints && d.constraints.length>0) {
    ct.innerHTML = d.constraints.map(c=>'<span class="tag'+(d.clamped?' clamped':'')+'">'+c+'</span>').join(' ');
  } else { ct.innerHTML = '<span style="color:var(--dim)">None</span>'; }

  // Chart data
  hist.s.push(solar); hist.l.push(load); hist.g.push(grid); hist.e.push(ess);
  for(const k in hist) if(hist[k].length>MAX) hist[k].shift();
  drawLiveChart();
}

// ==================== DEVICE POLLING ====================
async function updateDevices() {
  try {
    const res = await fetch('/api/devices');
    if (!res.ok) return;
    const d = await res.json();

    // BMS breakdown inside Storage widget
    const bl = document.getElementById('bmsList');
    if (bl && d.bms) {
      bl.innerHTML = d.bms.map(b =>
        '<div class="dev-row">' +
          '<span class="dev-id"><span class="dev-dot '+(b.online?'on':'off')+'"></span>'+b.id+'</span>' +
          '<span class="dev-val">'+b.soc.toFixed(1)+'% · '+b.temp_c.toFixed(1)+'°C</span>' +
        '</div>'
      ).join('');
    }

    // EV Charger breakdown
    const evStatuses = ['Available','Charging','SuspEV','SuspEVSE','Finishing','Fault'];
    const el = document.getElementById('evList');
    if (el && d.ev_charger) {
      el.innerHTML = d.ev_charger.map(ev => {
        const st = evStatuses[ev.status] || 'Unknown';
        const stColor = ev.status===1?'#06b6d4':ev.status===5?'var(--danger)':'var(--dim)';
        return '<div class="dev-row">' +
          '<span class="dev-id"><span class="dev-dot '+(ev.online?'on':'off')+'"></span>'+ev.id+'</span>' +
          '<span class="dev-val" style="color:'+stColor+'">'+st+(ev.status===1?' · '+kw(ev.active_power_w):'')+(ev.vehicle_connected?' · 🔌':'')+'</span>' +
        '</div>';
      }).join('');
    }
    // EV summary
    const evBadge = document.getElementById('wEVBadge');
    const evPower = document.getElementById('wEVPower');
    const evInfo = document.getElementById('wEVInfo');
    if (evBadge && d.ev_charger) {
      const charging = d.ev_charging || 0;
      const total = d.ev_charger.length;
      evBadge.textContent = charging + '/' + total + ' charging';
      evBadge.style.background = charging > 0 ? '#06b6d433' : 'var(--border)';
      evBadge.style.color = charging > 0 ? '#06b6d4' : 'var(--text)';
      evPower.textContent = kw(d.total_ev_power_w || 0);
      evInfo.textContent = 'Total EV load · ' + (d.ev_charger.filter(e=>e.vehicle_connected).length) + ' vehicles connected';
    }

    // Device grid (all devices)
    const dg = document.getElementById('devGrid');
    if (dg) {
      let html = '';
      // BMS chips
      (d.bms||[]).forEach(b => {
        html += '<a href="/component/'+b.id+'" class="dev-chip" style="text-decoration:none;color:inherit;">' +
          '<div class="dev-name"><span class="dev-dot '+(b.online?'on':'off')+'"></span>'+b.id+'</div>' +
          '<div class="dev-info">🔋 '+b.soc.toFixed(1)+'% · '+b.temp_c.toFixed(1)+'°C</div>' +
        '</a>';
      });
      // PCS chips
      (d.pcs||[]).forEach(p => {
        const st = p.status===1?'RUN':p.status===2?'FAULT':'STBY';
        const cls = p.status===2?' fault':'';
        html += '<a href="/component/'+p.id+'" class="dev-chip'+cls+'" style="text-decoration:none;color:inherit;">' +
          '<div class="dev-name"><span class="dev-dot '+(p.online?'on':'off')+'"></span>'+p.id+'</div>' +
          '<div class="dev-info">⚡ '+kw(p.active_power_w)+' · '+st+'</div>' +
        '</a>';
      });
      // Meter chips
      (d.meter||[]).forEach(m => {
        html += '<a href="/component/'+m.id+'" class="dev-chip" style="text-decoration:none;color:inherit;">' +
          '<div class="dev-name"><span class="dev-dot '+(m.online?'on':'off')+'"></span>'+m.id+'</div>' +
          '<div class="dev-info">📡 '+kw(m.active_power_w)+' · '+m.frequency_hz.toFixed(1)+'Hz</div>' +
        '</a>';
      });
      // EV Charger chips
      (d.ev_charger||[]).forEach(ev => {
        const st = evStatuses[ev.status] || '?';
        const cls = ev.status===5?' fault':'';
        html += '<a href="/component/'+ev.id+'" class="dev-chip'+cls+'" style="text-decoration:none;color:inherit;">' +
          '<div class="dev-name"><span class="dev-dot '+(ev.online?'on':'off')+'"></span>'+ev.id+'</div>' +
          '<div class="dev-info">🔌 '+st+(ev.status===1?' · '+kw(ev.active_power_w):'')+'</div>' +
        '</a>';
      });
      dg.innerHTML = html;
    }

    // Device total badge
    const total = (d.bms_online||0)+(d.pcs_online||0)+(d.meter_online||0)+(d.ev_charger_online||0);
    const totalAll = (d.bms||[]).length+(d.pcs||[]).length+(d.meter||[]).length+(d.ev_charger||[]).length;
    const badge = document.getElementById('wDevTotal');
    if (badge) {
      badge.textContent = total+'/'+totalAll+' online';
      badge.style.background = total===totalAll ? '#22c55e33' : '#dc262633';
      badge.style.color = total===totalAll ? '#22c55e' : '#f87171';
    }
  } catch(e) {}
}
// Poll devices every 2 seconds
setInterval(updateDevices, 2000);
updateDevices();

function drawLiveChart() {
  const c=document.getElementById('liveChart'), ctx=c.getContext('2d');
  const dpr=devicePixelRatio||1, r=c.getBoundingClientRect();
  c.width=r.width*dpr; c.height=r.height*dpr; ctx.scale(dpr,dpr);
  const W=r.width, H=r.height;
  ctx.clearRect(0,0,W,H);
  const all=[...hist.s,...hist.l,...hist.g,...hist.e];
  if(!all.length) return;
  const mx=Math.max(Math.abs(Math.max(...all)),Math.abs(Math.min(...all)),1000), mid=H/2;

  const cs = getComputedStyle(document.documentElement);
  const gridLine = cs.getPropertyValue('--border').trim();
  const axisLine = cs.getPropertyValue('--dim').trim();
  const legendText = cs.getPropertyValue('--muted').trim();

  ctx.strokeStyle=gridLine; ctx.lineWidth=.5;
  for(let i=0;i<=4;i++){const y=H/4*i; ctx.beginPath();ctx.moveTo(0,y);ctx.lineTo(W,y);ctx.stroke();}
  ctx.strokeStyle=axisLine; ctx.lineWidth=1; ctx.beginPath();ctx.moveTo(0,mid);ctx.lineTo(W,mid);ctx.stroke();

  [{d:hist.s,c:'#facc15'},{d:hist.l,c:'#f97316'},{d:hist.g,c:'#38bdf8'},{d:hist.e,c:'#22c55e'}].forEach(s=>{
    if(s.d.length<2)return; ctx.strokeStyle=s.c; ctx.lineWidth=1.5; ctx.beginPath();
    s.d.forEach((v,i)=>{const x=i/(MAX-1)*W, y=mid-v/mx*mid*.9; i?ctx.lineTo(x,y):ctx.moveTo(x,y);}); ctx.stroke();
  });

  ctx.font='11px sans-serif'; let lx=8;
  [{t:'Solar',c:'#facc15'},{t:'Load',c:'#f97316'},{t:'Grid',c:'#38bdf8'},{t:'ESS',c:'#22c55e'}].forEach(l=>{
    ctx.fillStyle=l.c; ctx.fillRect(lx,6,12,3); ctx.fillStyle=legendText; ctx.fillText(l.t,lx+16,12); lx+=ctx.measureText(l.t).width+32;
  });
}

// ==================== HISTORY ====================
document.getElementById('histDate').value = new Date().toISOString().slice(0,10);

async function loadHistory() {
  const date = document.getElementById('histDate').value;
  if(!date) return;
  const sumRes = await fetch('/api/history/summary?date='+date);
  const sum = await sumRes.json();
  if(sum.total_cycles===0) {
    document.getElementById('summaryGrid').style.display='none';
    document.getElementById('histChartWrap').style.display='none';
    document.getElementById('socChartWrap').style.display='none';
    document.getElementById('histEmpty').style.display='block';
    document.getElementById('histEmpty').innerHTML='No data for <b>'+date+'</b>.';
    return;
  }
  document.getElementById('histEmpty').style.display='none';
  document.getElementById('summaryGrid').style.display='grid';
  document.getElementById('histChartWrap').style.display='block';
  document.getElementById('socChartWrap').style.display='block';
  document.getElementById('hPeakSolar').textContent = sum.peak_solar_w.toLocaleString();
  document.getElementById('hPeakLoad').textContent = sum.peak_load_w.toLocaleString();
  document.getElementById('hPeakImport').textContent = sum.peak_grid_import.toLocaleString();
  document.getElementById('hPeakExport').textContent = Math.abs(sum.peak_grid_export).toLocaleString();
  document.getElementById('hSOCRange').textContent = sum.min_soc.toFixed(1)+'% — '+sum.max_soc.toFixed(1)+'%';
  document.getElementById('hAvgSOC').textContent = sum.avg_soc.toFixed(1);
  document.getElementById('hAvgTemp').textContent = sum.avg_temp_c.toFixed(1);
  document.getElementById('hCycles').textContent = sum.total_cycles.toLocaleString();
  const dataRes = await fetch('/api/history/date?date='+date);
  const data = await dataRes.json();
  document.getElementById('histChartDate').textContent = date;
  document.getElementById('socChartDate').textContent = date;
  drawHistoryChart(data);
  drawSOCChart(data);
}

function drawHistoryChart(data) {
  const c=document.getElementById('histChart'), ctx=c.getContext('2d');
  const dpr=devicePixelRatio||1, r=c.getBoundingClientRect();
  c.width=r.width*dpr; c.height=r.height*dpr; ctx.scale(dpr,dpr);
  const W=r.width, H=r.height; ctx.clearRect(0,0,W,H);
  if(!data||!data.length) return;
  const allV=[]; data.forEach(d=>{allV.push(d.SolarW||d.solar_w||0,d.LoadW||d.load_w||0,d.GridW||d.grid_w||0,d.ESSW||d.ess_w||0);});
  const mx=Math.max(Math.abs(Math.max(...allV)),Math.abs(Math.min(...allV)),1000), mid=H/2, N=data.length;
  const cs2 = getComputedStyle(document.documentElement);
  const gl2 = cs2.getPropertyValue('--border').trim();
  const al2 = cs2.getPropertyValue('--dim').trim();
  const lt2 = cs2.getPropertyValue('--muted').trim();
  ctx.strokeStyle=gl2; ctx.lineWidth=.5;
  for(let i=0;i<=4;i++){const y=H/4*i; ctx.beginPath();ctx.moveTo(0,y);ctx.lineTo(W,y);ctx.stroke();}
  ctx.strokeStyle=al2; ctx.lineWidth=1; ctx.beginPath();ctx.moveTo(0,mid);ctx.lineTo(W,mid);ctx.stroke();
  const step=Math.max(1,Math.floor(N/600));
  [{k1:'SolarW',k2:'solar_w',c:'#facc15'},{k1:'LoadW',k2:'load_w',c:'#f97316'},{k1:'GridW',k2:'grid_w',c:'#38bdf8'},{k1:'ESSW',k2:'ess_w',c:'#22c55e'}].forEach(s=>{
    ctx.strokeStyle=s.c; ctx.lineWidth=1.2; ctx.beginPath(); let f=true;
    for(let i=0;i<N;i+=step){const v=data[i][s.k1]||data[i][s.k2]||0,x=i/(N-1)*W,y=mid-v/mx*mid*.9; f?(ctx.moveTo(x,y),f=false):ctx.lineTo(x,y);} ctx.stroke();
  });
  ctx.font='11px sans-serif'; let lx=8;
  [{t:'Solar',c:'#facc15'},{t:'Load',c:'#f97316'},{t:'Grid',c:'#38bdf8'},{t:'ESS',c:'#22c55e'}].forEach(l=>{
    ctx.fillStyle=l.c; ctx.fillRect(lx,6,12,3); ctx.fillStyle=lt2; ctx.fillText(l.t,lx+16,12); lx+=ctx.measureText(l.t).width+32;
  });
  ctx.fillStyle=al2; ctx.font='10px sans-serif';
  [0,.25,.5,.75,1].forEach(p=>{const i=Math.min(Math.floor(p*(N-1)),N-1);const ts=data[i].Timestamp||data[i].timestamp||'';const t=ts.includes('T')?ts.split('T')[1].slice(0,5):ts.slice(11,16);const x=p*W; ctx.fillText(t,x<W-30?x+2:x-28,H-4);});
}

function drawSOCChart(data) {
  const c=document.getElementById('socChart'), ctx=c.getContext('2d');
  const dpr=devicePixelRatio||1, r=c.getBoundingClientRect();
  c.width=r.width*dpr; c.height=r.height*dpr; ctx.scale(dpr,dpr);
  const W=r.width, H=r.height; ctx.clearRect(0,0,W,H);
  if(!data||!data.length) return;
  const N=data.length, step=Math.max(1,Math.floor(N/600));
  const cs3 = getComputedStyle(document.documentElement);
  const gl3 = cs3.getPropertyValue('--border').trim();
  const al3 = cs3.getPropertyValue('--dim').trim();
  ctx.strokeStyle=gl3; ctx.lineWidth=.5;
  for(let i=0;i<=4;i++){const y=H/4*i; ctx.beginPath();ctx.moveTo(0,y);ctx.lineTo(W,y);ctx.stroke();ctx.fillStyle=al3;ctx.font='10px sans-serif';ctx.fillText((100-25*i)+'%',4,y+12);}
  ctx.beginPath(); ctx.moveTo(0,H); let f=true;
  for(let i=0;i<N;i+=step){const soc=data[i].SOC||data[i].soc||0,x=i/(N-1)*W,y=H-soc/100*H; f?(ctx.lineTo(x,y),f=false):ctx.lineTo(x,y);}
  ctx.lineTo(W,H); ctx.closePath();
  const grad=ctx.createLinearGradient(0,0,0,H); grad.addColorStop(0,'rgba(34,197,94,0.3)'); grad.addColorStop(1,'rgba(34,197,94,0.02)');
  ctx.fillStyle=grad; ctx.fill();
  ctx.strokeStyle='#22c55e'; ctx.lineWidth=2; ctx.beginPath(); f=true;
  for(let i=0;i<N;i+=step){const soc=data[i].SOC||data[i].soc||0,x=i/(N-1)*W,y=H-soc/100*H; f?(ctx.moveTo(x,y),f=false):ctx.lineTo(x,y);} ctx.stroke();
  ctx.fillStyle=al3; ctx.font='10px sans-serif';
  [0,.25,.5,.75,1].forEach(p=>{const i=Math.min(Math.floor(p*(N-1)),N-1);const ts=data[i].Timestamp||data[i].timestamp||'';const t=ts.includes('T')?ts.split('T')[1].slice(0,5):ts.slice(11,16);const x=p*W;ctx.fillText(t,x<W-30?x+2:x-28,H-4);});
}

connectWS();
</script>
</body>
</html>`
