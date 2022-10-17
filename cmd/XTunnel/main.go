package main

import (
	"app/internal"
	"app/internal/global"
	"app/internal/util"
	zapcore2 "go.uber.org/zap/zapcore"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const AppName = "XTunnel"

var debug = false

// 确保再最高优先级时初始化应用
var _ = func() interface{} {
	global.AppConfig().SetAppName(AppName)
	util.SetupLogsWithBinaryName(global.AppConfig().GetAppName(), debug)
	return nil
}()

var rootCmd = &cobra.Command{
	Use: "XTunnel",
}

func main() {
	util.SetupDefaultLogs(zapcore2.InfoLevel, "./log/XTunnel.log")
	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = internal.Version
}
