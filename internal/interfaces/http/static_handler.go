package http

import (
	nethttp "net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http/middleware"
)

// StaticHandler раздаёт SPA из каталога web и отдаёт index.html как fallback.
type StaticHandler struct {
	dir string
}

// NewStaticHandler создаёт StaticHandler.
func NewStaticHandler(dir string) *StaticHandler {
	return &StaticHandler{dir: filepath.Clean(dir)}
}

// ServeHTTP обслуживает статические файлы SPA.
func (h *StaticHandler) ServeHTTP(w nethttp.ResponseWriter, r *nethttp.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
		middleware.WriteError(w, nethttp.StatusNotFound, "not_found", "endpoint not found")
		return
	}
	if r.Method != nethttp.MethodGet && r.Method != nethttp.MethodHead {
		middleware.WriteError(w, nethttp.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	path := filepath.Join(h.dir, r.URL.Path)
	if !strings.HasPrefix(path, h.dir) || strings.Contains(r.URL.Path, "..") {
		middleware.WriteError(w, nethttp.StatusForbidden, "forbidden", "forbidden path")
		return
	}

	//nolint:gosec // путь ограничен каталогом h.dir (filepath.Join + HasPrefix + проверка "..")
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		nethttp.ServeFile(w, r, path)
		return
	}

	index := filepath.Join(h.dir, "index.html")
	if info, err := os.Stat(index); err != nil || info.IsDir() {
		middleware.WriteError(w, nethttp.StatusNotFound, "not_found", "not found")
		return
	}
	nethttp.ServeFile(w, r, index)
}
