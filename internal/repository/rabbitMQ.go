package repository

import (
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/model"
)

// почему-то в wbf не реализованы методы закрытия соединений

type RabbitMQ struct {
	Publisher    *rabbitmq.Publisher
	Consumer     *rabbitmq.Consumer
	Exchange     *rabbitmq.Exchange
	QueueManager *rabbitmq.QueueManager
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := rabbitmq.Connect(url, 5, 5*time.Second)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	var r RabbitMQ
	r.Exchange = rabbitmq.NewExchange("delayed_exchange", "x-delayed-message")
	r.Exchange.Durable = true
	args := amqp091.Table{
		"x-delayed-type": "direct",
	}

	r.Exchange.Args = args
	err = r.Exchange.BindToChannel(ch)
	if err != nil {
		return nil, err
	}

	r.QueueManager = rabbitmq.NewQueueManager(ch)
	cfgQueue := rabbitmq.QueueConfig{
		Durable: true,
	}

	_, err = r.QueueManager.DeclareQueue("delayed_queue", cfgQueue)
	if err != nil {
		return nil, err
	}
	cfgConsume := rabbitmq.ConsumerConfig{
		Queue:    "delayed_queue",
		Consumer: "1",
		AutoAck:  true,
	}

	err = ch.QueueBind("delayed_queue", "delayed_key", r.Exchange.Name(), false, nil)
	if err != nil {
		return nil, err
	}

	r.Consumer = rabbitmq.NewConsumer(ch, &cfgConsume)

	r.Publisher = rabbitmq.NewPublisher(ch, r.Exchange.Name())

	return &r, nil
}

func (r *RabbitMQ) PublishNotification(n model.CreateNotification, uid string) error {

	var ms int64
	if n.DateTime.Before(time.Now()) {
		ms = 0
	} else {
		delta := time.Until(n.DateTime)
		ms = delta.Milliseconds()

	}

	opts := rabbitmq.PublishingOptions{
		Headers: amqp091.Table{
			"x-delay": ms,
		},
	}

	retrнStrat := retry.Strategy{
		Attempts: 5,
		Delay:    5 * time.Second,
		Backoff:  2,
	}

	msg, err := json.Marshal(map[string]string{
		"uuid": uid,
	})
	if err != nil {
		return err
	}

	err = r.Publisher.PublishWithRetry(msg, "delayed_key", "application/json", retrнStrat, opts)
	if err != nil {
		return err
	}
	zlog.Logger.Info().Msgf("notification publish in queue: %s", uid)
	return nil
}
