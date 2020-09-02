package urlshortener

import (
	"net/http"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type Bitly struct {
	logger      logger.Logger
	client      *http.Client
	accessToken string
}

func NewBitly(logger logger.Logger, client *http.Client, accessToken string) Bitly {
	return Bitly{
		logger:      logger,
		client:      client,
		accessToken: accessToken,
	}
}

func (b *Bitly) GenerateLink(url string) (shortenedLink string, err error) {
	return "", nil
}
