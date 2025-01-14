package cmd

import (
	"github.com/Krzysztofz01/apikit/cmd/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "apikit",
	Short: "",
	Long:  "",
	Run: func(_ *cobra.Command, _ []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.FatalErr(err)
	}
}
