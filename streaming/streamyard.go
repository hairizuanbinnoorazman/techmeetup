package streaming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
)

type Stream struct {
	ID           string
	Destinations []Destination
	// Inputs
	StartDate   time.Time
	Name        string
	Description string
	ImagePath   string
	IsPublic    bool
}

type Destination struct {
	ID string
	// Type is youtube or facebook
	Type string
	Link string
}

type Streamyard struct {
	logger              logger.Logger
	client              *http.Client
	csrfToken           string
	jwt                 string
	userID              string
	youtubeDestination  string
	facebookDestination string
}

type StreamyardListResponse struct {
	HasMore    bool                          `json:"hasMore"`
	Broadcasts []StreamyardBroadcastResponse `json:"broadcasts"`
}

type StreamyardBroadcastResponse struct {
	ID      string                              `json:"id"`
	Status  string                              `json:"status"`
	Title   string                              `json:"title"`
	Outputs []StreamyardBroadcastOutputResponse `json:"outputs"`
}

type StreamyardDestinationResponse struct {
	Output StreamyardBroadcastOutputResponse `json:"output"`
}

type StreamyardBroadcastOutputResponse struct {
	ID               string `json:"id"`
	Privacy          string `json:"privacy"`
	PlannedStartTime string `json:"plannedStartTime"`
	Platform         string `json:"platform"`
	PlatformType     string `json:"platformType"`
	Description      string `json:"description"`
	PlatformUserName string `json:"platformUserName"`
	PlatformLink     string `json:"platformLink"`
	Image            string `json:"image"`
}

func NewStreamyard(logger logger.Logger, client *http.Client, csrfToken, jwt, userID, youtubeDestination, facebookDestination string) Streamyard {
	return Streamyard{
		logger:              logger,
		client:              client,
		csrfToken:           csrfToken,
		jwt:                 jwt,
		userID:              userID,
		youtubeDestination:  youtubeDestination,
		facebookDestination: facebookDestination,
	}
}

func (s Streamyard) GetStream(ctx context.Context, streamID string) (Stream, error) {
	if streamID == "" {
		return Stream{}, fmt.Errorf("StreamID is missing. Please provide streamID value first")
	}

	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return Stream{}, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := fmt.Sprintf("https://streamyard.com/api/broadcasts/%v", streamID)
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, finalURL.String(), nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://streamyard.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return Stream{}, err
	}
	raw, err := ioutil.ReadAll(resp.Body)

	s.logger.Info(string(raw))

	var aa StreamyardBroadcastResponse
	err = json.Unmarshal(raw, &aa)
	if err != nil {
		return Stream{}, fmt.Errorf("Unable to parse json output. Err: %v", err)
	}

	if len(aa.Outputs) == 0 {
		return Stream{}, fmt.Errorf("Output streams are missing")
	}
	isPublic := false
	if aa.Outputs[0].Privacy == "public" {
		isPublic = true
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05Z", aa.Outputs[0].PlannedStartTime)
	if err != nil {
		return Stream{}, fmt.Errorf("Unable to parse timeoutputs. Input time value: %v", aa.Outputs[0].PlannedStartTime)
	}
	loc, _ := time.LoadLocation("Asia/Singapore")
	parsedTime = parsedTime.In(loc)

	ds := []Destination{}
	for _, zz := range aa.Outputs {
		ds = append(ds, Destination{
			ID:   zz.ID,
			Type: zz.Platform,
			Link: zz.PlatformLink,
		})
	}

	return Stream{
		ID:           aa.ID,
		StartDate:    parsedTime,
		Name:         aa.Title,
		Description:  aa.Outputs[0].Description,
		IsPublic:     isPublic,
		Destinations: ds,
	}, nil
}

func (s Streamyard) UpdateStream(ctx context.Context, streamID, title string) error {
	if streamID == "" || title == "" {
		return fmt.Errorf("StreamID or title is missing. Please provide streamID value first")
	}

	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := fmt.Sprintf("https://streamyard.com/api/broadcasts/%v", streamID)
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	type createReq struct {
		CSRFToken string `json:"csrfToken"`
		Title     string `json:"title"`
	}
	createRequest := createReq{
		CSRFToken: s.csrfToken,
		Title:     title,
	}
	rawReq, _ := json.Marshal(createRequest)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), bytes.NewBuffer(rawReq))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://streamyard.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("Err while doing request. Err: %v", err)
	}
	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to extract out rawResp. Err: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status error code. StatusCode: %v RawResp: %v", resp.StatusCode, string(rawResp))
	}
	return nil
}

