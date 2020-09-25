package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"google.golang.org/api/calendar/v3"
)

type CalendarEvent struct {
	ID          string
	StartTime   time.Time
	EndTime     time.Time
	Title       string
	Description string
	Duration    float64
}

type GoogleCalendar struct {
	calendarSvc *calendar.Service
	logger      logger.Logger
}

func NewGoogleCalendar(calendarSvc *calendar.Service, logger logger.Logger) GoogleCalendar {
	return GoogleCalendar{
		calendarSvc: calendarSvc,
		logger:      logger,
	}
}

func (g *GoogleCalendar) GetEvent(ctx context.Context, calendarID, eventID string) (CalendarEvent, error) {
	eventGetReq := g.calendarSvc.Events.Get(calendarID, eventID)
	eventGetReq = eventGetReq.Context(ctx)
	resp, err := eventGetReq.Do()
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("Unable to retrieve calendar event. Err: %v", err)
	}
	startTime, err := time.Parse("2006-01-02T15:04:05-07:00", resp.Start.DateTime)
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("Unable to parse start time of calendar event. Err: %v", err)
	}
	endTime, err := time.Parse("2006-01-02T15:04:05-07:00", resp.End.DateTime)
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("Unable to parse start time of calendar event. Err: %v", err)
	}
	duration := endTime.Sub(startTime)
	return CalendarEvent{
		ID:          resp.Id,
		StartTime:   startTime,
		EndTime:     endTime,
		Title:       resp.Summary,
		Description: resp.Description,
		Duration:    duration.Minutes(),
	}, nil
}

func (g *GoogleCalendar) CreateEvent(ctx context.Context, calendarID string, c CalendarEvent) (CalendarEvent, error) {
	if c.StartTime.IsZero() || c.EndTime.IsZero() || c.Title == "" || c.Description == "" {
		return CalendarEvent{}, fmt.Errorf("Issue with input calendar event")
	}
	e := calendar.Event{
		Summary:     c.Title,
		Description: c.Description,
		Start: &calendar.EventDateTime{
			DateTime: c.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			TimeZone: "Asia/Singapore",
		},
		End: &calendar.EventDateTime{
			DateTime: c.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			TimeZone: "Asia/Singapore",
		},
	}
	g.logger.Infof("%+v", e)
	eventCreateReq := g.calendarSvc.Events.Insert(calendarID, &e)
	eventCreateReq = eventCreateReq.Context(ctx)
	resp, err := eventCreateReq.Do()
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("Unable to create event. Err: %v", err)
	}
	c.ID = resp.Id
	return c, nil
}

func (g *GoogleCalendar) UpdateEvent(ctx context.Context, calendarID string, c CalendarEvent) (CalendarEvent, error) {
	if c.StartTime.IsZero() || c.EndTime.IsZero() || c.Title == "" || c.Description == "" {
		return CalendarEvent{}, fmt.Errorf("Issue with input calendar event")
	}
	e := calendar.Event{
		Summary:     c.Title,
		Description: c.Description,
		Start: &calendar.EventDateTime{
			DateTime: c.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			TimeZone: "Asia/Singapore",
		},
		End: &calendar.EventDateTime{
			DateTime: c.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			TimeZone: "Asia/Singapore",
		},
	}
	g.logger.Infof("%+v", e)
	eventCreateReq := g.calendarSvc.Events.Update(calendarID, c.ID, &e)
	eventCreateReq = eventCreateReq.Context(ctx)
	resp, err := eventCreateReq.Do()
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("Unable to create event. Err: %v", err)
	}
	c.ID = resp.Id
	return c, nil
}
