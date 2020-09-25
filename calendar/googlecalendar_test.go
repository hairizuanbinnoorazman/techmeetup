package calendar

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type calendarConfig struct {
	CalendarID string `yaml:"calendar_id"`
}

type fullConfig struct {
	CalendarConfig calendarConfig `yaml:"calendar_config"`
}

func configHelper(configType string) string {
	raw, _ := ioutil.ReadFile("../config.yaml")
	var s fullConfig
	yaml.Unmarshal(raw, &s)
	if configType == "id" {
		return s.CalendarConfig.CalendarID
	}
	return ""
}

func googleCalendarServiceHelper(t *testing.T) *calendar.Service {
	type googleCreds struct {
		AccessToken string `yaml:"access_token"`
	}
	type creds struct {
		Google googleCreds `yaml:"google"`
	}

	var credentials creds
	rawCreds, _ := ioutil.ReadFile("../authstore.yaml")
	yaml.Unmarshal(rawCreds, &credentials)

	token := oauth2.Token{
		AccessToken: credentials.Google.AccessToken,
		TokenType:   "Bearer",
	}
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(&token))
	aa, _ := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	return aa
}

func timeHelper(timeStr string) time.Time {
	loc, _ := time.LoadLocation("Asia/Singapore")
	timeTime, _ := time.ParseInLocation("2006-01-02T15:04:05", timeStr, loc)
	return timeTime
}

func TestGoogleCalendar_CreateEvent(t *testing.T) {
	type fields struct {
		calendarSvc *calendar.Service
		logger      logger.Logger
	}
	type args struct {
		ctx        context.Context
		calendarID string
		c          CalendarEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CalendarEvent
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				calendarSvc: googleCalendarServiceHelper(t),
				logger:      logger.LoggerForTests{Tester: t},
			},
			args: args{
				ctx:        context.TODO(),
				calendarID: configHelper("id"),
				c: CalendarEvent{
					StartTime: timeHelper("2020-09-24T19:00:00"),
					EndTime:   timeHelper("2020-09-24T20:00:00"),
					Title:     "Test this stuff",
					Description: `This is another test
This is another test

This is another test`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleCalendar{
				calendarSvc: tt.fields.calendarSvc,
				logger:      tt.fields.logger,
			}
			got, err := g.CreateEvent(tt.args.ctx, tt.args.calendarID, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleCalendar.CreateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleCalendar.CreateEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleCalendar_GetEvent(t *testing.T) {
	type fields struct {
		calendarSvc *calendar.Service
		logger      logger.Logger
	}
	type args struct {
		ctx        context.Context
		calendarID string
		eventID    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CalendarEvent
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				calendarSvc: googleCalendarServiceHelper(t),
				logger:      logger.LoggerForTests{Tester: t},
			},
			args: args{
				ctx:        context.TODO(),
				calendarID: configHelper("id"),
				eventID:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleCalendar{
				calendarSvc: tt.fields.calendarSvc,
				logger:      tt.fields.logger,
			}
			got, err := g.GetEvent(tt.args.ctx, tt.args.calendarID, tt.args.eventID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleCalendar.GetEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleCalendar.GetEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleCalendar_UpdateEvent(t *testing.T) {
	type fields struct {
		calendarSvc *calendar.Service
		logger      logger.Logger
	}
	type args struct {
		ctx        context.Context
		calendarID string
		c          CalendarEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CalendarEvent
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				calendarSvc: googleCalendarServiceHelper(t),
				logger:      logger.LoggerForTests{Tester: t},
			},
			args: args{
				ctx:        context.TODO(),
				calendarID: configHelper("id"),
				c: CalendarEvent{
					ID:        "",
					StartTime: timeHelper("2020-09-24T19:00:00"),
					EndTime:   timeHelper("2020-09-24T21:00:00"),
					Title:     "Test this stuff",
					Description: `This is another test
This is another test

I will run another round of tests

This is another test`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleCalendar{
				calendarSvc: tt.fields.calendarSvc,
				logger:      tt.fields.logger,
			}
			got, err := g.UpdateEvent(tt.args.ctx, tt.args.calendarID, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleCalendar.UpdateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleCalendar.UpdateEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
