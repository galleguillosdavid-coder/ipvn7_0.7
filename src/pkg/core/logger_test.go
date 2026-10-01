package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger_LifecycleAndPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_ipvn7.log")

	err := InitLogger(logPath, true)
	if err != nil {
		t.Fatalf("InitLogger fallo: %v", err)
	}

	LogInfo("Prueba informativa %s", "OK")
	LogWarn("Prueba advertencia %d", 123)
	LogError("Prueba error %s", "CRIT")
	LogDebug("Prueba debug %s", "VERBOSE")

	CloseLogger()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("error leyendo log: %v", err)
	}
	str := string(content)

	if !strings.Contains(str, "Prueba informativa OK") {
		t.Errorf("log no contiene info: %s", str)
	}
	if !strings.Contains(str, "Prueba advertencia 123") {
		t.Errorf("log no contiene warn: %s", str)
	}
	if !strings.Contains(str, "Prueba error CRIT") {
		t.Errorf("log no contiene error: %s", str)
	}
	if !strings.Contains(str, "Prueba debug VERBOSE") {
		t.Errorf("log no contiene debug: %s", str)
	}
}
