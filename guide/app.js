/* ==========================================================================
   GUÍA INTERACTIVA SOBERANA IPVN7 - MOTOR REACTIVO DE APRENDIZAJE PROGRESIVO
   Axioma III: <250 líneas | Vanilla JS | 60 FPS | Zero External Libraries
   ========================================================================== */

let orbState = "connected", bpm = 72, ecgProgress = 0, ecgScanX = 0;
const ecgPoints = new Array(320).fill(20);

window.addEventListener("DOMContentLoaded", () => {
  initEcgCanvas(); initHighwayCanvas(); initMeshCanvas();
});

function scrollToSection(id) {
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: "smooth" });
}

// Control Interactivo del Orbe Zen y Electrocardiograma
function cycleOrb() {
  const b = document.getElementById("orbBtn"), r = document.getElementById("orbRing");
  const i = document.getElementById("orbIcon"), s = document.getElementById("orbStatus");
  const sub = document.getElementById("orbSub"), bp = document.getElementById("bpmVal"), st = document.getElementById("ecgStatus");

  if (orbState === "connected") {
    orbState = "disconnected"; b.className = "orb-btn orb-red"; r.className = "orb-ring red";
    i.innerText = "🔌"; s.innerText = "DESCONECTADO"; sub.innerText = "TOCA PARA ACTIVAR AUTOPISTA";
    bpm = 0; bp.innerText = "0"; st.innerText = "Asistolia · Sin Señal de Red"; st.style.color = "var(--red)";
  } else if (orbState === "disconnected") {
    orbState = "connecting"; b.className = "orb-btn orb-yellow"; r.className = "orb-ring yellow";
    i.innerText = "⏳"; s.innerText = "SINCRONIZANDO"; sub.innerText = "BLINDANDO CLAVES POST-CUÁNTICAS...";
    bpm = 90; bp.innerText = "90"; st.innerText = "Negociando ML-KEM-768..."; st.style.color = "var(--amber)";
    setTimeout(() => { if (orbState === "connecting") cycleOrb(); }, 1800);
  } else {
    orbState = "connected"; b.className = "orb-btn orb-green"; r.className = "orb-ring";
    i.innerText = "⚡"; s.innerText = "PROTEGIDO"; sub.innerText = "AUTOPISTA PROPIA ACTIVA";
    bpm = 72; bp.innerText = "72"; st.innerText = "Ritmo Sinusal · Salud Óptima"; st.style.color = "var(--green)";
  }
}

// Trazado del ECG en Tiempo Real
function initEcgCanvas() {
  const canvas = document.getElementById("ecgCanvas");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");

  function getEcgY(t) {
    if (t < 0.15) return 20;
    if (t < 0.22) return 20 - Math.sin(((t - 0.15) / 0.07) * Math.PI) * 4;
    if (t < 0.31) return 20;
    if (t < 0.35) return 23 - ((t - 0.31) / 0.04) * 20;
    if (t < 0.39) return 3 + ((t - 0.35) / 0.04) * 30;
    if (t < 0.43) return 33 - ((t - 0.39) / 0.04) * 13;
    if (t < 0.52) return 20;
    if (t < 0.65) return 20 - Math.sin(((t - 0.52) / 0.13) * Math.PI) * 5;
    return 20;
  }

  function loop() {
    requestAnimationFrame(loop);
    let speed = (bpm / 60) * 0.018, color = "#10b981", glow = "rgba(16, 185, 129, 0.8)", newY = 20;
    if (orbState === "connecting") {
      color = "#f59e0b"; glow = "rgba(245, 158, 11, 0.8)";
      ecgProgress = (ecgProgress + 0.03) % 1; newY = 20 + Math.sin(ecgProgress * Math.PI * 4) * 7;
    } else if (orbState === "disconnected") {
      color = "#ef4444"; glow = "rgba(239, 68, 68, 0.6)"; newY = 20 + (Math.random() - 0.5) * 1.5;
    } else {
      ecgProgress = (ecgProgress + speed) % 1; newY = getEcgY(ecgProgress);
    }
    for (let s = 0; s < 2; s++) { ecgPoints[ecgScanX] = newY; ecgScanX = (ecgScanX + 1) % ecgPoints.length; }

    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.beginPath(); ctx.lineWidth = 2; ctx.strokeStyle = color; ctx.shadowBlur = 6; ctx.shadowColor = glow;
    for (let i = 0; i < ecgPoints.length; i++) {
      if (Math.abs(i - ecgScanX) < 6) continue;
      if (i === 0 || Math.abs(i - 1 - ecgScanX) < 6) ctx.moveTo(i, ecgPoints[i]); else ctx.lineTo(i, ecgPoints[i]);
    }
    ctx.stroke();
    ctx.beginPath(); ctx.arc(ecgScanX, newY, 3, 0, Math.PI * 2); ctx.fillStyle = "#fff"; ctx.shadowBlur = 10; ctx.shadowColor = color; ctx.fill();
  }
  loop();
}

