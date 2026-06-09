package http

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"koji/internal/config"
)

func registerStaticRoutes(mux *http.ServeMux, cfg config.Config) error {
	if cfg.DevMode {
		return setupDevModeRouter(mux, cfg.DevProxyURL)
	}
	return setupProductionRouter(mux, cfg.StaticAssetDir)
}

func setupDevModeRouter(mux *http.ServeMux, devProxyURL string) error {
	viteURL, err := url.Parse(devProxyURL)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(viteURL)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	return nil
}

func setupProductionRouter(mux *http.ServeMux, staticAssetDir string) error {
	publicFS := os.DirFS(staticAssetDir)
	fileServer := http.FileServer(http.FS(publicFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		if !isSafeStaticRequestPath(r.URL.Path) {
			http.NotFound(w, r)
			return
		}

		if serveStaticFile(w, r, fileServer, publicFS) {
			return
		}

		serveProductionIndex(w, r, publicFS)
	})
	return nil
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, fileServer http.Handler, publicFS fs.FS) bool {
	staticPath := staticRequestPath(r.URL.Path)
	if staticPath == "" {
		staticPath = "index.html"
	}

	f, err := publicFS.Open(staticPath)
	if err != nil {
		return false
	}
	defer f.Close()

	if strings.HasPrefix(staticPath, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	fileServer.ServeHTTP(w, r)
	return true
}

func isSafeStaticRequestPath(requestPath string) bool {
	staticPath := staticRequestPath(requestPath)
	if staticPath == "" {
		return true
	}
	if strings.Contains(staticPath, "\\") {
		return false
	}
	if strings.Contains(staticPath, "\x00") {
		return false
	}
	return fs.ValidPath(staticPath) && path.Clean(staticPath) == staticPath
}

func staticRequestPath(requestPath string) string {
	return strings.TrimPrefix(requestPath, "/")
}

func serveProductionIndex(w http.ResponseWriter, r *http.Request, publicFS fs.FS) {
	index, err := publicFS.Open("index.html")
	if err != nil {
		http.Error(w, "Frontend assets not found", http.StatusInternalServerError)
		return
	}
	defer index.Close()

	stat, err := index.Stat()
	if err != nil {
		http.Error(w, "Frontend assets not found", http.StatusInternalServerError)
		return
	}

	body, err := fs.ReadFile(publicFS, "index.html")
	if err != nil {
		http.Error(w, "Frontend assets not found", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "index.html", stat.ModTime(), bytes.NewReader(body))
}
