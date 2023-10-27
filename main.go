package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const defaultPort = 8000

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	_ = initIndex()
	r := gin.Default()
	r.Static("/", appDirectory())

	port := portUse(parseArgPort())

	go func() {
		_ = r.Run(fmt.Sprintf("127.0.0.1:%d", port))
	}()
	go func() {
		cmd := exec.Command("cmd", "/c", "start", fmt.Sprintf("http://127.0.0.1:%d", port))
		_ = cmd.Run()
	}()

	select {
	case _sig := <-sig:
		log.Println(fmt.Sprintf("[stop] %v\n", _sig))
	}

}

func parseArgPort() int {
	var port = defaultPort
	fp := flag.Int("p", port, "http端口号")
	fport := flag.Int("port", port, "http端口号")
	flag.Parse()

	if len(os.Args) > 1 {
		port, _ = strconv.Atoi(os.Args[1])
	}
	if *fp != defaultPort && *fp > 0 {
		port = *fp
	}
	if *fport != defaultPort && *fport > 0 {
		port = *fport
	}
	return port
}

func appDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return dir
}

func portUse(port int) int {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return portUse(port + 1)
	}
	defer func() { _ = listener.Close() }()
	return port
}

func initIndex() error {
	f := filepath.Join(appDirectory(), "index.html")
	fi, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = fi.WriteString(strings.ReplaceAll(htmlTemplate, "__WEB_PATH__", appDirectory()))
	if err != nil {
		return err
	}
	return nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
    <style type="text/css">
        html,body,.c{height: 100%;}
        .c{
            display: flex;
            justify-content: center;
            align-items: center;

            font-family: "Microsoft JhengHei UI";
            text-align: center;
        }
        .c .title{
            font-size: 86px;
        }
        .c .info{
            color: #6e6e6e;
            margin-top: 20px;
            font-size: 38px;
        }
        .c .info  span{
            border-bottom: 2px solid #a1a1a1;
            padding: 2px;
            cursor: copy;
        }
    </style>
</head>
<body>

<div class="c">
    <div>
        <div class="title">Hello, 你好</div>
        <div class="info">file:///<span>__WEB_PATH__</span></div>
    </div>
</div>

</body>
</html>

`
