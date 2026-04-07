package nats

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	log  *slog.Logger
	conn *nats.Conn
}

func New(log *slog.Logger, brokerAdress string) (*Publisher, error) {
	nc, err := nats.Connect(brokerAdress)
	if err != nil {
		log.Error("cant connect to nats", "err", err)
		return nil, err
	}
	return &Publisher{log: log, conn: nc}, nil
}

func (p *Publisher) Publish(topic string, data string) error {
	if err := p.conn.Publish(topic, []byte(data)); err != nil {
		p.log.Error("cant publish", "err", err, "topic", topic)
		return err
	}
	if err := p.conn.Flush(); err != nil {
		p.log.Error("cant flush topic", "err", err, "topic", topic)
		return err
	}
	return nil
}

func (p *Publisher) Close() {
	p.conn.Close()
}
