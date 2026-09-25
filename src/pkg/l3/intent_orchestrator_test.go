package l3

import (
	"testing"
)

func TestParseNaturalLanguage(t *testing.T) {
	// 1. Energía
	p1 := ParseNaturalLanguage("Pon el PC de la casa en modo reposo")
	if p1.Category != CategoryHardwareOS || p1.Action != ActionPowerSleep {
		t.Fatalf("esperado sleep, obtenido: %+v", p1)
	}

	// 2. Bloqueo
	p2 := ParseNaturalLanguage("Bloquea el teclado y mouse de mi estación de trabajo")
	if p2.Category != CategoryHardwareOS || p2.Action != ActionPowerLock {
		t.Fatalf("esperado lock, obtenido: %+v", p2)
	}

	// 3. Monitor
	p3 := ParseNaturalLanguage("Comprueba si dejé el monitor encendido; apaga la pantalla")
	if p3.Category != CategoryHardwareOS || p3.Action != ActionMonitorOff {
		t.Fatalf("esperado monitor_off, obtenido: %+v", p3)
	}

	// 4. Papelera
	p4 := ParseNaturalLanguage("Abre descargas y limpia la papelera de reciclaje por completo")
	if p4.Category != CategoryHardwareOS || p4.Action != ActionCleanRecycleBin {
		t.Fatalf("esperado clean_recycle_bin, obtenido: %+v", p4)
	}

	// 5. Batería en notebook remoto
	p5 := ParseNaturalLanguage("Si la batería del notebook en casa baja de 20% avísame")
	if p5.Category != CategoryHardwareOS || p5.Action != ActionBatteryStatus || p5.TargetNode != "notebook" {
		t.Fatalf("esperado battery en notebook, obtenido: %+v", p5)
	}

	// 6. Dron
	p6 := ParseNaturalLanguage("Despegar dron de reconocimiento a 5 metros")
	if p6.Category != CategoryRobotDrone || p6.Action != "takeoff" {
		t.Fatalf("esperado takeoff en dron, obtenido: %+v", p6)
	}
}

