package eventstore

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/streaming"

	"github.com/hairizuanbinnoorazman/techmeetup/bannergen"
	"github.com/hairizuanbinnoorazman/techmeetup/calendar"
	"github.com/hairizuanbinnoorazman/techmeetup/eventmgmt"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"gopkg.in/yaml.v2"
)

type SubMeetupFeatureControl struct {
	DryRunMode              bool `yaml:"dryrun_mode"`
	StreamyardSync          bool `yaml:"streamyard_sync"`
	CalendarSync            bool `yaml:"calendar_sync"`
	MeetupSync              bool `yaml:"meetup_sync"`
	SlidesSync              bool `yaml:"slides_sync"`
	SheetsReporterSync      bool `yaml:"sheets_reporter_sync"`
	PostYoutubeSync         bool `yaml:"post_youtube_sync"`
	GenerateBannerImageSync bool `yaml:"generate_banner_image_sync"`
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
	TrackEvent             bool         `yaml:"track_event"`
	GenerateBannerImage    bool         `yaml:"generate_banner_image"`
	UpdateImageOnPlatforms bool         `yaml:"update_image_on_platforms"`
	FeaturedImagePath      string       `yaml:"featured_image_path"`
	StartDate              time.Time    `yaml:"start_date"`
	Title                  string       `yaml:"title"`
	Description            string       `yaml:"description"`
	IsOnline               bool         `yaml:"is_online"`
	IsPublic               bool         `yaml:"is_public"`
	YoutubeLink            string       `yaml:"youtube_link"`
	FacebookLink           string       `yaml:"facebook_link"`
	StreamyardID           string       `yaml:"streamyard_id"`
	MeetupID               string       `yaml:"meetup_id"`
	CalendarEventID        string       `yaml:"calendar_event_id"`
	Organizers             []Organizer  `yaml:"organizers"`
	Agenda                 []AgendaItem `yaml:"agenda"`
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
		TrackEvent             bool         `yaml:"track_event"`
		GenerateBannerImage    bool         `yaml:"generate_banner_image"`
		UpdateImageOnPlatforms bool         `yaml:"update_image_on_platforms"`
		FeaturedImagePath      string       `yaml:"featured_image_path"`
		StartDate              string       `yaml:"start_date"`
		Title                  string       `yaml:"title"`
		Description            string       `yaml:"description"`
		IsOnline               bool         `yaml:"is_online"`
		IsPublic               bool         `yaml:"is_public"`
		YoutubeLink            string       `yaml:"youtube_link"`
		FacebookLink           string       `yaml:"facebook_link"`
		StreamyardID           string       `yaml:"streamyard_id"`
		MeetupID               string       `yaml:"meetup_id"`
		CalendarEventID        string       `yaml:"calendar_event_id"`
		Organizers             []Organizer  `yaml:"organizers"`
		Agenda                 []AgendaItem `yaml:"agenda"`
		// In minutes
		Duration int `yaml:"duration"`
	}

	var tmp alias
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	startDate, err := time.ParseInLocation("2006-01-02T15:04:05Z07:00", tmp.StartDate, loc)
	if err != nil {
		return fmt.Errorf("Unable to parse dates: Err: %v", err)
	}

	e.TrackEvent = tmp.TrackEvent
	e.GenerateBannerImage = tmp.GenerateBannerImage
	e.FeaturedImagePath = tmp.FeaturedImagePath
	e.UpdateImageOnPlatforms = tmp.UpdateImageOnPlatforms
	e.StartDate = startDate
	e.Title = tmp.Title
	e.Description = tmp.Description
	e.IsPublic = tmp.IsPublic
	e.IsOnline = tmp.IsOnline
	e.YoutubeLink = tmp.YoutubeLink
	e.FacebookLink = tmp.FacebookLink
	e.StreamyardID = tmp.StreamyardID
	e.MeetupID = tmp.MeetupID
	e.CalendarEventID = tmp.CalendarEventID
	e.Organizers = tmp.Organizers
	e.Agenda = tmp.Agenda
	e.Duration = tmp.Duration
	return nil
}

