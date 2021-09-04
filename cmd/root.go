package cmd

import (
	"app/util"
	"bytes"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/util/gconv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
)

type TunnelInfo struct {
	Mode    string
	SrcHost string
	SrcPort int
	DstHost string
	DstPort int
}

type SessionInfo struct {
	SshUri        string   `yaml:"ssh_uri"`
	SshUser       string   `yaml:"ssh_user,omitempty"`
	SshHost       string   `yaml:"ssh_host,omitempty"`
	SshPort       int      `yaml:"ssh_port,omitempty"`
	SshIdentity   string   `yaml:"ssh_identity"`
	SshTunnelUris []string `yaml:"ssh_tunnels"`
	SshTunnels    []TunnelInfo
}

func publicKeysByPrivateKeyContent(privateKeyContent string) ssh.AuthMethod {

	buffer := gconv.Bytes(privateKeyContent)

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		zap.S().Fatalf("Cannot parse SSH public key content: %s", privateKeyContent)
		return nil
	}
	return ssh.PublicKeys(key)
}

var rootCmd = &cobra.Command{
	Use: "SshTunnel",
	RunE: func(cmd *cobra.Command, args []string) error {

		var sessionInfos []*SessionInfo

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

			// 忽略无效的记录
			if len(s.SshUri) == 0 || len(s.SshHost) == 0 {
				continue
			}

			sessionInfos = append(sessionInfos, &s)
		}

		for _, session := range sessionInfos {
			for _, tunnelUri := range session.SshTunnelUris {
				m := util.NamedStringSubMatch(
					regexp.MustCompile(
						`(?P<mode>(L|R))=>(?P<src_host>.*):(?P<src_port>\d+)=>(?P<dst_host>.+):(?P<dst_port>\d+)`,
					),
					tunnelUri)

				if len(m) == 0 {
					zap.S().Warnf("tunnelUri: %s, which is not valid.", gconv.Strings(tunnelUri))
					continue
				}

				zap.S().Debugf("%s", gconv.Strings(m))
				session.SshTunnels = append(session.SshTunnels, TunnelInfo{
					Mode:    m["mode"],
					SrcHost: util.GetAnyString(m["src_host"], "0.0.0.0"),
					SrcPort: gconv.Int(m["src_port"]),
					DstHost: m["dst_host"],
					DstPort: gconv.Int(m["dst_port"]),
				})
			}
		}
		zap.S().Infof("sessionInfos: %s", gconv.Strings(sessionInfos))

		for _, session := range sessionInfos {

			sshConfig := &ssh.ClientConfig{
				// SSH connection username
				User: session.SshUser,
				Auth: []ssh.AuthMethod{
					// put here your private key path
					publicKeysByPrivateKeyContent(session.SshIdentity),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			for _, tunnelInfo := range session.SshTunnels {

				switch tunnelInfo.Mode {
				case "R":
					err := retry.Do(func() error {
						// Connect to SSH remote server using serverEndpoint
						serverConn, err := ssh.Dial("tcp",
							fmt.Sprintf("%s:%d", session.SshHost, session.SshPort),
							sshConfig,
						)
						if err != nil {
							zap.S().Errorf("Dial INTO remote server error: %s", err)
							return err
						}

						// Listen on remote server port
						listener, err := serverConn.Listen("tcp",
							fmt.Sprintf("%s:%d", tunnelInfo.SrcHost, tunnelInfo.SrcPort),
						)
						if err != nil {
							zap.S().Errorf("Listen open port ON remote server error: %s", err)
							return err
						}

						defer func() {
							_ = listener.Close()
							_ = serverConn.Close()
						}()

						for {
							out, err := net.Dial("tcp",
								fmt.Sprintf("%s:%d", tunnelInfo.DstHost, tunnelInfo.DstPort))
							if err != nil {
								zap.S().Errorf("Dial INTO local service error: %s", err)
								return err
							}

							in, err := listener.Accept()
							if err != nil {
								return err
							}

							ch := make(chan error, 2)
							go func() { _, err := io.Copy(in, out); ioutil.NopCloser(out); ch <- err }()
							go func() { _, err := io.Copy(out, in); ioutil.NopCloser(in); ch <- err }()

						}

					},
						retry.Attempts(5),
					)
					if err != nil {
						zap.S().Fatalf(
							"tunnel(R=>%s:%d=>%s:%d) build failed: %s.",
							tunnelInfo.SrcHost,
							tunnelInfo.SrcPort,
							tunnelInfo.DstHost,
							tunnelInfo.DstPort,
							err,
						)
						return err
					}

					break
				case "L":

					break
				}

			}
		}

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
