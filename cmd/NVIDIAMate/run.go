package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {

		zap.S().Debug("Hello World")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

}
