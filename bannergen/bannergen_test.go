package bannergen

import "testing"

func Test_generate_banner(t *testing.T) {
	type args struct {
		outputPath   string
		seriesName   string
		webinarTitle string
		webinarDate  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successful case",
			args: args{
				outputPath:   "../yahoo.png",
				seriesName:   "Webinar #78",
				webinarTitle: "This is a test of a webinar",
				webinarDate:  "21st May 2020 - 7.30pm to 9.00pm",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generate_banner(tt.args.outputPath, tt.args.seriesName, tt.args.webinarTitle, tt.args.webinarDate)
		})
	}
}
