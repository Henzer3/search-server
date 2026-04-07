package nats

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type natsGeneral struct {
	sub *nats.Subscription
	f   func(msg *nats.Msg)
}

type chanSubscriber struct {
	log         *slog.Logger
	conn        *nats.Conn
	subscribers map[string]natsGeneral
	ch          chan *nats.Msg
	stopCh      chan struct{}
}

func New(log *slog.Logger, brokerAdress string, chanSize int) (*chanSubscriber, error) {
	if chanSize <= 0 {
		return nil, fmt.Errorf("chanSize must be > 0")
	}
	nc, err := nats.Connect(brokerAdress)
	if err != nil {
		log.Error("cant connect to nats", "err", err)
		return nil, err
	}
	return &chanSubscriber{log: log, conn: nc, subscribers: make(map[string]natsGeneral),
		ch: make(chan *nats.Msg, chanSize), stopCh: make(chan struct{})}, nil
}

func (s *chanSubscriber) Subscribe(subj string, f func(msg *nats.Msg)) error {
	sub, err := s.conn.ChanSubscribe(subj, s.ch)
	if err != nil {
		s.log.Error("cant subscribes", "err", err, "topic", subj)
		return err
	}
	general := natsGeneral{sub: sub, f: f}
	s.subscribers[subj] = general
	return nil
}

func (s *chanSubscriber) unsubscribe(subj string) error {
	if v, ok := s.subscribers[subj]; ok {
		if err := v.sub.Unsubscribe(); err != nil {
			s.log.Error("cant Unsubscribe", "err", err, "topic", subj)
			return err
		}
		delete(s.subscribers, subj)
	}
	return nil
}

func (s *chanSubscriber) unsubscribeAll() error {
	for key := range s.subscribers {
		if err := s.unsubscribe(key); err != nil {
			return err
		}
	}
	return nil
}

func (s *chanSubscriber) StartListen() {
	go func() {
		for {
			select {
			case <-s.stopCh:
				return
			case msg := <-s.ch:
				if h, ok := s.subscribers[msg.Subject]; ok {
					h.f(msg)
				} else {
					s.log.Warn("no handler for subject", "subject", msg.Subject)
				}
			}
		}
	}()
}

func (s *chanSubscriber) Stop() error {
	close(s.stopCh)

	if err := s.unsubscribeAll(); err != nil {
		return err
	}

	s.conn.Close()
	return nil
}
