package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	tslides "github.com/hairizuanbinnoorazman/techmeetup/slides"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"google.golang.org/api/slides/v1"
)

var (
	linkreplacerCmd = func() *cobra.Command {
		linkreplacercmd := &cobra.Command{
			Use:   "linkreplacer",
			Short: "Link replacer would replace links in particular assets with url shorten-ed links",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		linkreplacercmd.AddCommand(retrieveLinksCmd())
		linkreplacercmd.AddCommand(applyLinksCmd())
		return linkreplacercmd
	}

	retrieveLinksCmd = func() *cobra.Command {
		var authFile string
		retrievelinkscmd := &cobra.Command{
			Use:   "retrieve [Google Slide ID]",
			Short: "Retrive all link looking things",
			Long: `
This utility extracts all text items from Google Slides to be appended into a single yaml 
file for further configuration. There are a few things to note though if you're using this.
1. The whole textbox is the URL
2. URL is complete (contains schema etc - http or https exists at the front of it)

Further improvements can be added to this tool in the future`,
			Args: cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				presentationSlideID := args[0]
				credJSON, err := ioutil.ReadFile(authFile)
				if err != nil {
					logrus.Errorf("Unable to read auth file. We will not proceed. Err: %v", err)
					os.Exit(1)
				}
				slideService, err := slides.NewService(context.Background(), option.WithCredentialsJSON(credJSON))
				if err != nil {
					logrus.Errorf("Unable to create slide service. We will not proceed. Err: %v", err)
					os.Exit(1)
				}
				gslides := tslides.NewGoogleSlides(logrus.StandardLogger(), slideService)
				items, err := gslides.GetAllText(context.Background(), presentationSlideID)
				if err != nil {
					logrus.Errorf("Unable to fetch text data from slides. Err: %v", err)
					os.Exit(1)
				}
				items = tslides.FilterForURLs(items)
				raw, err := yaml.Marshal(items)
				if err != nil {
					logrus.Errorf("Yaml marshal error. Err: %v", err)
					os.Exit(1)
				}
				fmt.Println(string(raw))
			},
		}
		retrievelinkscmd.Flags().StringVar(&authFile, "authfile", "auth.json", "Authentication json needed for some platforms")
		return retrievelinkscmd
	}

	applyLinksCmd = func() *cobra.Command {
		var authFile string
		var configFile string
		applylinkscmd := &cobra.Command{
			Use:   "apply [Google Slide ID]",
			Short: "Apply the configuration to make the required changes to the links",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		applylinkscmd.Flags().StringVar(&authFile, "authfile", "auth.json", "Authentication json needed for some platforms")
		applylinkscmd.Flags().StringVar(&configFile, "config", "config.yaml", "Configuration file. Please utilize the fetcher to ensure the right format of config is used")
		return applylinkscmd
	}
)

// Configuration for linkfetcher specifically
// We are right now restricting this to only work with Google Slides (Generalizing this would come later)
type linkReplacerConfig struct {
	// Only one provider at the moment: bitly
	URLShortenerProvider string `yaml:"url_shortener_provider"`
	// Prefix adds to the front of the path
	// E.g. prefix = gdg-10, path = microservices, final_result = gdg-10-microservices
	// Will only be used if shortened path has a value, else, a randomized value provided by t
	// the shortener service will be provided
	Prefix string                 `yaml:"prefix"`
	Items  []linkReplaceSlideItem `yaml:"items"`
}

type linkReplaceSlideItem struct {
	SlidePageID string `yaml:"slide_page_id"`
	URL         string `yaml:"url"`
	// If this is empty, we would rely on shortener provider to provide a random endpoint
	ShortenedPath string `yaml:"shortened_path"`
}
