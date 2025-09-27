package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"DelayedNotifier/internal/model"

	"github.com/wb-go/wbf/zlog"
)

func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorInJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var n model.CreateNotification
	var err error
	n.Text = r.FormValue("text")
	n.DateTime, err = time.Parse(time.RFC3339, r.FormValue("datetime"))
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, "error in parse time", http.StatusInternalServerError)
		return
	}

	chatId, err := strconv.Atoi(r.FormValue("tg_chatid"))
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, "error in parse chatid", http.StatusInternalServerError)
		return
	}

	n.TgChatId = int64(chatId)

	nRepo, err := h.DB.CreateNotification(r.Context(), n)
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, "failed to create record in DB", http.StatusInternalServerError)
		return
	}

	err = h.Redis.Set(r.Context(), nRepo.Uid, nRepo.Status)
	if err != nil {
		zlog.Logger.Err(err)
	} else {
		zlog.Logger.Info().Msgf("Write notification %s in cache", nRepo.Uid)

	}

	err = h.RMQ.PublishNotification(n, nRepo.Uid)
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, "failed to publish in queue", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{"result": nRepo.Uid})
	if err != nil {
		zlog.Logger.Err(err)
		writeErrorInJSON(w, "failed to publish in queue", http.StatusInternalServerError)
		return
	}

}
