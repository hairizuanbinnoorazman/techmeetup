package main

import "github.com/spf13/cobra"

var (
	versionCmd = func() *cobra.Command {
		versioncmd := &cobra.Command{
			Use:   "version",
			Short: "Provide version of cli being run",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		return versioncmd
	}
)
