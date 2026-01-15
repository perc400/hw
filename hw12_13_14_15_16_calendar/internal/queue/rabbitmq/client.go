package rabbitmq

import (
	"context"
	"errors"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/queue" //nolint:depguard
	"github.com/streadway/amqp"                                   //nolint:depguard
)

const (
	exchangeName = "calendar.events"
	exchangeType = "direct"
	routingKey   = "notification"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	logger  Logger
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

func NewClient(queue string, url string, logger Logger) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err = ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err = ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	logger.Info("RabbitMQ connected, exchange and queue ready")

	return &Client{
		conn:    conn,
		channel: ch,
		queue:   queue,
		logger:  logger,
	}, nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, msg queue.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return c.channel.Publish(
		exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
		},
	)
}

func (c *Client) Consume(ctx context.Context, handler queue.MessageHandler) error {
	deliveries, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer string
		true,    // autoAck
		false,   // exclusive
		false,   // noLocal
		false,   // noWait
		nil,     // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("RabbitMQ consumer stopped")
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("deliveries channel closed")
			}

			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := handler(msgCtx, queue.Message{Body: d.Body})
			cancel()
			if err != nil {
				c.logger.Error("handler error: " + err.Error())
			}
		}
	}
}
