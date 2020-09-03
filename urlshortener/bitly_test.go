package urlshortener

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

func getBitlyAuth() string {
	type config struct {
		AuthToken string `yaml:"auth_token"`
	}
	var c config
	raw, _ := ioutil.ReadFile("../sampleLinkReplacer.yaml")
	yaml.Unmarshal(raw, c)
	return c.AuthToken
}

func TestBitly_GenerateLink(t *testing.T) {
	type fields struct {
		logger      logger.Logger
		client      *http.Client
		accessToken string
	}
	type args struct {
		ctx  context.Context
		url  string
		tags []string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantShortenedLink string
		wantErr           bool
	}{
		{
			name: "Successful case",
			fields: fields{
				logger:      logger.LoggerForTests{Tester: t},
				client:      http.DefaultClient,
				accessToken: getBitlyAuth(),
			},
			args: args{
				ctx:  context.TODO(),
				url:  "https://www.youtube.com",
				tags: []string{"zontext"},
			},
			wantShortenedLink: "aa",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitly{
				logger:      tt.fields.logger,
				client:      tt.fields.client,
				accessToken: tt.fields.accessToken,
			}
			gotShortenedLink, err := b.GenerateLink(tt.args.ctx, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bitly.GenerateLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShortenedLink != tt.wantShortenedLink {
				t.Errorf("Bitly.GenerateLink() = %v, want %v", gotShortenedLink, tt.wantShortenedLink)
			}
		})
	}
}
