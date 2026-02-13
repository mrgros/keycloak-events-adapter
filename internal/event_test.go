package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"keycloak-events-adapter/internal/mock"
)

func TestEventService_Push(t *testing.T) {
	eventCorrect := &Event{
		Id:        uuid.New(),
		Time:      time.Now(),
		Type:      EventTypeLogin,
		RealmId:   uuid.New(),
		RealmName: "test",
		ClientId:  "client id",
		UserId:    uuid.New(),
		SessionId: "session id",
		IpAddress: "127.0.0.1",
		Error:     "error",
		Details:   map[string]string{"key": "value"},
	}

	type args struct {
		event *Event
	}
	tests := []struct {
		name    string
		args    args
		prepare func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event])
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				event: eventCorrect,
			},
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				eventStorage.EXPECT().Push(eventCorrect).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "push_error",
			args: args{
				event: eventCorrect,
			},
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				eventStorage.EXPECT().Push(eventCorrect).Return(fmt.Errorf("push error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			adminEventStorage := mock.NewMockEventKeeper[AdminEvent](ctrl)
			eventStorage := mock.NewMockEventKeeper[Event](ctrl)

			tt.prepare(adminEventStorage, eventStorage)
			e := NewEventService(adminEventStorage, eventStorage)
			if err := e.Push(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventService_PushAdmin(t *testing.T) {
	adminEventCorrect := &AdminEvent{
		Id:             uuid.New(),
		Time:           time.Now(),
		RealmId:        uuid.New(),
		RealmName:      "test",
		ResourceType:   "user",
		OperationType:  OperationTypeCreate,
		ResourcePath:   "/users",
		Representation: "{}",
		Error:          "",
		Details:        map[string]string{"key": "value"},
	}

	type args struct {
		event *AdminEvent
	}
	tests := []struct {
		name    string
		args    args
		prepare func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event])
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				event: adminEventCorrect,
			},
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				adminEventStorage.EXPECT().Push(adminEventCorrect).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "push_admin_error",
			args: args{
				event: adminEventCorrect,
			},
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				adminEventStorage.EXPECT().Push(adminEventCorrect).Return(fmt.Errorf("push admin error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			adminEventStorage := mock.NewMockEventKeeper[AdminEvent](ctrl)
			eventStorage := mock.NewMockEventKeeper[Event](ctrl)

			tt.prepare(adminEventStorage, eventStorage)
			e := NewEventService(adminEventStorage, eventStorage)
			if err := e.PushAdmin(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("PushAdmin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventService_Read(t *testing.T) {
	tests := []struct {
		name       string
		numWorkers int
		prepare    func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event])
	}{
		{
			name:       "single_worker",
			numWorkers: 1,
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				adminEventStorage.EXPECT().Process(gomock.Any()).DoAndReturn(
					func(ctx context.Context) {
						<-ctx.Done()
					})
				eventStorage.EXPECT().Process(gomock.Any()).DoAndReturn(
					func(ctx context.Context) {
						<-ctx.Done()
					})
			},
		},
		{
			name:       "multiple_workers",
			numWorkers: 3,
			prepare: func(adminEventStorage *mock.MockEventKeeper[AdminEvent], eventStorage *mock.MockEventKeeper[Event]) {
				for i := 0; i < 3; i++ {
					adminEventStorage.EXPECT().Process(gomock.Any()).DoAndReturn(
						func(ctx context.Context) {
							<-ctx.Done()
						})
					eventStorage.EXPECT().Process(gomock.Any()).DoAndReturn(
						func(ctx context.Context) {
							<-ctx.Done()
						})
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			adminEventStorage := mock.NewMockEventKeeper[AdminEvent](ctrl)
			eventStorage := mock.NewMockEventKeeper[Event](ctrl)

			tt.prepare(adminEventStorage, eventStorage)
			e := NewEventService(adminEventStorage, eventStorage)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			e.Read(ctx, tt.numWorkers)
		})
	}
}
