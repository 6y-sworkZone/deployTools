package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"deploytools/internal/config"
	"deploytools/pkg/utils"
)

var (
	cfgManager     *config.ConfigManager
	configPath     string
	verbose        bool
	noBackup       bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deploytools",
		Short: "A powerful deployment tool for managing server deployments",
		Long:  `DeployTools is a command-line tool for managing server configurations and deploying projects to remote servers.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			cfgManager, err = config.NewConfigManager(configPath)
			if err != nil {
				fmt.Printf("Failed to initialize config: %v\n", err)
				os.Exit(1)
			}

			if verbose {
				utils.DefaultLogger.SetLevel(utils.LevelDebug)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noBackup, "no-backup", false, "Disable backup before deployment")

	rootCmd.AddCommand(NewServerCmd())
	rootCmd.AddCommand(NewProjectCmd())
	rootCmd.AddCommand(NewGroupCmd())
	rootCmd.AddCommand(NewDeployCmd())
	rootCmd.AddCommand(NewListCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
