package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed assets/web
var webAssets embed.FS

// sanitizeCSPHost strips any characters from host that are not valid in
// a CSP directive value (alphanumeric, dots, hyphens, colons for port).
// This prevents CSP injection via a crafted Host header.
func sanitizeCSPHost(host string) string {
	var b strings.Builder
	for _, r := range host {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == ':' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// spaHandler serves the React SPA. It serves static assets from /app/assets/
// and returns index.html for all other /app/ routes (client-side routing).
type spaHandler struct {
	indexHTML        []byte
	assetsFileServer http.Handler
}

func newSPAHandler(apiKey string) *spaHandler {
	webSub, err := fs.Sub(webAssets, "assets/web")
	if err != nil {
		panic("embedded web assets not found: " + err.Error())
	}

	// Read the SPA index.html for serving on all routes.
	idx, err := fs.ReadFile(webSub, "index.html")
	if err != nil {
		panic("embedded web index.html not found: " + err.Error())
	}

	// If an API key is configured, inject a meta tag so the frontend knows
	// authentication is required. This must happen at build-time since the
	// embedded FS is read-only. Validate that the injection succeeded;
	// strings.Replace returns the original string unchanged if </head> is not
	// found, which would leave auth-required deployments unprotected.
	if apiKey != "" {
		authMeta := `<meta name="cistern-auth" content="required" />`
		injected := strings.Replace(string(idx), "</head>", authMeta+"\n  </head>", 1)
		if injected == string(idx) {
			panic("dashboard SPA: failed to inject auth meta tag into index.html — </head> tag not found; auth-required deployments would be unprotected")
		}
		idx = []byte(injected)
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
		w.Header().Set("X-Content-Type-Options", "nosniff")
		h.assetsFileServer.ServeHTTP(w, r)
		return
	}

	// All other /app/ routes serve index.html for client-side routing.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// Restrict connect-src to same-origin WebSocket only (no ws:/wss: wildcards).
	// Build the ws/wss hosts dynamically from the request Host header.
	// Sanitize: only allow alphanumeric, dots, hyphens, colons (port).
	wsHost := sanitizeCSPHost(r.Host)
	connectSrc := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' ws://" + wsHost + " wss://" + wsHost + "; img-src 'self'; font-src 'self'"
	w.Header().Set("Content-Security-Policy", connectSrc)
	w.Write(h.indexHTML) //nolint:errcheck
}
