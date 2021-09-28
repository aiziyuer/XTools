package main

import (
	"app/internal"
	"app/internal/util"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use: "nvida-mate",
}

func main() {
	util.SetupLogs("./log/nvida-mate.log")
	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = internal.Version
}
