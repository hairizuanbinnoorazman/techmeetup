package eventstore

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/streaming"

	"github.com/hairizuanbinnoorazman/techmeetup/calendar"
	"github.com/hairizuanbinnoorazman/techmeetup/eventmgmt"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"gopkg.in/yaml.v2"
)

type SubMeetupFeatureControl struct {
	DryRunMode         bool `yaml:"dryrun_mode"`
	StreamyardSync     bool `yaml:"streamyard_sync"`
	CalendarSync       bool `yaml:"calendar_sync"`
	MeetupSync         bool `yaml:"meetup_sync"`
	SlidesSync         bool `yaml:"slides_sync"`
	SheetsReporterSync bool `yaml:"sheets_reporter_sync"`
	PostYoutubeSync    bool `yaml:"post_youtube_sync"`
}

type EventStore struct {
	eventstoreFile      string
	calendarID          string
	calendarEventInvite string
	meetupClient        eventmgmt.Meetup
	logger              logger.Logger
	calendarSvc         calendar.GoogleCalendar
	streamyardSvc       streaming.Streamyard
	featureControl      SubMeetupFeatureControl
}

func NewEventStore(l logger.Logger, eventMgmt eventmgmt.Meetup, calendarSvc calendar.GoogleCalendar, streamyardSvc streaming.Streamyard, eventStoreFile, calendarID, calendarEventInvite string, featureControl SubMeetupFeatureControl) EventStore {
	return EventStore{
		eventstoreFile:      eventStoreFile,
		calendarID:          calendarID,
		calendarEventInvite: calendarEventInvite,
		logger:              l,
		meetupClient:        eventMgmt,
		calendarSvc:         calendarSvc,
		streamyardSvc:       streamyardSvc,
		featureControl:      featureControl,
	}
}

type Event struct {
	TrackEvent      bool         `yaml:"track_event"`
	StartDate       time.Time    `yaml:"start_date"`
	Title           string       `yaml:"title"`
	Description     string       `yaml:"description"`
	IsOnline        bool         `yaml:"is_online"`
	YoutubeLink     string       `yaml:"youtube_link"`
	FacebookLink    string       `yaml:"facebook_link"`
	StreamyardLink  string       `yaml:"streamyard_link"`
	MeetupID        string       `yaml:"meetup_id"`
	CalendarEventID string       `yaml:"calendar_event_id"`
	Organizers      []Organizer  `yaml:"organizers"`
	Agenda          []AgendaItem `yaml:"agenda"`
	// In minutes
	Duration int `yaml:"duration"`
}

func (e Event) Validate() error {
	// Sanity check - skip if any of the fields below are empty or not set right
	if e.Title == "" || e.Description == "" {
		return fmt.Errorf("Empty values found for entry %v", e.Title)
	}
	if e.StartDate.IsZero() {
		return fmt.Errorf("Start date is not set for entry %v", e.Title)
	}
	if e.Duration == 0 {
		return fmt.Errorf("Duration of event is not set for entry %v", e.Title)
	}
	if len(e.Organizers) == 0 {
		return fmt.Errorf("No organizers found for event. Do ensure that these fvalues are set")
	}

	return nil
}

func (e *Event) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		TrackEvent      bool         `yaml:"track_event"`
		StartDate       string       `yaml:"start_date"`
		Title           string       `yaml:"title"`
		Description     string       `yaml:"description"`
		IsOnline        bool         `yaml:"is_online"`
		YoutubeLink     string       `yaml:"youtube_link"`
		FacebookLink    string       `yaml:"facebook_link"`
		StreamyardLink  string       `yaml:"streamyard_link"`
		MeetupID        string       `yaml:"meetup_id"`
		CalendarEventID string       `yaml:"calendar_event_id"`
		Organizers      []Organizer  `yaml:"organizers"`
		Agenda          []AgendaItem `yaml:"agenda"`
		// In minutes
		Duration int `yaml:"duration"`
	}

	var tmp alias
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	startDate, err := time.ParseInLocation("2006-01-02T15:04:05", tmp.StartDate, loc)
	if err != nil {
		return fmt.Errorf("Unable to parse dates: Err: %v", err)
	}

	e.TrackEvent = tmp.TrackEvent
	e.StartDate = startDate
	e.Title = tmp.Title
	e.Description = tmp.Description
	e.IsOnline = tmp.IsOnline
	e.YoutubeLink = tmp.YoutubeLink
	e.FacebookLink = tmp.FacebookLink
	e.StreamyardLink = tmp.StreamyardLink
	e.MeetupID = tmp.MeetupID
	e.CalendarEventID = tmp.CalendarEventID
	e.Organizers = tmp.Organizers
	e.Agenda = tmp.Agenda
	e.Duration = tmp.Duration
	return nil
}

type AgendaItem struct {
	// Type can be either break/speaker
	Type      string        `yaml:"type"`
	StartTime time.Time     `yaml:"start_time"`
	Duration  time.Duration `yaml:"duration"`
	Topic     string        `yaml:"topic"`
	Synopsis  string        `yaml:"synopsis"`
	Speakers  []Speaker     `yaml:"speakers"`
}

type Organizer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Speaker struct {
	Name         string `yaml:"name"`
	Email        string `yaml:"email"`
	Profile      string `yaml:"profile"`
	ProfileImage string `yaml:"profile_image"`
}

