package templates

const WebUITemplate = `<!DOCTYPE html><html lang="es"><head><meta charset="utf-8"><title>IPVN7</title>
<meta name="viewport" content="width=device-width,initial-scale=1"><style>
:root{--bg:#0b0f19;--card:#111827;--txt:#f3f4f6;--dim:#9ca3af;--red:#ef4444;--yellow:#f59e0b;--green:#10b981;--cyan:#06b6d4;--card-border:#1f2937}
*{box-sizing:border-box}body{background:var(--bg);color:var(--txt);font-family:'Segoe UI',-apple-system,BlinkMacSystemFont,sans-serif;margin:0;padding:15px;display:flex;justify-content:center;align-items:center;min-height:100vh}
.app{max-width:420px;width:100%%;background:var(--card);border:1px solid var(--card-border);border-radius:24px;padding:24px;box-shadow:0 20px 50px rgba(0,0,0,0.7);text-align:center;position:relative}
.device-name{font-size:24px;font-weight:700;margin-bottom:6px;color:var(--txt);letter-spacing:-0.5px}
.device-did{font-size:12px;color:var(--dim);margin-bottom:18px;font-family:'Consolas','Monaco',monospace;word-break:break-all;line-height:1.3;padding:0 8px}
.metrics-panel{display:flex;justify-content:space-around;margin-bottom:16px;padding:12px 10px;background:#0d1117;border-radius:12px;border:1px solid #21262d}
.metric-item{text-align:center}
.metric-label{font-size:10px;color:var(--dim);text-transform:uppercase;letter-spacing:0.5px;margin-bottom:3px}
.metric-value{font-size:15px;font-weight:600;color:var(--txt);font-family:'Consolas','Monaco',monospace}
.ecg-card{background:#060a12;border:1px solid #1a2333;border-radius:14px;padding:10px 14px;margin-bottom:20px;text-align:left;position:relative;overflow:hidden;box-shadow:inset 0 0 20px rgba(0,0,0,0.8)}
.ecg-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:6px}
.ecg-title{font-size:11px;font-weight:700;letter-spacing:1px;color:var(--dim);text-transform:uppercase;display:flex;align-items:center;gap:6px}
.ecg-title-dot{width:7px;height:7px;border-radius:50%%;background:var(--green);display:inline-block;box-shadow:0 0 8px var(--green);animation:heartbeat-blink 1.2s infinite}
.ecg-bpm{font-family:'Consolas',monospace;font-size:12px;font-weight:700;color:var(--green);background:rgba(16,185,129,0.12);padding:2px 8px;border-radius:20px;border:1px solid rgba(16,185,129,0.3);display:flex;align-items:center;gap:4px}
.ecg-bpm.yellow{color:var(--yellow);background:rgba(245,158,11,0.12);border-color:rgba(245,158,11,0.3)}
.ecg-bpm.red{color:var(--red);background:rgba(239,68,68,0.12);border-color:rgba(239,68,68,0.3)}
.ecg-canvas-container{position:relative;width:100%%;height:46px;background:radial-gradient(ellipse at center,rgba(16,185,129,0.04) 0%%,transparent 70%%);border-radius:6px;background-image:linear-gradient(rgba(255,255,255,0.03) 1px,transparent 1px),linear-gradient(90deg,rgba(255,255,255,0.03) 1px,transparent 1px);background-size:16px 16px}
.ecg-canvas{display:block;width:100%%;height:100%%}
.circle-container{position:relative;width:200px;height:200px;margin:0 auto 6px}
.circle-btn{width:200px;height:200px;border-radius:50%%;border:none;cursor:pointer;display:flex;flex-direction:column;justify-content:center;align-items:center;font-weight:700;transition:all 0.3s;outline:none;position:relative}
.circle-green{background:#0d231a;border:4px solid var(--green);color:var(--green);box-shadow:0 0 50px rgba(16,185,129,0.45);animation:cardiac-lubdub 1.2s ease-in-out infinite}
.circle-yellow{background:#231a0e;border:4px solid var(--yellow);color:var(--yellow);box-shadow:0 0 50px rgba(245,158,11,0.4);animation:spin-pulse 1.4s infinite}
.circle-red{background:#1f1315;border:4px solid var(--red);color:var(--red);box-shadow:0 0 50px rgba(239,68,68,0.3)}
.heartbeat-ring{position:absolute;top:-12px;left:-12px;width:calc(100%% + 24px);height:calc(100%% + 24px);border-radius:50%%;border:3px solid var(--green);pointer-events:none;opacity:0.8;animation:heartbeat-ripple 1.2s cubic-bezier(0.215,0.61,0.355,1) infinite;z-index:1}
.heartbeat-ring.yellow{border-color:var(--yellow);animation-duration:1.5s}
.heartbeat-ring.red{border-color:var(--red);opacity:0.2;animation:none}
@keyframes cardiac-lubdub{0%%{transform:scale(1)}14%%{transform:scale(1.05)}26%%{transform:scale(1)}40%%{transform:scale(1.03)}54%%{transform:scale(1)}100%%{transform:scale(1)}}
@keyframes heartbeat-ripple{0%%{transform:scale(0.96);opacity:0.8;box-shadow:0 0 10px rgba(16,185,129,0.5)}14%%{transform:scale(1.04);opacity:1;box-shadow:0 0 25px rgba(16,185,129,0.8)}26%%{transform:scale(0.99);opacity:0.6}40%%{transform:scale(1.02);opacity:0.9;box-shadow:0 0 20px rgba(16,185,129,0.7)}70%%{transform:scale(1.12);opacity:0}100%%{transform:scale(1.12);opacity:0}}
@keyframes heartbeat-blink{0%%,100%%{opacity:1;transform:scale(1)}50%%{opacity:0.4;transform:scale(0.85)}}
@keyframes spin-pulse{0%%,100%%{opacity:1;transform:scale(1)}50%%{opacity:0.85;transform:scale(0.97)}}
.circle-icon{font-size:46px;line-height:1;margin-bottom:6px}
.circle-status{font-size:15px;letter-spacing:1px;font-weight:700;text-transform:uppercase}
.circle-sub{font-size:11px;letter-spacing:0.5px;color:var(--dim);margin-top:4px;font-weight:500}
.close-btn{position:absolute;top:15px;right:15px;width:32px;height:32px;border-radius:50%%;border:1px solid #374151;background:#1f2937;color:#9ca3af;font-size:18px;cursor:pointer;display:flex;align-items:center;justify-content:center;transition:all 0.2s;padding:0;line-height:1}
.close-btn:hover{background:#374151;color:#f3f4f6;border-color:#4b5563}
.guide-btn{position:absolute;top:15px;left:15px;width:32px;height:32px;border-radius:50%%;border:1px solid #374151;background:#1f2937;color:var(--cyan);font-size:15px;cursor:pointer;display:flex;align-items:center;justify-content:center;text-decoration:none;transition:all 0.2s}
.guide-btn:hover{background:#374151;border-color:var(--cyan);transform:scale(1.05)}
</style></head><body><div class="app">
<a href="/guide/" target="_blank" class="guide-btn" title="Abrir Guía Interactiva Soberana">🧭</a>
<button class="close-btn" onclick="exitApp()" title="Cerrar aplicación">×</button>
<div class="device-name" id="deviceName">%s</div>
<div class="device-did" id="deviceDID">%s</div>
<div id="updateBanner" style="display:none;margin-bottom:14px;padding:6px 14px;background:rgba(6,182,212,0.12);border:1px solid rgba(6,182,212,0.35);border-radius:20px;font-size:11px;color:var(--cyan);cursor:pointer" onclick="applyUpdate()">✨ <span id="updateText">Nueva versión</span> · <b>Actualizar en 1-Clic</b></div>
<div class="metrics-panel">
<div class="metric-item"><div class="metric-label">Pares</div><div class="metric-value" id="metricPeers">0</div></div>
<div class="metric-item"><div class="metric-label">Tráfico</div><div class="metric-value" id="metricTraffic">0 B</div></div>
<div class="metric-item"><div class="metric-label">Tiempo</div><div class="metric-value" id="metricUptime">0s</div></div>
</div>
<div class="ecg-card">
<div class="ecg-header">
<div class="ecg-title"><span class="ecg-title-dot" id="ecgDot"></span><span id="ecgStatusText">Ritmo Cardíaco · Sinusal</span></div>
<div class="ecg-bpm" id="ecgBpm"><span id="heartIcon">❤️</span> <span id="bpmVal">72</span> BPM</div>
</div>
<div class="ecg-canvas-container">
<canvas class="ecg-canvas" id="ecgCanvas" width="370" height="46"></canvas>
</div>
</div>
<div class="circle-container">
<div class="heartbeat-ring" id="heartbeat"></div>
<button id="circleBtn" class="circle-btn circle-green" onclick="cycleState()">
<div class="circle-icon" id="circleIcon">⚡</div>
<div class="circle-status" id="circleStatus">CONECTADO</div>
<div class="circle-sub" id="circleSub">RED VITAL ACTIVA</div>
</button>
</div>
</div>
<script>
let curState="connected",bpm=72,lastTraffic={rx:0,tx:0,time:Date.now()};
const canvas=document.getElementById('ecgCanvas'),ctx=canvas.getContext('2d');
let scanX=0,points=[],maxPoints=370;
for(let i=0;i<maxPoints;i++)points.push(23);
function getEcgY(t){
if(t<0.15)return 23;
if(t<0.22){let p=(t-0.15)/0.07;return 23-Math.sin(p*Math.PI)*4;}
if(t<0.28)return 23;
if(t<0.31){let q=(t-0.28)/0.03;return 23+q*3;}
if(t<0.35){let r=(t-0.31)/0.04;return 26-r*22;}
if(t<0.39){let s=(t-0.35)/0.04;return 4+s*34;}
if(t<0.43){let b=(t-0.39)/0.04;return 38-b*15;}
if(t<0.52)return 23;
if(t<0.65){let tw=(t-0.52)/0.13;return 23-Math.sin(tw*Math.PI)*6;}
return 23;
}
let cycleProgress=0;
function drawEcg(){
requestAnimationFrame(drawEcg);
let speed=(bpm/60)*0.015,color='#10b981',glow='rgba(16,185,129,0.8)';
if(curState==='connecting'||curState==='disconnecting'){color='#f59e0b';glow='rgba(245,158,11,0.8)';}
else if(curState==='disconnected'){color='#ef4444';glow='rgba(239,68,68,0.6)';}
let newY=23;
if(curState==='connected'){cycleProgress=(cycleProgress+speed)%%1;newY=getEcgY(cycleProgress);}
else if(curState==='connecting'||curState==='disconnecting'){cycleProgress=(cycleProgress+0.03)%%1;newY=23+Math.sin(cycleProgress*Math.PI*4)*8;}
else{newY=23+(Math.random()-0.5)*1.2;}
for(let s=0;s<2;s++){points[scanX]=newY;scanX=(scanX+1)%%maxPoints;}
ctx.clearRect(0,0,canvas.width,canvas.height);
ctx.beginPath();ctx.lineWidth=2;ctx.strokeStyle=color;ctx.shadowBlur=8;ctx.shadowColor=glow;
for(let i=0;i<maxPoints;i++){
if(Math.abs(i-scanX)<8)continue;
if(i===0||Math.abs(i-1-scanX)<8){ctx.moveTo(i,points[i]);}else{ctx.lineTo(i,points[i]);}
}
ctx.stroke();
ctx.beginPath();ctx.arc(scanX,newY,3,0,Math.PI*2);ctx.fillStyle='#ffffff';ctx.shadowBlur=12;ctx.shadowColor=color;ctx.fill();
}
requestAnimationFrame(drawEcg);
async function exitApp(){try{await fetch('/api/v1/vpn/exit',{method:'POST'});window.close();}catch(e){}}
async function cycleState(){try{let r=await fetch('/api/v1/vpn/cycle',{method:'POST'});let d=await r.json();setUI(d.status);}catch(e){}}
function setUI(st){
curState=st;
let b=document.getElementById('circleBtn'),i=document.getElementById('circleIcon'),l=document.getElementById('circleStatus'),sub=document.getElementById('circleSub'),ring=document.getElementById('heartbeat'),bpmBadge=document.getElementById('ecgBpm'),stText=document.getElementById('ecgStatusText'),dot=document.getElementById('ecgDot');
b.className="circle-btn";ring.className="heartbeat-ring";bpmBadge.className="ecg-bpm";
if(st==="connected"){
b.classList.add("circle-green");i.innerText="⚡";l.innerText="CONECTADO";sub.innerText="RED VITAL ACTIVA";
dot.style.background="var(--green)";dot.style.boxShadow="0 0 8px var(--green)";stText.innerText="Ritmo Cardíaco · Sinusal";
}else if(st==="connecting"){
b.classList.add("circle-yellow");ring.classList.add("yellow");bpmBadge.classList.add("yellow");i.innerText="⏳";l.innerText="CONECTANDO";sub.innerText="SINCRONIZANDO...";
dot.style.background="var(--yellow)";dot.style.boxShadow="0 0 8px var(--yellow)";stText.innerText="Buscando Ritmo de Red...";bpm=88;document.getElementById('bpmVal').innerText=bpm;
}else if(st==="disconnecting"){
b.classList.add("circle-yellow");ring.classList.add("yellow");bpmBadge.classList.add("yellow");i.innerText="⏳";l.innerText="DESCONECTANDO";sub.innerText="CERRANDO ENLACE";
dot.style.background="var(--yellow)";dot.style.boxShadow="0 0 8px var(--yellow)";stText.innerText="Desacelerando...";
}else{
b.classList.add("circle-red");ring.classList.add("red");bpmBadge.classList.add("red");i.innerText="🔌";l.innerText="CONECTAR";sub.innerText="PRESIONA PARA INICIAR";
dot.style.background="var(--red)";dot.style.boxShadow="0 0 8px var(--red)";stText.innerText="Asistolia · Sin Señal";bpm=0;document.getElementById('bpmVal').innerText="0";
}
}
function fb(b){if(b<1024)return b+' B';let k=b/1024;if(k<1024)return k.toFixed(1)+' KB';let m=k/1024;if(m<1024)return m.toFixed(1)+' MB';return (m/1024).toFixed(2)+' GB';}
function updateHeartbeat(rx,tx){
let now=Date.now(),delta=(now-lastTraffic.time)/1000;if(delta<=0)delta=1;
let rxRate=(rx-lastTraffic.rx)/delta,txRate=(tx-lastTraffic.tx)/delta,totalRate=rxRate+txRate;
lastTraffic={rx:rx,tx:tx,time:now};
if(curState==='connected'){
let rateKB=totalRate/1024;
bpm=Math.round(68+Math.min(rateKB*1.5,52));
document.getElementById('bpmVal').innerText=bpm;
let hb=document.getElementById('heartbeat'),btn=document.getElementById('circleBtn');
let dur=(60/bpm).toFixed(2);
hb.style.animationDuration=dur+'s';btn.style.animationDuration=dur+'s';
}
}
async function poll(){
try{
let r=await fetch('/api/v1/status');let d=await r.json();
if(d.vpn_state&&d.vpn_state!==curState){setUI(d.vpn_state);}
if(d.peers_count!==undefined){document.getElementById('metricPeers').innerText=d.peers_count;}
if(d.uptime_sec!==undefined){document.getElementById('metricUptime').innerText=d.uptime_sec+'s';}
if(d.gateway){
let el=document.getElementById('metricTraffic');
if(el)el.innerText=fb(d.gateway.bytes_rx||0)+' / '+fb(d.gateway.bytes_tx||0);
updateHeartbeat(d.gateway.bytes_rx||0,d.gateway.bytes_tx||0);
}
}catch(e){}
}
setInterval(poll,2000);setUI('connected');poll();
async function checkUpdates(){try{let r=await fetch('/api/v1/update/check');if(r.ok){let d=await r.json();if(d.available){document.getElementById('updateText').innerText='Nueva versión '+d.latest_version;document.getElementById('updateBanner').style.display='inline-block';}}}catch(e){}}
async function applyUpdate(){let b=document.getElementById('updateBanner');b.innerHTML='⏳ <i>Descargando y verificando firma SHA-256...</i>';try{let r=await fetch('/api/v1/update/apply',{method:'POST'});let d=await r.json();if(d.ok){b.innerHTML='✅ <b>¡Actualizado a '+d.new_version+'! Reiniciando...</b>';b.style.borderColor='var(--green)';b.style.color='var(--green)';setTimeout(()=>location.reload(),3000);}else{b.innerHTML='⚠️ Error: '+(d.error||'Falla al actualizar');}}catch(e){b.innerHTML='⚠️ Error de conexión al actualizar';}}
setTimeout(checkUpdates,1500);
</script></body></html>`