type AgendaItem struct {
	// Type can be either break/speaker
	Type     string    `yaml:"type"`
	Topic    string    `yaml:"topic"`
	Synopsis string    `yaml:"synopsis"`
	Speakers []Speaker `yaml:"speakers"`
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
	err = yaml.Unmarshal(raw, &data)
	if err != nil {
		return fmt.Errorf("Issue with unmarshalling data. Err: %v", err)
	}

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

		tmpEvent = s.createOrUpdateYoutubeStreamyard(tmpEvent)
		data[idx].StreamyardID = tmpEvent.StreamyardID
		data[idx].YoutubeLink = tmpEvent.YoutubeLink

		tmpEvent = s.createOrUpdateMeetup(tmpEvent)
		data[idx].MeetupID = tmpEvent.MeetupID

		tmpEvent = s.createOrUpdateCalendar(tmpEvent)
		data[idx].CalendarEventID = tmpEvent.CalendarEventID

		// Cleanup for platform updates
		data[idx].UpdateImageOnPlatforms = false
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

	if e.YoutubeLink == "" || e.StreamyardID == "" {
		s.logger.Error("Streaming svc not setup and youtube link not available. Cannot setup meetup")
		return e
	}

	if e.FeaturedImagePath == "" {
		s.logger.Error("No featured image provided. Please provide it")
		return e
	}

	if e.MeetupID == "" {
		s.logger.Info("Detected that meetup link is not created for this event. Will recreate")
		meetupOrganizers := []string{}
		for _, o := range s.meetupClient.OrganizerMapping {
			meetupOrganizers = append(meetupOrganizers, o)
		}
		resp, err := s.meetupClient.CreateDraftEvent(context.TODO(), eventmgmt.Event{
			StartTime:   e.StartDate,
			Name:        e.Title,
			Description: e.Description,
			IsWebinar:   true,
			IsPublic:    e.IsPublic,
			WebinarLink: e.YoutubeLink,
			Duration:    120,
			Organizers:  meetupOrganizers,
		})
		if err != nil {
			s.logger.Errorf("Unable to create draft event. Err: %v", err)
			return e
		}
		e.MeetupID = resp.ID
		photoID, err := s.meetupClient.UploadPhoto(context.TODO(), resp.ID, e.FeaturedImagePath)
		if err != nil {
			s.logger.Errorf("Unable to upload photo. Err: %v", err)
			return e
		}
		_, err = s.meetupClient.UpdateEvent(context.TODO(), resp, eventmgmt.WithFeaturedPhoto(photoID))
		if err != nil {
			s.logger.Errorf("Unable to update event with featured photo")
			return e
		}
		return e
	}

	meetupEvent, err := s.meetupClient.GetEvent(context.TODO(), e.MeetupID)
	if err != nil {
		s.logger.Errorf("Unable to retrieve event details from meetup. Err: %v MeetupID: %v", err, e.MeetupID)
		return e
	}
	parsedDesc := eventmgmt.ConvertMeetupHTMLToText(meetupEvent.Description)
	s.logger.Infof(`Change Detection:
  Description: %v
  Title: %v
  UpdateImageOnPlatforms: %v
`, eventmgmt.AppendYoutubeLinktoDesc(e.Description, e.YoutubeLink) != parsedDesc, meetupEvent.Name != e.Title, e.UpdateImageOnPlatforms)
	if (eventmgmt.AppendYoutubeLinktoDesc(e.Description, e.YoutubeLink) != parsedDesc || meetupEvent.Name != e.Title) && !e.UpdateImageOnPlatforms {
		s.logger.Info("Begin update of meetup - no image update needed")
		meetupEvent.Description = e.Description
		meetupEvent.Name = e.Title
		meetupEvent.StartTime = e.StartDate
		meetupEvent.IsPublic = e.IsPublic
		_, err = s.meetupClient.UpdateEvent(context.TODO(), meetupEvent)
		if err != nil {
			s.logger.Errorf("Do check functionality to make sure all is working as expected. Err: %v", err)
			return e
		}
		s.logger.Info("End update of meetup - no image update needed")
		return e
	}

	if e.UpdateImageOnPlatforms {
		s.logger.Info("Begin update of meetup - with image update needed")
		meetupEvent.Description = e.Description
		meetupEvent.Name = e.Title
		photoID, err := s.meetupClient.UploadPhoto(context.TODO(), meetupEvent.ID, e.FeaturedImagePath)
		if err != nil {
			s.logger.Errorf("Unable to upload photo. Err: %v", err)
			return e
		}
		_, err = s.meetupClient.UpdateEvent(context.TODO(), meetupEvent, eventmgmt.WithFeaturedPhoto(photoID))
		if err != nil {
			s.logger.Errorf("Unable to update event with featured photo")
			return e
		}
		s.logger.Info("End update of meetup - with image update needed")
	}

	return e
}

