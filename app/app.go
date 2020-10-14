// Package app wraps all the functionality and keeps it out of cmd package
package app

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/streaming"

	calendarZ "github.com/hairizuanbinnoorazman/techmeetup/calendar"
	"github.com/hairizuanbinnoorazman/techmeetup/eventmgmt"
	"github.com/hairizuanbinnoorazman/techmeetup/eventstore"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type App struct {
	configStore ConfigStore
	logger      logger.Logger
	// Internal controls
	authStore           AuthStore
	config              Config
	eventMgr            eventmgmt.EventMgmt
	googleAuth          GoogleAuthRefresher
	meetupAuth          MeetupAuthRefresher
	eventMgmtTicker     *time.Ticker
	authRefresherTicker *time.Ticker
	calendarSvc         calendarZ.GoogleCalendar
}

func NewApp(c ConfigStore, l logger.Logger) App {
	return App{
		configStore: c,
		logger:      l,
	}
}

// If any error happens - all features will be disabled
func (a *App) Initialize() {
	a.logger.Info("Initialize application")
	defer a.logger.Info("Application Initialization Complete")
	a.config, _ = a.configStore.Get()
	authstore := NewBasicAuthStore(a.config.Authstore)
	a.authStore = &authstore
	a.eventMgmtTicker = time.NewTicker(time.Duration(a.config.Features.MeetupSync.IdleDuration) * time.Second)
	if !a.config.Features.MeetupSync.Enabled {
		a.eventMgmtTicker.Stop()
	}
	a.authRefresherTicker = time.NewTicker(60 * time.Second)
	a.RerunAuth()
}

func (a *App) RerunAuth() {
	a.config, _ = a.configStore.Get()
	authstore := NewBasicAuthStore(a.config.Authstore)
	a.googleAuth = GoogleAuthRefresher{
		client:       http.DefaultClient,
		logger:       a.logger,
		authStore:    &authstore,
		clientID:     a.config.Google.ClientID,
		clientSecret: a.config.Google.ClientSecret,
	}
	a.meetupAuth = MeetupAuthRefresher{
		client:       http.DefaultClient,
		logger:       a.logger,
		authStore:    &authstore,
		clientID:     a.config.Meetup.ClientID,
		clientSecret: a.config.Meetup.ClientSecret,
	}

	m, _ := authstore.GetGoogleToken()
	token := oauth2.Token{
		AccessToken: m.AccessToken,
		TokenType:   "Bearer",
	}
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(&token))
	aa, _ := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	a.calendarSvc = calendarZ.NewGoogleCalendar(aa, a.logger)
}

func (a *App) Run(notifyConfigChange chan bool, interrupts chan os.Signal) {
	a.logger.Info("Begin running sync loop")
	defer a.logger.Info("Sync loop ends")
	go server(a.config, a.authStore, notifyConfigChange)
	for {
		select {
		case <-notifyConfigChange:
			a.logger.Warning("Begin running initialization")
			time.Sleep(1 * time.Second)
			a.Initialize()
		case <-interrupts:
			a.logger.Warning("Application stopping")
			time.Sleep(1 * time.Second)
			os.Exit(1)
		case <-a.eventMgmtTicker.C:
			a.logger.Info("Begin event syncing")
			m, err := a.authStore.GetMeetupToken()
			if err != nil {
				a.logger.Errorf("Unable to retrieve meetup token. %v", err)
			}
			meetupClient := eventmgmt.NewMeetup(a.logger, http.DefaultClient, a.config.MeetupConfig.MeetupGroup, m.AccessToken, a.config.MeetupConfig.OrganizerMapping)
			streamyardClient := streaming.NewStreamyard(a.logger, http.DefaultClient, a.config.Streamyard.CSRFToken, a.config.Streamyard.JWT, a.config.StreamyardConfig.UserID, a.config.StreamyardConfig.YoutubeDestination, a.config.StreamyardConfig.FacebookGroupDestination)
			s := eventstore.NewEventStore(a.logger, meetupClient, a.calendarSvc, streamyardClient, a.config.EventStoreFile, a.config.CalendarConfig.CalendarID, a.config.CalendarConfig.CalendarEventInvitation, a.config.Features.MeetupSync.SubFeatures)
			err = s.CheckEvents(time.Now())

			if err != nil {
				a.logger.Errorf("Issue when checking events. %v", err)
			}

			time.Sleep(1 * time.Second)
		case <-a.authRefresherTicker.C:
			a.logger.Info("Begin refreshing tokens")
			err := a.googleAuth.Refresh()
			if err != nil {
				a.logger.Errorf("Unable to refresh Google Access Tokens. Err: %v", err)
			}
			err = a.meetupAuth.Refresh()
			if err != nil {
				a.logger.Errorf("Unable to refresh Meetup Access Tokens. Err: %v", err)
			}
			// Doublecheck for streamyard jwt token expiration
			err = streaming.JWTChecker(a.logger, a.config.Streamyard.JWT)
			if err != nil {
				a.logger.Errorf("Do check streamyard login creds to ensure no further issues with automation. Err: %v", err)
			}

			// Need to double check this - initialize may reset
			a.RerunAuth()
		}
	}
}
