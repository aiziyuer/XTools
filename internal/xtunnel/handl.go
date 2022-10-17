package xtunnel

import (
	util2 "app/internal/util"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"regexp"

	"github.com/avast/retry-go"
	"github.com/gogf/gf/util/gconv"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
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
	SshPassword   string   `yaml:"ssh_password"`
	SshIdentity   string   `yaml:"ssh_identity"`
	SshTunnelUris []string `yaml:"ssh_tunnels"`
	SshTunnels    []TunnelInfo
}

type TunnelHandler struct {
}

func (t *TunnelHandler) Do(sessionInfos []*SessionInfo) error {

	for _, session := range sessionInfos {
		for _, tunnelUri := range session.SshTunnelUris {
			m := util2.NamedStringSubMatch(
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
				SrcHost: util2.GetAnyString(m["src_host"], "0.0.0.0"),
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
				ssh.Password(session.SshPassword),
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
						"ssh_tunnel(R=>%s:%d=>%s:%d) build failed: %s.",
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
