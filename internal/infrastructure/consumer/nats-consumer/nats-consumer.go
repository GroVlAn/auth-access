package nats_consumer

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type service interface {
	BindUserRole(ctx context.Context, userID string) error
}

type Conf struct {
	NatsURL string
	Subj    string
}

type NatsConsumer struct {
	l  zerolog.Logger
	nc *nats.Conn
	s  service
	Conf
}

func New(l zerolog.Logger, s service, cfg Conf) (*NatsConsumer, error) {
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("connection to nats: %w", err)
	}

	return &NatsConsumer{
		l:    l,
		nc:   nc,
		s:    s,
		Conf: cfg,
	}, nil
}

func (c *NatsConsumer) SubscribeUserID(ctx context.Context) error {
	_, err := c.nc.Subscribe(c.Conf.Subj, func(msg *nats.Msg) {
		if err := c.s.BindUserRole(ctx, string(msg.Data)); err != nil {
			c.l.Error().Err(err).Msg("failed bind user role")
		}
	})

	return err
}

func (c *NatsConsumer) Close() error {
	if err := c.nc.Drain(); err != nil {
		return fmt.Errorf("closing nats consumer: %w", err)
	}

	return nil
}
