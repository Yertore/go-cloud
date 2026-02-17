package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type rootResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

func main() {
	//после установки godotenv - go get github.com/joho/godotenv
	//загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	// ---------------- mux — это http.Handler --------------------
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler)
	mux.HandleFunc("/", rootHandler)

	// Middleware
	middlewareHandler := loggingMiddleware(mux)

	log.Println("starting on :8080")
	if err := http.ListenAndServe(":8080", middlewareHandler); err != nil {
		log.Fatal(err)
	}

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := rootResponse{
		Service: "go-cloud",
		Version: "1.0.0",
	}
	_ = json.NewEncoder(w).Encode(resp)

	//можно так
	//json.NewEncoder(w).Encode(resp)
	//можно и так правилнее
	//if err := json.NewEncoder(w).Encode(resp); err != nil {
	//	log.Printf("encode error: %v", err)
	//}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	//статус - WriteHeader(code) выставляет HTTP статус-код в ответе.
	w.WriteHeader(http.StatusOK) //http.StatusOK = 200

	//тело
	//Write пишет тело ответа (response body).
	//HTTP тело — это байты.
	//строку "ok" нужно превратить в байты: []byte("ok").
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Println("write failed:", err)
	}
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {

	time.Sleep(5 * time.Millisecond)
	//strings.EqualFold делает сравнение без учета регистра (True/TRUE/true).
	//ready := strings.EqualFold(os.Getenv("APP_READY"), "true")
	ready := os.Getenv("APP_READY") == "true"
	if !ready {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Println("write failed:", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		latency := time.Since(start)

		log.Printf("method=%s path=%s latency=%dns", r.Method, r.URL.Path, latency.Microseconds())
	})
}
