package core

import "net/http"

// handleTVReceiver sirve la interfaz web interactiva para navegadores de Smart TVs con SSE en vivo
func (s *CoreServer) handleTVReceiver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <title>ipvn7 Smart TV — Receptor Soberano</title>
  <link rel="stylesheet" href="https://cdn.plyr.io/3.7.8/plyr.css">
  <style>
    body { margin: 0; background: #050811; color: #f1f5f9; font-family: sans-serif; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; overflow: hidden; }
    h1 { font-size: 1.8rem; margin: 0 0 4px; color: #38bdf8; }
    p { font-size: 0.95rem; color: #94a3b8; margin: 0 0 12px; }
    .badge { background: rgba(56,189,248,0.15); border: 1px solid #38bdf8; padding: 4px 12px; border-radius: 99px; font-weight: bold; color: #38bdf8; margin-top: 10px; font-size: 0.8rem; }
    .screen-container { width: 92vw; height: 74vh; background: #000; border: 2px solid #1e293b; border-radius: 12px; display: flex; align-items: center; justify-content: center; position: relative; overflow: hidden; }
    .plyr { width: 100% !important; height: 100% !important; border-radius: 10px; }
    .plyr--video { height: 100%; }
  </style>
</head>
<body>
  <h1>ipvn7 Sovereign TV Receiver</h1>
  <p>Conectado a la malla local · Reproductor Plyr de Alta Fidelidad</p>
  <div class="screen-container" id="tv-box">
    <div id="tv-status" style="color:#64748b; font-size:1.2rem;">Esperando video o señal desde el Panel de Control...</div>
  </div>
  <div class="badge">Nodo Local Activo · 0 Servidores Centrales</div>
  <script src="https://cdn.plyr.io/3.7.8/plyr.js"></script>
  <script>
    var player = null;
    function extractYt(u) {
      if (!u) return '';
      if (u.length === 11 && !u.includes('/') && !u.includes('.')) return u;
      var p = u.split('youtu.be/'); if (p.length > 1) return p[1].split('?')[0];
      p = u.split('v='); if (p.length > 1) return p[1].split('&')[0];
      return '';
    }
    function playMedia(url) {
      var box = document.getElementById('tv-box');
      if (player) { try { player.destroy(); } catch(e){} player = null; }
      var yt = extractYt(url);
      if (yt) {
        box.innerHTML = '<div id="plyr-el" data-plyr-provider="youtube" data-plyr-embed-id="' + yt + '"></div>';
      } else if (url) {
        box.innerHTML = '<video id="plyr-el" playsinline controls><source src="' + url + '"></video>';
      } else return;
      player = new Plyr('#plyr-el', { autoplay: true, muted: false, youtube: { noCookie: true, rel: 0 } });
      player.on('ready', function() { try { player.play(); } catch(e){} });
    }
    var p = new URLSearchParams(window.location.search);
    var initV = p.get('v') || p.get('media');
    if (initV) playMedia(initV);
    try {
      var es = new EventSource('/api/v1/events');
      es.onmessage = function(e) {
        try {
          var d = JSON.parse(e.data);
          if (d.type === 'HOME_CAST_STARTED' && d.payload && d.payload.media_url) playMedia(d.payload.media_url);
          else if (d.type === 'HOME_CAST_STOPPED') {
            if (player) { try { player.destroy(); } catch(err){} player = null; }
            document.getElementById('tv-box').innerHTML = '<div style="color:#64748b; font-size:1.2rem;">Transmisión finalizada. Esperando nueva señal...</div>';
          } else if (d.type === 'HOME_CAST_CONTROL' && player) {
            var a = d.payload.action, v = d.payload.value;
            if (a === 'pause') player.pause();
            if (a === 'play') player.play();
            if (a === 'volume') player.volume = Math.max(0, Math.min(1, v / 100));
            if (a === 'mute') player.muted = !player.muted;
          }
        } catch(err){}
      };
    } catch(e){}
  </script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}
