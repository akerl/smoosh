package cmd

import (
	"github.com/spf13/cobra"

	"github.com/akerl/smoosh/config"
)

func updateRunner(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()

	configFile, err := flags.GetString("config")
	if err != nil {
		return err
	}

	c, err := config.NewConfig(configFile)
	if err != nil {
		return err
	}

	c.Noop, err = flags.GetBool("noop")
	if err != nil {
		return err
	}

	return c.Sync()
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Run an update based on the provided configuration",
	RunE:  updateRunner,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("config", "c", "", "Config file for update")
	updateCmd.Flags().BoolP("noop", "n", false, "Don't actually install/update files")
}
