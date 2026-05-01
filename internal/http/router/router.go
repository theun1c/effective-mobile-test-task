package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/theun1c/effective-mobile-test-task/internal/http/handler"
)

type healthResponse struct {
	Status string `json:"status"`
}

func New(logger *slog.Logger, subscriptionService handler.SubscriptionService) http.Handler {
	mux := http.NewServeMux()

	subscriptionHandler := handler.NewSubscriptionHandlerWithLogger(subscriptionService, logger)

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

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		observer := &statusObserver{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(observer, r)
		logger.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status_code", observer.statusCode,
			"duration", time.Since(startedAt),
		)
	})
}

type statusObserver struct {
	http.ResponseWriter
	statusCode int
}

func (o *statusObserver) WriteHeader(statusCode int) {
	o.statusCode = statusCode
	o.ResponseWriter.WriteHeader(statusCode)
}
