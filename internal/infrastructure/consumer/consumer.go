package consumer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type service interface {
	BindUserRole(ctx context.Context, userID string) error
}

type Conf struct {
	Brokers []string
	Topic   string
	GroupID string
}

type EvenUserConsumer struct {
	cons *kafka.Reader
	s    service
}

func New(cfg Conf, s service) *EvenUserConsumer {
	return &EvenUserConsumer{
		cons: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
		s: s,
	}
}

func (c *EvenUserConsumer) SubscribeUserID(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.cons.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("reading user id from kafka: %s", err)
			}

			if err := c.s.BindUserRole(ctx, string(msg.Value)); err != nil {
				return fmt.Errorf("binding user role: %w", err)
			}

			return c.cons.CommitMessages(ctx, msg)
		}
	}
}

func (c *EvenUserConsumer) Close() error {
	if err := c.cons.Close(); err != nil {
		return fmt.Errorf("closing kafka consumer: %w", err)
	}

	return nil
}
