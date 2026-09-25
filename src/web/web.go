package web

import (
	"embed"
	"net/http"
)

// Assets contiene todos los archivos estáticos de la interfaz web embebidos directamente en el binario autónomo de Go
//
//go:embed index.html *.css *.js *.wasm
var Assets embed.FS

// GetFS retorna el sistema de archivos embebido listo para http.FileServer
func GetFS() http.FileSystem {
	return http.FS(Assets)
}

