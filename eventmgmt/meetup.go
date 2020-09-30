package eventmgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type Meetup struct {
	logger           logger.Logger
	client           *http.Client
	accessToken      string
	meetupGroup      string
	OrganizerMapping map[string]string
}

type MeetupEventResp struct {
	Created       int64  `json:"created"`
	Description   string `json:"description"`
	Duration      int64  `json:"duration"`
	ID            string `json:"id"`
	IsOnlineEvent bool   `json:"is_online_event"`
	Link          string `json:"link"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Time          int64  `json:"time"`
	HowToFindUs   string `json:"how_to_find_us"`
}

type MeetupEventHost struct {
	ID       string `json:"id"`
	Intro    string `json:"intro"`
	JoinDate int64  `json:"join_date"`
	Name     string `json:"name"`
}

type MeetupFeaturedPhoto struct {
	BaseURL     string `json:"base_url"`
	HighresLink string `json:"highres_link"`
	ID          string `json:"id"`
	PhotoLink   string `json:"photo_link"`
	ThumbLink   string `json:"thumb_link"`
	Type        string `json:"type"`
}

func NewMeetup(logger logger.Logger, client *http.Client, meetupGroup, accessToken string, organizerMapping map[string]string) Meetup {
	return Meetup{
		logger:           logger,
		client:           client,
		accessToken:      accessToken,
		meetupGroup:      meetupGroup,
		OrganizerMapping: organizerMapping,
	}
}

// ListUpcomingEvents list out all upcoming events on meetup page
func (m *Meetup) ListUpcomingEvents(ctx context.Context) ([]Event, error) {
	url := fmt.Sprintf("https://api.meetup.com/%v/events", m.meetupGroup)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := m.client.Do(req)
	if err != nil {
		return []Event{}, fmt.Errorf("Unable to fetch event. Err: %v", err)
	}
	raw, _ := ioutil.ReadAll(resp.Body)
	m.logger.Info(string(raw))
	return []Event{}, nil
}

// ListPastEvents list out all upcoming events on meetup page
func (m *Meetup) ListPastEvents(ctx context.Context) ([]Event, error) {
	rawURL := fmt.Sprintf("https://api.meetup.com/%v/events", m.meetupGroup)
	finalURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return []Event{}, err
	}
	queries := finalURL.Query()
	queries.Add("has_ended", "true")
	queries.Add("status", "past")
	finalURL.RawQuery = queries.Encode()
	m.logger.Info(finalURL.String())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, finalURL.String(), nil)
	resp, err := m.client.Do(req)
	if err != nil {
		return []Event{}, fmt.Errorf("Unable to fetch event. Err: %v", err)
	}
	raw, _ := ioutil.ReadAll(resp.Body)
	m.logger.Info(string(raw))
	return []Event{}, nil
}

func (m *Meetup) GetEvent(ctx context.Context, id string) (Event, error) {
	url := fmt.Sprintf("https://api.meetup.com/%v/events/%v", m.meetupGroup, id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := m.client.Do(req)
	if err != nil {
		return Event{}, fmt.Errorf("Unable to fetch event. Err: %v", err)
	}
	raw, _ := ioutil.ReadAll(resp.Body)
	// m.logger.Info(string(raw))
	var meetupResp MeetupEventResp
	err = json.Unmarshal(raw, &meetupResp)
	if err != nil {
		return Event{}, fmt.Errorf("Error in parsing response from meetup.com. Err: %v", err)
	}
	unixStartTime := meetupResp.Time / 1000
	startTime := time.Unix(unixStartTime, 0)
	return Event{
		ID:          meetupResp.ID,
		StartTime:   startTime,
		Name:        meetupResp.Name,
		Description: meetupResp.Description,
		IsWebinar:   meetupResp.IsOnlineEvent,
		WebinarLink: meetupResp.HowToFindUs,
		Duration:    int(meetupResp.Duration / (1000 * 60)),
	}, nil
}

func (m *Meetup) UploadPhoto(ctx context.Context, eventID, photoFilePath string) (string, error) {
	initialURL := fmt.Sprintf("https://api.meetup.com/%v/events/%v/photos", m.meetupGroup, eventID)
	file, err := os.Open(photoFilePath)
	if err != nil {
		return "", err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	fi, err := file.Stat()
	if err != nil {
		return "", err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("photo", fi.Name())
	if err != nil {
		return "", err
	}
	part.Write(fileContents)
	writer.Close()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, initialURL, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	m.logger.Info(string(rawResp))
	return "", nil
}

func (m *Meetup) CreateDraftEvent(ctx context.Context, e Event) (Event, error) {
	initialURl := fmt.Sprintf("https://api.meetup.com/%v/events", m.meetupGroup)
	data := url.Values{}

	// Modify description
	desc := e.Description + fmt.Sprintf("\n\nYou can watch the live video via the following link:\n%v", e.WebinarLink)
	desc = ConvertDescriptionToMeetupHTML(desc)

	data.Set("announce", "false")
	data.Set("duration", strconv.Itoa(e.Duration*60*1000))
	data.Set("event_hosts", strings.Join(e.Organizers, ","))
	data.Set("name", e.Name)
	data.Set("publish_status", "draft")
	data.Set("time", strconv.Itoa(int(e.StartTime.Unix()*1000)))
	data.Set("venue_id", "online")
	data.Set("description", desc)
	data.Set("how_to_find_us", e.WebinarLink)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, initialURl, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
	m.logger.Info(data.Encode())
	resp, err := m.client.Do(req)
	if err != nil {
		return Event{}, err
	}
	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Event{}, err
	}
	var meetupResp MeetupEventResp
	json.Unmarshal(rawResp, &meetupResp)
	e.ID = meetupResp.ID
	return e, nil
}

func WithFeaturedPhoto(photoID string) func(url.Values) {
	return func(d url.Values) {
		d.Add("featured_photo_id", photoID)
	}
}

func (m *Meetup) UpdateEvent(ctx context.Context, e Event, f ...func(url.Values)) (Event, error) {
	if e.ID == "" {
		return Event{}, fmt.Errorf("No meetup event identified. Update cancelled")
	}
	initialURl := fmt.Sprintf("https://api.meetup.com/%v/events/%v", m.meetupGroup, e.ID)
	data := url.Values{}

	// Modify description
	desc := e.Description + fmt.Sprintf("\n\nYou can watch the live video via the following link:\n%v", e.WebinarLink)
	desc = ConvertDescriptionToMeetupHTML(desc)

	data.Set("announce", "false")
	data.Set("duration", strconv.Itoa(e.Duration*60*1000))
	data.Set("event_hosts", strings.Join(e.Organizers, ","))
	data.Set("name", e.Name)
	data.Set("publish_status", "draft")
	data.Set("time", strconv.Itoa(int(e.StartTime.Unix()*1000)))
	data.Set("venue_id", "online")
	data.Set("description", desc)
	data.Set("how_to_find_us", e.WebinarLink)

	for _, z := range f {
		z(data)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPatch, initialURl, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
	m.logger.Info(data.Encode())
	resp, err := m.client.Do(req)
	if err != nil {
		return Event{}, err
	}
	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Event{}, err
	}
	var meetupResp MeetupEventResp
	json.Unmarshal(rawResp, &meetupResp)
	e.ID = meetupResp.ID
	return e, nil
}
