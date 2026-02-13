package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"keycloak-events-adapter/internal"
	"keycloak-events-adapter/internal/api/grpc/mock"
	eventv1 "keycloak-events-adapter/internal/specs/gen/keycloak/event/v1"
)

func TestEventServer_CreateAdmin(t *testing.T) {
	validUUID := uuid.New().String()
	validRealmID := uuid.New().String()
	validUserID := uuid.New().String()
	validClientID := uuid.New().String()

	tests := []struct {
		name        string
		req         *eventv1.CreateAdminRequest
		prepare     func(*mock.MockEventProvider)
		wantErrCode codes.Code
	}{
		{
			name: "Success",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				Time:          timestamppb.Now(),
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
				ResourceType:  "CLIENT",
				ResourcePath:  "clients/" + validUUID,
				AuthDetails: &eventv1.CreateAdminRequest_AuthDetails{
					RealmId:   validRealmID,
					ClientId:  validClientID,
					UserId:    validUserID,
					IpAddress: "127.0.0.1",
				},
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().PushAdmin(gomock.Any()).DoAndReturn(func(event *internal.AdminEvent) error {
					assert.Equal(t, validUUID, event.Id.String())
					assert.Equal(t, validRealmID, event.RealmId.String())
					assert.Equal(t, internal.OperationTypeCreate, event.OperationType)
					assert.Equal(t, validUserID, event.AuthDetails.UserId.String())
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
		{
			name: "Invalid ID",
			req: &eventv1.CreateAdminRequest{
				Id:            "invalid-uuid",
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Realm ID",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       "invalid-uuid",
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid AuthDetails User ID",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
				AuthDetails: &eventv1.CreateAdminRequest_AuthDetails{
					UserId: "invalid-uuid",
				},
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Push Error",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().PushAdmin(gomock.Any()).DoAndReturn(func(event *internal.AdminEvent) error {
					assert.Equal(t, validUUID, event.Id.String())
					return errors.New("storage error")
				})
			},
			wantErrCode: codes.Internal,
		},
		{
			name:        "Nil Request",
			req:         nil,
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Nil AuthDetails - Success",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				Time:          timestamppb.Now(),
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
				ResourceType:  "CLIENT",
				AuthDetails:   nil,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().PushAdmin(gomock.Any()).DoAndReturn(func(event *internal.AdminEvent) error {
					assert.Equal(t, validUUID, event.Id.String())
					assert.Nil(t, event.AuthDetails)
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
		{
			name: "Invalid AuthDetails Client ID",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
				AuthDetails: &eventv1.CreateAdminRequest_AuthDetails{
					ClientId: "invalid-client-id",
				},
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid AuthDetails Realm ID",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
				AuthDetails: &eventv1.CreateAdminRequest_AuthDetails{
					RealmId: "invalid-realm-id",
				},
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Operation Type",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				RealmId:       validRealmID,
				OperationType: 999, // Invalid type
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Valid Time conversion",
			req: &eventv1.CreateAdminRequest{
				Id:            validUUID,
				Time:          timestamppb.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)),
				RealmId:       validRealmID,
				OperationType: eventv1.OperationType_OPERATION_TYPE_CREATE,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().PushAdmin(gomock.Any()).DoAndReturn(func(event *internal.AdminEvent) error {
					expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
					assert.True(t, event.Time.Equal(expectedTime), "Expected %v, got %v", expectedTime, event.Time)
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mock.NewMockEventProvider(ctrl)
			if tt.prepare != nil {
				tt.prepare(mockProvider)
			}

			server := NewEventServer(mockProvider)
			resp, err := server.CreateAdmin(context.Background(), tt.req)

			if tt.wantErrCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			}
		})
	}
}

func TestEventServer_Create(t *testing.T) {
	validUUID := uuid.New().String()
	validRealmID := uuid.New().String()
	validUserID := uuid.New().String()

	tests := []struct {
		name        string
		req         *eventv1.CreateRequest
		prepare     func(*mock.MockEventProvider)
		wantErrCode codes.Code
	}{
		{
			name: "Success",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				Time:    timestamppb.Now(),
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().Push(gomock.Any()).DoAndReturn(func(event *internal.Event) error {
					assert.Equal(t, validUUID, event.Id.String())
					assert.Equal(t, validRealmID, event.RealmId.String())
					assert.Equal(t, validUserID, event.UserId.String())
					assert.Equal(t, internal.EventTypeLogin, event.Type)
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
		{
			name: "Invalid ID",
			req: &eventv1.CreateRequest{
				Id:      "invalid-uuid",
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Realm ID",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				RealmId: "invalid-uuid",
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid User ID",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				RealmId: validRealmID,
				UserId:  "invalid-uuid",
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Push Error",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().Push(gomock.Any()).DoAndReturn(func(event *internal.Event) error {
					assert.Equal(t, validUUID, event.Id.String())
					return errors.New("storage error")
				})
			},
			wantErrCode: codes.Internal,
		},
		{
			name:        "Nil Request",
			req:         nil,
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Event Type",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    999, // Invalid type
			},
			prepare:     nil,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Valid Time conversion",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				Time:    timestamppb.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)),
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().Push(gomock.Any()).DoAndReturn(func(event *internal.Event) error {
					expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
					assert.True(t, event.Time.Equal(expectedTime), "Expected %v, got %v", expectedTime, event.Time)
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
		{
			name: "Nil Time - should use zero time",
			req: &eventv1.CreateRequest{
				Id:      validUUID,
				Time:    nil,
				RealmId: validRealmID,
				UserId:  validUserID,
				Type:    eventv1.EventType_EVENT_TYPE_LOGIN,
			},
			prepare: func(m *mock.MockEventProvider) {
				m.EXPECT().Push(gomock.Any()).DoAndReturn(func(event *internal.Event) error {
					assert.True(t, event.Time.IsZero(), "Expected zero time, got %v", event.Time)
					return nil
				})
			},
			wantErrCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mock.NewMockEventProvider(ctrl)
			if tt.prepare != nil {
				tt.prepare(mockProvider)
			}

			server := NewEventServer(mockProvider)
			resp, err := server.Create(context.Background(), tt.req)

			if tt.wantErrCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			}
		})
	}
}
