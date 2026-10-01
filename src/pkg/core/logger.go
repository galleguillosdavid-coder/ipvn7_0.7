package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logMu       sync.Mutex
	logFile     *os.File
	multiWriter io.Writer = os.Stdout
	debugActive bool
)

// InitLogger inicializa el sistema de logging persistente dual (consola + archivo en disco)
func InitLogger(logPath string, debug bool) error {
	logMu.Lock()
	defer logMu.Unlock()

	debugActive = debug
	if logPath == "" {
		logPath = filepath.Join("data", "ipvn7.log")
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if logFile != nil {
		_ = logFile.Close()
	}
	logFile = f
	multiWriter = io.MultiWriter(os.Stdout, f)

	writeEntry("SYSTEM", fmt.Sprintf("=== Inicio de sesion ipvn7 (PID: %d, Debug: %v) ===", os.Getpid(), debug))
	return nil
}

// CloseLogger cierra limpiamente el archivo de log persistente
func CloseLogger() {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		writeEntry("SYSTEM", fmt.Sprintf("=== Cierre de sesion ipvn7 (PID: %d) ===", os.Getpid()))
		_ = logFile.Close()
		logFile = nil
		multiWriter = os.Stdout
	}
}

func writeEntry(level, msg string) {
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("[%s] [%-6s] %s\n", ts, level, msg)
	_, _ = fmt.Fprint(multiWriter, line)
}

// LogInfo registra un evento informativo
func LogInfo(format string, args ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	writeEntry("INFO", fmt.Sprintf(format, args...))
}

// LogWarn registra una advertencia
func LogWarn(format string, args ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	writeEntry("WARN", fmt.Sprintf(format, args...))
}

// LogError registra un error critico
func LogError(format string, args ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	writeEntry("ERROR", fmt.Sprintf(format, args...))
}

// LogDebug registra un detalle de depuracion si debugActive es verdadero
func LogDebug(format string, args ...any) {
	if !debugActive {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	writeEntry("DEBUG", fmt.Sprintf(format, args...))
}
