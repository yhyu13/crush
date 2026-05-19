package pubsub

import "context"

const (
	CreatedEvent EventType = "created"
	UpdatedEvent EventType = "updated"
	DeletedEvent EventType = "deleted"

	CriticReviewTriggeredEvent EventType = "critic.review.triggered"
	CriticVerdictRenderedEvent EventType = "critic.verdict.rendered"
	CriticLoopCompletedEvent   EventType = "critic.loop.completed"
)

// CriticLoopEvent is published when a critic review loop terminates.
type CriticLoopEvent struct {
	SessionID    string
	Iterations   int
	FinalVerdict string
}

type Subscriber[T any] interface {
	Subscribe(context.Context) <-chan Event[T]
}

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of a resource
	Event[T any] struct {
		Type    EventType
		Payload T
	}

	Publisher[T any] interface {
		Publish(EventType, T)
	}
)
