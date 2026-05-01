package router

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/theun1c/effective-mobile-test-task/internal/http/handler"
)

type healthResponse struct {
	Status string `json:"status"`
}

func New(logger *log.Logger, subscriptionService handler.SubscriptionService) http.Handler {
	mux := http.NewServeMux()

	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
	})

	mux.HandleFunc("POST /subscriptions", subscriptionHandler.Create)
	mux.HandleFunc("GET /subscriptions", subscriptionHandler.List)
	mux.HandleFunc("GET /subscriptions/{id}", subscriptionHandler.GetByID)
	mux.HandleFunc("PUT /subscriptions/{id}", subscriptionHandler.Update)
	mux.HandleFunc("DELETE /subscriptions/{id}", subscriptionHandler.Delete)

	return requestLogger(logger, mux)
}

func requestLogger(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("request method=%s path=%s remote_addr=%s duration=%s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(startedAt))
	})
}
