package main

import (
	"app/internal/util"
	"bufio"
	"fmt"
	zapcore2 "go.uber.org/zap/zapcore"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func handleConnection(c net.Conn) {
	defer c.Close()

	fmt.Printf("Client %s\n", c.RemoteAddr().String())

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		fileName := strings.TrimSpace(string(netData))
		fmt.Printf("%s: start send to client(%s).\n", fileName, c.RemoteAddr().String())

		file, err := os.Open(strings.TrimSpace(fileName))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		n, err := io.Copy(c, file)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("%s: %d bytes sent to client(%s).\n", fileName, n, c.RemoteAddr().String())
	}
}

var listenPort = 4444

var mainCmd = &cobra.Command{
	Use:   "FileStreamServer",
	Short: "File Server By Stream",
	Long:  `File Server By Stream.`,
	Run: func(cmd *cobra.Command, args []string) {

		l, err := net.Listen("tcp4", fmt.Sprintf(":%d", listenPort))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer l.Close()
		rand.Seed(time.Now().Unix())

		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			go handleConnection(c)
		}

	},
}

func init() {
	mainCmd.PersistentFlags().IntVarP(&listenPort, "port", "P", 4444, "listen server port")
}

func main() {
	util.SetupDefaultLogs(zapcore2.InfoLevel, "/var/log/FileStream/info.log")
	if err := mainCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
