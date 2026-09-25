// ipvn7 Universal Core Dashboard — Controlador del Botón Soberano VPN (app_vpn_button.js)
// Axioma III: Límite estricto de 400 líneas

let currentVPNMode = 'mesh_only';
let isToggling = false;

function initVPNButton() {
  const circleBtn = document.getElementById('hero-vpn-circle');
  const togglePanelBtn = document.getElementById('btn-toggle-ecosystem');

  if (circleBtn) {
    circleBtn.addEventListener('click', handleCircleClick);
  }

  if (togglePanelBtn) {
    togglePanelBtn.addEventListener('click', handleToggleEcosystem);
    // Iniciar siempre en modo colapsado Zen como pantalla principal pura
    setEcosystemExpanded(false);
  }

  // Polling del estado de la VPN
  fetchVPNStatus();
  setInterval(fetchVPNStatus, 3000);
}

async function fetchVPNStatus() {
  if (isToggling) return; // Evitar sobreescritura durante transición activa
  try {
    const res = await fetch('/api/v1/vpn/status');
    if (!res.ok) return;
    const data = await res.json();
    renderVPNState(data);
  } catch (err) {
    console.debug('Error consultando estado VPN:', err);
  }
}

function renderVPNState(data) {
  currentVPNMode = data.mode || 'mesh_only';
  const wrapper = document.getElementById('hero-circle-wrapper');
  const label = document.getElementById('hero-state-label');
  const subtext = document.getElementById('hero-mode-subtext');
  const badge = document.getElementById('hero-status-badge');
  const badgeText = document.getElementById('hero-status-badge-text');
  const desc = document.getElementById('hero-status-desc');
  const proxyDetails = document.getElementById('hero-proxy-details');
  const proxyStats = document.getElementById('hero-proxy-stats');

  if (!wrapper) return;

  wrapper.classList.remove('state-red', 'state-yellow', 'state-green');
  badge.classList.remove('badge-red', 'badge-yellow', 'badge-green');

  if (data.color_state === 'green') {
    wrapper.classList.add('state-green');
    badge.classList.add('badge-green');
    label.textContent = 'PROTEGIDO';
    subtext.textContent = 'INTERNET CANALIZADO';
    badgeText.textContent = data.title;
    desc.textContent = data.description;
    if (proxyDetails) proxyDetails.style.display = 'flex';
    if (proxyStats) {
      const txKb = ((data.bytes_tx || 0) / 1024).toFixed(1);
      const rxKb = ((data.bytes_rx || 0) / 1024).toFixed(1);
      proxyStats.textContent = `Tx: ${txKb} KB | Rx: ${rxKb} KB | Proxy: ${data.socks5_addr}`;
    }
  } else if (data.color_state === 'yellow') {
    wrapper.classList.add('state-yellow');
    badge.classList.add('badge-yellow');
    label.textContent = 'CONECTANDO';
    subtext.textContent = 'INICIANDO TÚNEL';
    badgeText.textContent = data.title;
    desc.textContent = data.description;
    if (proxyDetails) proxyDetails.style.display = 'none';
  } else {
    wrapper.classList.add('state-red');
    badge.classList.add('badge-red');
    label.textContent = 'MALLA P2P';
    subtext.textContent = 'SIN TÚNEL DE INTERNET';
    badgeText.textContent = data.title;
    desc.textContent = data.description;
    if (proxyDetails) proxyDetails.style.display = 'none';
  }
}

async function handleCircleClick() {
  if (isToggling) return;
  isToggling = true;

  const wrapper = document.getElementById('hero-circle-wrapper');
  const label = document.getElementById('hero-state-label');
  const subtext = document.getElementById('hero-mode-subtext');
  const badge = document.getElementById('hero-status-badge');
  const badgeText = document.getElementById('hero-status-badge-text');
  const desc = document.getElementById('hero-status-desc');

  // Transición visual inmediata al estado amarillo (conectando)
  wrapper.classList.remove('state-red', 'state-green');
  wrapper.classList.add('state-yellow');
  badge.classList.remove('badge-red', 'badge-green');
  badge.classList.add('badge-yellow');
  label.textContent = 'ENLACE...';
  subtext.textContent = 'CONMUTANDO MODO';
  badgeText.textContent = 'Conectando túnel cuántico...';
  desc.textContent = 'Reconfigurando canales de enrutamiento y gateway de protección...';

  try {
    const res = await fetch('/api/v1/vpn/toggle', { method: 'POST' });
    if (res.ok) {
      // Pequeña pausa háptica para asentar el socket
      await new Promise(r => setTimeout(r, 600));
      await fetchVPNStatus();
    }
  } catch (err) {
    console.error('Error conmutando estado VPN:', err);
  } finally {
    isToggling = false;
    fetchVPNStatus();
  }
}

function handleToggleEcosystem() {
  const panel = document.getElementById('advanced-ecosystem-panel');
  const isExpanded = panel.classList.contains('expanded');
  setEcosystemExpanded(!isExpanded);
}

function setEcosystemExpanded(expanded) {
  const panel = document.getElementById('advanced-ecosystem-panel');
  const toggleBtn = document.getElementById('btn-toggle-ecosystem');
  const btnText = document.getElementById('toggle-ecosystem-text');

  if (!panel || !toggleBtn) return;

  if (expanded) {
    panel.classList.remove('collapsed');
    panel.classList.add('expanded');
    toggleBtn.classList.add('expanded');
    if (btnText) btnText.textContent = 'Ocultar panel avanzado';
    localStorage.setItem('ipvn7_ecosystem_expanded', 'true');
  } else {
    panel.classList.remove('expanded');
    panel.classList.add('collapsed');
    toggleBtn.classList.remove('expanded');
    if (btnText) btnText.textContent = 'Expandir todas las bondades de ipvn7';
    localStorage.setItem('ipvn7_ecosystem_expanded', 'false');
  }
}
