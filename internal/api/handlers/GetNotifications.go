package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/zlog"
)

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorInJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	args := r.URL.Query()

	page, err := strconv.Atoi(args.Get("page"))
	if err != nil {
		writeErrorInJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	pageSize, err := strconv.Atoi(args.Get("size"))
	if err != nil {
		writeErrorInJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, err := h.DB.GetTotalNotifications(r.Context())
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := h.DB.GetNotifications(r.Context(), page, pageSize)
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respMap := map[string]any{
		"total":         total,
		"notifications": n,
	}

	err = json.NewEncoder(w).Encode(respMap)
	if err != nil {
		writeErrorInJSON(w, err.Error(), http.StatusInternalServerError)
	}

}
