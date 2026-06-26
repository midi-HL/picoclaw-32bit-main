package api

import (
	"net/http"
	"net/http/httputil"
)

func (h *Handler) registerChatRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/chat", h.handleChatProxy)
	mux.HandleFunc("/api/chat/stream", h.handleChatProxy)
}

// handleChatProxy reverse-proxies /api/chat and /api/chat/stream to the gateway.
func (h *Handler) handleChatProxy(w http.ResponseWriter, r *http.Request) {
	if !h.gatewayAvailableForProxy() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"gateway not available"}`))
		return
	}

	target := h.gatewayProxyURL()
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Scheme = target.Scheme
			req.Host = target.Host
		},
	}
	proxy.ServeHTTP(w, r)
}
