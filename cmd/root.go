package cmd

import (
	"bytes"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/util/gconv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type SessionInfo struct {
	SshUri      string   `yaml:"ssh_uri"`
	SshIdentity string   `yaml:"ssh_identity"`
	SshTunnels  []string `yaml:"ssh_tunnels"`
}

var rootCmd = &cobra.Command{
	Use: "SshTunnel",
	RunE: func(cmd *cobra.Command, args []string) error {

		var sessionInfos []SessionInfo

		decoder := yaml.NewDecoder(bytes.NewReader(gfile.GetBytes("sessions.yaml")))
		for {
			var s SessionInfo
			if err := decoder.Decode(&s); err != nil {
				// Break when there are no more documents to decode
				if err != io.EOF {
					return err
				}
				break
			}
			sessionInfos = append(sessionInfos, s)
		}

		zap.S().Infof("sessionInfos: %s", gconv.Strings(sessionInfos))

		return nil
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}

}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = version
}

func initConfig() {
	viper.SetConfigFile("config.yaml")

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		zap.S().Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}