// Lienzo Comparativo de la Autopista (Internet vs IPvN7)
function initHighwayCanvas() {
  const canvas = document.getElementById("highwayCanvas");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  let t = 0;

  function draw() {
    requestAnimationFrame(draw);
    t += 0.015;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const w = canvas.width;

    ctx.fillStyle = "rgba(255, 255, 255, 0.5)"; ctx.font = "11px sans-serif";
    ctx.fillText("INTERNET COMERCIAL (Con peajes, servidores de empresas y rastreo)", 20, 24);
    ctx.strokeStyle = "rgba(239, 68, 68, 0.3)"; ctx.lineWidth = 3; ctx.setLineDash([6, 6]);
    ctx.beginPath(); ctx.moveTo(40, 60); ctx.lineTo(160, 60); ctx.lineTo(240, 90); ctx.lineTo(360, 40); ctx.lineTo(480, 80); ctx.lineTo(w - 40, 60); ctx.stroke(); ctx.setLineDash([]);

    const checkpoints = [{ x: 160, y: 60, name: "ISP Local" }, { x: 240, y: 90, name: "Servidor Central" }, { x: 360, y: 40, name: "Empresa Espía" }, { x: 480, y: 80, name: "Ruta Saturada" }];
    checkpoints.forEach(cp => {
      ctx.beginPath(); ctx.arc(cp.x, cp.y, 6, 0, Math.PI * 2); ctx.fillStyle = "#ef4444"; ctx.fill();
      ctx.fillStyle = "#94a3b8"; ctx.font = "10px sans-serif"; ctx.fillText(cp.name, cp.x - 25, cp.y + 18);
    });

    let slowX = 40 + ((t * 80) % (w - 80));
    ctx.beginPath(); ctx.arc(slowX, 60 + Math.sin(slowX * 0.03) * 15, 5, 0, Math.PI * 2);
    ctx.fillStyle = "#f59e0b"; ctx.shadowBlur = 8; ctx.shadowColor = "#f59e0b"; ctx.fill();

    ctx.fillStyle = "var(--green)"; ctx.font = "11px sans-serif";
    ctx.fillText("IPVN7 AUTOPISTA PROPIA (Túnel directo P2P con blindaje cuántico)", 20, 134);
    ctx.strokeStyle = "rgba(16, 185, 129, 0.4)"; ctx.lineWidth = 4;
    ctx.beginPath(); ctx.moveTo(40, 165); ctx.lineTo(w - 40, 165); ctx.stroke();

    let fastX = 40 + ((t * 320) % (w - 80));
    ctx.beginPath(); ctx.arc(fastX, 165, 7, 0, Math.PI * 2);
    ctx.fillStyle = "#00e5ff"; ctx.shadowBlur = 15; ctx.shadowColor = "#00e5ff"; ctx.fill();

    ctx.beginPath(); ctx.arc(40, 165, 9, 0, Math.PI * 2); ctx.fillStyle = "#10b981"; ctx.fill(); ctx.fillText("Tu Dispositivo", 10, 190);
    ctx.beginPath(); ctx.arc(w - 40, 165, 9, 0, Math.PI * 2); ctx.fillStyle = "#10b981"; ctx.fill(); ctx.fillText("Destino Directo", w - 80, 190);
  }
  draw();
}

