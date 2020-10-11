package app

import (
	"html/template"
	"log"
	"net/http"
)

type webinarImageData struct {
	SeriesName   string
	WebinarTitle string
	WebinarDate  string
}

type image struct{}

func (i image) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	seriesName := r.URL.Query().Get("series_name")
	webinarTitle := r.URL.Query().Get("webinar_title")
	webinarDate := r.URL.Query().Get("webinar_date")

	tmpl, err := template.New("").ParseFiles("./templates/rocket_image.html")
	if err != nil {
		log.Printf("Error in parsing template. Err: %v", err)
	}
	data := webinarImageData{
		SeriesName:   seriesName,
		WebinarTitle: webinarTitle,
		WebinarDate:  webinarDate,
	}

	err = tmpl.ExecuteTemplate(w, "rocket_image.html", data)
	if err != nil {
		log.Printf("Error in parsing template. ERr: %v", err)
	}
}
