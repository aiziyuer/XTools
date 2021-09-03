package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use: "SshTunnel",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = version
}

func initConfig() {
	viper.SetConfigFile("config.yaml")

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		zap.S().Debugf("Using config file:", viper.ConfigFileUsed())
	}
}
