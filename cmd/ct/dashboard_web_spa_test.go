package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSPAHandler_ServesIndexHTML(t *testing.T) {
	handler := newSPAHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/app/")
	if err != nil {
		t.Fatalf("GET /app/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /app/: status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
}

func TestSPAHandler_ServesSubRoutes(t *testing.T) {
	handler := newSPAHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	for _, path := range []string{"/app/droplets", "/app/castellarius", "/app/doctor"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status = %d, want %d", path, resp.StatusCode, http.StatusOK)
		}

		ct := resp.Header.Get("Content-Type")
		if ct != "text/html; charset=utf-8" {
			t.Errorf("GET %s: Content-Type = %q, want %q", path, ct, "text/html; charset=utf-8")
		}
	}
}

func TestSPAHandler_RedirectsAppRoot(t *testing.T) {
	mux := http.NewServeMux()
	spa := newSPAHandler()
	mux.Handle("/app/", spa)
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusMovedPermanently)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(server.URL + "/app")
	if err != nil {
		t.Fatalf("GET /app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("GET /app: status = %d, want %d", resp.StatusCode, http.StatusMovedPermanently)
	}

	loc := resp.Header.Get("Location")
	if loc != "/app/" {
		t.Errorf("Location = %q, want %q", loc, "/app/")
	}
}
