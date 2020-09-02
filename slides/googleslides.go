package slides

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"google.golang.org/api/slides/v1"
)

type GoogleSlides struct {
	logger       logger.Logger
	slideService *slides.Service
}

func NewGoogleSlides(logger logger.Logger, slideService *slides.Service) GoogleSlides {
	return GoogleSlides{
		logger:       logger,
		slideService: slideService,
	}
}

// FilterForURLs - DO NOT USE
func FilterForURLs(items []TextOnSlide) []TextOnSlide {
	cleaned := []TextOnSlide{}
	for _, i := range items {
		_, err := url.ParseRequestURI(i.Text)
		if err == nil {
			cleaned = append(cleaned, TextOnSlide{SlidePageID: i.SlidePageID, Text: i.Text})
		}
	}
	return cleaned
}

type TextOnSlide struct {
	SlidePageID string `yaml:"slide_page_id"`
	Text        string `yaml:"text"`
}

// GetLinks iterates through all text objects in slides, retrieve the link data within it
func (g *GoogleSlides) GetAllText(ctx context.Context, slidesID string) ([]TextOnSlide, error) {
	getSlidesCall := g.slideService.Presentations.Get(slidesID)
	getSlidesCall = getSlidesCall.Context(ctx)
	slides, err := getSlidesCall.Do()
	if err != nil {
		return []TextOnSlide{}, fmt.Errorf("Unable to retrive presentation slides. Err: %v", err)
	}
	gatheredList := []TextOnSlide{}
	for _, s := range slides.Slides {
		for _, t := range s.PageElements {
			if t.Shape == nil {
				continue
			}
			if t.Shape.ShapeType == "TEXT_BOX" {
				for _, i := range t.Shape.Text.TextElements {
					if i.TextRun == nil {
						continue
					}
					rawText := i.TextRun.Content
					processedText := strings.TrimRight(rawText, "\n")
					processedText = strings.Trim(processedText, " ")
					gatheredList = append(gatheredList, TextOnSlide{SlidePageID: s.ObjectId, Text: processedText})
				}
			}
		}
	}
	return gatheredList, nil
}
