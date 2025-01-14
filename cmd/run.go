package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Krzysztofz01/apikit/cmd/log"
	"github.com/Krzysztofz01/apikit/internal/config"
	"github.com/Krzysztofz01/apikit/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if r := recover(); r != nil {
				log.FatalErr(r)
			}
		}()

		configuration, err := config.LoadServerConfigurationFromFile()
		if err != nil {
			log.FatalErr(err)
		}

		logger, loggerDispose, err := log.CreateBaseLogger(configuration.VerboseMode)
		if err != nil {
			log.FatalErr(err)
		}

		defer func() {
			if err := loggerDispose(); err != nil {
				log.FatalErr(err)
			}
		}()

		internalLogger := log.CreateInternalLogger(logger)

		server, err := server.CreateApiKitServer(http.DefaultClient, configuration, internalLogger)
		if err != nil {
			log.FatalErr(err)
		}

		go func() {
			if err := server.Start(); err != nil && err != http.ErrServerClosed {
				internalLogger.Errorf("Runtime", "Server runtime failure: %s", err.Error())
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			internalLogger.Errorf("Runtime", "Server shutdown failure: %s", err.Error())
		}
	},
}
