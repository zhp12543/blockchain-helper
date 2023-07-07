package main

import (
	"fmt"
	"github.com/zhp12543/fil-address/config"
	"github.com/zhp12543/fil-address/http/routers"
	"github.com/zhp12543/fil-address/logger"
	"os"
)

const (
	version = "1.0.0"
)

func init() {
	fmt.Println("Version: ", version)
	config.Init()
}

func main() {
	fmt.Println("http service starting... listen:", config.AppConfig.Tcp)
	if err := routers.Init().Run(config.AppConfig.Tcp); err != nil {
		logger.Errorf("start app occurs err: %v", err)
		os.Exit(0)
	}
}
