package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"keycloak-events-adapter/internal"
	eventv1 "keycloak-events-adapter/internal/specs/gen/keycloak/event/v1"
)

type EventServer struct {
	eventv1.UnimplementedEventAPIServer

	eventService internal.EventProvider
}

func NewEventServer(
	eventService internal.EventProvider,
) *EventServer {
	return &EventServer{
		eventService: eventService,
	}
}

func (e *EventServer) CreateAdmin(ctx context.Context, request *eventv1.CreateAdminRequest) (*eventv1.CreateAdminResponse, error) {
	adminEvent, err := mapCreateAdminRequestToAdminEvent(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't get request object: %s", err))
	}

	if adminEvent == nil {
		return nil, status.Error(codes.Internal, "empty admin event")
	}

	err = e.eventService.PushAdmin(adminEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("admin event: %s", err))
	}

	return &eventv1.CreateAdminResponse{}, nil
}

func (e *EventServer) Create(ctx context.Context, request *eventv1.CreateRequest) (*eventv1.CreateResponse, error) {
	event, err := mapCreateRequestToEvent(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't get request object: %s", err))
	}

	if event == nil {
		return nil, status.Error(codes.Internal, "empty event")
	}

	err = e.eventService.Push(event)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("event: %s", err))
	}

	return &eventv1.CreateResponse{}, nil
}
