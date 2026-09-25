// ipvn7 Universal Core Dashboard — Ecosistema Doméstico & Spooler (app_homehub.js)
// Google Cast Sender SDK + Plyr + Smart Home Hub (Axioma III: <= 400 líneas)

// 1. GOOGLE CAST SENDER SDK
window['__onGCastApiAvailable'] = function(isAvailable) {
  if (isAvailable && window.cast && window.cast.framework) {
    try {
      cast.framework.CastContext.getInstance().setOptions({
        receiverApplicationId: chrome.cast.media.DEFAULT_MEDIA_RECEIVER_APP_ID,
        autoJoinPolicy: chrome.cast.AutoJoinPolicy.ORIGIN_SCOPED
      });
      cast.framework.CastContext.getInstance().addEventListener(
        cast.framework.CastContextEventType.SESSION_STATE_CHANGED, function(e) {
          if (e.sessionState === cast.framework.SessionState.SESSION_STARTED) {
            showToast('📺 Conectado a Smart TV mediante Google Cast');
            castMediaDirectly();
          }
        });
    } catch (e) { console.debug('Cast init:', e); }
  }
};

function castMediaDirectly(mediaUrl) {
  if (!window.cast || !window.cast.framework) return false;
  const session = cast.framework.CastContext.getInstance()?.getCurrentSession();
  if (!session) return false;
  const url = mediaUrl || document.getElementById('cast-url')?.value.trim();
  if (!url) return false;
  const mi = new chrome.cast.media.MediaInfo(url, 'video/mp4');
  mi.metadata = new chrome.cast.media.GenericMediaMetadata();
  mi.metadata.title = 'Transmisión ipvn7 Soberana';
  const req = new chrome.cast.media.LoadRequest(mi);
  req.autoplay = true;
  session.loadMedia(req).then(
    () => showToast('▶️ Transmitiendo vía Google Cast'),
    (err) => console.debug('Cast loadMedia:', err)
  );
  return true;
}

