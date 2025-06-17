package eventstore_events

import (
	evercore "github.com/kernelplex/evercore/base"
	events "github.com/kernelplex/ubase/lib/evercoregen/events"
)

// evercore:event
type RoleCreatedEvent struct {
	Name string
}

func (a RoleCreatedEvent) GetEventType() string {
	return events.RoleCreatedEventType
}
func (a RoleCreatedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type RolePermissionAddedEvent struct {
	PermissionId int64
}

func (a RolePermissionAddedEvent) GetEventType() string {
	return events.RolePermissionAddedEventType
}
func (a RolePermissionAddedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type RolePermissionRemovedEvent struct {
	PermissionId int64
}

func (a RolePermissionRemovedEvent) GetEventType() string {
	return events.RolePermissionRemovedEventType
}

func (a RolePermissionRemovedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}
