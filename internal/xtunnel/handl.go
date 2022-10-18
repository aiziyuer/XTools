package xtunnel

import (
	util "app/internal/util"
	"context"
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/avast/retry-go"
	httpDialer "github.com/mwitkow/go-http-dialer"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"sync"

	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

type TunnelInfo struct {
	Mode       string
	LocalHost  string
	LocalPort  int
	RemoteHost string
	RemotePort int
}

type SessionInfo struct {
	SshUri              string   `yaml:"ssh_uri"`
	SshUser             string   `yaml:"ssh_user,omitempty"`
	SshHost             string   `yaml:"ssh_host,omitempty"`
	SshPort             int      `yaml:"ssh_port,omitempty"`
	SshPassword         string   `yaml:"ssh_password"`
	SshIdentity         string   `yaml:"ssh_identity"`
	SshProxyUri         string   `yaml:"ssh_proxy_uri"`
	SshTunnelStringList []string `yaml:"ssh_tunnels"`
	SshTunnels          []TunnelInfo
}

type TunnelHandler struct {
}

func (t *TunnelHandler) wrapperSessionInfo(sessionInfos []*SessionInfo) error {

	for _, session := range sessionInfos {
		// 处理session的uri

		if m := util.NamedStringSubMatch(
			regexp.MustCompile(
				`(?P<user>[^:]+)@(?P<host>[^:]+)(:(?P<port>\d+))?`,
			),
			session.SshUri); len(m) == 0 {
			zap.S().Warnf("session.SshUri: %s, which is not valid, ignore.", gconv.Strings(session.SshUri))
			continue
		} else {
			session.SshUser = m["user"]
			session.SshHost = m["host"]
			session.SshPort = gconv.Int(util.GetAnyString(m["port"], "22"))
		}

		for _, tunnelUri := range session.SshTunnelStringList {
			m := util.NamedStringSubMatch(
				regexp.MustCompile(
					`(?P<mode>(L|R|D))=>(?P<local_host>[^:]+):(?P<local_port>\d+)(=>(?P<remote_host>.+):(?P<remote_port>\d+))?`,
				),
				tunnelUri)

			if len(m) == 0 {
				zap.S().Warnf("tunnelUri: %s, which is not valid.", gconv.Strings(tunnelUri))
				continue
			}

			zap.S().Debugf("%s", gconv.Strings(m))
			session.SshTunnels = append(session.SshTunnels, TunnelInfo{
				Mode:       m["mode"],
				LocalHost:  util.GetAnyString(m["local_host"], "0.0.0.0"),
				LocalPort:  gconv.Int(m["local_port"]),
				RemoteHost: util.GetAnyString(m["remote_host"], ""),
				RemotePort: gconv.Int(util.GetAnyString(m["remote_port"], "-1")),
			})
		}
	}
	zap.S().Infof("sessionInfos: %s", gconv.Strings(sessionInfos))

	return nil
}

func (t *TunnelHandler) getServerConn(proxyUri, sshEndpoint string, sshConfig *ssh.ClientConfig) (serverConn *ssh.Client, err error) {

	if len(proxyUri) != 0 {

		var dialer proxy.Dialer

		if gstr.HasPrefix(proxyUri, "socks5://") {
			dialer, err = proxy.SOCKS5("tcp", gstr.ReplaceByMap(proxyUri, map[string]string{"socks5://": ""}), nil, proxy.Direct)
			if err != nil {
				return nil, err
			}
		}

		if gstr.HasPrefix(proxyUri, "http://") {
			proxyUrl, err := url.Parse(proxyUri)
			if err != nil {
				zap.S().Error(err)
				return nil, err
			}
			dialer = httpDialer.New(proxyUrl)
		}

		conn, err := dialer.Dial("tcp", sshEndpoint)
		if err != nil {
			return serverConn, err
		}

		c, chans, reqs, err := ssh.NewClientConn(conn, conn.RemoteAddr().String(), sshConfig)
		if err != nil {
			return serverConn, err
		}
		serverConn = ssh.NewClient(c, chans, reqs)

		return serverConn, nil

	} else {

		// ssh 链接建立
		serverConn, err = ssh.Dial("tcp", sshEndpoint, sshConfig)
		if err != nil {
			zap.S().Errorf("Dial INTO remote server error: %s", err)
			return serverConn, err
		}
	}

	return serverConn, err
}

func (t *TunnelHandler) Do(sessionInfos []*SessionInfo) error {

	_ = t.wrapperSessionInfo(sessionInfos)

	var wg sync.WaitGroup
	for _, session := range sessionInfos {

		var authMethods []ssh.AuthMethod

		if session.SshPassword != "" {
			authMethods = append(authMethods, ssh.Password(session.SshPassword))
		}

		if session.SshIdentity != "" {
			authMethods = append(authMethods, publicKeysByPrivateKeyContent(session.SshIdentity))
		}

		sshConfig := &ssh.ClientConfig{
			User:            session.SshUser,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		for _, tunnelInfo := range session.SshTunnels {
			wg.Add(1)

			switch tunnelInfo.Mode {
			case "R":
				go func(proxyUri, sshEndpoint, srcEndpoint, dstEndpoint string) {
					defer wg.Done()

					err := retry.Do(func() error {

						// ssh 链接建立
						serverConn, err := t.getServerConn(proxyUri, sshEndpoint, sshConfig)
						if err != nil {
							zap.S().Errorf("Dial INTO remote server error: %s", err)
							return err
						}

						// 维持远程连接
						go func() {
							for {
								session, err := serverConn.NewSession()
								if err != nil {
									zap.S().Fatal(err)
								}

								_, err = session.Output("echo '1' ")
								if err != nil {
									zap.S().Fatal(err)
								}
							}
						}()

						listener, err := serverConn.Listen("tcp", srcEndpoint)

						if err != nil {
							zap.S().Errorf("Listen open port ON remote server error: %s", err)
							return err
						}

						defer func() {
							_ = listener.Close()
							_ = serverConn.Close()
						}()

						for {
							out, err := net.Dial("tcp", dstEndpoint)
							if err != nil {
								zap.S().Errorf("Dial INTO local service error: %s", err)
								return err
							}

							zap.S().Infof("ssh_tunnel(R=>%s=>%s) build success.", srcEndpoint, dstEndpoint)

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

						zap.S().Errorf("ssh_tunnel(R=>%s=>%s) build failed: %s.",
							srcEndpoint,
							dstEndpoint,
							err,
						)

					}

				}(
					session.SshProxyUri,
					fmt.Sprintf("%s:%d", session.SshHost, session.SshPort),
					fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
					fmt.Sprintf("%s:%d", tunnelInfo.RemoteHost, tunnelInfo.RemotePort),
				)
				continue
			case "L":
				go func(proxyUri, sshEndpoint, srcEndpoint, dstEndpoint string) {
					defer wg.Done()

					err := retry.Do(func() error {

						// ssh 链接建立
						serverConn, err := t.getServerConn(proxyUri, sshEndpoint, sshConfig)
						if err != nil {
							zap.S().Errorf("Dial INTO remote server error: %s", err)
							return err
						}

						// 维持远程连接
						go func() {
							for {
								session, err := serverConn.NewSession()
								if err != nil {
									zap.S().Fatal(err)
								}

								_, err = session.Output("echo '1' ")
								if err != nil {
									zap.S().Fatal(err)
								}
							}
						}()

						listener, err := net.Listen("tcp", srcEndpoint)
						if err != nil {
							zap.S().Errorf("Listen open port(%s) ON local server error: %s",
								srcEndpoint,
								err)
							return err
						}

						defer func() {
							_ = listener.Close()
							_ = serverConn.Close()
						}()

						zap.S().Infof("ssh_tunnel(L=>%s=>%s) build success.", srcEndpoint, dstEndpoint)

						for {

							// 打开远端的端口
							out, err := serverConn.Dial("tcp", dstEndpoint)
							if err != nil {
								zap.S().Errorf(
									"open port(%s) ON remote server error: %s",
									dstEndpoint,
									err)
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
						zap.S().Errorf(
							"ssh_tunnel(L=>%s=>%s) build failed: %s.",
							srcEndpoint,
							dstEndpoint,
							err,
						)
					}

				}(
					session.SshProxyUri,
					fmt.Sprintf("%s:%d", session.SshHost, session.SshPort),
					fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
					fmt.Sprintf("%s:%d", tunnelInfo.RemoteHost, tunnelInfo.RemotePort),
				)
				continue
			case "D":
				go func(proxyUri, sshEndpoint, localEnpoint string) {
					defer wg.Done()

					err := retry.Do(func() error {

						serverConn, err := t.getServerConn(proxyUri, sshEndpoint, sshConfig)
						if err != nil {
							zap.S().Fatal(err)
						}

						go func(serverConn *ssh.Client) {
							for {
								session, err := serverConn.NewSession()
								if err != nil {
									zap.S().Error(err)
									return
								}

								_, err = session.Output("echo '1' ")
								if err != nil {
									zap.S().Error(err)
									return
								}
							}
						}(serverConn)

						defer func() {
							_ = serverConn.Close()
						}()

						serverSocks, err := socks5.New(&socks5.Config{
							Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
								return serverConn.Dial(network, addr)
							},
						})
						if err != nil {
							zap.S().Errorf("socks5 New error: %s", err)
						}

						zap.S().Infof("ssh_tunnel(D=>%s) build success.", localEnpoint)

						if err := serverSocks.ListenAndServe("tcp", localEnpoint); err != nil {
							zap.S().Errorf("failed to create socks5 server: %s", err)
						}

						return nil

					},
						retry.Attempts(5),
					)

					if err != nil {
						zap.S().Errorf(
							"ssh_tunnel(D=>%s) build failed: %s.",
							localEnpoint,
							err,
						)
					}

				}(
					session.SshProxyUri,
					fmt.Sprintf("%s:%d", session.SshHost, session.SshPort),
					fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
				)

				continue
			}
		}

	}
	wg.Wait()

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
