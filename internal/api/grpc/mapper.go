package grpc

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"keycloak-events-adapter/internal"
	eventv1 "keycloak-events-adapter/internal/specs/gen/keycloak/event/v1"
)

var eventTypeMap = map[eventv1.EventType]internal.EventType{
	eventv1.EventType_EVENT_TYPE_LOGIN:                                    internal.EventTypeLogin,
	eventv1.EventType_EVENT_TYPE_LOGIN_ERROR:                              internal.EventTypeLoginError,
	eventv1.EventType_EVENT_TYPE_REGISTER:                                 internal.EventTypeRegister,
	eventv1.EventType_EVENT_TYPE_REGISTER_ERROR:                           internal.EventTypeRegisterError,
	eventv1.EventType_EVENT_TYPE_LOGOUT:                                   internal.EventTypeLogout,
	eventv1.EventType_EVENT_TYPE_LOGOUT_ERROR:                             internal.EventTypeLogoutError,
	eventv1.EventType_EVENT_TYPE_CODE_TO_TOKEN:                            internal.EventTypeCodeToToken,
	eventv1.EventType_EVENT_TYPE_CODE_TO_TOKEN_ERROR:                      internal.EventTypeCodeToTokenError,
	eventv1.EventType_EVENT_TYPE_CLIENT_LOGIN:                             internal.EventTypeClientLogin,
	eventv1.EventType_EVENT_TYPE_CLIENT_LOGIN_ERROR:                       internal.EventTypeClientLoginError,
	eventv1.EventType_EVENT_TYPE_REFRESH_TOKEN:                            internal.EventTypeRefreshToken,
	eventv1.EventType_EVENT_TYPE_REFRESH_TOKEN_ERROR:                      internal.EventTypeRefreshTokenError,
	eventv1.EventType_EVENT_TYPE_INTROSPECT_TOKEN:                         internal.EventTypeIntrospectToken,
	eventv1.EventType_EVENT_TYPE_INTROSPECT_TOKEN_ERROR:                   internal.EventTypeIntrospectTokenError,
	eventv1.EventType_EVENT_TYPE_FEDERATED_IDENTITY_LINK:                  internal.EventTypeFederatedIdentityLink,
	eventv1.EventType_EVENT_TYPE_FEDERATED_IDENTITY_LINK_ERROR:            internal.EventTypeFederatedIdentityLinkError,
	eventv1.EventType_EVENT_TYPE_REMOVE_FEDERATED_IDENTITY:                internal.EventTypeRemoveFederatedIdentity,
	eventv1.EventType_EVENT_TYPE_REMOVE_FEDERATED_IDENTITY_ERROR:          internal.EventTypeRemoveFederatedIdentityError,
	eventv1.EventType_EVENT_TYPE_UPDATE_EMAIL:                             internal.EventTypeUpdateEmail,
	eventv1.EventType_EVENT_TYPE_UPDATE_EMAIL_ERROR:                       internal.EventTypeUpdateEmailError,
	eventv1.EventType_EVENT_TYPE_UPDATE_PROFILE:                           internal.EventTypeUpdateProfile,
	eventv1.EventType_EVENT_TYPE_UPDATE_PROFILE_ERROR:                     internal.EventTypeUpdateProfileError,
	eventv1.EventType_EVENT_TYPE_VERIFY_EMAIL:                             internal.EventTypeVerifyEmail,
	eventv1.EventType_EVENT_TYPE_VERIFY_EMAIL_ERROR:                       internal.EventTypeVerifyEmailError,
	eventv1.EventType_EVENT_TYPE_VERIFY_PROFILE:                           internal.EventTypeVerifyProfile,
	eventv1.EventType_EVENT_TYPE_VERIFY_PROFILE_ERROR:                     internal.EventTypeVerifyProfileError,
	eventv1.EventType_EVENT_TYPE_GRANT_CONSENT:                            internal.EventTypeGrantConsent,
	eventv1.EventType_EVENT_TYPE_GRANT_CONSENT_ERROR:                      internal.EventTypeGrantConsentError,
	eventv1.EventType_EVENT_TYPE_UPDATE_CONSENT:                           internal.EventTypeUpdateConsent,
	eventv1.EventType_EVENT_TYPE_UPDATE_CONSENT_ERROR:                     internal.EventTypeUpdateConsentError,
	eventv1.EventType_EVENT_TYPE_REVOKE_GRANT:                             internal.EventTypeRevokeGrant,
	eventv1.EventType_EVENT_TYPE_REVOKE_GRANT_ERROR:                       internal.EventTypeRevokeGrantError,
	eventv1.EventType_EVENT_TYPE_SEND_VERIFY_EMAIL:                        internal.EventTypeSendVerifyEmail,
	eventv1.EventType_EVENT_TYPE_SEND_VERIFY_EMAIL_ERROR:                  internal.EventTypeSendVerifyEmailError,
	eventv1.EventType_EVENT_TYPE_SEND_RESET_PASSWORD:                      internal.EventTypeSendResetPassword,
	eventv1.EventType_EVENT_TYPE_SEND_RESET_PASSWORD_ERROR:                internal.EventTypeSendResetPasswordError,
	eventv1.EventType_EVENT_TYPE_RESET_PASSWORD:                           internal.EventTypeResetPassword,
	eventv1.EventType_EVENT_TYPE_RESET_PASSWORD_ERROR:                     internal.EventTypeResetPasswordError,
	eventv1.EventType_EVENT_TYPE_RESTART_AUTHENTICATION:                   internal.EventTypeRestartAuthentication,
	eventv1.EventType_EVENT_TYPE_RESTART_AUTHENTICATION_ERROR:             internal.EventTypeRestartAuthenticationError,
	eventv1.EventType_EVENT_TYPE_INVALID_SIGNATURE:                        internal.EventTypeInvalidSignature,
	eventv1.EventType_EVENT_TYPE_INVALID_SIGNATURE_ERROR:                  internal.EventTypeInvalidSignatureError,
	eventv1.EventType_EVENT_TYPE_REGISTER_NODE:                            internal.EventTypeRegisterNode,
	eventv1.EventType_EVENT_TYPE_REGISTER_NODE_ERROR:                      internal.EventTypeRegisterNodeError,
	eventv1.EventType_EVENT_TYPE_UNREGISTER_NODE:                          internal.EventTypeUnregisterNode,
	eventv1.EventType_EVENT_TYPE_UNREGISTER_NODE_ERROR:                    internal.EventTypeUnregisterNodeError,
	eventv1.EventType_EVENT_TYPE_USER_INFO_REQUEST:                        internal.EventTypeUserInfoRequest,
	eventv1.EventType_EVENT_TYPE_USER_INFO_REQUEST_ERROR:                  internal.EventTypeUserInfoRequestError,
	eventv1.EventType_EVENT_TYPE_IMPERSONATE:                              internal.EventTypeImpersonate,
	eventv1.EventType_EVENT_TYPE_IMPERSONATE_ERROR:                        internal.EventTypeImpersonateError,
	eventv1.EventType_EVENT_TYPE_EXECUTE_ACTIONS:                          internal.EventTypeExecuteActions,
	eventv1.EventType_EVENT_TYPE_EXECUTE_ACTIONS_ERROR:                    internal.EventTypeExecuteActionsError,
	eventv1.EventType_EVENT_TYPE_EXECUTE_ACTION_TOKEN:                     internal.EventTypeExecuteActionToken,
	eventv1.EventType_EVENT_TYPE_EXECUTE_ACTION_TOKEN_ERROR:               internal.EventTypeExecuteActionTokenError,
	eventv1.EventType_EVENT_TYPE_CLIENT_INFO:                              internal.EventTypeClientInfo,
	eventv1.EventType_EVENT_TYPE_CLIENT_INFO_ERROR:                        internal.EventTypeClientInfoError,
	eventv1.EventType_EVENT_TYPE_CLIENT_REGISTER:                          internal.EventTypeClientRegister,
	eventv1.EventType_EVENT_TYPE_CLIENT_REGISTER_ERROR:                    internal.EventTypeClientRegisterError,
	eventv1.EventType_EVENT_TYPE_CLIENT_UPDATE:                            internal.EventTypeClientUpdate,
	eventv1.EventType_EVENT_TYPE_CLIENT_UPDATE_ERROR:                      internal.EventTypeClientUpdateError,
	eventv1.EventType_EVENT_TYPE_CLIENT_DELETE:                            internal.EventTypeClientDelete,
	eventv1.EventType_EVENT_TYPE_CLIENT_DELETE_ERROR:                      internal.EventTypeClientDeleteError,
	eventv1.EventType_EVENT_TYPE_TOKEN_EXCHANGE:                           internal.EventTypeTokenExchange,
	eventv1.EventType_EVENT_TYPE_TOKEN_EXCHANGE_ERROR:                     internal.EventTypeTokenExchangeError,
	eventv1.EventType_EVENT_TYPE_DELETE_ACCOUNT:                           internal.EventTypeDeleteAccount,
	eventv1.EventType_EVENT_TYPE_DELETE_ACCOUNT_ERROR:                     internal.EventTypeDeleteAccountError,
	eventv1.EventType_EVENT_TYPE_USER_DISABLED_BY_PERMANENT_LOCKOUT:       internal.EventTypeUserDisabledByPermanentLockout,
	eventv1.EventType_EVENT_TYPE_USER_DISABLED_BY_PERMANENT_LOCKOUT_ERROR: internal.EventTypeUserDisabledByPermanentLockoutError,
	eventv1.EventType_EVENT_TYPE_USER_DISABLED_BY_TEMPORARY_LOCKOUT:       internal.EventTypeUserDisabledByTemporaryLockout,
	eventv1.EventType_EVENT_TYPE_USER_DISABLED_BY_TEMPORARY_LOCKOUT_ERROR: internal.EventTypeUserDisabledByTemporaryLockoutError,
	eventv1.EventType_EVENT_TYPE_UPDATE_CREDENTIAL:                        internal.EventTypeUpdateCredential,
	eventv1.EventType_EVENT_TYPE_UPDATE_CREDENTIAL_ERROR:                  internal.EventTypeUpdateCredentialError,
	eventv1.EventType_EVENT_TYPE_REMOVE_CREDENTIAL:                        internal.EventTypeRemoveCredential,
	eventv1.EventType_EVENT_TYPE_REMOVE_CREDENTIAL_ERROR:                  internal.EventTypeRemoveCredentialError,
}

