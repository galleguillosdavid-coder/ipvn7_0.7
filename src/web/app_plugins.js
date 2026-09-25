// ipvn7 Sovereign Mesh OS — Gestor Universal de Plugins (app_plugins.js)
// Cumple Axioma III: <= 400 líneas. Control de 2 botones: [Instalar/Desinstalar] y [Encender/Apagar]

const IPVN7_PLUGINS = {
  'tv': { id: 'tv', name: 'Pantalla & Smart TV', desc: 'Transmite videos y proyecta tu pantalla a TVs o proyectores de tu red local.', installed: true, enabled: true, icon: '📺' },
  'print': { id: 'print', name: 'Impresoras', desc: 'Envía documentos a tus impresoras físicas o guárdalos en tu cola de impresión.', installed: true, enabled: true, icon: '🖨️' },
  'files': { id: 'files', name: 'Archivos & Buzón Local', desc: 'Comparte y recibe archivos cifrados con tus otros dispositivos sin nube intermedia.', installed: true, enabled: true, icon: '📦' },
  'wol': { id: 'wol', name: 'Encendido Remoto', desc: 'Despierta computadores o consolas apagadas en tu red con un solo clic.', installed: true, enabled: true, icon: '⚡' },
  'shield': { id: 'shield', name: 'Escudo IoT', desc: 'Aísla cámaras, enchufes y electrodomésticos para que nadie fuera de tu hogar pueda verlos.', installed: true, enabled: true, icon: '🛡️' },
  'chat': { id: 'chat', name: 'Chat Privado', desc: 'Mensajería directa dispositivo a dispositivo sin pasar por servidores de terceros.', installed: true, enabled: true, icon: '💬' },
  'tools': { id: 'tools', name: 'Herramientas y Satélites', desc: 'Acceso a túneles SSH protegidos, Docker, Robótica y control del sistema.', installed: true, enabled: true, icon: '⚙️' }
};

function loadPluginsState() {
  try {
    const saved = localStorage.getItem('ipvn7_plugins_state');
    if (saved) {
      const parsed = JSON.parse(saved);
      Object.keys(parsed).forEach(k => {
        if (IPVN7_PLUGINS[k]) {
          IPVN7_PLUGINS[k].installed = parsed[k].installed ?? true;
          IPVN7_PLUGINS[k].enabled = parsed[k].enabled ?? true;
        }
      });
    }
  } catch (e) {
    console.debug('Error loading plugins state:', e);
  }
}

function savePluginsState() {
  try {
    localStorage.setItem('ipvn7_plugins_state', JSON.stringify(IPVN7_PLUGINS));
  } catch (e) {
    console.debug('Error saving plugins state:', e);
  }
}

