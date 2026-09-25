// ipvn7 Guided Tour Engine (app_tour.js) — Experiencia de Usuario Radical
// Onboarding guiado paso a paso sin jerga técnica (Axioma III: <= 400 líneas)

const TOUR_STEPS = [
  {
    title: '¡Bienvenido a tu Red Privada!',
    content: 'Esta es tu red soberana. Aquí puedes interconectar y controlar todos tus dispositivos sin servidores intermedios y con máxima privacidad.',
    targetId: 'brand-logo-area'
  },
  {
    title: 'Conexión Segura en 1 Clic',
    content: 'Pulsa este botón para activar la red en un instante. No necesitas configurar puertos ni tener permisos de administrador.',
    targetId: 'btn-vpn-hero'
  },
  {
    title: 'Pregunta o Da Órdenes en tu Idioma',
    content: 'Escribe aquí como si hablaras con un asistente: "suspender equipo", "apagar monitor", o "limpiar papelera".',
    targetId: 'form-human-intent'
  },
  {
    title: 'Funciones Separadas y Claras',
    content: 'Cada servicio (Pantalla/TV, Impresoras, Archivos, Encendido) tiene su propia pestaña amplia y dedicada.',
    targetId: 'tab-btn-tv'
  },
  {
    title: 'Tú Tienes el Control Total',
    content: 'Todo lo que no es el núcleo es un complemento. Cada uno cuenta con dos botones: Instalar/Desinstalar y Encender/Apagar.',
    targetId: 'tab-btn-tools'
  }
];

let currentTourIndex = 0;

function startUserTour() {
  currentTourIndex = 0;
  showTourStep(currentTourIndex);
}

function showTourStep(index) {
  if (index < 0 || index >= TOUR_STEPS.length) {
    endUserTour();
    return;
  }
  currentTourIndex = index;
  const step = TOUR_STEPS[index];

  let overlay = document.getElementById('tour-backdrop-overlay');
  if (!overlay) {
    overlay = document.createElement('div');
    overlay.id = 'tour-backdrop-overlay';
    overlay.className = 'tour-backdrop';
    document.body.appendChild(overlay);
  }
  overlay.style.display = 'block';

  let popover = document.getElementById('tour-step-popover');
  if (!popover) {
    popover = document.createElement('div');
    popover.id = 'tour-step-popover';
    popover.className = 'tour-popover glass-card';
    document.body.appendChild(popover);
  }

  popover.innerHTML = `
    <div class="tour-header">
      <span class="tour-badge">Paso ${index + 1} de ${TOUR_STEPS.length}</span>
      <button class="tour-close-btn" onclick="endUserTour()">✕</button>
    </div>
    <h3 class="tour-title">${step.title}</h3>
    <p class="tour-text">${step.content}</p>
    <div class="tour-footer">
      <button class="btn-action secondary" onclick="endUserTour()">Saltar</button>
      <div style="display:flex; gap:8px;">
        ${index > 0 ? `<button class="btn-action secondary" onclick="showTourStep(${index - 1})">Anterior</button>` : ''}
        <button class="btn-action primary" onclick="showTourStep(${index + 1})">${index === TOUR_STEPS.length - 1 ? '¡Listo!' : 'Siguiente'}</button>
      </div>
    </div>
  `;
  popover.style.display = 'block';

  // Highlight element if available
  const el = document.getElementById(step.targetId);
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    const rect = el.getBoundingClientRect();
    popover.style.position = 'fixed';
    popover.style.top = `${Math.min(window.innerHeight - 250, Math.max(80, rect.bottom + 12))}px`;
    popover.style.left = `${Math.max(20, Math.min(window.innerWidth - 380, rect.left))}px`;
  } else {
    popover.style.position = 'fixed';
    popover.style.top = '40%';
    popover.style.left = '50%';
    popover.style.transform = 'translate(-50%, -50%)';
  }
}

function endUserTour() {
  const overlay = document.getElementById('tour-backdrop-overlay');
  if (overlay) overlay.style.display = 'none';
  const popover = document.getElementById('tour-step-popover');
  if (popover) popover.style.display = 'none';
  localStorage.setItem('ipvn7_tour_completed', 'true');
}

// Auto-inicio en primera visita
window.addEventListener('DOMContentLoaded', () => {
  if (!localStorage.getItem('ipvn7_tour_completed')) {
    setTimeout(() => startUserTour(), 1500);
  }
});
