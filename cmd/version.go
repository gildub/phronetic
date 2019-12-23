package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// BuildVersion contains the version at build time
	BuildVersion = "undefined"
	// BuildTime contains build time
	BuildTime = "undefined"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Provides analyze' version",
	Long:  `Returns the version of current analyze build`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(BuildVersion, BuildTime)
	},
}
