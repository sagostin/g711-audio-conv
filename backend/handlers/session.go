package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// SessionResponse is returned by the session endpoint.
// It never exposes the IP list — only the boolean result.
type SessionResponse struct {
	EasterEgg bool `json:"easterEgg"`
}

// easterEggIPs is loaded once from the EASTER_EGG_IPS environment variable.
// Format: comma-separated IPs, e.g. "192.168.1.100,10.0.0.5"
var easterEggIPs = loadEasterEggIPs()

func loadEasterEggIPs() map[string]bool {
	raw := os.Getenv("EASTER_EGG_IPS")
	if raw == "" {
		return nil
	}
	ips := make(map[string]bool)
	for _, ip := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(ip)
		if trimmed != "" {
			ips[trimmed] = true
		}
	}
	return ips
}

// SessionHandler returns session-level flags based on the client's IP.
// GET /api/session
func SessionHandler(proxyHeader string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIPFromRequest(r, proxyHeader)
		resp := SessionResponse{
			EasterEgg: easterEggIPs[clientIP],
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// getClientIPFromRequest extracts the client IP, checking the proxy header first.
func getClientIPFromRequest(r *http.Request, proxyHeader string) string {
	if proxyHeader != "" {
		if forwarded := r.Header.Get(proxyHeader); forwarded != "" {
			parts := strings.Split(forwarded, ",")
			return strings.TrimSpace(parts[0])
		}
	}
	addr := r.RemoteAddr
	// Strip port
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
