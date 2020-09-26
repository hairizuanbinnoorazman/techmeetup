package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type MeetupAuthorize struct {
	client      *http.Client
	logger      logger.Logger
	clientID    string
	redirectURI string
}

func (m MeetupAuthorize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorizeURL, _ := url.ParseRequestURI("https://secure.meetup.com/oauth2/authorize")
	query := authorizeURL.Query()
	query.Add("scope", "event_management")
	query.Add("client_id", m.clientID)
	query.Add("response_type", "code")
	query.Add("redirect_uri", m.redirectURI)
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusTemporaryRedirect)
}

type MeetupAuthRefresher struct {
	client       *http.Client
	logger       logger.Logger
	authStore    AuthStore
	clientID     string
	clientSecret string
}

func (m MeetupAuthRefresher) Refresh() error {
	meetupTokenInfo, err := m.authStore.GetMeetupToken()
	if err != nil {
		return err
	}

	if meetupTokenInfo.RefreshToken == "" {
		m.logger.Info("No refresh token available - will not refresh")
		return nil
	}

	if (time.Now().Unix() + 600) < meetupTokenInfo.ExpiryTime {
		m.logger.Info("Will not refresh - expiry time is more than 10 minutes away")
		return nil
	}

	m.logger.Info("Will refresh meetup token")
	accessURL, _ := url.ParseRequestURI("https://secure.meetup.com/oauth2/access")
	accessReqBody := url.Values{}
	accessReqBody["client_id"] = []string{m.clientID}
	accessReqBody["client_secret"] = []string{m.clientSecret}
	accessReqBody["grant_type"] = []string{"refresh_token"}
	accessReqBody["refresh_token"] = []string{meetupTokenInfo.RefreshToken}
	resp, err := m.client.PostForm(accessURL.String(), accessReqBody)
	if err != nil {
		return err
	}
	type authAccessResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	rawAuthAccessResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var a authAccessResp
	err = json.Unmarshal(rawAuthAccessResp, &a)
	if err != nil {
		return err
	}
	meetupTokenInfo.AccessToken = a.AccessToken
	meetupTokenInfo.ExpiryTime = time.Now().Unix() + a.ExpiresIn
	err = m.authStore.StoreMeetupToken(meetupTokenInfo)
	if err != nil {
		return err
	}
	return nil
}

type MeetupAccess struct {
	client             *http.Client
	logger             logger.Logger
	authStore          AuthStore
	clientID           string
	clientSecret       string
	redirectURI        string
	notifyConfigChange chan bool
}

func (m MeetupAccess) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	accessURL, _ := url.ParseRequestURI("https://secure.meetup.com/oauth2/access")
	accessReqBody := url.Values{}
	accessReqBody["client_id"] = []string{m.clientID}
	accessReqBody["client_secret"] = []string{m.clientSecret}
	accessReqBody["grant_type"] = []string{"authorization_code"}
	accessReqBody["redirect_uri"] = []string{m.redirectURI}
	accessReqBody["code"] = []string{code}
	resp, err := m.client.PostForm(accessURL.String(), accessReqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type authAccessResp struct {
		AccessToken   string `json:"access_token"`
		RefereshToken string `json:"refresh_token"`
		ExpiresIn     int64  `json:"expires_in"`
	}
	rawAuthAccessResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var a authAccessResp
	err = json.Unmarshal(rawAuthAccessResp, &a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = m.authStore.StoreMeetupToken(MeetupToken{
		RefreshToken: a.RefereshToken,
		AccessToken:  a.AccessToken,
		ExpiryTime:   time.Now().Unix() + a.ExpiresIn,
	})
	if err != nil {
		m.logger.Errorf("Failed to write to file but managed to get credentials. Will print it out for now")
		m.logger.Errorf("%+v", a)
	}
	defer func() {
		m.notifyConfigChange <- true
	}()
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
