package cmd

import (
	"fmt"
	"os"

	"github.com/Krzysztofz01/apikit/internal/constants"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "",
	Long:  "",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Fprintf(os.Stdout, "%s\n", constants.Version)
	},
}