func (s EventStore) CheckEvents(filterDate time.Time) error {
	raw, err := ioutil.ReadFile(s.eventstoreFile)
	if err != nil {
		return err
	}
	var data []Event
	yaml.Unmarshal(raw, &data)

	for idx, d := range data {
		if d.TrackEvent == false {
			s.logger.Warningf("CheckEvents is not run for the following event: %v as tracing is not turned on for it", d.Title)
			continue
		}

		err := d.Validate()
		if err != nil {
			s.logger.Errorf("Error in proceeding to evaluate the following event: %v", d.Title)
		}

		tmpEvent := d

		tmpEvent = s.createOrUpdateStreamyard(tmpEvent)
		data[idx].StreamyardLink = tmpEvent.StreamyardLink
		data[idx].YoutubeLink = tmpEvent.YoutubeLink

		tmpEvent = s.createOrUpdateMeetup(tmpEvent)
		data[idx].MeetupID = tmpEvent.MeetupID

		tmpEvent = s.createOrUpdateCalendar(tmpEvent)
		data[idx].CalendarEventID = tmpEvent.CalendarEventID
	}

	rawData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("Unable to marshall the yaml file. Operations may repeat")
	}
	err = ioutil.WriteFile(s.eventstoreFile, rawData, 0755)

	return nil
}

func (s *EventStore) createOrUpdateMeetup(e Event) Event {
	if !s.featureControl.MeetupSync {
		s.logger.Warning("Meetup sync is disabled")
		return e
	}

	if time.Now().After(e.StartDate) {
		s.logger.Warning("Start Date Time is already past. We will no longer track this event for this MeetupSync")
		return e
	}

	if !e.IsOnline {
		s.logger.Warning("Event is not online. We will skip this workflow for now")
		return e
	}

	if e.YoutubeLink == "" || e.StreamyardLink == "" {
		s.logger.Error("Streaming svc not setup and youtube link not available. Cannot setup meetup")
		return e
	}

	if e.MeetupID == "" {
		s.logger.Info("Detected that meetup link is not created for this event. Will recreate")
		resp, err := s.meetupClient.CreateDraftEvent(context.TODO(), eventmgmt.Event{
			StartTime:   e.StartDate,
			Name:        e.Title,
			Description: e.Description,
			IsWebinar:   true,
			WebinarLink: e.YoutubeLink,
			Duration:    120,
		})
		if err != nil {
			s.logger.Errorf("Unable to create draft event. Err: %v", err)
			return e
		}
		e.MeetupID = resp.ID
		return e
	}

	return e
}

func (s *EventStore) createOrUpdateStreamyard(e Event) Event {
	if !s.featureControl.StreamyardSync {
		s.logger.Warning("Streamyard sync is disabled")
		return e
	}

	if time.Now().After(e.StartDate) {
		s.logger.Warning("Start Date Time is already past. We will no longer track this event for this StreamyardSync")
		return e
	}

	if !e.IsOnline {
		s.logger.Warning("Event is not online. We will skip this workflow for now")
		return e
	}

	if e.YoutubeLink != "" && e.StreamyardLink == "" {
		s.logger.Error("Youtube link already available although streamyard link is still not available")
		return e
	}

	if e.StreamyardLink == "" {
		s.logger.Info("No streamyard link available. Begin to create streamyard link")
		streamCreateResp, err := s.streamyardSvc.CreateStream(context.TODO(), e.Title)
		if err != nil {
			s.logger.Error("Unable to create the stream on streamyard. Err: %v", err)
			return e
		}
		streamDestResp, err := s.streamyardSvc.CreateDestination(context.TODO(), "youtube", streamCreateResp)
		if err != nil {
			s.logger.Error("Unable to create the output on streamyard. Err: %v", err)
			return e
		}
		e.StreamyardLink = fmt.Sprintf("https://streamyard.com/%v", streamDestResp.ID)
		for _, dest := range streamDestResp.Destinations {
			if dest.Type == "youtube" {
				e.YoutubeLink = dest.Link
			}
		}
		return e
	}
	return e
}

func (s *EventStore) createOrUpdateCalendar(e Event) Event {
	if !s.featureControl.CalendarSync {
		s.logger.Warning("Calendar sync is disabled")
		return e
	}

	if time.Now().After(e.StartDate) {
		s.logger.Warning("Start Date Time is already past. We will no longer track this event for this CalendarSync")
		return e
	}

	if !e.IsOnline {
		s.logger.Warning("Event is not online. We will skip this workflow for now")
		return e
	}

	if e.StreamyardLink == "" || e.YoutubeLink == "" {
		s.logger.Warning("Streamyard link and youtube link missing. Due to this, we can't aren't able to set the right calendar invite description")
		return e
	}

	if e.CalendarEventID == "" {
		s.logger.Info("Detected that the calendar event id is not set - will create calendar event")
		resp, err := s.calendarSvc.CreateEvent(context.TODO(), s.calendarID, calendar.CalendarEvent{
			StartTime:   e.StartDate,
			EndTime:     e.StartDate.Add(time.Duration(e.Duration) * time.Minute),
			Title:       e.Title,
			Description: fmt.Sprintf(s.calendarEventInvite, e.StreamyardLink),
		})
		if err != nil {
			s.logger.Errorf("Unable to create calendar event. Err: %v", err)
			return e
		}
		e.CalendarEventID = resp.ID
		return e
	}
	return e
}
