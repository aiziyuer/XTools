package main

import (
	"app/internal/util"
	"fmt"
	zapcore2 "go.uber.org/zap/zapcore"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/jasonlvhit/gocron"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// --port 9900 rsync://127.0.0.1/scripts/
var appName = "XtraceExporter"
var listenPort = 9900
var center_host = "127.0.0.1"
var sessionMap sync.Map

// var sessionMap = make(map[string]*Session, 0)
var lock = sync.RWMutex{}

type Std_out_err_t struct {
	data io.ReadCloser
}

func (s Std_out_err_t) Read() string {
	data, _ := ioutil.ReadAll(s.data)
	return string(data)
}

type Session struct {
	Name   string
	Cmd    *exec.Cmd
	Stdout Std_out_err_t
	Stderr Std_out_err_t
}

var mainCmd = &cobra.Command{
	Use:        appName,
	Aliases:    []string{},
	SuggestFor: []string{},
	Short:      appName,
	Long:       appName,
	Example:    "",
	Run: func(cmd *cobra.Command, args []string) {
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		done := make(chan bool, 1)
		go func() {
			msg := <-sig
			zap.S().Warnf("receive msg: %s", gconv.String(msg))
			done <- true
		}()
		go func() {
			gocron.Every(1).Second().Do(func() {
				cmd := fmt.Sprintf("rsync -av rsync://%s/scripts/ /scripts &> /tmp/rsync.log", center_host)
				if _, err := exec.Command("bash", "-c", cmd).CombinedOutput(); err != nil {
					zap.S().Errorf("cmd: %s has errors:\n%s", cmd, err)
				}
			})
			gocron.Every(1).Second().Do(func() {
				if files, err := ioutil.ReadDir("/scripts/"); err != nil {
					log.Fatal(err)
				} else {
					for _, file := range files {
						if file.IsDir() || !gstr.HasSuffix(file.Name(), "stp") {
							continue
						}

						module_name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

						func() {

							lock.Lock()
							defer lock.Unlock()

							if _, ok := sessionMap.Load(module_name); !ok {
								cmd := exec.Command(
									"/usr/bin/stap",
									"-g",
									"-m",
									fmt.Sprintf("__xtrace_%s", module_name),
									fmt.Sprintf("/scripts/%s", file.Name()),
								)
								stdout, _ := cmd.StdoutPipe()
								stderr, _ := cmd.StderrPipe()
								cmd.SysProcAttr = &syscall.SysProcAttr{
									Pdeathsig: syscall.SIGTERM, // 子进程跟随父进程一起停止
								}
								zap.S().Debugf("create session: %s, command:\n%s", module_name, cmd.String())
								if err = cmd.Start(); err != nil {
									zap.S().Errorf("cmd: %s has error: %s.", cmd, err)
								}
								sessionMap.Store(module_name, &Session{Cmd: cmd, Stdout: Std_out_err_t{stdout}, Stderr: Std_out_err_t{stderr}})
							}

						}()

					}
				}
			})
			<-gocron.Start()
		}()
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal("NewWatcher failed: ", err)
		}
		defer watcher.Close()
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					switch event.Op {
					case fsnotify.Write, fsnotify.Rename, fsnotify.Create:
						if !gregex.IsMatchString(`^.+.stp$`, event.Name) {
							continue
						}
					default:
						continue
					}
					_, fileName := filepath.Split(event.Name)
					module_name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

					zap.S().Debugf("file changed: %s, event: %s", fileName, event.Op.String())

					func() {

						lock.Lock()
						defer lock.Unlock()

						if v, ok := sessionMap.LoadAndDelete(module_name); ok {

							if s, ok := v.(*Session); ok {
								zap.S().Debugf("delete session: %s", module_name)
								syscall.Kill(s.Cmd.Process.Pid, syscall.SIGTERM)
								ioutil.NopCloser(s.Stderr.data)
								ioutil.NopCloser(s.Stdout.data)
								s.Cmd.Process.Wait()
							}

						}

					}()

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					zap.S().Errorf("error:", err)
				}
			}
		}()
		if err = watcher.Add("/scripts"); err != nil {
			zap.S().Fatal("Add failed:", err)
		}
		go func() {
			r := gin.Default()
			r.GET("/:script_name", func(c *gin.Context) {
				script_name := c.Param("script_name")
				module_name := fmt.Sprintf("__xtrace_%s", gstr.TrimRightStr(script_name, ".stp"))
				file_name := fmt.Sprintf("/proc/systemtap/%s/metrics", module_name)
				zap.S().Infof("file_name: %s", file_name)
				c.String(http.StatusOK, gfile.GetContents(file_name))
			})
			r.Run(fmt.Sprintf(":%d", listenPort))
		}()
		daemon.SdNotify(false, daemon.SdNotifyReady)
		<-done
	},
}

func init() {
	mainCmd.PersistentFlags().IntVarP(&listenPort, "port", "P", 9900, "listen server port")
	mainCmd.PersistentFlags().StringVar(&center_host, "center_host", "127.0.0.1", "remote scripts center host to rsync")
}

func main() {
	util.SetupDefaultLogs(zapcore2.InfoLevel, fmt.Sprintf("/var/log/%s/info.log", appName))
	if err := mainCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
