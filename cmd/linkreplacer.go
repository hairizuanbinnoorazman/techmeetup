package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/techmeetup/urlshortener"

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
3. There is only one of such url in each page on the slide

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
		var dryMode bool
		var accessToken string
		applylinkscmd := &cobra.Command{
			Use:   "apply [Google Slide ID]",
			Short: "Apply the configuration to make the required changes to the links",
			Long:  ``,
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("Expected Google Slide ID here")
				}
				if dryMode == false {
					if accessToken == "" {
						return fmt.Errorf("bitly access token is missing. Add the --access-token <Bitly token> flag ")
					}
				}
				return nil
			},
			Run: func(cmd *cobra.Command, args []string) {
				presentationSlideID := args[0]
				raw, err := ioutil.ReadFile(configFile)
				if err != nil {
					fmt.Printf("Unable to read config file. Err: %v", err)
					os.Exit(1)
				}
				var links []tslides.TextOnSlideReplacer
				yaml.Unmarshal(raw, &links)
				var cleanedLinks []tslides.TextOnSlideReplacer
				for _, val := range links {
					if strings.Contains(val.Text, "bit.ly") || strings.Contains(val.Text, "bitly") {
						continue
					}
					cleanedLinks = append(cleanedLinks, val)
				}
				if dryMode {
					cleanedRaw, _ := yaml.Marshal(cleanedLinks)
					fmt.Println(string(cleanedRaw))
					os.Exit(0)
				}
				credJSON, err := ioutil.ReadFile(authFile)
				if err != nil {
					logrus.Errorf("Unable to read auth file. We will not proceed. Err: %v", err)
					os.Exit(1)
				}
				slideService, err := slides.NewService(context.Background(), option.WithCredentialsJSON(credJSON))
				if err != nil {
					logrus.Errorf("Unable to create slides service. We will not proceed. Err: %v", err)
					os.Exit(1)
				}
				gslides := tslides.NewGoogleSlides(logrus.StandardLogger(), slideService)
				bitlyClient := urlshortener.NewBitly(logrus.New(), http.DefaultClient, accessToken)
				for idx, val := range cleanedLinks {
					if val.ReplaceText != "" {
						continue
					}
					replaceURL, err := bitlyClient.GenerateLink(context.TODO(), val.Text)
					if err != nil {
						fmt.Printf("Early termination - error in generating new url. Please review. Err: %v\n", err)
						os.Exit(1)
					}
					cleanedLinks[idx].ReplaceText = replaceURL
					time.Sleep(1 * time.Second)
				}
				logrus.Infof("\nPrinting data to be applied:\n%+v\n", cleanedLinks)
				err = gslides.UpdateText(context.TODO(), presentationSlideID, cleanedLinks)
				if err != nil {
					fmt.Printf("Error in updating text on google slides. Err: %v", err)
					os.Exit(1)
				}
			},
		}
		applylinkscmd.Flags().StringVar(&authFile, "authfile", "auth.json", "Authentication json needed for some platforms")
		applylinkscmd.Flags().StringVar(&configFile, "config", "config.yaml", "Configuration file. Please utilize the fetcher to ensure the right format of config is used")
		applylinkscmd.Flags().StringVar(&accessToken, "access-token", "", "Access token for bit.ly")
		applylinkscmd.Flags().BoolVar(&dryMode, "drymode", false, "Runs through some filters and returns a yaml that contains actual links that would be replaced")
		return applylinkscmd
	}
)
