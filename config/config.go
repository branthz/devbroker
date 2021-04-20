package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/branthz/utarrow/lib/log"
)

type localConfig struct {
	Version   string
	Name      string
	Listen    string
	LogLevel  string
	LogPath   string
	QueueSize int
	DataPath  string
}

var (
	//LocalConfig ...
	LocalConfig localConfig
)

func NewConfig(path string) *localConfig {
	_, err := toml.DecodeFile(path, &LocalConfig)
	if err != nil {
		fmt.Println("config:", err)
		os.Exit(-1)
	}
	err = log.SetupRotate(LocalConfig.LogPath, LocalConfig.LogLevel)
	if err != nil {
		fmt.Println("log start", err)
		os.Exit(-1)
	}
	return &LocalConfig
}

func GetConfig() localConfig {
	return LocalConfig
}
