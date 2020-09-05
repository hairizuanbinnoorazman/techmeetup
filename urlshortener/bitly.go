package urlshortener

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type Bitly struct {
	logger      logger.Logger
	client      *http.Client
	accessToken string
	groupUID    string
}

func NewBitly(logger logger.Logger, client *http.Client, accessToken string) Bitly {
	return Bitly{
		logger:      logger,
		client:      client,
		accessToken: accessToken,
	}
}

type bitlyResource struct {
	CreatedAt      string   `json:"created_at"`
	ID             string   `json:"id"`
	Link           string   `json:"link"`
	LongURL        string   `json:"long_url"`
	Archived       bool     `json:"archived"`
	CustomBitlinks []string `json:"custom_bitlinks"`
	Tags           []string `json:"tags"`
}

// GenerateLink - Special case with bitly - bitly apparently will return the same shortened link
// if the same long url is provided to it - hence, there is less of a need to check that this
// is expected.
func (b *Bitly) GenerateLink(ctx context.Context, url string) (shortenedLink string, err error) {
	type bitlyReq struct {
		Title   string `json:"title"`
		LongURL string `json:"long_url"`
		Domain  string `json:"domain"`
	}

	body := bitlyReq{
		LongURL: url,
		Domain:  "bit.ly",
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("Unable to marshal bitly request item. Err: %v", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api-ssl.bitly.com/v4/shorten", bytes.NewBuffer(rawBody))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", b.accessToken))
	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Unable to request for shortened url. Err: %v", err)
	}
	respRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Unable to parse request body")
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Unexpected status code. StatusCode: %v Response: %v", resp.StatusCode, string(respRaw))
	}

	var br bitlyResource
	err = json.Unmarshal(respRaw, &br)
	if err != nil {
		return "", fmt.Errorf("Unable to parse request")
	}

	return br.Link, nil
}
