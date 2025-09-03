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
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const defaultPort = 8000

var port = defaultPort

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	port = portUse(parseArgPort())
	_ = initIndex()
	r := gin.Default()
	r.StaticFS("/", gin.Dir(appDirectory(), true))

	go func() {
		_ = r.Run(fmt.Sprintf(":%d", port))
	}()
	go func() {
		var osName = strings.ToLower(runtime.GOOS)
		switch osName {
		case "windows":
			cmd := exec.Command("cmd", "/c", "start", fmt.Sprintf("http://127.0.0.1:%d/readme.html", port))
			_ = cmd.Run()
		case "darwin":
			cmd := exec.Command("open", fmt.Sprintf("http://127.0.0.1:%d/readme.html", port))
			_ = cmd.Run()
		}
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

func parseNetworkList() map[string]string {
	var m = make(map[string]string)
	interfaceList, _ := net.Interfaces()
	for _, tmpInterface := range interfaceList {
		// 只看Up和Running状态，不是广播和多播
		if tmpInterface.Flags&net.FlagUp == 0 || tmpInterface.Flags&net.FlagRunning == 0 {
			continue
		}
		addrList, _ := tmpInterface.Addrs()
		for _, tmpAddr := range addrList {
			tmpIp, _, err := net.ParseCIDR(tmpAddr.String())
			if err != nil {
				continue
			}
			var tmpIpV4 = net.ParseIP(tmpIp.String()).To4()
			if tmpIpV4 == nil || len(tmpIpV4) != 4 {
				continue
			}
			if tmpIpV4[3] == 0x1 {
				continue
			}
			m[tmpInterface.Name] = tmpIp.String()
		}
	}
	return m
}

func generateNetworkHtml() string {
	var m = parseNetworkList()
	if m == nil || len(m) <= 0 {
		return ""
	}
	var s = "<ul>"
	for tmpName, tmpIp := range m {
		var tmpUrl = fmt.Sprintf("http://%s:%d", tmpIp, port)
		s += fmt.Sprintf(`<li><a href="%s" title="%s">%s</a> (%s)</li>`, tmpUrl, tmpName, tmpUrl, strings.ToUpper(tmpName))
	}
	s += "</ul>"

	return s
}

func initIndex() error {
	f := filepath.Join(appDirectory(), "readme.html")
	//_, err := os.Stat(f)
	//if err == nil {
	//	return nil
	//}

	var htmlString = strings.ReplaceAll(htmlTemplate, "__WEB_PATH__", appDirectory())
	htmlString = strings.ReplaceAll(htmlString, "__HOST_LIST__", generateNetworkHtml())

	fi, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_APPEND|syscall.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	_ = fi.Truncate(0)
	_, err = fi.WriteString(htmlString)
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
    <title>fs-server</title>
    <style type="text/css">
        html, body, .c {
            height: 100%;
            margin: 0;
        }

        body {
            margin: 0;
        }

        .c {
            display: flex;
            justify-content: center;
            align-items: center;
            flex-direction: column;

            font-family: "Microsoft JhengHei UI";
            text-align: center;
        }

        .c .title {
            font-size: 86px;
            margin-bottom: 10px;
        }

        .c .info {
            color: #6e6e6e;
            font-size: 28px;
        }

        .c .info span {
            border-bottom: 2px solid #a1a1a1;
            padding: 2px;
            cursor: copy;
        }
        ul {
            list-style-type: square;
            line-height: 200%;
            font-size: 22px;
            text-align: left;
        }
    </style>
</head>
<body>

<div class="c">
    <div class="title">你好, Hello</div>
    <a class="info">file://<span>__WEB_PATH__</span></a>
	__HOST_LIST__
</div>

</body>
</html>

`
