package eventmgmt

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type meetup struct {
	logger       logger.Logger
	client       *http.Client
	refreshToken string
	accessToken  string
	expiryTime   int
	meetupGroup  string
}

func NewMeetup(logger logger.Logger, client *http.Client, meetupGroup, refreshToken, accessToken string, expiryTime int) meetup {
	return meetup{
		logger:       logger,
		client:       client,
		refreshToken: refreshToken,
		accessToken:  accessToken,
		expiryTime:   expiryTime,
		meetupGroup:  meetupGroup,
	}
}

// ListUpcomingEvents list out all upcoming events on meetup page
func (m *meetup) ListUpcomingEvents(ctx context.Context) ([]Event, error) {
	return []Event{}, nil
}

func (m *meetup) ListPastEvents(ctx context.Context) ([]Event, error) {
	return []Event{}, nil
}

func (m *meetup) GetEvent(ctx context.Context, id string) (Event, error) {
	url := fmt.Sprintf("https://api.meetup.com/%v/events/%v", m.meetupGroup, id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := m.client.Do(req)
	if err != nil {
		return Event{}, fmt.Errorf("Unable to fetch event. Err: %v", err)
	}
	raw, _ := ioutil.ReadAll(resp.Body)
	m.logger.Info(string(raw))
	return Event{}, nil
}

func (m *meetup) CreateEvent(ctx context.Context, e Event) (Event, error) {
	return Event{}, nil
}
