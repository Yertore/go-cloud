package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type rootResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

func main() {
	//после установки godotenv - go get github.com/joho/godotenv
	//загрузка .env
	//err := godotenv.Load()
	/*if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}*/

	// ---------------- mux — это http.Handler --------------------
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler)
	mux.HandleFunc("/", rootHandler)

	// Middleware
	handler := loggingMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Server instance (нужна, чтобы уметь делать Shutdown)
	//srv.ListenAndServe() — старт
	//srv.Shutdown(ctx) — мягкая остановка
	//srv.Close() — жёсткая остановка
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Запуск сервера в горутине, чтобы main мог ждать сигнал
	go func() {
		log.Printf("starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Ждём SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	//Это как подписка.
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received")

	// Даём 5 секунд на завершение текущих запросов
	//context.Background() — базовый пустой контекст
	//WithTimeout создаёт новый контекст, который сам отменится через 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//освобождает внутренние таймеры/ресурсы контекста
	//хорошая практика: всегда вызывать cancel, даже если таймаут сам сработает
	defer cancel()

	//srv.Shutdown(ctx) делает “мягкое завершение”:
	//перестаёт принимать новые соединения
	//закрывает “listeners”
	//ждёт пока текущие запросы закончатся
	//если ctx истёк (5 секунд прошли) — возвращает ошибку
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		//Если Shutdown не успел (таймаут) или завис, ты делаешь “force close”:
		//немедленно закрывает соединения
		//текущие запросы могут оборваться
		//Это “план Б”.
		_ = srv.Close() // жёстко закрыть, если не успели
	}

	log.Println("server stopped")

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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//strings.EqualFold делает сравнение без учета регистра (True/TRUE/true).
	ready := strings.EqualFold(os.Getenv("APP_READY"), "true")
	///ready := os.Getenv("APP_READY") == "true"
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

		log.Printf("method=%s path=%s latency=%s", r.Method, r.URL.Path, latency)
	})
}
