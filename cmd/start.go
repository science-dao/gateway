package cmd

import (
	"github.com/science-dao/gateway/internal/handlers"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		go handlers.Listen()
		handlers.Receive()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
