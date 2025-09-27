package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/repository"
)

type Handler struct {
	RMQ   *repository.RabbitMQ
	DB    *repository.Storage
	Redis *redis.Client
}

func NewHandler(db *repository.Storage, rbmq *repository.RabbitMQ, rs *redis.Client) Handler {
	return Handler{DB: db, RMQ: rbmq, Redis: rs}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zlog.Logger.Info().Msgf("%s %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func CacheMiddleware(next http.Handler, rs *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			path := r.URL.Path
			prefix := "/notify/"

			if !strings.HasPrefix(path, prefix) {
				writeErrorInJSON(w, "not found", http.StatusNotFound)
				return
			}

			nUid := strings.TrimPrefix(path, prefix)
			if nUid == "" {
				writeErrorInJSON(w, "not specified id", http.StatusBadRequest)
				return
			}

			val, err := rs.Get(r.Context(), nUid)
			if err == nil {
				_ = json.NewEncoder(w).Encode(map[string]string{"status": val})
				zlog.Logger.Info().Msg("Get notification from cache")
				return
			}

		}
		next.ServeHTTP(w, r)
	})
}

func writeErrorInJSON(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
}
