package internal

import "go.uber.org/zap"

type EventSender[T Event | AdminEvent] interface {
	Send(event *T) error
}

type Dummy[T Event | AdminEvent] struct {
	logger *zap.Logger
}

func NewDummy[T Event | AdminEvent](logger *zap.Logger) *Dummy[T] {
	return &Dummy[T]{logger: logger}
}

func (d *Dummy[T]) Send(event *T) error {
	d.logger.Debug("event took", zap.Any("event", event))
	return nil
}
