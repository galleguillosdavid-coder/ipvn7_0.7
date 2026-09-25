package web

import (
	"io"
	"testing"
)

func TestEmbeddedAssets(t *testing.T) {
	files := []string{
		"index.html", "style.css", "app.js",
		"app_core.js", "app_net.js", "app_homehub.js",
		"app_chat.js", "app_vpn_button.js", "app_intent.js", "app_files_spooler.js",
		"app_plugins.js", "app_tour.js",
		"style_base.css", "style_dashboard.css", "style_dispatcher.css",
		"style_homehub.css", "style_chat.css", "style_hero_button.css",
		"style_plugins.css", "style_tour.css",
	}
	for _, filename := range files {
		f, err := Assets.Open(filename)
		if err != nil {
			t.Fatalf("No se pudo abrir el archivo embebido '%s': %v", filename, err)
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			t.Fatalf("Error leyendo el archivo embebido '%s': %v", filename, err)
		}
		if len(data) == 0 {
			t.Fatalf("El archivo embebido '%s' está vacío", filename)
		}
	}

	fs := GetFS()
	if fs == nil {
		t.Fatal("GetFS() retornó nil")
	}
}
