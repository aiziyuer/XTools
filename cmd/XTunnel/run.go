package main

import (
	"app/internal/xtunnel"
	"github.com/spf13/cobra"
)

var (
	SshUri        = ""
	SshPassword   = ""
	SshTunnelUris []string
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
					SshUri:        SshUri,
					SshPassword:   SshPassword,
					SshTunnelUris: SshTunnelUris,
				},
			})

			return nil
		},
	}

	rootCmd.AddCommand(runCommand)

	runCommand.PersistentFlags().StringVar(&SshUri, "ssh_uri", "", "ssh uri")
	runCommand.PersistentFlags().StringVar(&SshUri, "ssh_password", "", "ssh password")
	runCommand.PersistentFlags().StringSliceVar(&SshTunnelUris, "ssh_tunnel_uri", []string{}, "ssh tunnel uri")

}
