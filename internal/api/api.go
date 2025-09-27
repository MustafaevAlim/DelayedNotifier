package api

import (
	"net/http"

	"DelayedNotifier/internal/api/handlers"
)

func Routes(h *handlers.Handler) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/notify/", h.CreateNotification)
	m.Handle("/notify/{id}", handlers.CacheMiddleware(http.HandlerFunc(h.GetOrDelNotification), h.Redis))
	m.HandleFunc("/notifications", h.GetNotifications)

	fs := http.FileServer(http.Dir("web"))
	m.Handle("/", http.StripPrefix("/", fs))

	loggingMux := handlers.LoggingMiddleware(m)
	return loggingMux
}
