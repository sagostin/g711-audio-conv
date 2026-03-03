package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"tops-audio-conv/handlers"
	"tops-audio-conv/middleware"
)

func main() {
	// Configuration from environment
	port := getEnv("PORT", "8080")
	proxyHeader := getEnv("PROXY_HEADER", "X-Forwarded-For")
	maxUploadMB := getEnvInt("MAX_UPLOAD_MB", 200)
	staticDir := getEnv("STATIC_DIR", "./static")
	conversionsDir := getEnv("CONVERSIONS_DIR", "./conversions")

	// Ensure conversions directory exists
	os.MkdirAll(conversionsDir, 0755)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", handlers.HealthHandler)
	mux.HandleFunc("/api/convert", handlers.ConvertHandler(int64(maxUploadMB)))
	mux.HandleFunc("/api/convert/bulk", handlers.BulkConvertHandler(int64(maxUploadMB)))
	mux.HandleFunc("/api/formats", handlers.FormatsHandler)
	mux.HandleFunc("/api/prefixes", handlers.PrefixesHandler)
	mux.HandleFunc("/api/session", handlers.SessionHandler(proxyHeader))

	// Serve Vue static files in production
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Serve static file if it exists, otherwise serve index.html (SPA routing)
			path := filepath.Join(staticDir, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		})
		log.Printf("Serving static files from %s", staticDir)
	}

	// Apply middleware
	handler := middleware.LoggingMiddleware(proxyHeader)(corsMiddleware(mux))

	log.Printf("Starting Audio Converter on :%s", port)
	log.Printf("Proxy header: %s", proxyHeader)
	log.Printf("Max upload size: %d MB", maxUploadMB)
	log.Printf("Conversions dir: %s", conversionsDir)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, X-File-Type, X-Normalization-DB, X-Job-ID, X-Input-Loudness, X-Input-Peak, X-Input-LRA, X-Output-Loudness, X-Output-Peak, X-Output-LRA")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[audio-converter] ")
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║      Audio Converter Server          ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
}
