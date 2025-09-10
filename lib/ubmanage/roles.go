package ubmanage

import (
	"time"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

type RoleAggregate struct {
	evercore.StateAggregate[RoleState]
}

func (t *RoleAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {

	switch ev := eventState.(type) {
	case RoleDeletedEvent:
		t.State.Deleted = true
		return nil
	case RoleUndeletedEvent:
		t.State.Deleted = false
		return nil

	case RolePermissionAddedEvent:
		if t.State.Permissions == nil {
			t.State.Permissions = make([]string, 0, 1)
		}
		t.State.Permissions = append(t.State.Permissions, ev.Permission)
		return nil

	case RolePermissionRemovedEvent:
		if t.State.Permissions == nil {
			return nil
		}
		for i, p := range t.State.Permissions {
			if p == ev.Permission {
				t.State.Permissions = append(t.State.Permissions[:i], t.State.Permissions[i+1:]...)
				return nil
			}
		}
		return nil
	default:
		return t.StateAggregate.ApplyEventState(eventState, eventTime, reference)
	}

}

type RoleState struct {
	Name           string `json:"name"`
	OrganizationId int64  `json:"organizationId"`
	SystemName     string `json:"systemName"`
	Deleted        bool   `json:"deleted"`
	Permissions    []string
}

// ============================================================================
// Commands
// ============================================================================

// RoleCreateCommand is a command to create a role.
type RoleCreateCommand struct {
	Name           string `json:"name"`
	SystemName     string `json:"systemName"`
	OrganizationId int64  `json:"organizationId"`
}

func (c RoleCreateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateField("name", c.Name, true, 1)
	validationTracker.ValidateSystemName("systemName", &c.SystemName, true)
	validationTracker.ValidateIntMinValue("organizationId", c.OrganizationId, 1)

	return validationTracker.Valid()
}

type RoleUpdateCommand struct {
	Id         int64   `json:"id"`
	Name       *string `json:"name"`
	SystemName *string `json:"systemName"`
}

func (c RoleUpdateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("id", c.Id, 1)
	validationTracker.ValidateOptionalField("name", c.Name, 1)
	validationTracker.ValidateOptionalField("systemName", c.SystemName, 1)

	return validationTracker.Valid()
}

type RoleDeleteCommand struct {
	Id int64 `json:"id"`
}

type RoleUndeleteCommand struct {
	Id int64 `json:"id"`
}

type RolePermissionAddCommand struct {
	Id         int64  `json:"id"`
	Permission string `json:"permission"`
}

func (c RolePermissionAddCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateField("permission", c.Permission, true, 1)

	return validationTracker.Valid()
}

type RolePermissionRemoveCommand struct {
	Id         int64  `json:"id"`
	Permission string `json:"permission"`
}

func (c RolePermissionRemoveCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("id", c.Id, 1)
	validationTracker.ValidateField("permission", c.Permission, true, 1)

	return validationTracker.Valid()
}

// ============================================================================
// Events
// ============================================================================

// evercore:state-event
type RoleCreatedEvent struct {
	Id             int64  `json:"id"`
	OrganizationId int64  `json:"organizationId"`
	Name           string `json:"name"`
	SystemName     string `json:"systemName"`
}

// evercore:state-event
type RoleUpdatedEvent struct {
	Id         int64   `json:"id"`
	Name       *string `json:"name"`
	SystemName *string `json:"systemName"`
}

// evercore:event
type RoleDeletedEvent struct {
}

func (a RoleDeletedEvent) GetEventType() string {
	return "RoleDeletedEvent"
}

func (a RoleDeletedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type RoleUndeletedEvent struct {
}

func (a RoleUndeletedEvent) GetEventType() string {
	return "RoleUndeletedEvent"
}

func (a RoleUndeletedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type RolePermissionAddedEvent struct {
	Permission string `json:"permission"`
}

func (a RolePermissionAddedEvent) GetEventType() string {
	return "RolePermissionAddedEvent"
}

func (a RolePermissionAddedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type RolePermissionRemovedEvent struct {
	Permission string `json:"permission"`
}

func (a RolePermissionRemovedEvent) GetEventType() string {
	return "RolePermissionRemovedEvent"
}

func (a RolePermissionRemovedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}
