package ubmanage

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	ev "github.com/kernelplex/ubase/internal/evercoregen/events"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ubalgorithms"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

type PrefectService interface {
	UserBelongsToRole(ctx context.Context, userId int64, roleId int64) (bool, error)

	UserHasPermission(ctx context.Context, userId int64, orgId int64, permission string) (bool, error)

	GroupInvalidation(ctx context.Context, roleId int64) error

	UserInvalidation(ctx context.Context, userId int64) error

	ApiKeyToUser(ctx context.Context, apiKey string) (ApiKeyData, error)

	Start() error
	Stop() error
}
type ApiKeyData struct {
	UserId         int64  `json:"userId"`
	Email          string `json:"email,omitempty"`
	OrganizationId int64  `json:"organizationId,omitempty"`
	ExpiresAt      int64  `json:"expiresAt,omitempty"`
}

type UserData struct {
	Id    int64   `json:"id"`
	Email string  `json:"email"`
	Roles []int64 `json:"groups"`
}

type GroupPermissions struct {
	GroupId        int64    `json:"groupId"`
	OrganizationId int64    `json:"organizationId"`
	Permissions    []string `json:"permissions"`
}

type PrefectServiceImpl struct {
	managementService ManagementService
	userCache         *ubalgorithms.LRUCache[int64, *UserData]
	groupCache        *ubalgorithms.LRUCache[int64, *GroupPermissions]
	apiKeyCache       *ubalgorithms.LRUCache[string, *ApiKeyData]
	store             *evercore.EventStore
	ctx               context.Context
	cancel            context.CancelFunc
	started           bool
}

func NewPrefectService(
	managementService ManagementService,
	store *evercore.EventStore,
	userCacheSize int,
	groupCacheSize int,
) PrefectService {
	ensure.That(managementService != nil, "managementService cannot be nil")
	ensure.That(store != nil, "store cannot be nil")
	ensure.That(userCacheSize > 0, "userCacheSize must be greater than 0")
	ensure.That(groupCacheSize > 0, "groupCacheSize must be greater than 0")

	userCache := ubalgorithms.NewLRUCache[int64, *UserData](userCacheSize)
	groupCache := ubalgorithms.NewLRUCache[int64, *GroupPermissions](groupCacheSize)
	apiKeyCache := ubalgorithms.NewLRUCache[string, *ApiKeyData](groupCacheSize)
	return &PrefectServiceImpl{
		managementService: managementService,
		userCache:         userCache,
		groupCache:        groupCache,
		apiKeyCache:       apiKeyCache,
		store:             store,
	}
}

func (p *PrefectServiceImpl) getUserData(ctx context.Context, userId int64) (*UserData, error) {
	userData, found := p.userCache.Get(userId)
	slog.Info("getUserData", "userId", userId, "foundInCache", found)
	if !found {
		// Load user data from management service
		userResp, err := p.managementService.UserGetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		if userResp.Status != ubstatus.Success {
			return nil, fmt.Errorf("failed to get user data: %s", userResp.Message)
		}

		roleList, err := p.managementService.UserGetAllOrganizationRoles(ctx, userId)

		if err != nil {
			return nil, err
		}

		roles := make([]int64, len(roleList.Data))
		for i, role := range roleList.Data {
			roles[i] = role.RoleID
		}

		userData = &UserData{
			Id:    userResp.Data.Id,
			Email: userResp.Data.State.Email,
			Roles: roles,
		}
		p.userCache.Put(userId, userData)
	}
	return userData, nil
}

func (p *PrefectServiceImpl) UserBelongsToRole(ctx context.Context, userId int64, groupId int64) (bool, error) {

	if started := p.started; !started {
		slog.Error("prefect service not started")
		return false, fmt.Errorf("prefect service not started")
	}

	userData, err := p.getUserData(ctx, userId)
	if err != nil {
		return false, err
	}

	if slices.Contains(userData.Roles, groupId) {
		return true, nil
	}

	return false, nil
}

func (p *PrefectServiceImpl) getGroupPermissions(ctx context.Context, groupId int64) (*GroupPermissions, error) {
	if started := p.started; !started {
		slog.Error("prefect service not started")
		return nil, fmt.Errorf("prefect service not started")
	}

	groupData, found := p.groupCache.Get(groupId)
	if !found {
		// Load group data from management service
		groupResp, err := p.managementService.RoleGetById(ctx, groupId)
		if err != nil {
			return nil, err
		}
		if groupResp.Status != ubstatus.Success {
			return nil, fmt.Errorf("failed to get group data: %s", groupResp.Message)
		}
		groupData = &GroupPermissions{
			GroupId:        groupResp.Data.Id,
			OrganizationId: groupResp.Data.State.OrganizationId,
			Permissions:    groupResp.Data.State.Permissions,
		}
		p.groupCache.Put(groupId, groupData)
	}
	return groupData, nil
}

func (p *PrefectServiceImpl) UserHasPermission(ctx context.Context, userId int64, orgId int64, permission string) (bool, error) {
	slog.Info("Checking if user has permission", "userId", userId, "orgId", orgId, "permission", permission)
	if started := p.started; !started {
		slog.Error("prefect service not started")
		return false, fmt.Errorf("prefect service not started")
	}

	userData, err := p.getUserData(ctx, userId)
	if err != nil {
		return false, err
	}

	for _, roleId := range userData.Roles {
		groupData, err := p.getGroupPermissions(ctx, roleId)
		if err != nil {
			return false, err
		}

		if groupData.OrganizationId != orgId {
			continue
		}

		if slices.Contains(groupData.Permissions, permission) {
			return true, nil
		}
	}

	return false, nil
}

