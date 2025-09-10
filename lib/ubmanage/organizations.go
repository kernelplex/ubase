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

	validationTracker.ValidateField("name", c.Name, true, 0)
	validationTracker.ValidateSystemName("systemName", &c.SystemName, true)
	validationTracker.ValidateField("status", c.Status, true, 0)

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
	validationTracker.ValidateIntMinValue("id", c.Id, 1)

	validationTracker.ValidateOptionalField("name", c.Name, 1)
	validationTracker.ValidateOptionalField("systemName", c.SystemName, 1)
	validationTracker.ValidateOptionalField("status", c.Status, 1)
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
