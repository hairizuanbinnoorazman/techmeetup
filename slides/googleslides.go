package slides

import (
	"github.com/hairizuanbinnoorazman/techmeetup/logger"
	"google.golang.org/api/slides/v1"
)

type GoogleSlides struct {
	logger       logger.Logger
	slideService *slides.Service
}

func (g *GoogleSlides) GetLinks(slidesID string) ([]string, error) {
	return []string{}, nil
}