function initHomeHubActions() {
  document.getElementById('cast-target')?.addEventListener('change', (e) => {
    const c = document.getElementById('cast-custom-ip');
    if (c) c.style.display = e.target.value === 'custom:ip' ? 'block' : 'none';
  });

  const btnCastProj = document.getElementById('btn-cast-project');
  if (btnCastProj) {
    btnCastProj.addEventListener('click', async () => {
      btnCastProj.disabled = true;
      try {
        const res = await fetch(`${API_BASE}/api/v1/home/cast/project`, { method: 'POST' });
        const d = await res.json();
        showToast(`🖥️ ${d.message}`);
      } catch (err) { showToast('Error abriendo menú de proyección', true); }
      finally { btnCastProj.disabled = false; }
    });
  }

  async function performCast(target, mediaUrl, title) {
    if (!mediaUrl) return;
    if (castMediaDirectly(mediaUrl)) return;
    try {
      const res = await fetch(`${API_BASE}/api/v1/home/cast`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target_tv: target || '192.168.1.65', media_url: mediaUrl, title: title || 'Video Soberano' })
      });
      const d = await res.json();
      if (!res.ok) { showToast(`❌ ${d.message || d.error || 'Error TV'}`, true); return; }
      showToast(`📺 ${d.message}`);
    } catch (err) { showToast('Error de red al comunicar con la TV', true); }
  }

  document.getElementById('btn-cast-action')?.addEventListener('click', async () => {
    let t = document.getElementById('cast-target')?.value.trim() || '';
    if (t === 'custom:ip') t = document.getElementById('cast-custom-ip')?.value.trim() || '';
    const url = document.getElementById('cast-url')?.value.trim();
    if (!url) { showToast('Ingresa un enlace de video o favorito', true); return; }
    await performCast(t, url, 'Video Soberano');
  });

  document.getElementById('btn-cast-stop')?.addEventListener('click', async () => {
    try { window.cast?.framework?.CastContext?.getInstance()?.getCurrentSession()?.endSession(true); } catch (e) {}
    try {
      await fetch(`${API_BASE}/api/v1/home/cast/stop`, { method: 'POST' });
      showToast('⏹️ Transmisión detenida');
    } catch (err) { showToast('Error deteniendo transmisión', true); }
  });

  // Control Remoto Dual (Google Cast + Plyr /tv vía SSE)
  async function sendTvControl(action, value = 0) {
    try {
      const s = window.cast?.framework?.CastContext?.getInstance()?.getCurrentSession();
      if (s) {
        if (action === 'volume') s.setVolume(Math.max(0, Math.min(1, value / 100)));
        if (action === 'mute') s.setMute(!s.isMute());
        const m = s.getMediaSession();
        if (m && action === 'pause') m.pause();
        if (m && action === 'play') m.play();
      }
    } catch (e) {}
    try {
      await fetch(`${API_BASE}/api/v1/home/cast/control`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action, value })
      });
    } catch (e) {}
  }

  const volSlider = document.getElementById('tv-vol-slider');
  const volLabel = document.getElementById('tv-vol-label');
  if (volSlider) {
    volSlider.addEventListener('input', () => {
      if (volLabel) volLabel.textContent = `Volumen: ${volSlider.value}%`;
      sendTvControl('volume', parseInt(volSlider.value));
    });
  }
  document.getElementById('btn-tv-mute')?.addEventListener('click', () => { sendTvControl('mute'); showToast('🔇 Sonido conmutado'); });
  document.getElementById('btn-tv-vol-up')?.addEventListener('click', () => {
    if (!volSlider) return;
    volSlider.value = Math.min(100, parseInt(volSlider.value) + 10);
    volSlider.dispatchEvent(new Event('input'));
  });
  document.getElementById('btn-tv-pause')?.addEventListener('click', () => { sendTvControl('pause'); showToast('⏸️ Video pausado'); });
  document.getElementById('btn-tv-play')?.addEventListener('click', () => { sendTvControl('play'); showToast('▶️ Video reanudado'); });

  // Canales y Favoritos
  const defaultFavs = [
    { name: '🌲 4K Naturaleza', url: 'https://www.youtube.com/watch?v=LXb3EKWsInQ' },
    { name: '🎧 Lofi Chill', url: 'https://www.youtube.com/watch?v=jfKfPfyJRdk' },
    { name: '🚀 NASA Live', url: 'https://www.youtube.com/watch?v=21X5lGlDOfg' },
    { name: '🎵 Synthwave', url: 'https://www.youtube.com/watch?v=4xDzrJKXOOY' }
  ];
  function getFavorites() {
    try {
      const s = localStorage.getItem('ipvn7_tv_favs');
      return s ? JSON.parse(s) : defaultFavs;
    } catch (e) { return defaultFavs; }
  }
  function renderFavorites() {
    const bar = document.getElementById('favorites-bar');
    const cnt = document.getElementById('fav-count');
    if (!bar) return;
    const favs = getFavorites();
    if (cnt) cnt.textContent = `${favs.length} disponibles`;
    bar.innerHTML = favs.map((f, i) => `
      <button class="btn-action secondary btn-compact fav-chip" data-idx="${i}" style="font-size:0.75rem; padding:3px 8px; border-radius:6px;" title="${escapeHtml(f.url)}">
        <span>${escapeHtml(f.name)}</span>
      </button>`).join('');
    bar.querySelectorAll('.fav-chip').forEach(btn => {
      btn.addEventListener('click', () => {
        const fav = favs[parseInt(btn.dataset.idx)];
        const inp = document.getElementById('cast-url');
        if (inp) inp.value = fav.url;
        let t = (castSelect ? castSelect.value : '').trim();
        if (t === 'custom:ip' && castCustomIp) t = castCustomIp.value.trim();
        performCast(t, fav.url, fav.name);
      });
    });
  }
  renderFavorites();

  const btnSaveFav = document.getElementById('btn-save-favorite');
  if (btnSaveFav) {
    btnSaveFav.addEventListener('click', () => {
      const url = document.getElementById('cast-url')?.value.trim();
      if (!url) return;
      const name = prompt('Nombre para este favorito:', '⭐ Mi Canal');
      if (!name) return;
      const f = getFavorites();
      f.push({ name, url });
      localStorage.setItem('ipvn7_tv_favs', JSON.stringify(f));
      renderFavorites();
      showToast(`⭐ '${name}' guardado`);
    });
  }

  // Impresora y Spooler
  document.getElementById('print-target')?.addEventListener('change', (e) => {
    const c = document.getElementById('print-custom-ip');
    if (c) c.style.display = e.target.value === 'custom:ip' ? 'block' : 'none';
  });

  document.getElementById('btn-print-action')?.addEventListener('click', async () => {
    let t = document.getElementById('print-target')?.value.trim() || 'local:spool';
    if (t === 'custom:ip') t = document.getElementById('print-custom-ip')?.value.trim() || 'local:spool';
    const doc = document.getElementById('print-doc-name')?.value.trim() || 'comprobante_soberano.pdf';
    const content = document.getElementById('print-doc-content')?.value || '';
    try {
      const res = await fetch(`${API_BASE}/api/v1/home/print`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target_printer: t, doc_name: doc, content, copies: 1 })
      });
      const d = await res.json();
      showToast(res.ok ? `🖨️ ${d.message}` : `❌ ${d.message || d.error}`, !res.ok);
      if (res.ok) fetchPrintJobs();
    } catch (err) { showToast('Error de red al imprimir', true); }
  });

  document.getElementById('btn-print-native')?.addEventListener('click', () => {
    const doc = document.getElementById('print-doc-name')?.value.trim() || 'Comprobante';
    const w = window.open('', '_blank');
    if (w) {
      w.document.write(`<title>${escapeHtml(doc)}</title><pre style="padding:20px;font-family:sans-serif;">${escapeHtml(document.getElementById('print-doc-content')?.value || '')}</pre>`);
      w.document.close();
      w.print();
      showToast('🖨️ Diálogo de impresión local abierto');
    }
  });

  document.getElementById('btn-toggle-print-jobs')?.addEventListener('click', () => {
    const d = document.getElementById('print-jobs-drawer');
    if (d) { d.style.display = d.style.display === 'block' ? 'none' : 'block'; if (d.style.display === 'block') fetchPrintJobs(); }
  });

  // Wake-on-LAN con autodescubrimiento ARP
  async function populateWoLDevices() {
    try {
      const res = await fetch(`${API_BASE}/api/v1/home/wol`);
      if (!res.ok) return;
      const devs = await res.json() || [];
      const sel = document.getElementById('wol-mac');
      if (sel && devs.length > 0) {
        sel.innerHTML = devs.map(d => `<option value="${escapeHtml(d.mac)}">⚡ ${escapeHtml(d.ip)} [${escapeHtml(d.mac)}]</option>`).join('');
      }
    } catch (e) {}
  }
  populateWoLDevices();

  const btnWoL = document.getElementById('btn-wol-action');
  if (btnWoL) {
    btnWoL.addEventListener('click', async () => {
      const mac = document.getElementById('wol-mac')?.value.trim();
      if (!mac) { showToast('Selecciona un dispositivo con dirección MAC', true); return; }
      btnWoL.disabled = true;
      try {
        const res = await fetch(`${API_BASE}/api/v1/home/wol`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ mac_address: mac })
        });
        const d = await res.json();
        showToast(res.ok ? `⚡ ${d.message}` : `❌ ${d.message || d.error}`, !res.ok);
      } catch (err) { showToast('Error enviando Wake-on-LAN', true); }
      finally { btnWoL.disabled = false; }
    });
  }

  // Blindaje IoT ZTNA
  const btnIoT = document.getElementById('btn-iot-shield-toggle');
  const txtIoT = document.getElementById('txt-iot-shield');
  if (btnIoT && txtIoT) {
    btnIoT.addEventListener('click', async () => {
      iotShieldActive = !iotShieldActive;
      btnIoT.disabled = true;
      try {
        await fetch(`${API_BASE}/api/v1/home/iot-shield`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ enabled: iotShieldActive })
        });
        txtIoT.textContent = iotShieldActive ? '🛡️ Blindaje IoT ACTIVO' : '🛡️ Activar Blindaje IoT';
        btnIoT.className = `btn-action ${iotShieldActive ? 'primary' : 'secondary'} btn-hub full-width`;
        showToast(iotShieldActive ? '🛡️ Blindaje ZTNA Activado' : 'Blindaje ZTNA en modo estándar');
      } catch (err) { showToast('Error en blindaje IoT', true); }
      finally { btnIoT.disabled = false; }
    });
  }

  // Gestión de Archivos y Spooler desacoplados en app_files_spooler.js
}
