// Package youtube handles logic to be able to pull information from youtube
// Future versions of this pkg would allow for alternations to youtube description
package youtube

import (
	"context"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	ID          string
	Title       string
	Description string
}

type Youtube struct {
	logger     logger.Logger
	channelID  string
	youtubeSvc *youtube.Service
}

func (y Youtube) GetVideos(ctx context.Context, videoIDs ...string) ([]Video, error) {
	youtubeVideoCall := y.youtubeSvc.Videos.List([]string{"id", "snippet"})
	youtubeVideoCall = youtubeVideoCall.Id(videoIDs...)
	youtubeVideoCall = youtubeVideoCall.Context(ctx)
	resp, err := youtubeVideoCall.Do()
	if err != nil {
		return []Video{}, err
	}
	lol := []Video{}
	for _, v := range resp.Items {
		lol = append(lol, Video{ID: v.Id, Title: v.Snippet.Title, Description: v.Snippet.Description})
	}
	return lol, nil
}

func (y Youtube) ListVideos(ctx context.Context) ([]Video, error) {
	youtubeSearchCall := y.youtubeSvc.Search.List([]string{"id", "snippet"})
	youtubeSearchCall = youtubeSearchCall.ChannelId(y.channelID)
	youtubeSearchCall = youtubeSearchCall.Order("date")
	youtubeSearchCall = youtubeSearchCall.Type("video")
	youtubeSearchCall = youtubeSearchCall.Context(ctx)
	youtubeSearchCall = youtubeSearchCall.MaxResults(50)
	resp, err := youtubeSearchCall.Do()
	if err != nil {
		return []Video{}, err
	}
	lol := []Video{}
	for _, v := range resp.Items {
		lol = append(lol, Video{ID: v.Id.VideoId, Title: v.Snippet.Title, Description: v.Snippet.Description})
	}
	return lol, nil
}
