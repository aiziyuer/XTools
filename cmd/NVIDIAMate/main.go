package main

import (
	"app/internal"
	"app/internal/util"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use: "NVIDIAMate",
}

func main() {

	util.SetupLogs("./log/NVIDIAMate.log")
	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}

}

func init() {
	rootCmd.Version = internal.Version
}
