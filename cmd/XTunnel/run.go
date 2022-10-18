package main

import (
	"app/internal/xtunnel"
	"github.com/spf13/cobra"
)

var (
	SshUri              = ""
	SshPassword         = ""
	SshTunnelStringList []string
)

func init() {

	runCommand := &cobra.Command{
		Use:   "run",
		Short: "run",
		Long:  "run",
		RunE: func(cmd *cobra.Command, args []string) error {

			h := &xtunnel.TunnelHandler{}
			_ = h.Do([]*xtunnel.SessionInfo{
				{
					SshUri:              SshUri,
					SshPassword:         SshPassword,
					SshTunnelStringList: SshTunnelStringList,
				},
			})

			return nil
		},
	}

	rootCmd.AddCommand(runCommand)

	runCommand.PersistentFlags().StringVar(&SshUri, "ssh_uri", "", "ssh uri")
	runCommand.PersistentFlags().StringVar(&SshPassword, "ssh_password", "", "ssh password")
	runCommand.PersistentFlags().StringSliceVar(&SshTunnelStringList, "ssh_tunnels", []string{}, "ssh tunnel")

}
