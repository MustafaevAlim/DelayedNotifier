package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/wb-go/wbf/zlog"
)

func (h *Handler) GetOrDelNotification(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
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

	if r.Method == http.MethodGet {

		n, err := h.DB.GetNotificationByUUID(r.Context(), nUid)
		if err != nil {
			zlog.Logger.Err(err)
			if errors.Is(err, sql.ErrNoRows) {
				writeErrorInJSON(w, "not found notifications with this id", http.StatusNotFound)
				return
			}

			writeErrorInJSON(w, "failed get notification", http.StatusInternalServerError)
			return
		}

		err = h.Redis.Set(r.Context(), n.Uid, n.Status)
		if err != nil {
			zlog.Logger.Err(err)
		} else {
			zlog.Logger.Info().Msgf("Write notification %s in cache", n.Uid)

		}

		err = json.NewEncoder(w).Encode(map[string]string{"status": n.Status})
		if err != nil {
			zlog.Logger.Err(err)
			writeErrorInJSON(w, "failed encode", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodDelete {
		err := h.DB.UpdateStatusNotification(r.Context(), nUid, "canceled")
		if err != nil {
			zlog.Logger.Err(err)
			writeErrorInJSON(w, "failed canceled notification", http.StatusInternalServerError)
			return
		}
	}

}
