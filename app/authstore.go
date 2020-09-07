package app

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type AuthStore interface {
	StoreMeetupToken(m MeetupToken) error
	GetMeetupToken() (MeetupToken, error)
	StoreGoogleToken(g GoogleToken) error
	GetGoogleToken() (GoogleToken, error)
}

type MeetupToken struct {
	RefreshToken string `yaml:"refresh_token"`
	AccessToken  string `yaml:"access_token"`
}

type GoogleToken struct {
	RefreshToken string `yaml:"refresh_token"`
	AccessToken  string `yaml:"access_token"`
}

type BasicAuthStore struct {
	filePath string
}

func NewBasicAuthStore(f string) BasicAuthStore {
	return BasicAuthStore{
		filePath: f,
	}
}

type internalAuthStore struct {
	Meetup MeetupToken `yaml:"meetup"`
	Google GoogleToken `yaml:"google"`
}

func (b *BasicAuthStore) StoreMeetupToken(m MeetupToken) error {
	raw, err := ioutil.ReadFile(b.filePath)
	if err != nil {
		return err
	}
	var a internalAuthStore
	yaml.Unmarshal(raw, &a)
	a.Meetup = m
	newRaw, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(b.filePath, newRaw, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *BasicAuthStore) GetMeetupToken() (MeetupToken, error) {
	raw, err := ioutil.ReadFile(b.filePath)
	if err != nil {
		return MeetupToken{}, err
	}
	var a internalAuthStore
	yaml.Unmarshal(raw, &a)
	return a.Meetup, nil
}

func (b *BasicAuthStore) StoreGoogleToken(g GoogleToken) error {
	raw, err := ioutil.ReadFile(b.filePath)
	if err != nil {
		return err
	}
	var a internalAuthStore
	yaml.Unmarshal(raw, &a)
	a.Google = g
	newRaw, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(b.filePath, newRaw, 0644)
	if err != nil {
		return err
	}
	return nil
}
func (b *BasicAuthStore) GetGoogleToken() (GoogleToken, error) {
	raw, err := ioutil.ReadFile(b.filePath)
	if err != nil {
		return GoogleToken{}, err
	}
	var a internalAuthStore
	yaml.Unmarshal(raw, &a)
	return a.Google, nil
}
