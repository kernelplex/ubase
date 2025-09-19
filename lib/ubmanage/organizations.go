package ubmanage

import (
	"maps"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	ee "github.com/kernelplex/ubase/internal/evercoregen/events"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

// evercore:aggregate
type OrganizationState struct {
	Name       string            `json:"name"`
	SystemName string            `json:"systemName"`
	Status     string            `json:"status"`
	Settings   map[string]string `json:"settings"`
}

// evercore:aggregate
type OrganizationAggregate struct {
	evercore.StateAggregate[OrganizationState]
}

func (t *OrganizationAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {

	switch ev := eventState.(type) {
	case OrganizationSettingsAddedEvent:
		if t.State.Settings == nil {
			t.State.Settings = make(map[string]string)
		}
		maps.Copy(t.State.Settings, ev.Settings)
		return nil
	case OrganizationSettingsRemovedEvent:
		if t.State.Settings == nil {
			return nil
		}
		for _, key := range ev.SettingKeys {
			delete(t.State.Settings, key)
		}
		return nil
	}

	return t.StateAggregate.ApplyEventState(eventState, eventTime, reference)
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
	validationTracker.ValidateSystemName("systemName", c.SystemName, false)
	validationTracker.ValidateOptionalField("status", c.Status, 1)
	return validationTracker.Valid()
}

type OrganizationSettingsAddCommand struct {
	Id       int64             `json:"id"`
	Settings map[string]string `json:"settings"`
}

func (c OrganizationSettingsAddCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	v := ubvalidation.NewValidationTracker()
	v.ValidateIntMinValue("id", c.Id, 1)
	if len(c.Settings) == 0 {
		v.AddIssue("settings", "settings cannot be empty")
	}
	for k := range c.Settings {
		if k == "" {
			v.AddIssue("settings", "settings cannot contain empty keys")
			break
		}
	}
	return v.Valid()
}

type OrganizationSettingsRemoveCommand struct {
	Id          int64    `json:"id"`
	SettingKeys []string `json:"settingKeys"`
}

func (c OrganizationSettingsRemoveCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	v := ubvalidation.NewValidationTracker()
	v.ValidateIntMinValue("id", c.Id, 1)
	if len(c.SettingKeys) == 0 {
		v.AddIssue("settingKeys", "settingKeys cannot be empty")
	}
	for _, k := range c.SettingKeys {
		if k == "" {
			v.AddIssue("settingKeys", "settingKeys cannot contain empty values")
			break
		}
	}
	return v.Valid()
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

// evercore:event
type OrganizationSettingsAddedEvent struct {
	Settings map[string]string `json:"settings"`
}

func (e OrganizationSettingsAddedEvent) GetEventType() string {
	return ee.OrganizationSettingsAddedEventType
}

func (e OrganizationSettingsAddedEvent) Serialize() string {
	return evercore.SerializeToJson(e)
}

// evercore:event
type OrganizationSettingsRemovedEvent struct {
	SettingKeys []string `json:"settingKeys"`
}

func (e OrganizationSettingsRemovedEvent) GetEventType() string {
	return ee.OrganizationSettingsRemovedEventType
}

func (e OrganizationSettingsRemovedEvent) Serialize() string {
	return evercore.SerializeToJson(e)
}