func (s *EventStore) createOrUpdateYoutubeStreamyard(e Event) Event {
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

	if e.YoutubeLink != "" && e.StreamyardID == "" {
		s.logger.Error("Youtube link already available although streamyard link is still not available")
		return e
	}

	if e.FeaturedImagePath == "" {
		s.logger.Error("No featured image provided. Please provide it")
		return e
	}

	if e.StreamyardID == "" {
		s.logger.Info("No streamyard link available. Begin to create streamyard link")
		streamCreateResp, err := s.streamyardSvc.CreateStream(context.TODO(), e.Title)
		if err != nil {
			s.logger.Error("Unable to create the stream on streamyard. Err: %v", err)
			return e
		}
		streamCreateResp.StartDate = e.StartDate
		streamCreateResp.ImagePath = e.FeaturedImagePath
		streamCreateResp.Description = e.Description
		streamCreateResp.IsPublic = e.IsPublic
		s.logger.Infof("Created streamyard: %+v", streamCreateResp)
		streamDestResp, err := s.streamyardSvc.CreateDestination(context.TODO(), "youtube", streamCreateResp)
		if err != nil {
			s.logger.Error("Unable to create the output on streamyard. Err: %v", err)
			return e
		}
		e.StreamyardID = streamDestResp.ID
		for _, dest := range streamDestResp.Destinations {
			if dest.Type == "youtube" {
				e.YoutubeLink = dest.Link
			}
		}
		s.logger.Infof("Create of streamyard youtube stream complete. Event: %v", e)
		return e
	}

	streamyardStream, err := s.streamyardSvc.GetStream(context.TODO(), e.StreamyardID)
	if err != nil {
		s.logger.Errorf("Unable to retrieve stream from streamyard. %v", err)
		return e
	}

	s.logger.Infof("Change Detection:\n  DescriptionChange: %v\n  TitleChange: %v\n  ImageChange: %v", streamyardStream.Description != e.Description, streamyardStream.Name != e.Title, e.UpdateImageOnPlatforms)
	if streamyardStream.Description != e.Description || streamyardStream.Name != e.Title || e.UpdateImageOnPlatforms {
		s.logger.Info("Begin update of streamyard")
		streamyardStream.Description = e.Description
		streamyardStream.Name = e.Title
		streamyardStream.ImagePath = e.FeaturedImagePath
		streamyardStream.StartDate = e.StartDate
		streamyardStream.IsPublic = e.IsPublic
		s.streamyardSvc.UpdateDestination(context.TODO(), "youtube", streamyardStream, e.UpdateImageOnPlatforms)

		if streamyardStream.Name != e.Title {
			err = s.streamyardSvc.UpdateStream(context.TODO(), e.StreamyardID, e.Title)
			if err != nil {
				s.logger.Errorf("Unable to update stream. Err: %v", err)
			}
		}
		s.logger.Info("End update of streamyard")
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

	if e.StreamyardID == "" || e.YoutubeLink == "" {
		s.logger.Warning("Streamyard link and youtube link missing. Due to this, we can't aren't able to set the right calendar invite description")
		return e
	}

	if e.CalendarEventID == "" {
		s.logger.Info("Detected that the calendar event id is not set - will create calendar event")
		zz := make(map[string]bool)
		for _, organizer := range e.Organizers {
			zz[organizer.Email] = true
		}
		for _, agenda := range e.Agenda {
			for _, speaker := range agenda.Speakers {
				zz[speaker.Email] = true
			}
		}
		yy := []string{}
		for k := range zz {
			yy = append(yy, k)
		}

		resp, err := s.calendarSvc.CreateEvent(context.TODO(), s.calendarID, calendar.CalendarEvent{
			StartTime:   e.StartDate,
			EndTime:     e.StartDate.Add(time.Duration(e.Duration) * time.Minute),
			Title:       e.Title,
			Description: fmt.Sprintf(s.calendarEventInvite, fmt.Sprintf("https://streamyard.com/%v", e.StreamyardID)),
			Attendees:   yy,
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

func (s *EventStore) createWebinarBannerImage(e Event) Event {
	if !s.featureControl.GenerateBannerImageSync {
		s.logger.Warning("Generate Banner Image sync is disabled")
		return e
	}

	if time.Now().After(e.StartDate) {
		s.logger.Warning("Start Date Time is already past. We will no longer track this event for this Autogenerating banner image")
		return e
	}

	if !e.IsOnline {
		s.logger.Warning("Event is not online. We will skip this workflow for now")
		return e
	}

	if !e.GenerateBannerImage {
		s.logger.Warning("Generating Banner Image is disabled")
		return e
	}

	if e.Title == "" || e.StartDate.IsZero() || e.Duration == 0 {
		s.logger.Warning("Missing title or start date is zero")
		return e
	}

	items := strings.Split(e.Title, "-")
	if len(items) != 2 {
		s.logger.Warning("Unable to split title to series name + title name. Title: %v", e.Title)
		return e
	}

	endTime := e.StartDate.Add(time.Duration(e.Duration) * time.Minute)
	seriesName := strings.Trim(items[0], " ")
	webinarTitle := strings.Trim(items[1], " ")
	formattedTime := fmt.Sprintf("%v to %v", e.StartDate.Format("2 January 2006 - 15:04pm"), endTime.Format("15:04pm"))
	outputPath := time.Now().Format("20060102_1504") + ".png"

	var streamData streaming.Stream
	var err error
	if e.StreamyardID != "" {
		streamData, err = s.streamyardSvc.GetStream(context.TODO(), e.StreamyardID)
		if err != nil {
			s.logger.Errorf("Unable to receive stream. Err: %v", err)
			return e
		}
	}

	// Handle the following case:
	// When streamyardID is not defined - new event being created
	// When title on streamyard is not the same as the one being identified as the one we have
	if streamData.Name != e.Title || e.StreamyardID == "" {
		err = bannergen.Generate_banner(outputPath, seriesName, webinarTitle, formattedTime)
		if err != nil {
			s.logger.Errorf("Generating banner failed.\n  Err: %v\n  seriesName: %v\n  webinarTitle: %v\n  formattedTime: %v", err, seriesName, webinarTitle, formattedTime)
			return e
		}
	}

	e.FeaturedImagePath = outputPath
	e.UpdateImageOnPlatforms = true
	return e
}