func (s Streamyard) CreateStream(ctx context.Context, title string) (Stream, error) {
	if title == "" {
		return Stream{}, fmt.Errorf("No title provided to stream. Please relook at the inputs for this")
	}

	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return Stream{}, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := "https://streamyard.com/api/broadcasts"
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	type createReq struct {
		CSRFToken       string `json:"csrfToken"`
		RecordOnly      bool   `json:"recordOnly"`
		SelectedBrandID string `json:"selectedBrandId"`
		Title           string `json:"title"`
	}
	createRequest := createReq{
		CSRFToken:       s.csrfToken,
		RecordOnly:      false,
		SelectedBrandID: s.userID,
		Title:           title,
	}
	rawReq, _ := json.Marshal(createRequest)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), bytes.NewBuffer(rawReq))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://streamyard.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return Stream{}, err
	}
	raw, err := ioutil.ReadAll(resp.Body)

	var aa StreamyardBroadcastResponse
	err = json.Unmarshal(raw, &aa)
	if err != nil {
		return Stream{}, fmt.Errorf("Unable to retrieve parse response from streamyard. Err: %v", err)
	}

	// s.logger.Info(string(raw))

	return Stream{
		ID:   aa.ID,
		Name: aa.Title,
	}, nil
}

func (s Streamyard) CreateDestination(ctx context.Context, destinationStreamType string, ss Stream) (Stream, error) {
	if destinationStreamType == "" || ss.ID == "" || ss.Name == "" || ss.Description == "" || ss.ImagePath == "" || ss.StartDate.IsZero() {
		return Stream{}, fmt.Errorf("Empty inputs detected:\nstream: %v\ndestinationStreamType: %v", ss, destinationStreamType)
	}

	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return ss, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := fmt.Sprintf("https://streamyard.com/api/broadcasts/%v/outputs", ss.ID)
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	rawImage, err := ioutil.ReadFile(ss.ImagePath)
	if err != nil {
		return ss, fmt.Errorf("Unable to load image file. Please check path to ensure correct. Err: %v", err)
	}

	imageContentType, err := imageTypeDetector(ss.ImagePath)
	if err != nil {
		return ss, fmt.Errorf("Unexpected Image Type found. Please review file type. Err: %v", err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormField("title")
	part.Write([]byte(ss.Name))
	part, _ = writer.CreateFormField("description")
	part.Write([]byte(ss.Description))
	if destinationStreamType == "youtube" {
		if ss.IsPublic {
			part, _ = writer.CreateFormField("privacy")
			part.Write([]byte("public"))
		} else {
			part, _ = writer.CreateFormField("privacy")
			part.Write([]byte("private"))
		}
	} else {
		part, _ = writer.CreateFormField("privacy")
		part.Write([]byte("public"))
	}

	// Special code to cope with custom type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			"image", "blob"))
	h.Set("Content-Type", imageContentType)
	part, _ = writer.CreatePart(h)
	part.Write(rawImage)
	// Special code to cope with custom type

	part, _ = writer.CreateFormField("plannedStartTime")
	part.Write([]byte(streamyardCompatibleTimeFormat(ss.StartDate)))
	if destinationStreamType == "youtube" {
		part, _ = writer.CreateFormField("destinationId")
		part.Write([]byte(s.youtubeDestination))
	} else if destinationStreamType == "facebook" {
		part, _ = writer.CreateFormField("destinationId")
		part.Write([]byte(s.facebookDestination))
	} else {
		return ss, fmt.Errorf("Bad destination location")
	}

	part, _ = writer.CreateFormField("csrfToken")
	part.Write([]byte(s.csrfToken))

	writer.Close()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), body)
	req.Header.Add("content-type", writer.FormDataContentType())
	req.Header.Add("origin", "https://streamyard.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
	resp, err := s.client.Do(req)
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ss, err
	}

	// s.logger.Info(string(raw))

	var aa StreamyardDestinationResponse
	err = json.Unmarshal(raw, &aa)
	if err != nil {
		return ss, fmt.Errorf("Unable to parse streamyard destination response")
	}

	ss.Destinations = append(ss.Destinations, Destination{
		ID:   aa.Output.ID,
		Type: aa.Output.Platform,
		Link: aa.Output.PlatformLink,
	})

	return ss, nil
}

