package domain

import (
	"github.com/opencorex-org/openrevenue/pkg/id"
	"time"
)

type EventTag struct{}
type EventID = id.ID[EventTag]
type Event struct {
	ID            EventID           `json:"id"`
	Action        string            `json:"action"`
	Actor         string            `json:"actor"`
	ResourceType  string            `json:"resourceType"`
	ResourceID    string            `json:"resourceId"`
	OccurredAt    time.Time         `json:"occurredAt"`
	CorrelationID string            `json:"correlationId"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

func New(action, actor, resourceType, resourceID, correlation string, now time.Time) Event {
	return Event{ID: id.New[EventTag](), Action: action, Actor: actor, ResourceType: resourceType, ResourceID: resourceID, CorrelationID: correlation, OccurredAt: now}
}
