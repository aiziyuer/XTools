package xtunnel

import (
	util "app/internal/util"
	"context"
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/avast/retry-go"
	"github.com/elazarl/goproxy"
	httpDialer "github.com/mwitkow/go-http-dialer"
	"io"
	"net"
	"net/http"
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
	SshUri              string `yaml:"ssh_uri"`
	SshUser             string `yaml:"ssh_user,omitempty"`
	SshHost             string `yaml:"ssh_host,omitempty"`
	SshPort             int    `yaml:"ssh_port,omitempty"`
	SshPassword         string `yaml:"ssh_password"`
	SshIdentity         string `yaml:"ssh_identity"`
	SshProxyUri         string `yaml:"ssh_proxy_uri,omitempty"`
	SshConfig           *ssh.ClientConfig
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

		var authMethods []ssh.AuthMethod

		if session.SshPassword != "" {
			authMethods = append(authMethods, ssh.Password(session.SshPassword))
		}

		if session.SshIdentity != "" {
			authMethods = append(authMethods, publicKeysByPrivateKeyContent(session.SshIdentity))
		}

		session.SshConfig = &ssh.ClientConfig{
			User:            session.SshUser,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		for _, tunnelUri := range session.SshTunnelStringList {
			m := util.NamedStringSubMatch(
				regexp.MustCompile(
					`(?P<mode>(L|RH|RD|R|LD|LH))=>(?P<host1>[^:]+):(?P<host1_port>\d+)(=>(?P<host2>.+):(?P<host2_port>\d+))?`,
				),
				tunnelUri)

			if len(m) == 0 {
				zap.S().Warnf("tunnelUri: %s, which is not valid.", gconv.Strings(tunnelUri))
				continue
			}

			zap.S().Debugf("%s", gconv.Strings(m))

			if gstr.Contains(m["mode"], "R") {
				session.SshTunnels = append(session.SshTunnels, TunnelInfo{
					Mode:       m["mode"],
					LocalHost:  util.GetAnyString(m["host2"], "0.0.0.0"),
					LocalPort:  gconv.Int(util.GetAnyString(m["host2_port"], "22")),
					RemoteHost: util.GetAnyString(m["host1"], ""),
					RemotePort: gconv.Int(util.GetAnyString(m["host1_port"], "22")),
				})
			} else {
				session.SshTunnels = append(session.SshTunnels, TunnelInfo{
					Mode:       m["mode"],
					LocalHost:  util.GetAnyString(m["host1"], "0.0.0.0"),
					LocalPort:  gconv.Int(util.GetAnyString(m["host1_port"], "22")),
					RemoteHost: util.GetAnyString(m["host2"], ""),
					RemotePort: gconv.Int(util.GetAnyString(m["host2_port"], "22")),
				})
			}

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

func (t *TunnelHandler) reverseTunnel(wg *sync.WaitGroup, sessionInfo *SessionInfo, srcEndpoint, dstEndpoint string) {

	defer wg.Done()

	err := retry.Do(func() error {

		// ssh 链接建立
		serverConn, err := t.getServerConn(
			sessionInfo.SshProxyUri,
			fmt.Sprintf("%s:%d", sessionInfo.SshHost, sessionInfo.SshPort),
			sessionInfo.SshConfig,
		)
		if err != nil {
			zap.S().Errorf("Dial INTO remote server error: %s", err)
			return err
		}

		// 维持远程连接
		go func() {
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

			in, err := listener.Accept()
			if err != nil {
				return err
			}

			out, err := net.Dial("tcp", dstEndpoint)
			if err != nil {
				zap.S().Errorf("Dial INTO local service error: %s", err)
				return err
			}

			zap.S().Infof("ssh_tunnel(R=>%s=>%s) build success.", srcEndpoint, dstEndpoint)

			ch := make(chan error, 2)
			go func() { _, err := io.Copy(in, out); io.NopCloser(out); ch <- err }()
			go func() { _, err := io.Copy(out, in); io.NopCloser(in); ch <- err }()

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
}

func (t *TunnelHandler) forwardTunnel(wg *sync.WaitGroup, sessionInfo *SessionInfo, srcEndpoint, dstEndpoint string) {
	defer wg.Done()

	err := retry.Do(func() error {

		// ssh 链接建立
		serverConn, err := t.getServerConn(
			sessionInfo.SshProxyUri,
			fmt.Sprintf("%s:%d", sessionInfo.SshHost, sessionInfo.SshPort),
			sessionInfo.SshConfig,
		)
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

			in, err := listener.Accept()
			if err != nil {
				return err
			}

			// 打开远端的端口
			out, err := serverConn.Dial("tcp", dstEndpoint)
			if err != nil {
				zap.S().Errorf(
					"open port(%s) ON remote server error: %s",
					dstEndpoint,
					err)
				return err
			}

			ch := make(chan error, 2)
			go func() { _, err := io.Copy(in, out); io.NopCloser(out); ch <- err }()
			go func() { _, err := io.Copy(out, in); io.NopCloser(in); ch <- err }()

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
}

func (t *TunnelHandler) localSocksProxy(wg *sync.WaitGroup, sessionInfo *SessionInfo, localEndpoint string) {

	defer wg.Done()

	err := retry.Do(func() error {

		serverConn, err := t.getServerConn(
			sessionInfo.SshProxyUri,
			fmt.Sprintf("%s:%d", sessionInfo.SshHost, sessionInfo.SshPort),
			sessionInfo.SshConfig,
		)
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

		zap.S().Infof("ssh_tunnel(D=>%s) build success.", localEndpoint)

		if err := serverSocks.ListenAndServe("tcp", localEndpoint); err != nil {
			zap.S().Errorf("failed to create socks5 server: %s", err)
		}

		return nil

	},
		retry.Attempts(5),
	)

	if err != nil {
		zap.S().Errorf(
			"ssh_tunnel(D=>%s) build failed: %s.",
			localEndpoint,
			err,
		)
	}
}

func (t *TunnelHandler) Do(sessionInfos []*SessionInfo) error {

	_ = t.wrapperSessionInfo(sessionInfos)

	var wg sync.WaitGroup
	for _, sessionInfo := range sessionInfos {
		for _, tunnelInfo := range sessionInfo.SshTunnels {
			wg.Add(1)

			switch tunnelInfo.Mode {
			case "R":

				go t.reverseTunnel(
					&wg,
					sessionInfo,
					fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
					fmt.Sprintf("%s:%d", tunnelInfo.RemoteHost, tunnelInfo.RemotePort),
				)

				continue
			case "RD":

				// 远端socks5代理
				continue
			case "RH": // 远端http代理

				go func(endpoint string) {

					err := retry.Do(func() error {

						// ssh 链接建立
						serverConn, err := t.getServerConn(
							sessionInfo.SshProxyUri,
							fmt.Sprintf("%s:%d", sessionInfo.SshHost, sessionInfo.SshPort),
							sessionInfo.SshConfig,
						)
						if err != nil {
							zap.S().Errorf("Dial INTO remote server error: %s", err)
							return err
						}

						listener, err := serverConn.Listen("tcp", endpoint)

						if err != nil {
							zap.S().Errorf("Listen open port ON remote server error: %s", err)
							return err
						}

						defer func() {
							_ = listener.Close()
							_ = serverConn.Close()
						}()

						listenerServer, err := net.Listen("tcp", "127.0.0.1:0")
						if err != nil {
							zap.S().Fatal(err)
						}

						go func() {

							httpProxy := goproxy.NewProxyHttpServer()
							httpProxy.Verbose = true

							if err := http.Serve(listenerServer, httpProxy); err != nil {
								zap.S().Fatal(err)
							}

						}()

						for {

							in, err := listener.Accept()
							if err != nil {
								return err
							}

							out, err := net.Dial("tcp", listenerServer.Addr().String())
							if err != nil {
								zap.S().Errorf("Dial INTO local service error: %s", err)
								return err
							}

							ch := make(chan error, 2)
							go func() { _, err := io.Copy(in, out); io.NopCloser(out); ch <- err }()
							go func() { _, err := io.Copy(out, in); io.NopCloser(in); ch <- err }()

						}

					}, retry.Attempts(5))

					if err != nil {
						zap.S().Errorf("ssh_tunnel(RH=>%s) build failed: %s.",
							endpoint,
							err,
						)
					}

				}(fmt.Sprintf("%s:%d", tunnelInfo.RemoteHost, tunnelInfo.RemotePort))

				continue
			case "L":

				go t.forwardTunnel(
					&wg,
					sessionInfo,
					fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
					fmt.Sprintf("%s:%d", tunnelInfo.RemoteHost, tunnelInfo.RemotePort),
				)

				continue
			case "LD":

				// 本地socks5代理
				go t.localSocksProxy(
					&wg,
					sessionInfo, fmt.Sprintf("%s:%d", tunnelInfo.LocalHost, tunnelInfo.LocalPort),
				)

				continue

			case "LH":

				// 本地http代理
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
