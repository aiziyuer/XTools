package main

import (
	"app/internal/util"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/os/gfile"
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

var mainCmd = &cobra.Command{
	Use:   appName,
	Short: "File Server By Stream",
	Long:  `File Server By Stream.`,
	Run: func(cmd *cobra.Command, args []string) {

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		done := make(chan bool, 1)

		go func() {
			msg := <-sig
			zap.S().Warnf("receive msg: %s", gconv.String(msg))
			done <- true
		}()

		// 定时任务
		go func() {

			gocron.Every(1).Second().Do(func() {
				fmt.Println("I am running task.")
				cmd := fmt.Sprintf("rsync -av rsync://%s/scripts/ /scripts &> /tmp/rsync.log", center_host)
				if _, err := exec.Command("bash", "-c", cmd).CombinedOutput(); err != nil {
					zap.S().Errorf("cmd: %s has errors:\n%s", cmd, err)
				}
			})

			<-gocron.Start()

		}()

		// watch文件服务

		// web  服务
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

		// 正常通知systemd服务已经启动
		daemon.SdNotify(false, daemon.SdNotifyReady)

		<-done

	},
}

func init() {
	mainCmd.PersistentFlags().IntVarP(&listenPort, "port", "P", 9900, "listen server port")
	mainCmd.PersistentFlags().StringVar(&center_host, "center_host", "127.0.0.1", "remote scripts center host to rsync")
}

func main() {
	util.SetupLogs(fmt.Sprintf("/var/log/%s/info.log", appName))
	if err := mainCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
