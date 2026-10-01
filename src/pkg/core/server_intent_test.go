package core

import (
	"testing"
)

// TestServer_IntentAPI deshabilitado temporalmente
// El servidor HTTP CoreServer fue simplificado/archivado en DEC-107
// Este test depende de APIs del servidor HTTP que no existen en el núcleo mínimo actual
func TestServer_IntentAPI(t *testing.T) {
	t.Skip("CoreServer HTTP API simplified in DEC-107 - test disabled until restoration")
}
