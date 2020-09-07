package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

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
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusPermanentRedirect)
}

type GoogleAccess struct {
	client       *http.Client
	logger       logger.Logger
	clientID     string
	clientSecret string
	redirectURI  string
}

func (g GoogleAccess) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	accessURL, _ := url.ParseRequestURI("https://secure.meetup.com/oauth2/access")
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
	g.logger.Infof("Response: %v", a)
}
