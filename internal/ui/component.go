package ui

// componentHTML is the standalone component detail page.
var componentHTML = `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Component Detail — Local EMS</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  :root {
    --bg:#0f172a; --card:#1e293b; --border:#334155; --text:#f1f5f9;
    --muted:#94a3b8; --dim:#475569; --ess:#22c55e; --solar:#facc15;
    --grid:#3b82f6; --load:#f97316; --danger:#ef4444; --warn:#f59e0b;
    --success:#22c55e;
  }
  body { background:var(--bg); color:var(--text); font-family:-apple-system,system-ui,sans-serif; min-height:100vh; }

  .header { display:flex; align-items:center; gap:12px; padding:12px 20px; background:var(--card); border-bottom:1px solid var(--border); }
  .header a { color:var(--muted); text-decoration:none; font-size:20px; }
  .header a:hover { color:var(--text); }
  .header h1 { font-size:16px; font-weight:600; }
  .header .comp-badge { font-size:11px; padding:3px 10px; border-radius:12px; font-weight:600; text-transform:uppercase; }
  .badge-bms { background:#22c55e33; color:#22c55e; }
  .badge-pcs { background:#3b82f633; color:#3b82f6; }
  .badge-meter { background:#f9731633; color:#f97316; }
  .badge-ev { background:#06b6d433; color:#06b6d4; }

  .comp-page { padding:20px; max-width:900px; margin:0 auto; }
  .status-bar { display:flex; gap:12px; margin-bottom:20px; flex-wrap:wrap; }
  .status-chip { background:var(--card); border:1px solid var(--border); border-radius:8px; padding:12px 16px; flex:1; min-width:120px; }
  .status-chip .label { font-size:11px; color:var(--muted); text-transform:uppercase; margin-bottom:4px; }
  .status-chip .value { font-size:20px; font-weight:700; font-variant-numeric:tabular-nums; }

  .detail-section { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:16px 20px; margin-bottom:16px; }
  .detail-section h3 { font-size:13px; color:var(--solar); margin-bottom:12px; text-transform:uppercase; }
  .detail-row { display:flex; justify-content:space-between; padding:6px 0; border-bottom:1px solid var(--border); font-size:13px; }
  .detail-row:last-child { border-bottom:none; }
  .detail-row .key { color:var(--muted); }
  .detail-row .val { color:var(--text); font-weight:500; font-variant-numeric:tabular-nums; }

  .cell-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(42px,1fr)); gap:3px; margin-top:8px; }
  .cell { text-align:center; font-size:9px; padding:4px 2px; border-radius:4px; font-variant-numeric:tabular-nums; font-weight:600; }

  .chart-wrap { background:var(--card); border:1px solid var(--border); border-radius:10px; padding:16px; margin-bottom:16px; }
  .chart-wrap .title { font-size:12px; color:var(--muted); margin-bottom:8px; }

  .online-dot { display:inline-block; width:8px; height:8px; border-radius:50%; margin-right:6px; vertical-align:middle; }
  .online-dot.on { background:var(--success); }
  .online-dot.off { background:var(--danger); }

  footer { text-align:center; padding:12px; font-size:11px; color:var(--dim); border-top:1px solid var(--border); }
</style>
</head><body>

<div class="header">
  <a href="/live" title="Back to Live">←</a>
  <h1 id="compTitle">Loading...</h1>
  <span class="comp-badge" id="compBadge"></span>
</div>

<div class="comp-page">
  <div class="status-bar" id="statusBar"></div>
  <div id="detailContent"></div>
</div>

<footer>Local EMS | Version 2026.10.0</footer>

<script>
const COMP_ID = /*COMP_ID*/ || '';
const compType = COMP_ID.startsWith('bms') ? 'bms' : COMP_ID.startsWith('pcs') ? 'pcs' : COMP_ID.startsWith('ev') ? 'ev' : 'meter';

function kw(w) { return Math.abs(w)>=1000 ? (Math.abs(w)/1000).toFixed(1)+' kW' : Math.abs(w)+' W'; }

// Set header
document.getElementById('compTitle').textContent = COMP_ID.toUpperCase();
const badge = document.getElementById('compBadge');
badge.textContent = compType.toUpperCase();
badge.className = 'comp-badge badge-' + compType;

// Cell voltage color
function cellColor(mv) {
  if (mv < 2900) return 'background:#ef444488;color:#fca5a5';
  if (mv < 3100) return 'background:#f59e0b44;color:#fcd34d';
  if (mv > 3600) return 'background:#ef444488;color:#fca5a5';
  if (mv > 3500) return 'background:#f59e0b44;color:#fcd34d';
  return 'background:#22c55e22;color:#86efac';
}

let powerHist = [];
const MAX_HIST = 60;

function renderBMS(b) {
  // Status bar
  document.getElementById('statusBar').innerHTML =
    '<div class="status-chip"><div class="label">SOC</div><div class="value" style="color:var(--ess)">'+b.soc.toFixed(1)+'%</div></div>' +
    '<div class="status-chip"><div class="label">Temperature</div><div class="value" style="color:'+(b.temp_c>50?'var(--danger)':'var(--text)')+'">'+b.temp_c.toFixed(1)+'°C</div></div>' +
    '<div class="status-chip"><div class="label">Cell Min</div><div class="value">'+b.min_cell_mv+' mV</div></div>' +
    '<div class="status-chip"><div class="label">Cell Max</div><div class="value">'+b.max_cell_mv+' mV</div></div>' +
    '<div class="status-chip"><div class="label">Status</div><div class="value"><span class="online-dot '+(b.online?'on':'off')+'"></span>'+(b.online?'Online':'Offline')+'</div></div>';

  // Detail rows
  let html = '<div class="detail-section"><h3>Battery Information</h3>';
  html += row('Capacity', b.capacity_wh + ' Wh');
  html += row('Max Charge', b.max_charge_w + ' W');
  html += row('Max Discharge', b.max_discharge_w + ' W');
  html += row('Alarm Flags', b.alarm_flags === 0 ? '✅ None' : '⚠️ 0x' + b.alarm_flags.toString(16).toUpperCase());
  html += row('Last Updated', new Date(b.updated_at).toLocaleTimeString());
  html += '</div>';

  // Simulated cell grid (fake cells based on min/max voltage range)
  html += '<div class="detail-section"><h3>Cell Voltage Heatmap</h3>';
  html += '<div style="font-size:11px;color:var(--muted);margin-bottom:8px;">Simulated 64 cells (range: '+b.min_cell_mv+'–'+b.max_cell_mv+' mV)</div>';
  html += '<div class="cell-grid">';
  const range = b.max_cell_mv - b.min_cell_mv;
  for (let i = 0; i < 64; i++) {
    const mv = b.min_cell_mv + Math.round(Math.random() * range);
    html += '<div class="cell" style="'+cellColor(mv)+'">'+mv+'</div>';
  }
  html += '</div></div>';

  document.getElementById('detailContent').innerHTML = html;
}

function renderPCS(p) {
  const statusNames = ['Standby','Running','Fault','Offline'];
  const st = statusNames[p.status] || 'Unknown';
  const stColor = p.status===1?'var(--success)':p.status===2?'var(--danger)':'var(--warn)';

  document.getElementById('statusBar').innerHTML =
    '<div class="status-chip"><div class="label">Active Power</div><div class="value" style="color:var(--grid)">'+kw(p.active_power_w)+'</div></div>' +
    '<div class="status-chip"><div class="label">Setpoint</div><div class="value" style="color:var(--ess)">'+kw(p.setpoint_w)+'</div></div>' +
    '<div class="status-chip"><div class="label">Max Power</div><div class="value">'+kw(p.max_power_w)+'</div></div>' +
    '<div class="status-chip"><div class="label">Status</div><div class="value" style="color:'+stColor+'">'+st+'</div></div>' +
    '<div class="status-chip"><div class="label">Online</div><div class="value"><span class="online-dot '+(p.online?'on':'off')+'"></span>'+(p.online?'Yes':'No')+'</div></div>';

  let html = '<div class="detail-section"><h3>PCS Information</h3>';
  html += row('Fault Code', p.fault_code === 0 ? '✅ None' : '⚠️ 0x' + p.fault_code.toString(16).toUpperCase());
  html += row('Utilization', p.max_power_w > 0 ? (Math.abs(p.active_power_w)/p.max_power_w*100).toFixed(1) + '%' : '—');
  html += row('Last Updated', new Date(p.updated_at).toLocaleTimeString());
  html += '</div>';

  // Power history chart
  powerHist.push(p.active_power_w);
  if (powerHist.length > MAX_HIST) powerHist.shift();
  html += '<div class="chart-wrap"><div class="title">Power Output — last '+MAX_HIST+' readings</div><canvas id="pcsChart"></canvas></div>';

  document.getElementById('detailContent').innerHTML = html;
  drawPCSChart();
}

function renderMeter(m) {
  document.getElementById('statusBar').innerHTML =
    '<div class="status-chip"><div class="label">Active Power</div><div class="value" style="color:var(--grid)">'+kw(m.active_power_w)+'</div></div>' +
    '<div class="status-chip"><div class="label">Voltage</div><div class="value">'+m.voltage_v.toFixed(1)+' V</div></div>' +
    '<div class="status-chip"><div class="label">Frequency</div><div class="value">'+m.frequency_hz.toFixed(2)+' Hz</div></div>' +
    '<div class="status-chip"><div class="label">Status</div><div class="value"><span class="online-dot '+(m.online?'on':'off')+'"></span>'+(m.online?'Online':'Offline')+'</div></div>';

  let html = '<div class="detail-section"><h3>Meter Information</h3>';
  html += row('Current', m.current_a.toFixed(2) + ' A');
  html += row('Energy Import', (m.energy_wh_in/1000).toFixed(1) + ' kWh');
  html += row('Energy Export', (m.energy_wh_out/1000).toFixed(1) + ' kWh');
  html += row('Last Updated', new Date(m.updated_at).toLocaleTimeString());
  html += '</div>';

  document.getElementById('detailContent').innerHTML = html;
}

function renderEVCharger(ev) {
  const evStatuses = ['Available','Charging','Suspended (EV)','Suspended (EVSE)','Finishing','Fault'];
  const st = evStatuses[ev.status] || 'Unknown';
  const stColor = ev.status===1?'#06b6d4':ev.status===5?'var(--danger)':ev.status===4?'var(--warn)':'var(--muted)';

  document.getElementById('statusBar').innerHTML =
    '<div class="status-chip"><div class="label">Charging Power</div><div class="value" style="color:#06b6d4">'+kw(ev.active_power_w)+'</div></div>' +
    '<div class="status-chip"><div class="label">Status</div><div class="value" style="color:'+stColor+'">'+st+'</div></div>' +
    '<div class="status-chip"><div class="label">Vehicle</div><div class="value">'+(ev.vehicle_connected?'🔌 Connected':'⚪ No vehicle')+'</div></div>' +
    '<div class="status-chip"><div class="label">Max Limit</div><div class="value">'+kw(ev.max_power_limit)+'</div></div>' +
    '<div class="status-chip"><div class="label">Online</div><div class="value"><span class="online-dot '+(ev.online?'on':'off')+'"></span>'+(ev.online?'Yes':'No')+'</div></div>';

  let html = '<div class="detail-section"><h3>EV Charger Information</h3>';
  html += row('Energy Delivered', (ev.energy_delivered/1000).toFixed(2) + ' kWh');
  html += row('Phase L1 Current', ev.current_l1.toFixed(1) + ' A');
  html += row('Phase L2 Current', ev.current_l2.toFixed(1) + ' A');
  html += row('Phase L3 Current', ev.current_l3.toFixed(1) + ' A');
  html += row('Max Power Limit', kw(ev.max_power_limit));
  html += row('Last Updated', new Date(ev.updated_at).toLocaleTimeString());
  html += '</div>';

  // Power history chart
  powerHist.push(ev.active_power_w);
  if (powerHist.length > MAX_HIST) powerHist.shift();
  html += '<div class="chart-wrap"><div class="title">Charging Power — last '+MAX_HIST+' readings</div><canvas id="pcsChart"></canvas></div>';

  document.getElementById('detailContent').innerHTML = html;
  drawPCSChart();
}

function row(k, v) {
  return '<div class="detail-row"><span class="key">'+k+'</span><span class="val">'+v+'</span></div>';
}

function drawPCSChart() {
  const c = document.getElementById('pcsChart');
  if (!c) return;
  const ctx = c.getContext('2d');
  const dpr = devicePixelRatio||1, r = c.getBoundingClientRect();
  c.width = r.width*dpr; c.height = 120*dpr;
  c.style.height = '120px';
  ctx.scale(dpr, dpr);
  const W = r.width, H = 120;

  ctx.clearRect(0,0,W,H);

  if (powerHist.length < 2) return;
  const maxV = Math.max(100, ...powerHist.map(Math.abs));
  const mid = H/2;

  ctx.beginPath();
  ctx.strokeStyle = '#3b82f6';
  ctx.lineWidth = 2;
  for (let i = 0; i < powerHist.length; i++) {
    const x = (i/(powerHist.length-1)) * W;
    const y = mid - (powerHist[i]/maxV) * (H/2 - 10);
    i === 0 ? ctx.moveTo(x,y) : ctx.lineTo(x,y);
  }
  ctx.stroke();

  // Zero line
  ctx.beginPath();
  ctx.strokeStyle = '#334155';
  ctx.lineWidth = 1;
  ctx.setLineDash([4,4]);
  ctx.moveTo(0, mid); ctx.lineTo(W, mid);
  ctx.stroke();
  ctx.setLineDash([]);
}

async function pollDevice() {
  try {
    const res = await fetch('/api/devices');
    if (!res.ok) return;
    const d = await res.json();

    // Find our component
    const allDevices = [...(d.bms||[]), ...(d.pcs||[]), ...(d.meter||[]), ...(d.ev_charger||[])];
    const dev = allDevices.find(x => x.id === COMP_ID);

    if (!dev) {
      document.getElementById('detailContent').innerHTML =
        '<div class="detail-section"><h3>Device Not Found</h3><p style="color:var(--danger)">Component "'+COMP_ID+'" not found in system.</p></div>';
      return;
    }

    if (compType === 'bms') renderBMS(dev);
    else if (compType === 'pcs') renderPCS(dev);
    else if (compType === 'ev') renderEVCharger(dev);
    else renderMeter(dev);
  } catch(e) {}
}

// Poll every 2s
setInterval(pollDevice, 2000);
pollDevice();
</script>
</body></html>
`
