package initiator

import (
	"fmt"
	"log/slog"
	"time"
)

type initiator struct {
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewInitiator(log *slog.Logger, period time.Duration, run func() error) (*initiator, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive")
	}

	i := &initiator{
		ticker: time.NewTicker(period),
		stopCh: make(chan struct{}),
	}

	if err := run(); err != nil {
		log.Error("cant do action in initiator during creating", "err", err)
	}

	go func() {
		for {
			select {
			case <-i.stopCh:
				return
			case <-i.ticker.C:
				if err := run(); err != nil {
					log.Error("cant do action in initiator", "err", err)
				}
			}
		}
	}()

	return i, nil
}

func (i *initiator) Stop() {
	i.ticker.Stop()
	close(i.stopCh)
}
