// ipvn7 Universal Core Dashboard — Chat Soberano E2EE (app_chat.js)
// Mensajería Criptográfica L4 Ed25519 (Axioma III: <= 400 líneas)

// ==========================================================================
// 10. GESTIÓN DEL CHAT SOBERANO E2EE (L4)
// ==========================================================================

let activeChatPeerDID = null;
let chatContacts = [];

function initChatApp() {
  const form = document.getElementById('form-chat-send');
  if (form) {
    form.addEventListener('submit', sendChatMessage);
  }
}

async function fetchChatContacts() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/chat/contacts`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    chatContacts = await res.json() || [];
    renderChatContacts(chatContacts);
    if (chatContacts.length > 0 && !activeChatPeerDID) {
      selectChatContact(chatContacts[0]);
    }
  } catch (err) {
    console.debug('Error consultando contactos de chat:', err);
  }
}

function selectChatContact(contact) {
  activeChatPeerDID = contact.did;
  const avatarEl = document.getElementById('conv-peer-avatar');
  const nameEl = document.getElementById('conv-peer-name');
  const didEl = document.getElementById('conv-peer-did');

  if (avatarEl) avatarEl.textContent = contact.avatar || '💻';
  if (nameEl) nameEl.textContent = contact.name || contact.did;
  if (didEl) didEl.textContent = contact.did;

  const cards = document.querySelectorAll('.chat-contact-card');
  cards.forEach(c => {
    if (c.dataset.did === contact.did) {
      c.classList.add('active');
    } else {
      c.classList.remove('active');
    }
  });

  fetchChatMessages(contact.did);
}

function renderChatContacts(contacts) {
  const container = document.getElementById('chat-contacts-container');
  if (!container) return;

  if (!contacts || contacts.length === 0) {
    container.innerHTML = '<div style="padding:16px; color:#94a3b8; font-size:0.8rem;">Sin contactos en la malla</div>';
    return;
  }

  container.innerHTML = contacts.map(c => {
    const isOnline = c.status === 'online';
    const activeClass = (activeChatPeerDID === c.did) ? 'active' : '';
    return `
      <div class="chat-contact-card ${activeClass}" data-did="${escapeHtml(c.did)}" onclick="selectChatContactByDID('${escapeHtml(c.did)}')">
        <span class="contact-avatar">${c.avatar || '💻'}</span>
        <div class="contact-info">
          <div class="contact-name">${escapeHtml(c.name || c.did)}</div>
          <div class="contact-status-line">
            <span class="contact-dot ${isOnline ? 'online' : ''}"></span>
            <span>${isOnline ? 'En línea' : 'Desconectado'}</span>
          </div>
        </div>
      </div>
    `;
  }).join('');
}

function selectChatContactByDID(did) {
  const c = chatContacts.find(item => item.did === did);
  if (c) selectChatContact(c);
}

async function fetchChatMessages(peerDID) {
  if (!peerDID) return;
  try {
    const res = await fetch(`${API_BASE}/api/v1/chat/messages?peer=${encodeURIComponent(peerDID)}`);
    if (!res.ok) return;
    const messages = await res.json() || [];
    renderChatMessages(messages);
  } catch (e) {
    console.debug('Error consultando mensajes:', e);
  }
}

function renderChatMessages(messages) {
  const container = document.getElementById('chat-messages-container');
  if (!container) return;

  if (!messages || messages.length === 0) {
    container.innerHTML = `
      <div class="chat-empty-hint">
        <span style="font-size:2.5rem;">💬</span>
        <p>Inicia una conversación soberana protegida por firmas Ed25519 y persistencia DAG.</p>
      </div>
    `;
    return;
  }

  const myDID = document.getElementById('val-did') ? document.getElementById('val-did').textContent.trim() : '';

  container.innerHTML = messages.map(m => {
    const isOutgoing = m.author_did === myDID || m.author_did !== 'did:ipvn7:local:daemon';
    const roleClass = isOutgoing ? 'outgoing' : 'incoming';
    const timeStr = new Date(m.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    const shortSig = m.signature ? m.signature.slice(0, 10) + '...' : 'Ed25519';
    const latency = m.delivery_latency_ms ? `${m.delivery_latency_ms.toFixed(2)} ms` : '<1 ms';

    return `
      <div class="chat-bubble-wrapper ${roleClass}">
        <div class="chat-bubble">${escapeHtml(m.text)}</div>
        <div class="chat-bubble-meta">
          <span>${timeStr}</span>
          <span class="chip-sig">🔑 ${shortSig}</span>
          <span class="chip-latency">⚡ ${latency}</span>
        </div>
      </div>
    `;
  }).join('');

  container.scrollTop = container.scrollHeight;
}

async function sendChatMessage(e) {
  e.preventDefault();
  const input = document.getElementById('chat-input-text');
  const btn = document.getElementById('btn-chat-submit');
  const text = input ? input.value.trim() : '';
  if (!text) return;

  const targetDID = activeChatPeerDID || (document.getElementById('val-did') ? document.getElementById('val-did').textContent.trim() : '');

  if (btn) btn.disabled = true;

  try {
    const res = await fetch(`${API_BASE}/api/v1/chat/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_did: targetDID, text: text })
    });
    if (!res.ok) {
      const errTxt = await res.text();
      throw new Error(errTxt);
    }
    input.value = '';
    fetchChatMessages(targetDID);
  } catch (err) {
    showToast(`Error al enviar mensaje: ${err.message}`, true);
  } finally {
    if (btn) btn.disabled = false;
  }
}

// ==========================================================================