package cmd

import (
	"os"

	"github.com/science-dao/gateway/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "gateway",
	Version: version,
	Short:   "Gateway node for DeSci DAOs",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		config.ConfigLogger()
		config.LoadClient()
		config.LoadAuth()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Error starting CLI: ", err)
		os.Exit(1)
	}
}
