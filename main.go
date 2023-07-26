package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"os"
	"path/filepath"
)

func main() {
	r := gin.Default()
	r.Static("/", appDirectory())
	_ = r.Run(fmt.Sprintf(":%d", portUse(8000)))
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
