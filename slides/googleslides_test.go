package slides

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"google.golang.org/api/option"
	"google.golang.org/api/slides/v1"
)

func slidesClientHelper() *slides.Service {
	credJSON, _ := ioutil.ReadFile("../auth.json")
	xClient, _ := slides.NewService(context.Background(), option.WithCredentialsJSON(credJSON))
	return xClient
}

func TestGoogleSlides_GetAllText(t *testing.T) {
	type fields struct {
		logger       logger.Logger
		slideService *slides.Service
	}
	type args struct {
		ctx      context.Context
		slidesID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				logger:       logger.LoggerForTests{Tester: t},
				slideService: slidesClientHelper(),
			},
			args: args{
				ctx:      context.TODO(),
				slidesID: "1A8tyh0MoV4BvWvEJLS3DgtdBWiZGF-fWehMehUqFCvQ",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleSlides{
				logger:       tt.fields.logger,
				slideService: tt.fields.slideService,
			}
			got, err := g.GetAllText(tt.args.ctx, tt.args.slidesID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleSlides.GetAllText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleSlides.GetAllText() = %v, want %v", got, tt.want)
			}
		})
	}
}
