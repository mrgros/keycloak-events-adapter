package tarantool

//go:generate mockgen -destination=mock/queue.go -package=mock github.com/tarantool/go-tarantool/queue Queue

import (
	"context"
	"errors"
	"github.com/tarantool/go-tarantool/queue"
	"go.uber.org/zap"
	"keycloak-events-adapter/internal"
	"time"
)

const EventsQueueName = "events"
const AdminEventsQueueName = "admin_events"

type Event[T internal.Event | internal.AdminEvent] struct {
	queue       queue.Queue
	eventSender internal.EventSender[T]
	logger      *zap.Logger
}

func NewEvent[T internal.Event | internal.AdminEvent](
	queue queue.Queue,
	eventSender internal.EventSender[T],
	logger *zap.Logger,
) *Event[T] {
	return &Event[T]{
		queue:       queue,
		eventSender: eventSender,
		logger:      logger,
	}
}

func (e *Event[T]) Push(event *T) error {
	_, err := e.queue.PutWithOpts(event, queue.Opts{
		Ttl: 4 * time.Hour,
	})
	if err != nil {
		e.logger.Error("failed to push", zap.Error(err))
		return errors.New("can't put to queue")
	}

	return nil
}

func (e *Event[T]) Process(ctx context.Context) {
	var event *T
	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Queue worker has been shutdown")
			return
		default:
		}

		task, err := e.queue.TakeTypedTimeout(1*time.Second, &event)
		if err != nil {
			e.logger.Error("can't take task", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}
		if task == nil {
			continue
		}

		err = e.eventSender.Send(event)
		if err != nil {
			e.logger.Error("can't send event", zap.Error(err))
			err = task.ReleaseCfg(queue.Opts{
				Delay: 10 * time.Second,
			})
			if err != nil {
				e.logger.Error("can't release task", zap.Error(err))
			}

			continue
		}

		err = task.Ack()
		if err != nil {
			e.logger.Error("can't release task, trying to delete", zap.Error(err))

			err = task.Delete()
			if err != nil {
				e.logger.Error("can't delete task", zap.Error(err))
			}
		}
	}
}
