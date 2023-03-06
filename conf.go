package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var conf *Conf

type Conf struct {
	App struct {
		Version string `toml:"version"`
		Title   string `toml:"title"`
	} `toml:"app"`
	Output struct {
		Directory      string `toml:"directory"`
		LogDir         string `toml:"logDir"`
		OutputTerminal bool   `toml:"outputTerminal"`
	} `toml:"output"`
	Task struct {
		Workers   int `toml:"workers"`
		Savepipe  int `toml:"savepipe"`
		Timedelay int `toml:"timedelay"`
		BufSize   int `toml:"bufSize"`
	} `toml:"task"`
	BreakPoint struct {
		SaveFilePath string `toml:"saveFilePath"`
	} `toml:"breakPoint"`
	Tm struct {
		Name   string `toml:"name"`
		Min    int    `toml:"min"`
		Max    int    `toml:"max"`
		Format string `toml:"format"`
		URL    string `toml:"url"`
	} `toml:"tm"`
	Lrs []struct {
		Min     int    `toml:"min"`
		Max     int    `toml:"max"`
		Geojson string `toml:"geojson"`
	} `toml:"lrs"`
}

// initConf 初始化配置
func InitConf(cfgFile string) {
	if cfgFile == "" {
		cfgFile = "conf.toml"
	}
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		fmt.Printf("config file(%s) not exist", cfgFile)
		os.Exit(1)
	}
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match
	err := viper.ReadInConfig()
	if err != nil {
		log.Warnf("read config file(%s) error, details: %s", viper.ConfigFileUsed(), err)
	}
	// 设置默认值
	viper.SetDefault("app.version", "v 0.1.0")
	viper.SetDefault("app.title", "MapCloud Tiler")
	viper.SetDefault("output.format", "mbtiles")
	viper.SetDefault("output.directory", "output")
	viper.SetDefault("task.workers", 4)
	viper.SetDefault("task.savepipe", 1)
	viper.SetDefault("task.timedelay", 0)

	err = viper.Unmarshal(&conf)
	if err != nil {
		panic("配置文件解析失败")
	}
}
