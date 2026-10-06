package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the version number and build information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, "tele version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
