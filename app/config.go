package app

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ConfigStore interface {
	Get() (Config, error)
}

func NewBasicConfigStore(f string) BasicConfigStore {
	return BasicConfigStore{
		Filename: f,
	}
}

type BasicConfigStore struct {
	Filename string
}

func (b BasicConfigStore) Get() (Config, error) {
	raw, err := ioutil.ReadFile(b.Filename)
	if err != nil {
		return Config{}, err
	}
	var a Config
	err = yaml.Unmarshal(raw, &a)
	if err != nil {
		return Config{}, err
	}
	return a, nil
}

type Config struct {
	Authstore        string                `yaml:"authstore"`
	EventStoreFile   string                `yaml:"eventstore"`
	Features         Features              `yaml:"features"`
	Meetup           MeetupCredentials     `yaml:"meetup_credentials"`
	Google           GoogleCredentials     `yaml:"google_credentials"`
	Streamyard       StreamyardCredentials `yaml:"streamyard_credentials"`
	SpreadsheetStats string                `yaml:"spreadsheet_stats"`
	CalendarConfig   CalendarConfig        `yaml:"calendar_config"`
	MeetupConfig     MeetupConfig          `yaml:"meetup_config"`
	StreamyardConfig StreamyardConfig      `yaml:"streamyard_config"`
}

type Features struct {
	MeetupSync  MeetupFeatureControl `yaml:"meetup_sync"`
	AuthRefresh FeatureControl       `yaml:"auth_refresh"`
}

type FeatureControl struct {
	Enabled      bool `yaml:"enabled"`
	IdleDuration int  `yaml:"idle_duration"`
}

type MeetupFeatureControl struct {
	Enabled      bool                    `yaml:"enabled"`
	IdleDuration int                     `yaml:"idle_duration"`
	SubFeatures  SubMeetupFeatureControl `yaml:"subfeatures"`
}

type SubMeetupFeatureControl struct {
	StreamyardSync     bool `yaml:"streamyard_sync"`
	CalendarSync       bool `yaml:"calendar_sync"`
	MeetupSync         bool `yaml:"meetup_sync"`
	SlidesSync         bool `yaml:"slides_sync"`
	SheetsReporterSync bool `yaml:"sheets_reporter_sync"`
	PostYoutubeSync    bool `yaml:"post_youtube_sync"`
}

type MeetupCredentials struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

type GoogleCredentials struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Scope        string `yaml:"scope"`
	RedirectURI  string `yaml:"redirect_uri"`
}

type StreamyardCredentials struct {
	CSRFToken string `yaml:"csrf_token"`
	JWT       string `yaml:"jwt"`
}

type CalendarConfig struct {
	CalendarID              string `yaml:"calendar_id"`
	CalendarEventInvitation string `yaml:"calendar_event_invitation"`
}

type MeetupConfig struct {
	MeetupGroup      string            `yaml:"meetup_group"`
	OrganizerMapping map[string]string `yaml:"organizer_mapping"`
}

type StreamyardConfig struct {
	UserID                   string `yaml:"user_id"`
	YoutubeDestination       string `yaml:"youtube_destination"`
	FacebookGroupDestination string `yaml:"facebook_group_destination"`
}
