package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "techmeetup",
			Short: "techmeetup is a cli that provide quick utility capabilities for meetup groups",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		cmd.AddCommand(serverCmd())
		cmd.AddCommand(linkreplacerCmd())
		cmd.AddCommand(versionCmd())
		return cmd
	}
)

func main() {
	rootCmd().Execute()
}