func (s Streamyard) UpdateDestination(ctx context.Context, destinationStreamType string, ss Stream, forceImageUpdate bool) (Stream, error) {
	if ss.ID == "" || ss.Description == "" || ss.StartDate.IsZero() || ss.Name == "" || len(ss.Destinations) == 0 {
		return ss, fmt.Errorf("No stream ID reference or destination ID reference provided. Please recheck inputs")
	}
	if forceImageUpdate {
		if ss.ImagePath == "" {
			return ss, fmt.Errorf("Attempt to update the image but no image path provided")
		}
	}

	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return Stream{}, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	destinationID := ""
	for _, dest := range ss.Destinations {
		if strings.ToLower(dest.Type) == strings.ToLower(destinationStreamType) {
			destinationID = dest.ID
		}
	}

	if destinationID == "" {
		return ss, fmt.Errorf("No matching stream types - please review data. %v", ss)
	}

	initialURL := fmt.Sprintf("https://streamyard.com/api/broadcasts/%v/outputs/%v", ss.ID, destinationID)
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormField("title")
	part.Write([]byte(ss.Name))
	part, _ = writer.CreateFormField("description")
	part.Write([]byte(ss.Description))

	if forceImageUpdate {
		rawImage, err := ioutil.ReadFile(ss.ImagePath)
		if err != nil {
			return ss, fmt.Errorf("Unable to load image file. Please check path to ensure correct. Err: %v", err)
		}

		imageContentType, err := imageTypeDetector(ss.ImagePath)
		if err != nil {
			return ss, fmt.Errorf("Unexpected Image Type found. Please review file type. Err: %v", err)
		}

		// Special code to cope with custom type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				"image", "blob"))
		h.Set("Content-Type", imageContentType)
		part, _ = writer.CreatePart(h)
		part.Write(rawImage)
		// Special code to cope with custom type
	}

	part, _ = writer.CreateFormField("plannedStartTime")
	part.Write([]byte(streamyardCompatibleTimeFormat(ss.StartDate)))
	part, _ = writer.CreateFormField("csrfToken")
	part.Write([]byte(s.csrfToken))

	writer.Close()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), body)
	req.Header.Add("content-type", writer.FormDataContentType())
	req.Header.Add("origin", "https://streamyard.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
	resp, err := s.client.Do(req)
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ss, err
	}

	s.logger.Info(string(raw))

	var aa StreamyardBroadcastOutputResponse
	err = json.Unmarshal(raw, &aa)
	if err != nil {
		return ss, fmt.Errorf("Unable to parse json response from Streamyard. Err: %v", err)
	}

	return ss, nil
}

type StreamyardDestinationResp struct {
	Destinations []StreamyardDestination `json:"destinations"`
}

type StreamyardDestination struct {
	PlatformUsername string `json:"platformUsername"`
	Platform         string `json:"platform"`
	ID               string `json:"id"`
}

func (s Streamyard) GetDestinations(ctx context.Context) ([]StreamyardDestination, error) {
	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return nil, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := "https://streamyard.com/api/destinations"
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, finalURL.String(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ssResp StreamyardDestinationResp
	// s.logger.Info(string(raw))
	json.Unmarshal(raw, &ssResp)
	return ssResp.Destinations, nil
}

func (s Streamyard) ListStreams(ctx context.Context) ([]Stream, error) {
	err := JWTChecker(s.logger, s.jwt)
	if err != nil {
		return []Stream{}, fmt.Errorf("Error while checking jwt. Err: %v", err)
	}

	initialURL := "https://streamyard.com/api/broadcasts?limit=10&isAvailable=true&isComplete=false"
	finalURL, _ := url.ParseRequestURI(initialURL)

	cj := s.createCookiejar(finalURL)
	s.client.Jar = cj

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, finalURL.String(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return []Stream{}, err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Stream{}, err
	}
	// s.logger.Info(string(raw))
	ss := []Stream{}
	var ssList StreamyardListResponse
	json.Unmarshal(raw, &ssList)
	for _, item := range ssList.Broadcasts {
		if len(item.Outputs) == 0 {
			continue
		}
		isPublic := false
		if item.Outputs[0].Privacy == "public" {
			isPublic = true
		}

		parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", item.Outputs[0].PlannedStartTime)

		ds := []Destination{}
		for _, zz := range item.Outputs {
			ds = append(ds, Destination{
				ID:   zz.ID,
				Type: zz.Platform,
				Link: zz.PlatformLink,
			})
		}

		ss = append(ss, Stream{
			ID:           item.ID,
			StartDate:    parsedTime,
			Name:         item.Title,
			Description:  item.Outputs[0].Description,
			IsPublic:     isPublic,
			Destinations: ds,
		})
	}
	return ss, nil
}

func (s *Streamyard) createCookiejar(reqUrl *url.URL) *cookiejar.Jar {
	cj, _ := cookiejar.New(nil)
	cj.SetCookies(reqUrl, []*http.Cookie{
		&http.Cookie{
			Name:  "csrfToken",
			Value: s.csrfToken,
		},
		&http.Cookie{
			Name:  "jwt",
			Value: s.jwt,
		},
	})
	return cj
}

func JWTChecker(logger logger.Logger, jwtToken string) error {
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("Unable to parse jwt token provided")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("Unable to generate the claims from jwt token")
	}
	value := claims["exp"]
	tm := time.Unix(int64(value.(float64)), 0)
	logger.Infof("Expiry date: %v", tm)
	if time.Now().After(tm) {
		return fmt.Errorf("JWT expired - don't proceed with request. It will fail")
	}
	if time.Now().Add(-72 * time.Hour).After(tm) {
		logger.Warning("JWT is expiring soon - within 3 days. Please logout and login once more for streamyard")
	}
	return nil
}

func streamyardCompatibleTimeFormat(t time.Time) string {
	return t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (Singapore Standard Time)")
}

func imageTypeDetector(f string) (string, error) {
	file := filepath.Base(f)
	parts := strings.SplitN(file, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("File extension wasn't available")
	}
	if parts[1] == "jpeg" || parts[1] == "jpg" {
		return "image/jpeg", nil
	} else if parts[1] == "png" {
		return "image/png", nil
	} else {
		return "", fmt.Errorf("Unable to detect file type")
	}
}
