// Package eventmgmt to handle requests the various event management tools out there
package eventmgmt

import (
	"context"
	"time"
)

type EventMgmt interface {
	ListUpcomingEvents(ctx context.Context) ([]Event, error)
	ListPastEvents(ctx context.Context) ([]Event, error)
	GetEvent(ctx context.Context) (Event, error)
	CreateEvent(ctx context.Context, e Event) (Event, error)
}

type Event struct {
	StartTime   time.Time
	EventName   string
	Description string
	Organizers  []string
	IsWebinar   bool
	WebinarLink string
	Agenda      []AgendaItem
}

type AgendaItem struct {
	// Type can be either break/speaker
	Type           string
	Name           string
	ProfilePicture string
	Profile        string
	StartTime      time.Time
	Duration       time.Duration
	Topic          string
	Synopsis       string
}
