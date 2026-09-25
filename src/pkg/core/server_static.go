package core

import (
	"net/http"
	"os"
	"path/filepath"

	"ipvn7/web"
)

// FindStaticDir localiza el directorio web relativo al ejecutable o al directorio de trabajo
func FindStaticDir() string {
	execDir := ""
	if execPath, err := os.Executable(); err == nil {
		execDir = filepath.Dir(execPath)
	}

	candidates := []string{
		"web",
		filepath.Join("..", "web"),
		filepath.Join(".", "web"),
	}
	if execDir != "" {
		candidates = append(candidates,
			filepath.Join(execDir, "web"),
			filepath.Join(execDir, "..", "web"),
		)
	}

	for _, c := range candidates {
		if fi, err := os.Stat(filepath.Join(c, "index.html")); err == nil && !fi.IsDir() {
			return c
		}
	}
	return ""
}

// getStaticFileSystem retorna el sistema de archivos para los recursos estáticos (disco o embebido)
func (s *CoreServer) getStaticFileSystem() http.FileSystem {
	if s.staticDir != "" {
		if fi, err := os.Stat(s.staticDir); err == nil && fi.IsDir() {
			return http.Dir(s.staticDir)
		}
	}
	return web.GetFS()
}
