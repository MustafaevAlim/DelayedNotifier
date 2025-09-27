package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/model"
)

type Storage struct {
	DB *dbpg.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := dbpg.New(dsn, nil, &dbpg.Options{
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Minute,
	})
	if err != nil {
		zlog.Logger.Error().Msgf("failed get new storage: %v\n", err)
		return nil, err
	}
	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Master.Close()
}

func (s *Storage) CreateNotification(ctx context.Context, n model.CreateNotification) (*model.NotificationInRepo, error) {
	var nInRepo model.NotificationInRepo
	nInRepo.DateTime = n.DateTime
	nInRepo.Text = n.Text
	nInRepo.Status = model.StatusWait
	nInRepo.TgChatId = n.TgChatId
	nInRepo.Uid = uuid.NewString()

	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO notifications (uuid, text_message, publish_time, status_publish, tg_chatid, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nInRepo.Uid, nInRepo.Text, nInRepo.DateTime, nInRepo.Status, nInRepo.TgChatId, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}

	zlog.Logger.Info().Msgf("Create record in DB: %s", nInRepo.Uid)

	return &nInRepo, nil

}

func (s *Storage) GetTotalNotifications(ctx context.Context) (int, error) {

	res, err := s.DB.QueryContext(ctx, "SELECT COUNT(*) FROM notifications")
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Err(err)
		}
	}()

	res.Next()
	var n int
	err = res.Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Storage) GetNotifications(ctx context.Context, page, pageSize int) ([]model.NotificationInResponse, error) {

	query := `SELECT uuid, text_message, status_publish, publish_time
			  FROM notifications
			  ORDER BY uuid DESC
			  LIMIT $1 OFFSET $2 `
	res, err := s.DB.QueryContext(ctx, query, pageSize, pageSize*(page-1))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Err(err)
		}
	}()

	nRes := make([]model.NotificationInResponse, 0)

	for res.Next() {
		n := model.NotificationInResponse{}
		err := res.Scan(&n.Uid, &n.Text, &n.Status, &n.DateTime)
		if err != nil {
			return nil, err
		}
		nRes = append(nRes, n)
	}

	return nRes, nil
}

func (s *Storage) UpdateStatusNotification(ctx context.Context, uid string, newStatus string) error {

	res, err := s.DB.ExecContext(ctx,
		`UPDATE notifications SET status_publish=$1, updated_at=$2 WHERE uuid=$3`, newStatus, time.Now(), uid)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (s *Storage) GetNotificationByUUID(ctx context.Context, uid string) (*model.NotificationInRepo, error) {
	var n model.NotificationInRepo
	res, err := s.DB.QueryContext(ctx,
		`SELECT uuid, text_message, publish_time, status_publish, tg_chatid
		  FROM notifications WHERE uuid=$1`, uid)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		err = res.Scan(&n.Uid, &n.Text, &n.DateTime, &n.Status, &n.TgChatId)
		if err != nil {
			return nil, err
		}
	}

	n.DateTime = n.DateTime.In(time.Local)
	return &n, nil

}
