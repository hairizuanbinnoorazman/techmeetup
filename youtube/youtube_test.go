// Package youtube handles logic to be able to pull information from youtube
// Future versions of this pkg would allow for alternations to youtube description
package youtube

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/yaml.v2"
)

func YoutubeServiceHelper(t *testing.T) *youtube.Service {
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
	aa, _ := youtube.NewService(context.TODO(), option.WithHTTPClient(client))
	return aa
}

func TestYoutube_ListVideos(t *testing.T) {
	type fields struct {
		logger      logger.Logger
		accessToken string
		channelID   string
		youtubeSvc  *youtube.Service
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Video
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				logger:     logger.LoggerForTests{Tester: t},
				channelID:  "",
				youtubeSvc: YoutubeServiceHelper(t),
			},
			args: args{
				ctx: context.TODO(),
			},
			want: []Video{
				Video{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := Youtube{
				logger:     tt.fields.logger,
				channelID:  tt.fields.channelID,
				youtubeSvc: tt.fields.youtubeSvc,
			}
			got, err := y.ListVideos(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Youtube.ListVideos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Youtube.ListVideos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYoutube_GetVideos(t *testing.T) {
	type fields struct {
		logger      logger.Logger
		accessToken string
		channelID   string
		youtubeSvc  *youtube.Service
	}
	type args struct {
		ctx      context.Context
		videoIDs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Video
		wantErr bool
	}{
		{
			name: "successful case",
			fields: fields{
				logger:     logger.LoggerForTests{Tester: t},
				youtubeSvc: YoutubeServiceHelper(t),
			},
			args: args{
				ctx:      context.TODO(),
				videoIDs: []string{""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := Youtube{
				logger:     tt.fields.logger,
				channelID:  tt.fields.channelID,
				youtubeSvc: tt.fields.youtubeSvc,
			}
			got, err := y.GetVideos(tt.args.ctx, tt.args.videoIDs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Youtube.GetVideos() error = %+v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Youtube.GetVideos() = %+v, want %v", got, tt.want)
			}
		})
	}
}