// Simulador de Ataque Cuántico
function runQuantumAttack() {
  const log = document.getElementById("quantumLog"), btn = document.getElementById("quantumBtn");
  btn.disabled = true; btn.innerText = "⚡ Computadora Cuántica Ejecutando Algoritmo de Shor...";
  log.innerHTML = "<span style='color:var(--amber)'>[INICIANDO ATAQUE CUÁNTICO] Simulando procesador de 10.000 Qubits lógicos...</span>";
  setTimeout(() => { log.innerHTML += "<br><span style='color:var(--red)'>❌ CIFRADO TRADICIONAL (RSA-2048 / ECC): Factorizado en 1.8 segundos. ¡CLAVES EXPUESTAS!</span>"; }, 1200);
  setTimeout(() => {
    log.innerHTML += "<br><span style='color:var(--green)'>🛡️ IPVN7 BLINDAJE CUÁNTICO (NIST ML-KEM-768): Estructura reticular intacta. Esfuerzo de ruptura: 2^192 operaciones. ¡INQUEBRANTABLE!</span>";
    btn.disabled = false; btn.innerText = "🧪 Repetir Simulación de Ataque Cuántico";
  }, 2600);
}

// Simulador de Malla Indestructible Kleinberg (Corte y Reconexión)
let meshCableCut = false;
function toggleMeshCut() {
  meshCableCut = !meshCableCut;
  const btn = document.getElementById("meshBtn"), log = document.getElementById("meshLog");
  if (meshCableCut) {
    btn.className = "btn-action green"; btn.innerText = "🔧 Restaurar Cable Principal";
    log.innerHTML = "<span style='color:var(--red)'>⚠️ Cable Principal ROTO.</span> Kleinberg O(log² N) reconectó tráfico por el Nodo B en <b>0.8 ms</b>. ¡Cero desconexión!";
  } else {
    btn.className = "btn-action red"; btn.innerText = "✂️ Cortar Cable Principal (Simular Caída)";
    log.innerHTML = "<span style='color:var(--green)'>✅ Todos los enlaces activos.</span> Tráfico fluyendo por la ruta física más directa.";
  }
}

function initMeshCanvas() {
  const canvas = document.getElementById("meshCanvas");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  let t = 0;

  function draw() {
    requestAnimationFrame(draw);
    t += 0.02;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const w = canvas.width, h = canvas.height;
    const nA = { x: 80, y: h / 2, name: "Tu Celular (A)" }, nB = { x: w / 2, y: 40, name: "Nodo Amigo (B)" }, nC = { x: w - 80, y: h / 2, name: "Tu PC Hogar (C)" };

    ctx.beginPath(); ctx.lineWidth = 3;
    if (meshCableCut) {
      ctx.strokeStyle = "rgba(239, 68, 68, 0.4)"; ctx.setLineDash([5, 5]);
      ctx.moveTo(nA.x, nA.y); ctx.lineTo(nC.x, nC.y); ctx.stroke(); ctx.setLineDash([]);
      ctx.fillStyle = "#ef4444"; ctx.font = "bold 16px sans-serif"; ctx.fillText("❌ ROTO", w / 2 - 25, h / 2 + 5);
    } else {
      ctx.strokeStyle = "rgba(16, 185, 129, 0.6)"; ctx.moveTo(nA.x, nA.y); ctx.lineTo(nC.x, nC.y); ctx.stroke();
      let pX = nA.x + ((t * 120) % (nC.x - nA.x));
      ctx.beginPath(); ctx.arc(pX, h / 2, 5, 0, Math.PI * 2); ctx.fillStyle = "#00e5ff"; ctx.fill();
    }

    ctx.beginPath(); ctx.lineWidth = meshCableCut ? 4 : 2;
    ctx.strokeStyle = meshCableCut ? "rgba(0, 229, 255, 0.8)" : "rgba(255, 255, 255, 0.15)";
    ctx.moveTo(nA.x, nA.y); ctx.lineTo(nB.x, nB.y); ctx.lineTo(nC.x, nC.y); ctx.stroke();

    if (meshCableCut) {
      let prog = (t * 1.5) % 2, curX = (prog < 1) ? nA.x + (nB.x - nA.x) * prog : nB.x + (nC.x - nB.x) * (prog - 1);
      let curY = (prog < 1) ? nA.y + (nB.y - nA.y) * prog : nB.y + (nC.y - nB.y) * (prog - 1);
      ctx.beginPath(); ctx.arc(curX, curY, 6, 0, Math.PI * 2); ctx.fillStyle = "#00e5ff"; ctx.fill();
    }

    [nA, nB, nC].forEach((n, idx) => {
      ctx.beginPath(); ctx.arc(n.x, n.y, 10, 0, Math.PI * 2);
      ctx.fillStyle = idx === 1 ? "#f59e0b" : "#10b981"; ctx.shadowBlur = 10; ctx.shadowColor = ctx.fillStyle; ctx.fill();
      ctx.fillStyle = "#f8fafc"; ctx.font = "12px sans-serif"; ctx.fillText(n.name, n.x - 35, n.y + (idx === 1 ? -18 : 26));
    });
  }
  draw();
}

