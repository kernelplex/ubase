package domain

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/eventstore_events"
	"github.com/kernelplex/ubase/lib/evercoregen"
	aggregates "github.com/kernelplex/ubase/lib/evercoregen/aggregates"
)

type RoleState struct {
	Name        string
	Premissions []int64
}

// evercore:aggregate
type RoleAggregate struct {
	Id       int64
	Sequence int64
	State    *RoleState
}

func NewRoleAggregate() RoleAggregate {
	return RoleAggregate{
		State: &RoleState{},
	}
}

func (a *RoleAggregate) AsAggregate() evercore.Aggregate {
	return a
}

func (a *RoleAggregate) GetId() int64 {
	return a.Id
}

func (a *RoleAggregate) SetId(id int64) {
	a.Id = id
}

func (a *RoleAggregate) GetSequence() int64 {
	return a.Sequence
}

func (a *RoleAggregate) SetSequence(seq int64) {
	a.Sequence = seq
}

func (a *RoleAggregate) GetAggregateType() string {
	return aggregates.RoleAggregateType
}

func (a *RoleAggregate) GetSnapshotFrequency() int64 {
	return 10
}

func (a *RoleAggregate) GetSnapshotState() (*string, error) {
	state := evercore.SerializeToJson(a.State)
	return &state, nil
}

func (a *RoleAggregate) DecodeEvent(ev evercore.SerializedEvent) (evercore.EventState, error) {
	eventState, err := evercoregen.EventDecoder(ev)
	if err != nil {
		return nil, err
	}

	if eventState == nil {
		return nil, fmt.Errorf("unknown event type: %s", ev.EventType)
	}
	return eventState, nil
}

func (a *RoleAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {

	// Print the actual event go type

	switch eventState := eventState.(type) {
	case eventstore_events.RoleCreatedEvent:
		a.State.Name = eventState.Name
	case eventstore_events.RolePermissionAddedEvent:
		a.State.Premissions = append(a.State.Premissions, eventState.PermissionId)
	case eventstore_events.RolePermissionRemovedEvent:
		for i, permissionId := range a.State.Premissions {
			if permissionId == eventState.PermissionId {
				a.State.Premissions = slices.Delete(a.State.Premissions, i, i+1)
				break
			}
		}
	default:
		slog.Info("unknown role event", "event", eventState.GetEventType(), "state", eventState.Serialize())
	}
	return nil
}

func (a *RoleAggregate) ApplySnapshot(snapshot *evercore.Snapshot) error {
	err := evercore.DeserializeFromJson(snapshot.State, &a.State)
	return err
}
