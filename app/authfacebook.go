package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type FacebookAuthorize struct {
	logger      logger.Logger
	clientID    string
	redirectURI string
}

func (f FacebookAuthorize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorizeURL, _ := url.ParseRequestURI("https://www.facebook.com/v8.0/dialog/oauth")
	query := authorizeURL.Query()
	query.Add("client_id", f.clientID)
	query.Add("redirect_uri", f.redirectURI)
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusPermanentRedirect)
}

type FacebookAccess struct {
	client       *http.Client
	logger       logger.Logger
	clientID     string
	clientSecret string
	redirectURI  string
}

func (f FacebookAccess) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	accessURL, _ := url.ParseRequestURI("https://graph.facebook.com/v8.0/oauth/access_token")
	accessReqBody := url.Values{}
	accessReqBody["client_id"] = []string{f.clientID}
	accessReqBody["client_secret"] = []string{f.clientSecret}
	accessReqBody["redirect_uri"] = []string{f.redirectURI}
	accessReqBody["code"] = []string{code}
	resp, err := f.client.PostForm(accessURL.String(), accessReqBody)
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
	f.logger.Infof("Response: %v", a)
}
