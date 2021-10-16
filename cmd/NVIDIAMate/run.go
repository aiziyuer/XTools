package main

import (
	"app/internal/util"

	"github.com/ghodss/yaml"
	"github.com/gogf/gf/os/gfile"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type MigVgpu struct {
	bus_id string `yaml:"bus_id"`
}

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {

		input, err := yaml.YAMLToJSON(gfile.GetBytes("configs/NVIDIAMate.yaml"))
		if err != nil {
			return err
		}

		mig_vgpus, err := util.JsonPathQueryInYamlContent(input, "$.mig_vgpus")
		if err != nil {
			return err
		}

		for _, mig_vgpu := range mig_vgpus.([]*MigVgpu) {
			zap.S().Debugf("bus_id: %s", mig_vgpu.bus_id)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
