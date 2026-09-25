// app_intent.js — Orquestador de Intenciones en Lenguaje Natural para el Panel ipvn7
(function() {
  'use strict';

  window.executeHumanIntent = async function(promptText, targetNode) {
    if (!promptText || !promptText.trim()) return;
    targetNode = targetNode || 'local';

    var statusEl = document.getElementById('intent-status-feedback');
    if (statusEl) {
      statusEl.style.display = 'block';
      statusEl.innerHTML = '<span style="color:#38bdf8;">⏳ Procesando:</span> "' + promptText + '"...';
    }

    try {
      var resp = await fetch('/api/v1/intent', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: promptText, target_node: targetNode })
      });
      var data = await resp.json();

      if (statusEl) {
        if (data.success) {
          statusEl.innerHTML = '<span style="color:#10b981;">✅ Éxito:</span> ' + (data.human_message || 'Orden ejecutada');
        } else {
          statusEl.innerHTML = '<span style="color:#ef4444;">❌ Fallo:</span> ' + (data.error || data.human_message || 'No se pudo ejecutar');
        }
      }

      if (window.showToast) {
        window.showToast(data.success ? 'success' : 'error', data.human_message || data.error);
      }
    } catch (err) {
      if (statusEl) {
        statusEl.innerHTML = '<span style="color:#ef4444;">❌ Error de conexión:</span> ' + err.message;
      }
    }
  };

  window.setupIntentBar = function() {
    var form = document.getElementById('form-human-intent');
    var input = document.getElementById('input-human-intent');
    if (form && input) {
      form.addEventListener('submit', function(e) {
        e.preventDefault();
        var val = input.value;
        if (val) {
          window.executeHumanIntent(val);
          input.value = '';
        }
      });
    }
  };

  document.addEventListener('DOMContentLoaded', window.setupIntentBar);
})();
