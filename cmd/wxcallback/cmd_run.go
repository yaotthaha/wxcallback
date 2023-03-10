package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"wxcallback/core"
	"wxcallback/lib/log"
)

var runCommand = &cobra.Command{
	Use:   "run",
	Short: "Run Service",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(run())
	},
}

var (
	paramConfig  string
	paramLogFile string
	paramDebug   bool
)

func init() {
	RootCommand.AddCommand(runCommand)
	runCommand.PersistentFlags().StringVarP(&paramConfig, "config", "c", "config.json", "Config File")
	runCommand.PersistentFlags().StringVarP(&paramLogFile, "log", "l", "", "Log File")
	runCommand.PersistentFlags().BoolVarP(&paramDebug, "debug", "d", false, "Debug Mode")
}

func run() int {
	logger := log.NewLogger(nil, nil)
	logger.SetDebug(paramDebug)
	logger.Info("Global", fmt.Sprintf("version %s", Version))
	defer logger.Info("Global", "Bye!!")
	if paramLogFile != "" {
		f, err := os.OpenFile(paramLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatal("Global", fmt.Sprintf("open log file fail: %s", err.Error()))
			return 1
		}
		defer func() {
			f.Close()
			logger.SetOutput(os.Stdout)
			logger.Info("Global", "close log file")
		}()
		logger.Info("Global", fmt.Sprintf("write log to file: %s", paramLogFile))
		logger.SetOutput(f)
	}
	f, err := os.ReadFile(paramConfig)
	if err != nil {
		logger.Fatal("Global", err.Error())
		return 1
	}
	var config core.Config
	err = json.Unmarshal(f, &config)
	if err != nil {
		logger.Fatal("Global", err.Error())
		return 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rvSIG(logger, cancel)
	server := config.NewServer(&core.ServerOption{
		Context: ctx,
		Logger:  logger,
	})
	server.RunWithContext(nil)
	return 0
}

func rvSIG(logger *log.Logger, cancel context.CancelFunc) {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(osSignals)
		close(osSignals)
	}()
	s, loaded := <-osSignals
	if loaded {
		logger.Warn("Global", fmt.Sprintf("Receive Signal: %s", s.String()))
		cancel()
	}
}
