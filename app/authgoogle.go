package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type GoogleAuthorize struct {
	client      *http.Client
	logger      logger.Logger
	clientID    string
	redirectURI string
	scope       string
}

func (g GoogleAuthorize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorizeURL, _ := url.ParseRequestURI("https://accounts.google.com/o/oauth2/v2/auth")
	query := authorizeURL.Query()
	query.Add("scope", g.scope)
	query.Add("access_type", "offline")
	query.Add("client_id", g.clientID)
	query.Add("response_type", "code")
	query.Add("redirect_uri", g.redirectURI)
	query.Add("prompt", "consent")
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusPermanentRedirect)
}

type GoogleAuthRefresher struct {
	client       *http.Client
	logger       logger.Logger
	authStore    AuthStore
	clientID     string
	clientSecret string
}

func (g GoogleAuthRefresher) Refresh() error {
	googleTokenInfo, err := g.authStore.GetGoogleToken()
	if err != nil {
		return err
	}

	if googleTokenInfo.RefreshToken == "" {
		g.logger.Info("No refresh token available - will not refresh")
		return nil
	}

	if (time.Now().Unix() + 600) < googleTokenInfo.ExpiryTime {
		g.logger.Info("Will not refresh - expiry time is more than 10 minutes away")
		return nil
	}

	g.logger.Info("Will refresh google token")
	accessURL, _ := url.ParseRequestURI("https://oauth2.googleapis.com/token")
	accessReqBody := url.Values{}
	accessReqBody["client_id"] = []string{g.clientID}
	accessReqBody["client_secret"] = []string{g.clientSecret}
	accessReqBody["refresh_token"] = []string{googleTokenInfo.RefreshToken}
	accessReqBody["grant_type"] = []string{"refresh_token"}
	resp, err := g.client.PostForm(accessURL.String(), accessReqBody)
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
	googleTokenInfo.AccessToken = a.AccessToken
	googleTokenInfo.ExpiryTime = time.Now().Unix() + a.ExpiresIn
	err = g.authStore.StoreGoogleToken(googleTokenInfo)
	return nil
}

type GoogleAccess struct {
	client             *http.Client
	logger             logger.Logger
	authStore          AuthStore
	clientID           string
	clientSecret       string
	redirectURI        string
	notifyConfigChange chan bool
}

func (g GoogleAccess) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	accessURL, _ := url.ParseRequestURI("https://oauth2.googleapis.com/token")
	accessReqBody := url.Values{}
	accessReqBody["code"] = []string{code}
	accessReqBody["client_id"] = []string{g.clientID}
	accessReqBody["client_secret"] = []string{g.clientSecret}
	accessReqBody["redirect_uri"] = []string{g.redirectURI}
	accessReqBody["grant_type"] = []string{"authorization_code"}
	resp, err := g.client.PostForm(accessURL.String(), accessReqBody)
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
	googleToken, err := g.authStore.GetGoogleToken()
	googleToken.AccessToken = a.AccessToken
	googleToken.ExpiryTime = time.Now().Unix() + a.ExpiresIn
	if a.RefereshToken != "" {
		googleToken.RefreshToken = a.RefereshToken
	}
	err = g.authStore.StoreGoogleToken(googleToken)
	if err != nil {
		g.logger.Errorf("Failed to write to file but managed to get credentials. Will print it out for now")
		g.logger.Errorf("%+v", a)
	}
	defer func() {
		g.notifyConfigChange <- true
	}()
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
