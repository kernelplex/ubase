package ubase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/kernelplex/ubase/lib/ubalgorithms"
	"github.com/kernelplex/ubase/lib/ubdata"
)

type UserRoles struct {
	UserId int64
	Roles  []int64
}

type RolePermissions struct {
	RoleId      int64
	Permissions []int64
}

type PermissionService interface {
	UserHasPermission(ctx context.Context, userId int64, permission string) (bool, error)
	GetPermissionId(name string) (int64, error)
	GetPermissions(ctx context.Context) (map[string]int64, error)
	GetPermissionsForRole(ctx context.Context, roleId int64) (map[string]bool, error)
	HasPermission(ctx context.Context, userID int64, permission string) (bool, error)
	Warmup(ctx context.Context, allPermissions []string) error
}

type PermissionServiceImpl struct {
	dbadapter         ubdata.DataAdapter
	roleService       RoleService
	userRoleCache     ubalgorithms.LRUCache[int64, UserRoles]
	rolePermissionMap map[int64]map[int64]bool
	permissionIdMap   map[string]int64
	lock              sync.RWMutex
}

func NewPermissionService(dbadapter ubdata.DataAdapter, roleService RoleService) PermissionService {
	userRoleCache := ubalgorithms.NewLRUCache[int64, UserRoles](100)
	return &PermissionServiceImpl{
		dbadapter:         dbadapter,
		userRoleCache:     *userRoleCache,
		permissionIdMap:   make(map[string]int64),
		roleService:       roleService,
		rolePermissionMap: make(map[int64]map[int64]bool),
	}
}

func (p *PermissionServiceImpl) AsPermissionService() PermissionService {
	return p
}

func (p *PermissionServiceImpl) subsribeToEvents() {
	// TODO: Provide a way to subscribe to events
	/*
		p.eventBus.Subscribe("role.permission.added", p.rolePermissionAddedHandler)
		p.eventBus.Subscribe("role.permission.removed", p.rolePermissionRemovedHandler)
		p.eventBus.Subscribe("user.role.added", p.userRoleAddedHandler)
		p.eventBus.Subscribe("user.role.removed", p.userRoleRemovedHandler)
	*/
}

func (p *PermissionServiceImpl) Warmup(ctx context.Context, allPermissions []string) error {
	permissionList, err := p.dbadapter.GetPermissions(ctx)
	if err != nil {
		slog.Error("Failed to get permissions", "error", err)
	}
	for _, permission := range permissionList {
		p.permissionIdMap[permission.Name] = permission.PermissionID
	}

	// Ensure all permissions exist in the map, if not add them to the database.
	for _, permission := range allPermissions {
		if _, ok := p.permissionIdMap[permission]; !ok {
			slog.Info("Permission not found in map, adding to database", "permission", permission)
			id, err := p.dbadapter.CreatePermission(ctx, permission)
			if err != nil {
				return err
			}
			p.permissionIdMap[permission] = id
		}
	}
	// Load all roles and their permissions
	roles, err := p.roleService.GetRoleList(ctx)
	if err != nil {
		slog.Error("Failed to get role list during warmup", "error", err)
		return fmt.Errorf("failed to get role list: %w", err)
	}

	for _, roleId := range roles {
		perms, err := p.GetPermissionsForRole(ctx, roleId)
		if err != nil {
			slog.Error("Failed to get permissions for role during warmup",
				"roleId", roleId,
				"error", err)
			continue
		}

		// Store permissions in rolePermissionMap
		permIds := make(map[int64]bool)
		for permName := range perms {
			if permId, ok := p.permissionIdMap[permName]; ok {
				permIds[permId] = true
			}
		}
		p.rolePermissionMap[roleId] = permIds
	}

	p.subsribeToEvents()

	return nil
}

func (p *PermissionServiceImpl) HasPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	return p.UserHasPermission(ctx, userID, permission)
}

func (p *PermissionServiceImpl) UserHasPermission(ctx context.Context, userId int64, permission string) (bool, error) {

	// Get user roles (from cache or database)
	userRoles, err := p.getUserRoles(ctx, userId)
	if err != nil {
		return false, err
	}

	// Check if any role has the requested permission
	for _, roleId := range userRoles.Roles {
		if perms, ok := p.rolePermissionMap[roleId]; ok {
			if permId, ok := p.permissionIdMap[permission]; ok && perms[permId] {
				return true, nil
			}
		}
	}

	return false, nil
}

func (p *PermissionServiceImpl) GetPermissions(ctx context.Context) (map[string]int64, error) {
	permissions, err := p.dbadapter.GetPermissions(ctx)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]int64)
	for _, permission := range permissions {
		permissionMap[permission.Name] = permission.PermissionID
	}
	return permissionMap, nil
}

func (p *PermissionServiceImpl) GetPermissionId(name string) (int64, error) {
	if id, ok := p.permissionIdMap[name]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("permission not found")
}

func (p *PermissionServiceImpl) getUserRoles(ctx context.Context, userId int64) (UserRoles, error) {
	// First try cache
	if cachedRoles, exists := p.userRoleCache.Get(userId); exists {
		return cachedRoles, nil
	}

	// Not in cache, get from database
	roles, err := p.dbadapter.GetUserRoles(ctx, userId)
	if err != nil {
		return UserRoles{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Update cache
	userRoles := UserRoles{
		UserId: userId,
		Roles:  make([]int64, 0, len(roles)),
	}
	for _, role := range roles {
		userRoles.Roles = append(userRoles.Roles, role.RoleID)
	}
	p.userRoleCache.Put(userId, userRoles)

	return userRoles, nil
}

func (p *PermissionServiceImpl) GetPermissionsForRole(ctx context.Context, roleId int64) (map[string]bool, error) {
	permissions, err := p.dbadapter.GetRolePermissions(ctx, roleId)
	if err != nil {
		slog.Error("failed to get role permissions", "roleId", roleId, "error", err)
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissionMap := make(map[string]bool)
	for _, perm := range permissions {
		permissionMap[perm.Name] = true
	}

	return permissionMap, nil
}
