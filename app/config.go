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
	Authstore string            `yaml:"authstore"`
	Features  Features          `yaml:"features"`
	Meetup    MeetupCredentials `yaml:"meetup_credentials"`
	Google    GoogleCredentials `yaml:"google_credentials"`
}

type Features struct {
	MeetupSync FeatureControl `yaml:"meetup_sync"`
}

type FeatureControl struct {
	Enabled      bool `yaml:"enabled"`
	IdleDuration int  `yaml:"idle_duration"`
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
