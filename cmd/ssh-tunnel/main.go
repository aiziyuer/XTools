package main

import (
	"app/internal"
	"app/internal/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "XTunnel",
}

func main() {
	util.SetupLogs("./log/info.log")
	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = internal.Version
}
