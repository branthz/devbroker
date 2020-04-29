package main

import (
	"flag"
	"fmt"

	"github.com/branthz/devbroker/config"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "c", "./broker.toml", "config path")
	flag.Parse()
	c := config.NewConfig(configPath)
	fmt.Println(*c)
	//s:=service.NewService()
}
