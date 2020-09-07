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
		client:       http.DefaultClient,
		logger:       logrus.New(),
		authStore:    a,
		clientID:     c.Meetup.ClientID,
		clientSecret: c.Meetup.ClientSecret,
		redirectURI:  c.Meetup.RedirectURI,
	}
	http.Handle("/auth/meetup/authorize", meetupAuthorize)
	http.Handle("/auth/meetup/access", meetupAccess)
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
