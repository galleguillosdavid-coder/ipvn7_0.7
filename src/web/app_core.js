// ipvn7 Universal Core Dashboard — Módulo Central (app_core.js)
// Polling de Estado, Métricas y Utilidades de UI (Axioma III: <= 400 líneas)

const API_BASE = '';
let sseConnection = null;
let activePeers = [];

// 1. Polling de Estado del Núcleo y Componentes
async function fetchCoreStatus() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/status`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    updateIdentityCard(data);
    updateMetrics(data);
    renderComponentsList(data.attached_components || []);
  } catch (err) {
    document.getElementById('txt-core-status').textContent = 'Conectando al Núcleo...';
    document.querySelector('.status-pill').classList.remove('active');
  }
}

async function fetchPeers() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/peers`);
    if (res.ok) {
      activePeers = await res.json() || [];
      document.getElementById('txt-peers-count').textContent = activePeers.length;
      document.getElementById('radar-peers-num').textContent = activePeers.length;
    }
  } catch (e) {
    console.debug('Error consultando pares:', e);
  }
}

function startCorePolling() {
  fetchCoreStatus();
  fetchPeers();
  fetchChatContacts();
  setInterval(fetchCoreStatus, 3000);
  setInterval(fetchPeers, 3000);
  setInterval(fetchChatContacts, 3000);
  setInterval(() => {
    if (activeChatPeerDID) {
      fetchChatMessages(activeChatPeerDID);
    }
  }, 2500);
}


// 2. Actualización de Interfaz
function updateIdentityCard(data) {
  document.getElementById('val-did').textContent = data.did || 'did:ipvn7:...';
  document.getElementById('val-ipv6').textContent = data.ipv6 ? `${data.ipv6}/64` : 'fd07::...';
  document.getElementById('val-ipv4').textContent = data.ipv4 ? `${data.ipv4}/16` : '10.7.0.0/16';
  
  document.getElementById('txt-core-status').textContent = 'Núcleo Operativo (v0.7)';
  document.querySelector('.status-pill').classList.add('active');

  const uptimeSec = data.uptime_seconds || 0;
  const mins = Math.floor(uptimeSec / 60);
  const secs = uptimeSec % 60;
  document.getElementById('txt-uptime').textContent = `${mins}m ${secs}s`;
}

function updateMetrics(data) {
  document.getElementById('val-tx-pkts').textContent = (data.total_tx_bytes || 0).toLocaleString();
  document.getElementById('val-rx-pkts').textContent = (data.total_rx_bytes || 0).toLocaleString();
  document.getElementById('val-drop-pkts').textContent = (data.total_drops || 0).toLocaleString();
  document.getElementById('val-attached-count').textContent = (data.attached_components || []).length;
}

function renderComponentsList(components) {
  const container = document.getElementById('components-list');
  if (!components || components.length === 0) {
    container.innerHTML = `
      <div class="empty-placeholder">
        <span>No hay componentes externos acoplados aún.</span>
        <button class="btn-action secondary" onclick="document.getElementById('btn-open-attach-modal').click()">
          ⚡ Acoplar el primero
        </button>
      </div>`;
    return;
  }

  container.innerHTML = components.map(c => {
    const stateClass = c.state === 'ACTIVE' ? 'active' : 'idle';
    const capsBadges = (c.capabilities || []).map(cap => `<span class="cap-tag">${escapeHtml(cap)}</span>`).join('');

    return `
      <div class="component-card">
        <div class="comp-main">
          <div class="comp-title">
            <span class="comp-name">${escapeHtml(c.name)}</span>
            <span class="comp-id">${escapeHtml(c.id)}</span>
          </div>
          <div class="comp-caps">${capsBadges}</div>
        </div>
        <div class="comp-meta">
          <span class="comp-state ${stateClass}">${c.state}</span>
          <span class="badge-lockfree" style="font-size:0.68rem;">v${escapeHtml(c.version || '1.0')}</span>
        </div>
      </div>
    `;
  }).join('');

  syncHomeHubTargets(components);
}

