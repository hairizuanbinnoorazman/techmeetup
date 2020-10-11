package app

import (
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

func server(c Config, a AuthStore, notifyConfigChange chan bool) {
	meetupAuthorize := MeetupAuthorize{
		client:      http.DefaultClient,
		logger:      logrus.New(),
		clientID:    c.Meetup.ClientID,
		redirectURI: c.Meetup.RedirectURI,
	}
	meetupAccess := MeetupAccess{
		client:             http.DefaultClient,
		logger:             logrus.New(),
		authStore:          a,
		clientID:           c.Meetup.ClientID,
		clientSecret:       c.Meetup.ClientSecret,
		redirectURI:        c.Meetup.RedirectURI,
		notifyConfigChange: notifyConfigChange,
	}
	googleAuthorize := GoogleAuthorize{
		client:      http.DefaultClient,
		logger:      logrus.New(),
		clientID:    c.Google.ClientID,
		redirectURI: c.Google.RedirectURI,
		scope:       c.Google.Scope,
	}
	googleAccess := GoogleAccess{
		client:             http.DefaultClient,
		logger:             logrus.New(),
		authStore:          a,
		clientID:           c.Google.ClientID,
		clientSecret:       c.Google.ClientSecret,
		redirectURI:        c.Google.RedirectURI,
		notifyConfigChange: notifyConfigChange,
	}

	http.Handle("/image", image{})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./assets"))))
	http.Handle("/auth/meetup/authorize", meetupAuthorize)
	http.Handle("/auth/meetup/access", meetupAccess)
	http.Handle("/auth/google/authorize", googleAuthorize)
	http.Handle("/auth/google/access", googleAccess)
	http.Handle("/", index{})
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func ConfigFileWatcher(watcher *fsnotify.Watcher, notifyConfigChange chan bool) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				notifyConfigChange <- true
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("error: %v\n", err)
		}
	}
}
