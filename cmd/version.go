package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/akerl/smoosh/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of smoosh",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("%s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
