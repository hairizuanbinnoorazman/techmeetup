package main

import "github.com/spf13/cobra"

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
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
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
