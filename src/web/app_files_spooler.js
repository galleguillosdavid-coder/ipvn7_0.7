// app_files_spooler.js — Gestión de Archivos Locales (Cloud Drop) y Spooler de Impresión
(function() {
  'use strict';

  function initFilesAndDrop() {
    var fileInp = document.getElementById('file-drop-input');
    var dropTxt = document.getElementById('drop-status-text');
    var dropZone = document.getElementById('drop-zone');

    if (dropZone && fileInp) {
      dropZone.addEventListener('click', function() { fileInp.click(); });
      dropZone.addEventListener('dragover', function(e) { e.preventDefault(); dropZone.style.borderColor = '#38bdf8'; });
      dropZone.addEventListener('dragleave', function() { dropZone.style.borderColor = '#334155'; });
      dropZone.addEventListener('drop', function(e) {
        e.preventDefault();
        dropZone.style.borderColor = '#334155';
        if (e.dataTransfer.files && e.dataTransfer.files.length) {
          fileInp.files = e.dataTransfer.files;
          fileInp.dispatchEvent(new Event('change'));
        }
      });
    }

    if (fileInp && dropTxt) {
      fileInp.addEventListener('change', async function(e) {
        if (!e.target.files || !e.target.files.length) return;
        var f = e.target.files[0];
        dropTxt.textContent = 'Transfiriendo ' + f.name + '...';
        try {
          var fd = new FormData();
          fd.append('file', f);
          var res = await fetch(API_BASE + '/api/v1/home/drop', { method: 'POST', body: fd });
          var d = await res.json();
          var spd = typeof d.speed_mbps === 'number' ? d.speed_mbps.toFixed(2) : d.speed_mbps;
          dropTxt.textContent = '✓ ' + f.name + ' @ ' + spd + ' Mbps';
          if (window.showToast) window.showToast('success', "'" + f.name + "' transferido a " + spd + ' Mbps');
          window.fetchDropFiles();
        } catch (err) {
          dropTxt.textContent = 'Error transfiriendo';
          if (window.showToast) window.showToast('error', 'Error en Local Drop');
        }
      });
    }

    var btnToggleDrop = document.getElementById('btn-toggle-drop-files');
    if (btnToggleDrop) {
      btnToggleDrop.addEventListener('click', function() {
        var drawer = document.getElementById('drop-files-drawer');
        if (drawer) {
          var open = drawer.style.display === 'block';
          drawer.style.display = open ? 'none' : 'block';
          if (!open) window.fetchDropFiles();
        }
      });
    }

    var btnOpenDrop = document.getElementById('btn-open-drop-folder');
    if (btnOpenDrop) {
      btnOpenDrop.addEventListener('click', async function() {
        try {
          await fetch(API_BASE + '/api/v1/home/drop/open', { method: 'POST' });
          if (window.showToast) window.showToast('success', '📂 Carpeta de descargas abierta localmente');
        } catch (err) {
          if (window.showToast) window.showToast('error', 'No se pudo abrir la carpeta local');
        }
      });
    }

    window.fetchDropFiles();
  }

  window.fetchDropFiles = async function() {
    try {
      var res = await fetch(API_BASE + '/api/v1/home/drop');
      if (!res.ok) return;
      var files = (await res.json()) || [];
      var b = document.getElementById('drop-files-badge');
      var d = document.getElementById('drop-files-drawer');
      if (b) b.textContent = files.length;
      if (!d) return;
      if (files.length === 0) {
        d.innerHTML = '<div style="color:#94a3b8; text-align:center; padding:4px;">Bandeja vacía</div>';
        return;
      }
      d.innerHTML = files.map(function(f) {
        return '<div style="display:flex; justify-content:space-between; align-items:center; padding:3px 0; border-bottom:1px solid #1e293b;">' +
          '<span style="overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width:140px; color:#cbd5e1;" title="' + escapeHtml(f.name) + '">📄 ' + escapeHtml(f.name) + '</span>' +
          '<a href="' + escapeHtml(f.url) + '" download class="btn-action secondary btn-compact" style="font-size:0.68rem; padding:2px 6px;">⬇️ Bajar</a>' +
          '</div>';
      }).join('');
    } catch (e) {}
  };

  window.initPrintJobs = function() {
    window.fetchPrintJobs();
    var btnToggle = document.getElementById('btn-toggle-print-jobs');
    if (btnToggle) {
      btnToggle.addEventListener('click', function() {
        var d = document.getElementById('print-jobs-drawer');
        if (d) {
          d.style.display = d.style.display === 'block' ? 'none' : 'block';
          if (d.style.display === 'block') window.fetchPrintJobs();
        }
      });
    }
  };

  window.fetchPrintJobs = async function() {
    try {
      var res = await fetch(API_BASE + '/api/v1/home/print/jobs');
      if (!res.ok) return;
      var jobs = (await res.json()) || [];
      var b = document.getElementById('print-jobs-badge');
      var el = document.getElementById('print-jobs-list');
      if (b) b.textContent = jobs.length;
      if (!el) return;
      if (jobs.length === 0) {
        el.innerHTML = '<div style="color:#94a3b8; font-size:0.75rem; text-align:center; padding:8px;">Sin trabajos encolados.</div>';
        return;
      }
      el.innerHTML = jobs.map(function(j) {
        return '<div class="spool-job-card">' +
          '<div>' +
            '<div class="spool-job-title">' + escapeHtml(j.doc_name || 'doc') + '</div>' +
            '<div class="spool-job-sub">' + escapeHtml(j.job_id || '') + ' · ' + escapeHtml(j.print_method || 'Spooler') + ' (' + (j.payload_bytes || 0) + 'B)</div>' +
          '</div>' +
          '<div style="display:flex; align-items:center; gap:4px;">' +
            '<a href="/api/v1/home/print/jobs?view=' + encodeURIComponent(j.job_id || '') + '" target="_blank" class="btn-action secondary btn-compact" style="font-size:0.68rem; padding:2px 6px;" title="Ver imagen de comprobante">👁️ Ver</a>' +
            '<span class="spool-job-status ' + (j.printed_phys ? 'ok' : '') + '">' + (j.printed_phys ? 'Enviado' : 'Encolado') + '</span>' +
          '</div>' +
        '</div>';
      }).join('');
    } catch (e) {}
  };

  document.addEventListener('DOMContentLoaded', function() {
    initFilesAndDrop();
    window.initPrintJobs();
  });
})();
