package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/paulmach/orb/maptile"
)

var BreakPointInst *BreakPoint

func InitBreakPoint() {
	path := filepath.Base(conf.BreakPoint.SaveFilePath)
	os.MkdirAll(path, os.ModePerm)
	filapath := filepath.Join(path, fmt.Sprintf("%s.log", conf.Tm.Name))
	file, err := os.OpenFile(filapath, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		panic("break point file open is error")
	}

	// 获取断点记录
	successMap := getBackPoint(file)

	saveChan := make(chan maptile.Tile, conf.Task.Workers)
	BreakPointInst = &BreakPoint{
		file,
		saveChan,
		successMap,
		false,
	}

	SafeExitInst.Register(BreakPointInst.BreakPointSafeFun)

	// 开始断点任务
	go BreakPointInst.Start()
}

// 初始化断点文件
func getBackPoint(file *os.File) map[string]struct{} {
	res := make(map[string]struct{})

	br := bufio.NewReader(file)
	for {
		line, isPrefix, err := br.ReadLine()
		if isPrefix {
			continue
		}
		if err == io.EOF {
			break
		}
		res[string(line)] = struct{}{}
	}
	return res
}

type BreakPoint struct {
	file       *os.File
	saveChan   chan maptile.Tile
	successMap map[string]struct{}
	isClose    bool
}

func (b *BreakPoint) IsSuccessed(tile maptile.Tile) bool {
	key := fmt.Sprintf("%d-%d-%d", tile.X, tile.Y, tile.Z)
	_, ok := b.successMap[key]
	return ok
}

func (b *BreakPoint) SetSuccessed(tile maptile.Tile) {
	if b.isClose {
		return
	}
	b.saveChan <- tile
}

func (b *BreakPoint) Start() {
	log.Infof("断点记录任务已开始")
	for tile := range b.saveChan {
		key := fmt.Sprintf("%d-%d-%d\n", tile.X, tile.Y, tile.Z)
		b.file.WriteString(key)
	}
}

func (b *BreakPoint) BreakPointSafeFun() {
	b.file.Close()
	log.Infof("断点记录任务已安全退出")
}
