// ipvn7 Universal Core Dashboard — Módulo de Red & Eventos (app_net.js)
// Radar Kleinberg, SSE y Envío de Datagramas (Axioma III: <= 400 líneas)

// 3. Radar de Kleinberg en Canvas
function initRadarAnimation() {
  const canvas = document.getElementById('radar-canvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  let angle = 0;

  function renderRadar() {
    const w = canvas.width;
    const h = canvas.height;
    const cx = w / 2;
    const cy = h / 2;
    const maxRadius = Math.min(cx, cy) - 15;

    ctx.clearRect(0, 0, w, h);

    // Fondo del radar
    ctx.fillStyle = 'rgba(7, 10, 19, 0.7)';
    ctx.fillRect(0, 0, w, h);

    // 12 Anillos concéntricos logarítmicos
    for (let i = 1; i <= 12; i++) {
      const r = (maxRadius / 12) * i;
      ctx.beginPath();
      ctx.arc(cx, cy, r, 0, Math.PI * 2);
      ctx.strokeStyle = i === 12 ? 'rgba(0, 240, 255, 0.35)' : 'rgba(168, 85, 247, 0.12)';
      ctx.lineWidth = i === 12 ? 1.5 : 1;
      ctx.stroke();
    }

    // Ejes cardinales
    ctx.beginPath();
    ctx.moveTo(cx, cy - maxRadius);
    ctx.lineTo(cx, cy + maxRadius);
    ctx.moveTo(cx - maxRadius, cy);
    ctx.lineTo(cx + maxRadius, cy);
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
    ctx.lineWidth = 1;
    ctx.stroke();

    // Haz giratorio del radar
    angle += 0.02;
    ctx.save();
    ctx.translate(cx, cy);
    ctx.rotate(angle);
    const grad = ctx.createLinearGradient(0, 0, maxRadius, 0);
    grad.addColorStop(0, 'rgba(0, 240, 255, 0.4)');
    grad.addColorStop(1, 'rgba(0, 240, 255, 0)');
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.arc(0, 0, maxRadius, 0, Math.PI / 4);
    ctx.closePath();
    ctx.fillStyle = grad;
    ctx.fill();
    ctx.restore();

    // Centro del nodo
    ctx.beginPath();
    ctx.arc(cx, cy, 5, 0, Math.PI * 2);
    ctx.fillStyle = '#00f0ff';
    ctx.shadowColor = '#00f0ff';
    ctx.shadowBlur = 10;
    ctx.fill();
    ctx.shadowBlur = 0;

    // Pares dibujados sobre los anillos
    activePeers.forEach((p, idx) => {
      const ring = (p.ring_index !== undefined ? p.ring_index : (idx % 12)) + 1;
      const r = (maxRadius / 12) * Math.min(ring, 12);
      const peerAngle = (idx * 137.5) * (Math.PI / 180); // Distribución áurea
      const px = cx + r * Math.cos(peerAngle);
      const py = cy + r * Math.sin(peerAngle);

      ctx.beginPath();
      ctx.arc(px, py, 4, 0, Math.PI * 2);
      ctx.fillStyle = '#10b981';
      ctx.shadowColor = '#10b981';
      ctx.shadowBlur = 8;
      ctx.fill();
      ctx.shadowBlur = 0;
    });

    requestAnimationFrame(renderRadar);
  }

  requestAnimationFrame(renderRadar);
}

// 4. Conexión de Eventos en Tiempo Real (SSE)
function initSSE() {
  if (sseConnection) sseConnection.close();

  sseConnection = new EventSource(`${API_BASE}/api/v1/events`);
  const terminal = document.getElementById('events-terminal');

  function appendLog(text, type = 'info') {
    const time = new Date().toLocaleTimeString();
    const line = document.createElement('div');
    line.className = `log-line ${type}`;
    line.textContent = `[${time}] ${text}`;
    terminal.appendChild(line);
    terminal.scrollTop = terminal.scrollHeight;

    while (terminal.children.length > 50) {
      terminal.removeChild(terminal.firstChild);
    }
  }

  sseConnection.addEventListener('init', e => {
    appendLog('Canal de eventos conectado con el núcleo.', 'info');
  });

  sseConnection.addEventListener('PEER_JOINED', e => {
    const data = JSON.parse(e.data);
    appendLog(`Nuevo par conectado a la malla: ${data.source}`, 'peer');
    fetchPeers();
  });

  sseConnection.addEventListener('PACKET_TX', e => {
    const data = JSON.parse(e.data);
    appendLog(`Datagrama transmitido hacia ${data.payload.target_did} (${data.payload.bytes}B)`, 'tx');
  });

  sseConnection.addEventListener('ZTNA_ALERT', e => {
    const data = JSON.parse(e.data);
    appendLog(`Alerta ZTNA: Paquete descartado de ${data.source}`, 'ztna');
  });

  sseConnection.addEventListener('COMPONENT_BOUND', e => {
    const data = JSON.parse(e.data);
    appendLog(`Componente externo acoplado: ${data.payload.name} (${data.payload.component_id})`, 'comp');
    showToast(`⚡ Componente acoplado: ${data.payload.name}`);
    fetchCoreStatus();
  });

  sseConnection.addEventListener('CHAT_MESSAGE_SENT', e => {
    const data = JSON.parse(e.data);
    appendLog(`Chat E2EE enviado a ${data.payload.target_did}: "${data.payload.text}"`, 'tx');
    if (currentAppView === 'chat') {
      fetchChatMessages(activeChatPeerDID);
    }
  });

  sseConnection.addEventListener('CHAT_MESSAGE_RECEIVED', e => {
    const data = JSON.parse(e.data);
    appendLog(`Chat E2EE recibido de ${data.payload.author_did}: "${data.payload.text}"`, 'info');
    if (currentAppView === 'chat') {
      fetchChatMessages(activeChatPeerDID);
    } else {
      showToast(`💬 Mensaje recibido de ${data.payload.author_did.slice(0, 16)}...`);
    }
  });

  sseConnection.onerror = () => {
    // Reconexión automática nativa de SSE
  };
}

// 5. Gestión del Modal de Acoplamiento
function initModal() {
  const modal = document.getElementById('modal-attach');
  const btnOpen = document.getElementById('btn-open-attach-modal');
  const btnClose = document.getElementById('btn-close-attach-modal');
  const btnCancel = document.getElementById('btn-cancel-attach');
  const form = document.getElementById('form-attach-component');

  const openModal = () => modal.classList.add('open');
  const closeModal = () => modal.classList.remove('open');

  btnOpen.addEventListener('click', openModal);
  btnClose.addEventListener('click', closeModal);
  btnCancel.addEventListener('click', closeModal);
  modal.addEventListener('click', e => {
    if (e.target === modal) closeModal();
  });

  form.addEventListener('submit', async e => {
    e.preventDefault();
    const payload = {
      id: document.getElementById('comp-id').value.trim(),
      name: document.getElementById('comp-name').value.trim(),
      version: document.getElementById('comp-version').value.trim(),
      transport: document.getElementById('comp-transport').value,
      capabilities: document.getElementById('comp-capabilities').value.split(',').map(s => s.trim()).filter(Boolean),
      endpoint: document.getElementById('comp-endpoint').value.trim(),
    };

    try {
      const res = await fetch(`${API_BASE}/api/v1/components/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        const txt = await res.text();
        throw new Error(txt);
      }
      showToast('Componente registrado con éxito');
      closeModal();
      form.reset();
      fetchCoreStatus();
    } catch (err) {
      alert(`Error registrando componente: ${err.message}`);
    }
  });
}

// 6. Envío de Datagramas
function initDatagramForm() {
  const form = document.getElementById('form-send-datagram');
  const didInput = document.getElementById('input-target-did');

  const btnLoopback = document.getElementById('quick-did-loopback');
  if (btnLoopback) {
    btnLoopback.addEventListener('click', () => {
      const myDID = document.getElementById('val-did').textContent.trim();
      didInput.value = myDID || 'did:ipvn7:local:daemon';
      document.getElementById('input-protocol').value = 'mesh:echo';
      showToast('Destino fijado en bucle local (Loopback)');
    });
  }

  const btnDaemon = document.getElementById('quick-did-daemon');
  if (btnDaemon) {
    btnDaemon.addEventListener('click', () => {
      didInput.value = 'did:ipvn7:local:daemon';
      document.getElementById('input-protocol').value = 'mesh:echo';
      showToast('Destino fijado en Daemon Soberano');
    });
  }

  const btnBroadcast = document.getElementById('quick-did-broadcast');
  if (btnBroadcast) {
    btnBroadcast.addEventListener('click', () => {
      didInput.value = 'did:ipvn7:broadcast';
      document.getElementById('input-protocol').value = 'mesh:discovery';
      showToast('Destino fijado en Broadcast de Malla');
    });
  }

  form.addEventListener('submit', async e => {
    e.preventDefault();
    const btn = document.getElementById('btn-send-pkt');
    btn.disabled = true;

    const payloadText = document.getElementById('input-payload').value;
    const encoder = new TextEncoder();
    const payloadBytes = Array.from(encoder.encode(payloadText));

    const datagram = {
      target_did: document.getElementById('input-target-did').value.trim(),
      protocol: document.getElementById('input-protocol').value.trim(),
      priority: parseInt(document.getElementById('input-priority').value, 10),
      payload: payloadBytes
    };

    try {
      const res = await fetch(`${API_BASE}/api/v1/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(datagram)
      });
      if (!res.ok) {
        const errTxt = await res.text();
        throw new Error(errTxt);
      }
      showToast('Datagrama inyectado a través de la malla');
    } catch (err) {
      showToast(`Error de transmisión: ${err.message}`, true);
    } finally {
      btn.disabled = false;
    }
  });
}

// 7. Acciones del Ecosistema Doméstico (Smart Home Hub)
let iotShieldActive = false;
let activeScreenStream = null;
