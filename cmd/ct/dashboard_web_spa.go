package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/web
var webAssets embed.FS

// spaHandler serves the React SPA. It serves static assets from /app/assets/
// and returns index.html for all other /app/ routes (client-side routing).
type spaHandler struct {
	indexHTML        []byte
	assetsFileServer http.Handler
}

func newSPAHandler() *spaHandler {
	webSub, err := fs.Sub(webAssets, "assets/web")
	if err != nil {
		panic("embedded web assets not found: " + err.Error())
	}

	// Read the SPA index.html for serving on all routes.
	idx, err := fs.ReadFile(webSub, "index.html")
	if err != nil {
		panic("embedded web index.html not found: " + err.Error())
	}

	assetsSub, err := fs.Sub(webSub, "assets")
	if err != nil {
		panic("embedded web/assets not found: " + err.Error())
	}

	return &spaHandler{
		indexHTML:        idx,
		assetsFileServer: http.StripPrefix("/app/assets/", http.FileServer(http.FS(assetsSub))),
	}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Serve static assets under /app/assets/
	if len(path) >= len("/app/assets/") && path[:len("/app/assets/")] == "/app/assets/" {
		h.assetsFileServer.ServeHTTP(w, r)
		return
	}

	// All other /app/ routes serve index.html for client-side routing.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(h.indexHTML) //nolint:errcheck
}
