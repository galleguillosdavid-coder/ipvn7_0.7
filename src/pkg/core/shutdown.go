package core

import (
	"context"
	"time"
)

// GracefulShutdown ejecuta una función de cleanup con un timeout determinista
// Unifica múltiples time.Sleep dispersos en un solo patrón coordinado
func GracefulShutdown(fn func(), timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		if fn != nil {
			fn()
		}
		close(done)
	}()

	select {
	case <-done:
		// Cleanup completado exitosamente
	case <-ctx.Done():
		// Timeout forzado - continuar con shutdown de todos modos
	}
}
