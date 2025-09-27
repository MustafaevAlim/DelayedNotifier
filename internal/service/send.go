package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
)

const maxRetry = 5

type RetryMessage struct {
	uid      string
	attempts int
	msg      tgbotapi.MessageConfig
}

func SendMessage(ctx context.Context, retryQueue chan RetryMessage, msgChan chan []byte, db *repository.Storage) {
	for msg := range msgChan {
		var msgMap map[string]string
		err := json.Unmarshal(msg, &msgMap)
		if err != nil {
			zlog.Logger.Err(err)
			continue
		}

		n, err := db.GetNotificationByUUID(ctx, msgMap["uuid"])
		if err != nil {
			zlog.Logger.Err(err)
			continue
		}

		if n.Status == model.StatusCanceled {
			zlog.Logger.Info().Msgf("Notification %s canceled", n.Uid)
			continue
		}

		msgConfig := tgbotapi.NewMessage(n.TgChatId, n.Text)
		retryQueue <- RetryMessage{
			msg:      msgConfig,
			attempts: 0,
			uid:      n.Uid,
		}

	}
	close(retryQueue)
}

func RetryWorker(ctx context.Context, retryQueue chan RetryMessage, db *repository.Storage, tgbot *TelegramBot) {
	for rm := range retryQueue {
		if tgbot == nil {
			log.Println(rm.msg.Text)
			continue
		}
		err := tgbot.Send(rm.msg.ChatID, rm.msg.Text)
		if err != nil {
			rm.attempts++

			if rm.attempts < maxRetry {
				go func(r RetryMessage) {
					select {
					case <-time.After(time.Second * 3 * time.Duration(r.attempts)):
						select {
						case <-ctx.Done():
							return
						case retryQueue <- r:
						}
					case <-ctx.Done():
						return
					}
				}(rm)
			} else {

				zlog.Logger.Warn().Msgf("Failed to send message after %d attempts: %v", rm.attempts, err)
				err = db.UpdateStatusNotification(ctx, rm.uid, model.StatusNotDelivered)
				if err != nil {
					zlog.Logger.Err(err)
				}
			}

		} else {

			err = db.UpdateStatusNotification(ctx, rm.uid, model.StatusSent)
			if err != nil {
				zlog.Logger.Err(err)
			}
		}
	}
}
