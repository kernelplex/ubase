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
	switch eventState.(type) {
	case RoleDeletedEvent:
		t.State.Deleted = true
		return nil
	case RoleUndeletedEvent:
		t.State.Deleted = false
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

	validationTracker.ValidateField("Name", c.Name, true, 1)
	validationTracker.ValidateField("SystemName", c.SystemName, true, 1)
	validationTracker.ValidateIntMinValue("OrganizationId", c.OrganizationId, 1)

	return validationTracker.Valid()
}

type RoleUpdateCommand struct {
	Id         int64   `json:"id"`
	Name       *string `json:"name"`
	SystemName *string `json:"systemName"`
}

func (c RoleUpdateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("Id", c.Id, 1)
	validationTracker.ValidateOptionalField("Name", c.Name, 1)
	validationTracker.ValidateOptionalField("SystemName", c.SystemName, 1)

	return validationTracker.Valid()
}

type RoleDeleteCommand struct {
	Id int64 `json:"id"`
}

type RoleUndeleteCommand struct {
	Id int64 `json:"id"`
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
