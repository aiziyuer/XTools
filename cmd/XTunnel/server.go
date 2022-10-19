package main

import (
	"app/internal/global"
	"app/internal/xtunnel"
	"bytes"
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
)

func init() {

	serverCommand := &cobra.Command{
		Use:   "server",
		Short: "server",
		Long:  "server",
		RunE: func(cmd *cobra.Command, args []string) error {

			var sessionInfos []*xtunnel.SessionInfo

			appName := global.AppConfig().GetAppName()
			configFilePath, err := homedir.
				Expand(fmt.Sprintf("~/.config/%s/%s.yaml", appName, appName))
			if err != nil {
				zap.S().Fatal(err)
			}

			decoder := yaml.NewDecoder(bytes.NewReader(gfile.GetBytes(configFilePath)))
			for {
				var s xtunnel.SessionInfo
				if err := decoder.Decode(&s); err != nil {
					// Break when there are no more documents to decode
					if err != io.EOF {
						zap.S().Fatal(err)
					}

					break
				}

				sessionInfos = append(sessionInfos, &s)
			}

			h := &xtunnel.TunnelHandler{}
			_ = h.Do(sessionInfos)

			return nil
		},
	}

	rootCmd.AddCommand(serverCommand)

}
