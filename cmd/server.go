package main

import "github.com/spf13/cobra"

var (
	serverCmd = func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "server",
			Short: "server will start a bunch of operations to handle workflows in mgmt of tech meetups",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		return cmd
	}
)
