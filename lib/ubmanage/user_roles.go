package ubmanage

import (
	"fmt"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	events "github.com/kernelplex/ubase/internal/evercoregen/events"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

// evercore:aggregate
type UserRolesAggregate struct {
	Id       int64
	Sequence int64
}

type test struct {
	evercore.StateAggregate[UserRolesAggregate]
}

// Note the UserRolesAggregsate is just a receiver for the related events. It does not record state.
func NewUserRolesAggregate() evercore.Aggregate {
	return &UserRolesAggregate{}
}

func (t *UserRolesAggregate) GetAggregateType() string {
	return "UserRolesAggregate"
}

func (t *UserRolesAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {
	return nil
}

func (t *UserRolesAggregate) GetSequence() int64 {
	return t.Sequence
}

func (t *UserRolesAggregate) SetId(id int64) {
	t.Id = id
}

func (t *UserRolesAggregate) GetId() int64 {
	return t.Id
}

func (t *UserRolesAggregate) SetSequence(seq int64) {
	t.Sequence = seq
}

func (t *UserRolesAggregate) GetSnapshotFrequency() int64 {
	return 0
}

func (t *UserRolesAggregate) GetSnapshotState() (*string, error) {
	return nil, fmt.Errorf("Cannot snapshot UserRolesAggregate")
}

func (t *UserRolesAggregate) ApplySnapshot(snapshot *evercore.Snapshot) error {
	return fmt.Errorf("Cannot apply snapshot to UserRolesAggregate")
}

// ============================================================================
// Commands
// ============================================================================

type UserAddToRoleCommand struct {
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func (c UserAddToRoleCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("UserId", c.UserId, 1)
	validationTracker.ValidateIntMinValue("RoleId", c.RoleId, 1)

	return validationTracker.Valid()
}

type UserRemoveFromRoleCommand struct {
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func (c UserRemoveFromRoleCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("UserId", c.UserId, 1)
	validationTracker.ValidateIntMinValue("RoleId", c.RoleId, 1)

	return validationTracker.Valid()
}

// ============================================================================
// Events
// ============================================================================

// evercore:event
type UserAddedToRoleEvent struct {
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func (a UserAddedToRoleEvent) GetEventType() string {
	return events.UserAddedToRoleEventType
}

func (a UserAddedToRoleEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserRemovedFromRoleEvent struct {
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func (a UserRemovedFromRoleEvent) GetEventType() string {
	return events.UserRemovedFromRoleEventType
}

func (a UserRemovedFromRoleEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}
