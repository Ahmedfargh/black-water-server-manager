package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Host string

var rootCmd = &cobra.Command{
	Use:   "bwcli",
	Short: "Blackwater CLI - Server Management & Monitoring",
	Long:  `A professional CLI to interact with the Blackwater server API.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&Host, "host", "H", "http://localhost:8080", "Blackwater Server URL")
}
