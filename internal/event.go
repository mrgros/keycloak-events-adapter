package internal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type OperationType uint8

type EventType uint8

const (
	OperationTypeCreate OperationType = 1
	OperationTypeUpdate OperationType = 2
	OperationTypeDelete OperationType = 3
	OperationTypeAction OperationType = 4
)

const (
	EventTypeLogin                               EventType = 1
	EventTypeLoginError                          EventType = 2
	EventTypeRegister                            EventType = 3
	EventTypeRegisterError                       EventType = 4
	EventTypeLogout                              EventType = 5
	EventTypeLogoutError                         EventType = 6
	EventTypeCodeToToken                         EventType = 7
	EventTypeCodeToTokenError                    EventType = 8
	EventTypeClientLogin                         EventType = 9
	EventTypeClientLoginError                    EventType = 10
	EventTypeRefreshToken                        EventType = 11
	EventTypeRefreshTokenError                   EventType = 12
	EventTypeIntrospectToken                     EventType = 15
	EventTypeIntrospectTokenError                EventType = 16
	EventTypeFederatedIdentityLink               EventType = 17
	EventTypeFederatedIdentityLinkError          EventType = 18
	EventTypeRemoveFederatedIdentity             EventType = 19
	EventTypeRemoveFederatedIdentityError        EventType = 20
	EventTypeUpdateEmail                         EventType = 21
	EventTypeUpdateEmailError                    EventType = 22
	EventTypeUpdateProfile                       EventType = 23
	EventTypeUpdateProfileError                  EventType = 24
	EventTypeVerifyEmail                         EventType = 29
	EventTypeVerifyEmailError                    EventType = 30
	EventTypeVerifyProfile                       EventType = 31
	EventTypeVerifyProfileError                  EventType = 32
	EventTypeGrantConsent                        EventType = 35
	EventTypeGrantConsentError                   EventType = 36
	EventTypeUpdateConsent                       EventType = 37
	EventTypeUpdateConsentError                  EventType = 38
	EventTypeRevokeGrant                         EventType = 39
	EventTypeRevokeGrantError                    EventType = 40
	EventTypeSendVerifyEmail                     EventType = 41
	EventTypeSendVerifyEmailError                EventType = 42
	EventTypeSendResetPassword                   EventType = 43
	EventTypeSendResetPasswordError              EventType = 44
	EventTypeResetPassword                       EventType = 47
	EventTypeResetPasswordError                  EventType = 48
	EventTypeRestartAuthentication               EventType = 49
	EventTypeRestartAuthenticationError          EventType = 50
	EventTypeInvalidSignature                    EventType = 51
	EventTypeInvalidSignatureError               EventType = 52
	EventTypeRegisterNode                        EventType = 53
	EventTypeRegisterNodeError                   EventType = 54
	EventTypeUnregisterNode                      EventType = 55
	EventTypeUnregisterNodeError                 EventType = 56
	EventTypeUserInfoRequest                     EventType = 57
	EventTypeUserInfoRequestError                EventType = 58
	EventTypeImpersonate                         EventType = 71
	EventTypeImpersonateError                    EventType = 72
	EventTypeExecuteActions                      EventType = 75
	EventTypeExecuteActionsError                 EventType = 76
	EventTypeExecuteActionToken                  EventType = 77
	EventTypeExecuteActionTokenError             EventType = 78
	EventTypeClientInfo                          EventType = 79
	EventTypeClientInfoError                     EventType = 80
	EventTypeClientRegister                      EventType = 81
	EventTypeClientRegisterError                 EventType = 82
	EventTypeClientUpdate                        EventType = 83
	EventTypeClientUpdateError                   EventType = 84
	EventTypeClientDelete                        EventType = 85
	EventTypeClientDeleteError                   EventType = 86
	EventTypeTokenExchange                       EventType = 89
	EventTypeTokenExchangeError                  EventType = 90
	EventTypeDeleteAccount                       EventType = 101
	EventTypeDeleteAccountError                  EventType = 102
	EventTypeUserDisabledByPermanentLockout      EventType = 105
	EventTypeUserDisabledByPermanentLockoutError EventType = 106
	EventTypeUserDisabledByTemporaryLockout      EventType = 107
	EventTypeUserDisabledByTemporaryLockoutError EventType = 108
	EventTypeUpdateCredential                    EventType = 113
	EventTypeUpdateCredentialError               EventType = 114
	EventTypeRemoveCredential                    EventType = 115
	EventTypeRemoveCredentialError               EventType = 116
)

type AuthDetails struct {
	RealmId   uuid.UUID
	RealmName string
	ClientId  uuid.UUID
	UserId    uuid.UUID
	IpAddress string
}

type AdminEvent struct {
	Id             uuid.UUID
	Time           time.Time
	RealmId        uuid.UUID
	RealmName      string
	AuthDetails    *AuthDetails
	ResourceType   string
	OperationType  OperationType
	ResourcePath   string
	Representation string
	Error          string
	Details        map[string]string
}

type Event struct {
	Id        uuid.UUID
	Time      time.Time
	Type      EventType
	RealmId   uuid.UUID
	RealmName string
	ClientId  string
	UserId    uuid.UUID
	SessionId string
	IpAddress string
	Error     string
	Details   map[string]string
}

type EventKeeper[T Event | AdminEvent] interface {
	Push(event *T) error
	Process(ctx context.Context)
}

type EventProvider interface {
	PushAdmin(event *AdminEvent) error
	Push(event *Event) error
	Read(ctx context.Context, numWorkers int)
}

type EventService struct {
	adminEventStorage EventKeeper[AdminEvent]
	eventStorage      EventKeeper[Event]
}

func NewEventService(
	adminEventStorage EventKeeper[AdminEvent],
	eventStorage EventKeeper[Event],
) *EventService {
	return &EventService{
		adminEventStorage: adminEventStorage,
		eventStorage:      eventStorage,
	}
}

func (e *EventService) PushAdmin(event *AdminEvent) error {
	err := e.adminEventStorage.Push(event)
	if err != nil {
		return fmt.Errorf("push admin event: %w", err)
	}

	return nil
}

func (e *EventService) Push(event *Event) error {
	err := e.eventStorage.Push(event)
	if err != nil {
		return fmt.Errorf("push event: %w", err)
	}

	return nil
}

func (e *EventService) Read(ctx context.Context, numWorkers int) {
	wg := &sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(fctx context.Context) {
			defer wg.Done()
			e.adminEventStorage.Process(fctx)
		}(ctx)

		wg.Add(1)
		go func(fctx context.Context) {
			defer wg.Done()
			e.eventStorage.Process(fctx)
		}(ctx)
	}

	wg.Wait()
}
