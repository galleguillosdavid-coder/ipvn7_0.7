// ipvn7 Universal Core Dashboard v0.7.0 — Orquestador Principal (app.js)
// Punto de Entrada Reactivo y Ciclo de Vida (Axioma III: <= 400 líneas)

document.addEventListener('DOMContentLoaded', () => {
  initCopyButtons();
  initModal();
  initDatagramForm();
  initHomeHubActions();
  startCorePolling();
  initRadarAnimation();
  initSSE();
  initTabs();
  initChatApp();
  initPrintJobs();
  initVPNButton();
});
