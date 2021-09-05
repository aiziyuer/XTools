package main

import (
	"app/internal"
	"app/internal/ssh_tunnel"
	"app/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "SshTunnel",
	RunE: func(cmd *cobra.Command, args []string) error {

		h := &ssh_tunnel.TunnelHandler{}

		return h.Do()
	},
}

func main() {
	util.SetupLogs("./log/info.log")
	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = internal.Version
}

func initConfig() {
	viper.SetConfigFile("config.yaml")

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		zap.S().Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}