func (p *PrefectServiceImpl) GroupInvalidation(ctx context.Context, groupId int64) error {
	p.groupCache.Remove(groupId)
	return nil
}

func (p *PrefectServiceImpl) UserInvalidation(ctx context.Context, userId int64) error {
	p.userCache.Remove(userId)
	return nil
}

func (p *PrefectServiceImpl) ApiKeyToUser(ctx context.Context, apiKey string) (ApiKeyData, error) {
	if started := p.started; !started {
		slog.Error("prefect service not started")
		return ApiKeyData{}, fmt.Errorf("prefect service not started")
	}

	apiKeyData, found := p.apiKeyCache.Get(apiKey)
	if found {
		// Check if the API key is expired
		currentTime := time.Now().Unix()
		if apiKeyData.ExpiresAt < currentTime {
			// API key is expired, remove it from cache
			p.apiKeyCache.Remove(apiKey)
			slog.Info("Api key", "key", apiKey, "expiredAt", apiKeyData.ExpiresAt, "currentTime", currentTime)
			return ApiKeyData{}, fmt.Errorf("api key expired")
		}

		return *apiKeyData, nil
	}

	apiKeyResp, err := p.managementService.UserGetByApiKey(ctx, apiKey)
	if err != nil {
		return ApiKeyData{}, err
	}
	if apiKeyResp.Status != ubstatus.Success {
		return ApiKeyData{}, fmt.Errorf("failed to get api key data: %s", apiKeyResp.Message)
	}

	for _, key := range apiKeyResp.Data.State.ApiKeys {
		// Ensure apiKey starts with key.Id
		if strings.HasPrefix(apiKey, key.Id) {
			apiKeyData = &ApiKeyData{
				UserId:         apiKeyResp.Data.Id,
				Email:          apiKeyResp.Data.State.Email,
				OrganizationId: key.OrganizationId,
				ExpiresAt:      key.ExpiresAt,
			}
			p.apiKeyCache.Put(apiKey, apiKeyData)
			return *apiKeyData, nil
		}
	}
	return ApiKeyData{}, fmt.Errorf("api key not found")
}

func (p *PrefectServiceImpl) main() {
	slog.Info("********** Starting Prefect Service **********")
	filter := evercore.SubscriptionFilter{
		EventTypes: []string{},
	}

	start := evercore.StartFrom{
		Kind: evercore.StartEnd,
	}
	options := evercore.Options{
		BatchSize:    100,
		PollInterval: 1 * time.Second,
		Lease:        30 * time.Second,
	}

	err := p.store.RunEphemeralSubscription(p.ctx, filter, start, options,
		func(ctx context.Context, evs []evercore.SerializedEvent) error {
			for _, e := range evs {
				slog.Info("PrefectService processing event", "eventType", e.EventType, "aggregateId", e.AggregateId, "sequence", e.Sequence)
				switch e.EventType {
				case ev.RoleDeletedEventType,
					ev.RoleUndeletedEventType,
					ev.RolePermissionAddedEventType,
					ev.RolePermissionRemovedEventType:
					p.GroupInvalidation(ctx, e.AggregateId)
				case ev.UserAddedToRoleEventType:
					_, state, err := evercore.DecodeEvent(e)
					if err != nil {
						return fmt.Errorf("failed to decode event: %w", err)
					}

					addedToRoleEvent, ok := state.(UserAddedToRoleEvent)
					if !ok {
						return fmt.Errorf("failed to cast event")
					}
					p.UserInvalidation(ctx, addedToRoleEvent.UserId)

				case ev.UserRemovedFromRoleEventType:
					_, state, err := evercore.DecodeEvent(e)
					if err != nil {
						return fmt.Errorf("failed to decode event: %w", err)
					}

					addedToRoleEvent, ok := state.(UserRemovedFromRoleEvent)
					if !ok {
						return fmt.Errorf("failed to cast event")
					}
					p.UserInvalidation(ctx, addedToRoleEvent.UserId)
				case ev.UserApiKeyDeletedEventType:
					// Invalidate all API keys for the user
					_, es, err := evercore.DecodeEvent(e)
					if err != nil {
						return fmt.Errorf("failed to decode event: %w", err)
					}

					apiKeyEvent, ok := es.(UserApiKeyDeletedEvent)
					if !ok {
						return fmt.Errorf("failed to cast event")
					}
					p.apiKeyCache.Remove(apiKeyEvent.Id)
				}
			}
			return nil
		})

	// If we get here, either the context was cancelled or there was an error.If there
	// was an error, something about our understanding of the system is wrong.
	if err != nil {
		slog.Error("error running ephemeral subscription", "error", err)
	}

	// We stop processing here to avoid giving incorrect permissions.
	p.started = false
}

func (p *PrefectServiceImpl) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancel = cancel
	go p.main()
	p.started = true
	return nil

	/*
	   // No-op


	   err := p.store.RunEphemeralSubscription("ubmanage.PrefectService", filter, func(event evercore.Event) {
	   });

	   	if (err != nil) {
	   		return fmt.Errorf("failed to start event subscription: %w", err)
	   	}
	*/
}

func (p *PrefectServiceImpl) Stop() error {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.started = false
	return nil
}
