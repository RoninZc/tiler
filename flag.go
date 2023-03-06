package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	hf         bool
	configPath string
	logLevel   string
)

func InitFlag() {
	flag.BoolVar(&hf, "h", false, "this help")
	flag.StringVar(&configPath, "c", "./conf/conf.toml", "set config `file`")
	flag.StringVar(&logLevel, "l", "info", "set log level (default: info)")
	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = usage
	flag.Parse()

	if hf {
		flag.Usage()
		os.Exit(0)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `tiler version: tiler/v0.1.0
Usage: tiler [-h] [-c filename] [-l logLevel]
`)
	flag.PrintDefaults()
}
