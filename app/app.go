// Package app wraps all the functionality and keeps it out of cmd package
package app

import (
	"os"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/eventmgmt"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type App struct {
	configStore ConfigStore
	logger      logger.Logger
	// Internal controls
	authStore       AuthStore
	config          Config
	eventMgr        eventmgmt.EventMgmt
	eventMgmtTicker *time.Ticker
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
	if a.config.Features.MeetupSync.Enabled {
		a.eventMgmtTicker = time.NewTicker(time.Duration(a.config.Features.MeetupSync.IdleDuration) * time.Second)
	} else {
		a.eventMgmtTicker = nil
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
		}
	}
}
