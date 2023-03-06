package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/shiena/ansicolor"
	"github.com/spf13/viper"
)

var log *logrus.Logger

// InitLog 初始化日志
func InitLog() {
	log = logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		ShowFullLevel:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	// then wrap the log output with it
	logDir := viper.GetString("output.logDir")
	logIO := make([]io.Writer, 0)
	if logDir != "" {
		os.MkdirAll(logDir, os.ModePerm)
		filename := filepath.Join(logDir, time.Now().Format("2006-01-02.log"))
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			panic("日志文件打开失败")
		}
		logIO = append(logIO, file)
	}
	if conf.Output.OutputTerminal {
		logIO = append(logIO, os.Stdout)
	}

	// 融合日志输出
	log.SetOutput(ansicolor.NewAnsiColorWriter(io.MultiWriter(logIO...)))

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
