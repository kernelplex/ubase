package ubmanage

import (
	"context"
	"fmt"
	"slices"

	"github.com/kernelplex/ubase/lib/ubalgorithms"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

type PrefectService interface {
	UserBelongsToRole(ctx context.Context, userId int64, roleId int64) (bool, error)

	UserHasPermission(ctx context.Context, userId int64, orgId int64, permission string) (bool, error)

	GroupInvalidation(ctx context.Context, roleId int64) error

	UserInvalidation(ctx context.Context, userId int64) error
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
}

func NewPrefectService(
	managementService ManagementService,
	userCacheSize int,
	groupCacheSize int,
) PrefectService {
	userCache := ubalgorithms.NewLRUCache[int64, *UserData](userCacheSize)
	groupCache := ubalgorithms.NewLRUCache[int64, *GroupPermissions](groupCacheSize)
	return &PrefectServiceImpl{
		managementService: managementService,
		userCache:         userCache,
		groupCache:        groupCache,
	}
}

func (p *PrefectServiceImpl) getUserData(ctx context.Context, userId int64) (*UserData, error) {
	userData, found := p.userCache.Get(userId)
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
