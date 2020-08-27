package main

import (
	"flag"
	"fmt"

	"github.com/branthz/devbroker/config"
	"github.com/branthz/devbroker/service"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "c", "./broker.toml", "config path")
	flag.Parse()
	c := config.NewConfig(configPath)
	fmt.Println(*c)
	s, err := service.NewService()
	if err != nil {
		panic(err)
	}
	s.Run()
}
