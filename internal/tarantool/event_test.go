package tarantool

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool/queue"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"keycloak-events-adapter/internal"
	"keycloak-events-adapter/internal/tarantool/mock"
	"testing"
	"time"
)

func TestEvent_Process(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type testCase[T interface {
		internal.Event | internal.AdminEvent
	}] struct {
		name    string
		args    args
		prepare func(queueMock *mock.MockQueue)
	}
	tests := []testCase[internal.Event]{
		{
			name: "take error",
			prepare: func(queueMock *mock.MockQueue) {
				var eventPtr *internal.Event
				queueMock.EXPECT().TakeTypedTimeout(1*time.Second, &eventPtr).Return(nil, errors.New("take error")).Times(1)
			},
		},
		{
			name: "nil task",
			prepare: func(queueMock *mock.MockQueue) {
				var eventPtr *internal.Event
				queueMock.EXPECT().TakeTypedTimeout(1*time.Second, &eventPtr).Return(nil, nil).AnyTimes()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			queueMock := mock.NewMockQueue(ctrl)
			logger := zap.NewNop()

			tt.prepare(queueMock)

			event := NewEvent[internal.Event](queueMock, internal.NewDummy[internal.Event](logger), logger)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer close(done)
				event.Process(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
			<-done
		})
	}

	adminTests := []testCase[internal.AdminEvent]{
		{
			name: "admin take error",
			prepare: func(queueMock *mock.MockQueue) {
				var eventPtr *internal.AdminEvent
				queueMock.EXPECT().TakeTypedTimeout(1*time.Second, &eventPtr).Return(nil, errors.New("take error")).Times(1)
			},
		},
		{
			name: "admin nil task",
			prepare: func(queueMock *mock.MockQueue) {
				var eventPtr *internal.AdminEvent
				queueMock.EXPECT().TakeTypedTimeout(1*time.Second, &eventPtr).Return(nil, nil).AnyTimes()
			},
		},
	}
	for _, tt := range adminTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			queueMock := mock.NewMockQueue(ctrl)
			logger := zap.NewNop()

			tt.prepare(queueMock)

			event := NewEvent[internal.AdminEvent](queueMock, internal.NewDummy[internal.AdminEvent](logger), logger)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer close(done)
				event.Process(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
			<-done
		})
	}
}

func TestEvent_Push(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	realmId := uuid.New()
	userId := uuid.New()

	type args[T interface {
		internal.Event | internal.AdminEvent
	}] struct {
		event *T
	}
	type testCase[T interface {
		internal.Event | internal.AdminEvent
	}] struct {
		name    string
		args    args[T]
		prepare func(queueMock *mock.MockQueue)
		wantErr bool
	}

	eventCorrect := &internal.Event{
		Id:        id,
		Time:      now,
		Type:      internal.EventTypeLogin,
		RealmId:   realmId,
		RealmName: "test",
		ClientId:  "client id",
		UserId:    userId,
		SessionId: "session id",
		IpAddress: "127.0.0.1",
		Error:     "error",
		Details:   map[string]string{"key": "value"},
	}

	emptyEvent := &internal.Event{
		Id:        uuid.UUID{},
		Time:      time.Time{},
		Type:      0,
		RealmId:   uuid.UUID{},
		RealmName: "",
		ClientId:  "",
		UserId:    uuid.UUID{},
		SessionId: "",
		IpAddress: "",
		Error:     "",
		Details:   nil,
	}

	adminEvent := &internal.AdminEvent{
		Id:             id,
		Time:           now,
		RealmId:        realmId,
		RealmName:      "admin realm",
		AuthDetails:    &internal.AuthDetails{RealmId: realmId, RealmName: "auth realm", ClientId: uuid.New(), UserId: userId, IpAddress: "192.168.1.1"},
		ResourceType:   "resource",
		OperationType:  internal.OperationTypeCreate,
		ResourcePath:   "/path",
		Representation: "{}",
		Error:          "",
		Details:        map[string]string{"admin": "detail"},
	}

	testsEvent := []testCase[internal.Event]{
		{
			name: "event correct",
			args: args[internal.Event]{
				event: eventCorrect,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(eventCorrect, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "queue put error",
			args: args[internal.Event]{
				event: eventCorrect,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(eventCorrect, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, errors.New("queue error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "nil event",
			args: args[internal.Event]{
				event: nil,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(nil, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, errors.New("nil event")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "empty event fields",
			args: args[internal.Event]{
				event: emptyEvent,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(emptyEvent, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, nil).Times(1)
			},
			wantErr: false,
		},
	}

	testsAdmin := []testCase[internal.AdminEvent]{
		{
			name: "admin event correct",
			args: args[internal.AdminEvent]{
				event: adminEvent,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(adminEvent, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "admin queue put error",
			args: args[internal.AdminEvent]{
				event: adminEvent,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(adminEvent, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, errors.New("queue error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "admin nil event",
			args: args[internal.AdminEvent]{
				event: nil,
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(nil, queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, errors.New("nil event")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "admin empty event fields",
			args: args[internal.AdminEvent]{
				event: &internal.AdminEvent{
					Id:             uuid.UUID{},
					Time:           time.Time{},
					RealmId:        uuid.UUID{},
					RealmName:      "",
					AuthDetails:    nil,
					ResourceType:   "",
					OperationType:  0,
					ResourcePath:   "",
					Representation: "",
					Error:          "",
					Details:        nil,
				},
			},
			prepare: func(queueMock *mock.MockQueue) {
				queueMock.EXPECT().PutWithOpts(gomock.Any(), queue.Opts{
					Ttl: 4 * time.Hour,
				}).Return(nil, nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range testsEvent {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			queueMock := mock.NewMockQueue(ctrl)
			logger := zap.NewNop()
			eventSender := internal.NewDummy[internal.Event](logger)

			tt.prepare(queueMock)
			eventStorage := NewEvent[internal.Event](queueMock, eventSender, logger)
			if err := eventStorage.Push(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	for _, tt := range testsAdmin {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			queueMock := mock.NewMockQueue(ctrl)
			logger := zap.NewNop()
			eventSender := internal.NewDummy[internal.AdminEvent](logger)

			tt.prepare(queueMock)
			eventStorage := NewEvent[internal.AdminEvent](queueMock, eventSender, logger)
			if err := eventStorage.Push(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
