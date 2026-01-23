package queue

import "context"

type Message struct {
	Body []byte
}

type MessageHandler func(ctx context.Context, msg Message) error

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}