function syncHomeHubTargets(components) {
  const castSelect = document.getElementById('cast-target');
  const printSelect = document.getElementById('print-target');
  const wolSelect = document.getElementById('wol-mac');

  if (castSelect) {
    const currentVal = castSelect.value;
    const tvs = components.filter(c => (c.capabilities || []).some(k => k.startsWith('tv:')));
    let optionsHtml = '';
    tvs.forEach(t => {
      optionsHtml += `<option value="${escapeHtml(t.endpoint)}">📺 ${escapeHtml(t.name)}</option>`;
    });
    optionsHtml += `
      <option value="screen:webrtc">🖥️ Capturar Pantalla en Navegador (/tv)</option>
      <option value="video:local">📺 Reproductor de Video Local</option>
      <option value="custom:ip">🌐 Ingresar IP Manual de Smart TV...</option>
    `;
    castSelect.innerHTML = optionsHtml;
    if (currentVal && (currentVal.startsWith('screen:') || currentVal.startsWith('video:') || currentVal === 'custom:ip' || tvs.some(t => t.endpoint === currentVal))) {
      castSelect.value = currentVal;
    } else if (tvs.length > 0) {
      castSelect.value = tvs[0].endpoint;
    }
  }

  if (printSelect) {
    const currentVal = printSelect.value;
    const printers = components.filter(c => (c.capabilities || []).some(k => k.startsWith('ipp:') || k.startsWith('printer:') || k.startsWith('raw:')));
    let optionsHtml = '';
    printers.forEach(p => {
      optionsHtml += `<option value="${escapeHtml(p.endpoint)}">🖨️ ${escapeHtml(p.name)} [Física / Red]</option>`;
    });
    optionsHtml += `<option value="local:spool">🖨️ Spooler Local (Comprobante Virtual)</option>`;
    optionsHtml += `<option value="custom:ip">🌐 Ingresar IP Manual (ej. 192.168.1.167:631)</option>`;
    printSelect.innerHTML = optionsHtml;
    if (currentVal && (currentVal === 'local:spool' || currentVal === 'custom:ip' || printers.some(p => p.endpoint === currentVal))) {
      printSelect.value = currentVal;
    } else if (printers.length > 0) {
      printSelect.value = printers[0].endpoint;
    }
  }

  if (wolSelect) {
    const wolDevices = components.filter(c => (c.capabilities || []).some(k => k.startsWith('mac:')));
    const currentVal = wolSelect.value;
    if (wolDevices.length > 0) {
      wolSelect.innerHTML = wolDevices.map(d => {
        const mac = (d.capabilities.find(k => k.startsWith('mac:')) || '').replace('mac:', '');
        return `<option value="${escapeHtml(mac)}">⚡ ${escapeHtml(d.name)} [${escapeHtml(mac)}]</option>`;
      }).join('');
      if (currentVal && wolDevices.some(d => d.capabilities.includes(`mac:${currentVal}`))) {
        wolSelect.value = currentVal;
      } else {
        const mac = (wolDevices[0].capabilities.find(k => k.startsWith('mac:')) || '').replace('mac:', '');
        wolSelect.value = mac;
      }
    } else {
      wolSelect.innerHTML = '<option value="0c:79:55:d5:75:20">📺 Smart TV TCL [0c:79:55:d5:75:20]</option>';
    }
  }

  const shieldStatus = document.getElementById('txt-iot-shield-status');
  if (shieldStatus) {
    const iotDevices = components.filter(c => (c.capabilities || []).some(k => k.startsWith('shadow_did:')));
    shieldStatus.textContent = `Monitoreando ${iotDevices.length} dispositivos en la LAN`;
  }
}


// 8. Utilidades
function initCopyButtons() {
  const setup = (btnId, valId, label) => {
    const btn = document.getElementById(btnId);
    if (!btn) return;
    btn.addEventListener('click', () => {
      const text = document.getElementById(valId).textContent;
      navigator.clipboard.writeText(text).then(() => {
        showToast(`${label} copiado al portapapeles`);
      });
    });
  };
  setup('btn-copy-did', 'val-did', 'DID Soberano');
  setup('btn-copy-ipv6', 'val-ipv6', 'IPv6');
  setup('btn-copy-ipv4', 'val-ipv4', 'IPv4');
}

function showToast(message, isError = false) {
  const box = document.getElementById('toast-box');
  const toast = document.createElement('div');
  toast.className = 'toast';
  if (isError) toast.style.borderColor = 'rgba(244, 63, 94, 0.4)';
  toast.textContent = message;
  box.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

// 9. Navegación de Vistas y Pestañas Dedicadas (Regla 18 / DEC-086)
let currentAppView = 'network';

function handleUrlHash() {
  const hash = window.location.hash.replace('#', '');
  const validViews = ['network', 'tv', 'print', 'files', 'wol', 'shield', 'chat', 'tools'];
  if (validViews.includes(hash)) {
    setAppView(hash);
  } else if (hash === 'homehub') {
    setAppView('tv');
  }
}

function setAppView(view) {
  if (view === 'dashboard' || view === 'kuzu') view = 'network';
  if (view === 'homehub') view = 'tv';
  currentAppView = view;
  const views = ['network', 'tv', 'print', 'files', 'wol', 'shield', 'chat', 'tools'];

  views.forEach(v => {
    const btn = document.getElementById(`tab-btn-${v}`);
    const el = document.getElementById(`view-${v}`);
    if (btn) btn.classList.toggle('active', v === view);
    if (el) el.style.display = (v === view) ? (v === 'chat' ? 'flex' : 'block') : 'none';
  });

  if (window.location.hash !== `#${view}`) {
    window.history.replaceState(null, '', `#${view}`);
  }

  if (view === 'chat') {
    fetchChatContacts();
  }
  if (typeof refreshAllPluginHeaders === 'function') {
    refreshAllPluginHeaders();
  }
}

function initTabs() {
  window.addEventListener('hashchange', handleUrlHash);
  handleUrlHash();
}
