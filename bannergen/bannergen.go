package bannergen

import (
	"context"
	"io/ioutil"
	"math"
	"net/url"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func Generate_banner(outputPath, seriesName, webinarTitle, webinarDate string) error {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// capture screenshot of an element
	bannerEndpoint, _ := url.Parse("http://localhost:9000/image")
	val := bannerEndpoint.Query()
	val.Add("series_name", seriesName)
	val.Add("webinar_title", webinarTitle)
	val.Add("webinar_date", webinarDate)
	bannerEndpoint.RawQuery = val.Encode()
	var buf []byte
	if err := chromedp.Run(ctx, elementScreenshot(bannerEndpoint.String(), `#banner`, &buf)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(outputPath, buf, 0644); err != nil {
		return err
	}
	return nil
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)

			return err
		}),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}