// Filtro de Tarjetas de Ámbitos
function selectSphere(sphereId) {
  document.querySelectorAll(".sphere-card").forEach(c => c.classList.remove("active"));
  const selected = document.getElementById("sphere-" + sphereId);
  if (selected) selected.classList.add("active");

  const details = {
    personal: "📱 <b>En tu vida:</b> Comparte videos pesados y fotos en calidad original entre tu celular y tu PC al instante. Sin Google One ni iCloud. <br><button class='btn-more' onclick=\"showExplanation('p2p')\">+ Saber más: ¿Cómo viaja la luz de tú a tú?</button>",
    iot: "🖨️ <b>En tus dispositivos:</b> Tu Smart TV, impresora o cámara IP obtienen una identidad soberana protegida (Shadow DID). <br><button class='btn-more' onclick=\"showExplanation('shadow')\">+ Saber más: ¿Cómo protegen los Shadow DIDs?</button>",
    work: "💻 <b>En tu trabajo:</b> Entra a tu equipo de trabajo desde cualquier lugar con 1 clic, atravesando CGNAT y cortafuegos. <br><button class='btn-more' onclick=\"showExplanation('routers')\">+ Saber más: ¿Qué ve y qué no ve el router?</button>",
    community: "🤝 <b>En tu comunidad:</b> Si tu internet anda lento, navegas por la ruta limpia de tu vecino con 1 clic. <br><button class='btn-more' onclick=\"showExplanation('egress')\">+ Saber más: ¿Cómo funciona la Salida Soberana?</button>"
  };
  const box = document.getElementById("sphereDetailBox");
  if (box && details[sphereId]) box.innerHTML = details[sphereId];
}

// Manejador de Pestañas Técnicas
function switchTechTab(tabId) {
  document.querySelectorAll(".tech-tab-btn").forEach(b => b.classList.remove("active"));
  document.querySelectorAll(".tech-panel").forEach(p => p.classList.remove("active"));
  const btn = document.getElementById("tab-btn-" + tabId);
  const panel = document.getElementById("tech-" + tabId);
  if (btn) btn.classList.add("active");
  if (panel) panel.classList.add("active");
}

// Manejador de Módulos Expansibles
function toggleAccordion(id) {
  const item = document.getElementById("acc-" + id);
  if (item) item.classList.toggle("open");
}

// Manejador del Modal de Explicaciones Profundas Tri-Nivel (+ Más)
let currentExpKey = "p2p";
let currentExpLevel = "simple";

function setExpLevel(level) {
  currentExpLevel = level;
  document.querySelectorAll(".level-btn").forEach(btn => btn.classList.remove("active"));
  const activeBtn = document.getElementById("lvl-" + level);
  if (activeBtn) activeBtn.classList.add("active");
  renderExpBody();
}

function renderExpBody() {
  const data = (typeof EXPLANATIONS !== "undefined") ? EXPLANATIONS[currentExpKey] : null;
  const body = document.getElementById("modalBody");
  if (!data || !body) return;
  body.innerHTML = data[currentExpLevel] || data.simple || "<p>Información no disponible.</p>";
}

function showExplanation(key, preferredLevel) {
  const modal = document.getElementById("explanationModal");
  currentExpKey = key || "p2p";
  if (preferredLevel) currentExpLevel = preferredLevel;
  const data = (typeof EXPLANATIONS !== "undefined") ? EXPLANATIONS[currentExpKey] : null;
  if (!modal || !data) return;

  document.getElementById("modalTitle").innerText = data.title;
  document.getElementById("modalSubtitle").innerText = data.subtitle;
  setExpLevel(currentExpLevel);
  modal.classList.add("open");
}

function closeExplanation(e) {
  const modal = document.getElementById("explanationModal");
  if (!modal) return;
  if (!e || e.target === modal || e.target.classList.contains("modal-close")) {
    modal.classList.remove("open");
  }
}



