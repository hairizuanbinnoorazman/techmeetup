package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hairizuanbinnoorazman/techmeetup/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/fsnotify.v1"
)

var (
	serverCmd = func() *cobra.Command {
		var configFile string
		cmd := &cobra.Command{
			Use:   "server",
			Short: "Server will start a bunch of operations to handle workflows in mgmt of tech meetups",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				notifyConfigChange := make(chan bool)
				notifyInterrupts := make(chan os.Signal, 1)
				signal.Notify(notifyInterrupts, syscall.SIGINT, syscall.SIGTERM)

				watcher, err := fsnotify.NewWatcher()
				if err != nil {
					logrus.Error("Unable to start file watch")
				}
				defer watcher.Close()
				watcher.Add(configFile)
				go app.ConfigFileWatcher(watcher, notifyConfigChange)

				configStore := app.NewBasicConfigStore(configFile)
				runner := app.NewApp(configStore, logrus.New())
				runner.Initialize()
				runner.Run(notifyConfigChange, notifyInterrupts)
			},
		}
		cmd.Flags().StringVar(&configFile, "config", "config.yaml", "Configuration file. Please utilize the fetcher to ensure the right format of config is used")
		return cmd
	}
)