var operationTypeMap = map[eventv1.OperationType]internal.OperationType{
	eventv1.OperationType_OPERATION_TYPE_CREATE: internal.OperationTypeCreate,
	eventv1.OperationType_OPERATION_TYPE_UPDATE: internal.OperationTypeUpdate,
	eventv1.OperationType_OPERATION_TYPE_DELETE: internal.OperationTypeDelete,
	eventv1.OperationType_OPERATION_TYPE_ACTION: internal.OperationTypeAction,
}

func mapCreateAdminRequestToAdminEvent(request *eventv1.CreateAdminRequest) (*internal.AdminEvent, error) {
	if request == nil {
		return nil, errors.New("request is empty")
	}

	id, err := uuid.Parse(request.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	var eventTime time.Time
	if request.GetTime() != nil {
		eventTime = request.GetTime().AsTime()
	}

	realmId, err := uuid.Parse(request.GetRealmId())
	if err != nil {
		return nil, fmt.Errorf("invalid realm id: %w", err)
	}

	var authDetails *internal.AuthDetails
	if request.GetAuthDetails() != nil {
		authDetails = &internal.AuthDetails{
			RealmName: request.GetAuthDetails().GetRealmName(),
			UserId:    uuid.Nil,
			IpAddress: request.GetAuthDetails().GetIpAddress(),
		}

		if request.GetAuthDetails().GetRealmId() != "" {
			authDetails.RealmId, err = uuid.Parse(request.GetAuthDetails().GetRealmId())
			if err != nil {
				return nil, fmt.Errorf("invalid auth details realm id: %w", err)
			}
		}

		if request.GetAuthDetails().GetClientId() != "" {
			authDetails.ClientId, err = uuid.Parse(request.GetAuthDetails().GetClientId())
			if err != nil {
				return nil, fmt.Errorf("invalid auth details client id: %w", err)
			}
		}

		if request.GetAuthDetails().GetUserId() != "" {
			authDetails.UserId, err = uuid.Parse(request.GetAuthDetails().GetUserId())
			if err != nil {
				return nil, fmt.Errorf("invalid auth details user id: %w", err)
			}
		}
	}

	operationType, ok := operationTypeMap[request.OperationType]
	if !ok {
		return nil, fmt.Errorf("invalid operation type: %s", request.OperationType)
	}

	return &internal.AdminEvent{
		Id:             id,
		Time:           eventTime,
		RealmId:        realmId,
		RealmName:      request.GetRealmName(),
		AuthDetails:    authDetails,
		ResourceType:   request.GetResourceType(),
		OperationType:  operationType,
		ResourcePath:   request.GetResourcePath(),
		Representation: request.GetRepresentation(),
		Error:          request.GetError(),
		Details:        request.GetDetails(),
	}, nil
}

func mapCreateRequestToEvent(request *eventv1.CreateRequest) (*internal.Event, error) {
	if request == nil {
		return nil, errors.New("request is empty")
	}

	id, err := uuid.Parse(request.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	var eventTime time.Time
	if request.GetTime() != nil {
		eventTime = request.GetTime().AsTime()
	}

	realmId, err := uuid.Parse(request.GetRealmId())
	if err != nil {
		return nil, fmt.Errorf("invalid realm id: %w", err)
	}

	userId, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	eventType, ok := eventTypeMap[request.Type]
	if !ok {
		return nil, fmt.Errorf("invalid event type: %s", request.Type)
	}

	return &internal.Event{
		Id:        id,
		Time:      eventTime,
		Type:      eventType,
		RealmId:   realmId,
		RealmName: request.GetRealmName(),
		ClientId:  request.GetClientId(),
		UserId:    userId,
		SessionId: request.GetSessionId(),
		IpAddress: request.GetIpAddress(),
		Error:     request.GetError(),
		Details:   request.GetDetails(),
	}, nil
}
