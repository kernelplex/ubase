package ubmanage

import (
	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

// evercore:aggregate
type OrganizationState struct {
	Name       string `json:"name"`
	SystemName string `json:"systemName"`
	Status     string `json:"status"`
}

// evercore:aggregate
type OrganizationAggregate struct {
	evercore.StateAggregate[OrganizationState]
}

// ============================================================================
// Commands
// ============================================================================

// OrganizationCreateCommand is a command to create an organization.
type OrganizationCreateCommand struct {
	Name       string `json:"name"`
	SystemName string `json:"systemName"`
	Status     string `json:"status"`
}

func (c OrganizationCreateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateField("Name", c.Name, true, 0)
	validationTracker.ValidateSystemName("SystemName", &c.SystemName, true)
	validationTracker.ValidateField("Status", c.Status, true, 0)

	return validationTracker.Valid()
}

type OrganizationUpdateCommand struct {
	Id         int64   `json:"id"`
	Name       *string `json:"name"`
	SystemName *string `json:"systemName"`
	Status     *string `json:"status"`
}

func (c OrganizationUpdateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	// Validate required ID field
	validationTracker.ValidateIntMinValue("Id", c.Id, 1)

	validationTracker.ValidateOptionalField("Name", c.Name, 1)
	validationTracker.ValidateOptionalField("SystemName", c.SystemName, 1)
	validationTracker.ValidateOptionalField("Status", c.Status, 1)
	return validationTracker.Valid()
}

// ============================================================================
// Queries
// ============================================================================

// ============================================================================
// Events
// ============================================================================

// evercore:state-event
type OrganizationAddedEvent struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	SystemName string `json:"systemName"`
	Status     string `json:"status"`
}

// evercore:state-event
type OrganizationUpdatedEvent struct {
	Id         int64   `json:"id"`
	Name       *string `json:"name"`
	SystemName *string `json:"systemName"`
	Status     *string `json:"status"`
}