async function syncPluginWithBackend(pluginId, action) {
  try {
    const pl = IPVN7_PLUGINS[pluginId];
    await fetch(`${API_BASE}/api/v1/components/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        id: `plugin_${pluginId}`,
        action: action,
        installed: pl.installed,
        enabled: pl.enabled
      })
    });
  } catch (e) {
    console.debug('Backend sync failed (continuing in local state):', e);
  }
}

function renderPluginHeader(pluginId, containerId) {
  const pl = IPVN7_PLUGINS[pluginId];
  if (!pl) return;
  const container = document.getElementById(containerId);
  if (!container) return;

  let existing = document.getElementById(`plugin-banner-${pluginId}`);
  if (!existing) {
    existing = document.createElement('div');
    existing.id = `plugin-banner-${pluginId}`;
    existing.className = 'plugin-control-card glass-card';
    container.prepend(existing);
  }

  const isInstalled = pl.installed;
  const isEnabled = pl.enabled && isInstalled;

  existing.innerHTML = `
    <div class="plugin-banner-top">
      <div class="plugin-title-block">
        <span class="plugin-icon">${pl.icon}</span>
        <div>
          <h3 class="plugin-heading">${pl.name} <span class="plugin-badge">Complemento Oficial</span></h3>
          <p class="plugin-subdesc">${pl.desc}</p>
        </div>
      </div>
      <div class="plugin-dual-controls">
        <!-- Botón 1: Instalar / Desinstalar -->
        <button type="button" class="btn-action ${isInstalled ? 'danger' : 'primary'} btn-plugin-install" onclick="togglePluginInstall('${pluginId}')">
          <span>${isInstalled ? '🗑️ Desinstalar' : '📦 Instalar'}</span>
        </button>
        <!-- Botón 2: Encender / Apagar -->
        <button type="button" class="btn-action ${isEnabled ? 'secondary' : 'accent'} btn-plugin-power" ${!isInstalled ? 'disabled' : ''} onclick="togglePluginPower('${pluginId}')">
          <span>${isEnabled ? '⏻ Apagar' : '⏻ Encender'}</span>
        </button>
      </div>
    </div>
    <div class="plugin-status-row">
      <div class="plugin-status-indicator">
        <span class="status-dot ${!isInstalled ? 'uninstalled' : (isEnabled ? 'running' : 'stopped')}"></span>
        <span class="status-label">${!isInstalled ? 'Desinstalado (Espacio liberado)' : (isEnabled ? 'Encendido y listo' : 'Apagado (En reposo sin consumir recursos)')}</span>
      </div>
      <div class="plugin-meta-info">Versión 0.7.0 • Privado y Local</div>
    </div>
  `;

  // Gestionar visibilidad del cuerpo del plugin
  const bodySelector = `.plugin-body-${pluginId}`;
  const bodies = container.querySelectorAll(bodySelector);
  let overlay = document.getElementById(`plugin-overlay-${pluginId}`);

  if (!isInstalled || !isEnabled) {
    bodies.forEach(b => b.style.display = 'none');
    if (!overlay) {
      overlay = document.createElement('div');
      overlay.id = `plugin-overlay-${pluginId}`;
      overlay.className = 'plugin-disabled-overlay glass-card';
      container.appendChild(overlay);
    }
    overlay.style.display = 'block';
    overlay.innerHTML = !isInstalled ? `
      <div class="plugin-disabled-content">
        <span class="overlay-icon">📦</span>
        <h4>${pl.name} no está instalado</h4>
        <p>Puedes instalar este complemento en cualquier momento con un solo clic.</p>
        <button class="btn-action primary" onclick="togglePluginInstall('${pluginId}')">📦 Instalar Ahora</button>
      </div>
    ` : `
      <div class="plugin-disabled-content">
        <span class="overlay-icon">💤</span>
        <h4>${pl.name} está apagado</h4>
        <p>El complemento está instalado pero en reposo. No consume memoria ni red.</p>
        <button class="btn-action primary" onclick="togglePluginPower('${pluginId}')">⏻ Encender Complemento</button>
      </div>
    `;
  } else {
    bodies.forEach(b => b.style.display = '');
    if (overlay) overlay.style.display = 'none';
  }
}

async function togglePluginInstall(pluginId) {
  const pl = IPVN7_PLUGINS[pluginId];
  if (!pl) return;
  pl.installed = !pl.installed;
  if (!pl.installed) {
    pl.enabled = false;
    showToast(`🗑️ ${pl.name} desinstalado`);
  } else {
    pl.enabled = true;
    showToast(`📦 ${pl.name} instalado con éxito`);
  }
  savePluginsState();
  refreshAllPluginHeaders();
  await syncPluginWithBackend(pluginId, pl.installed ? 'install' : 'uninstall');
}

async function togglePluginPower(pluginId) {
  const pl = IPVN7_PLUGINS[pluginId];
  if (!pl || !pl.installed) return;
  pl.enabled = !pl.enabled;
  if (pl.enabled) {
    showToast(`⚡ ${pl.name} encendido`);
  } else {
    showToast(`💤 ${pl.name} apagado`);
  }
  savePluginsState();
  refreshAllPluginHeaders();
  await syncPluginWithBackend(pluginId, pl.enabled ? 'power_on' : 'power_off');
}

function refreshAllPluginHeaders() {
  const mappings = [
    { id: 'tv', container: 'view-tv' },
    { id: 'print', container: 'view-print' },
    { id: 'files', container: 'view-files' },
    { id: 'wol', container: 'view-wol' },
    { id: 'shield', container: 'view-shield' },
    { id: 'chat', container: 'view-chat' },
    { id: 'tools', container: 'view-tools' }
  ];
  mappings.forEach(m => renderPluginHeader(m.id, m.container));
}

document.addEventListener('DOMContentLoaded', () => {
  loadPluginsState();
  refreshAllPluginHeaders();
});
