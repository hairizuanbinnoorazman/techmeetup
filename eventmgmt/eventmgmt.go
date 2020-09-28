// Package eventmgmt to handle requests the various event management tools out there
package eventmgmt

import (
	"context"
	"regexp"
	"strings"
	"time"
)

type EventMgmt interface {
	ListUpcomingEvents(ctx context.Context) ([]Event, error)
	ListPastEvents(ctx context.Context) ([]Event, error)
	GetEvent(ctx context.Context) (Event, error)
	CreateEvent(ctx context.Context, e Event) (Event, error)
}

type Event struct {
	ID          string
	StartTime   time.Time
	Name        string
	Description string
	IsWebinar   bool
	WebinarLink string
	// Meetup organizer
	Organizers []string
	// Time in minutes
	// This is temporarily set
	Duration int
}

func NewEvent(name, description, startTime string) (Event, error) {
	loc, _ := time.LoadLocation("Asia/Singapore")
	zz, err := time.ParseInLocation("2006-01-02T15:04:05", startTime, loc)
	if err != nil {
		return Event{}, err
	}
	return Event{
		StartTime:   zz,
		Name:        name,
		Description: description,
		IsWebinar:   true,
		Duration:    120,
	}, nil
}

func convertDescriptionToMeetupHTML(desc string) string {
	re := regexp.MustCompile(`((http|https):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?)`)
	output := re.ReplaceAllString(desc, "<a href=\"$1\">$1</a>")
	output = strings.ReplaceAll(output, "\n", "</br>")
	return output
}
