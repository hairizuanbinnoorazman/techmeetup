// Package app wraps all the functionality and keeps it out of cmd package
package app

import (
	"net/http"
	"os"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/eventmgmt"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
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
	a.authRefresherTicker = time.NewTicker(300 * time.Second)
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
			time.Sleep(1 * time.Second)
		case <-a.authRefresherTicker.C:
			err := a.googleAuth.Refresh()
			if err != nil {
				a.logger.Errorf("Unable to refresh Google Access Tokens. Err: %v", err)
			}
			err = a.meetupAuth.Refresh()
			if err != nil {
				a.logger.Errorf("Unable to refresh Meetup Access Tokens. Err: %v", err)
			}
			a.Initialize()
		}
	}
}
