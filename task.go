package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
	"github.com/teris-io/shortid"
	pb "gopkg.in/cheggaaa/pb.v1"
)

func InitTask() {
	start := time.Now()

	tm := TileMap{
		Name:   conf.Tm.Name,
		Min:    conf.Tm.Min,
		Max:    conf.Tm.Max,
		Format: conf.Tm.Format,
		URL:    conf.Tm.URL,
	}
	var layers []Layer
	for _, lrs := range conf.Lrs {
		for z := lrs.Min; z <= lrs.Max; z++ {
			c := loadCollection(lrs.Geojson)
			layer := Layer{
				URL:        conf.Tm.URL,
				Zoom:       z,
				Collection: c,
			}
			layers = append(layers, layer)
		}
	}

	task := NewTask(layers, tm)
	// 注册安全退出
	SafeExitInst.Register(task.AbortFun)

	// 开始下载
	task.Download()

	secs := time.Since(start).Seconds()
	log.Printf("\n%.3fs finished...", secs)
}

// Task 下载任务
type Task struct {
	ID           string
	Name         string
	Description  string
	File         string
	Min          int
	Max          int
	Layers       []Layer
	TileMap      TileMap
	Total        int64
	Current      int64
	Bar          *pb.ProgressBar
	workerCount  int
	savePipeSize int
	timeDelay    int
	bufSize      int
	tileWG       sync.WaitGroup
	abort        chan struct{}
	workers      chan struct{}
}

// NewTask 创建下载任务
func NewTask(layers []Layer, m TileMap) *Task {
	if len(layers) == 0 {
		return nil
	}
	id, _ := shortid.Generate()

	task := Task{
		ID:      id,
		Name:    m.Name,
		Layers:  layers,
		Min:     m.Min,
		Max:     m.Max,
		TileMap: m,
	}

	for i := 0; i < len(layers); i++ {
		if layers[i].URL == "" {
			layers[i].URL = m.URL
		}
		layers[i].Count = tilecover.CollectionCount(layers[i].Collection, maptile.Zoom(layers[i].Zoom))
		log.Printf("zoom: %d, tiles: %d \n", layers[i].Zoom, layers[i].Count)
		task.Total += layers[i].Count
	}

	task.workerCount = conf.Task.Workers
	task.savePipeSize = conf.Task.Savepipe
	task.timeDelay = conf.Task.Timedelay
	task.bufSize = conf.Task.BufSize

	task.abort = make(chan struct{})
	task.workers = make(chan struct{}, task.workerCount)

	return &task
}

// Bound 范围
func (task *Task) Bound() orb.Bound {
	bound := orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{-1, -1}}
	for _, layer := range task.Layers {
		for _, g := range layer.Collection {
			bound = bound.Union(g.Bound())
		}
	}
	return bound
}

// Center 中心点
func (task *Task) Center() orb.Point {
	layer := task.Layers[len(task.Layers)-1]
	bound := orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{-1, -1}}
	for _, g := range layer.Collection {
		bound = bound.Union(g.Bound())
	}
	return bound.Center()
}

func (task *Task) SetupFile() error {
	if task.File == "" {
		outdir := conf.Output.Directory
		os.MkdirAll(outdir, os.ModePerm)
		task.File = outdir
	}
	return nil
}

// 结束任务
func (task *Task) AbortFun() {
	task.abort <- struct{}{}
}

// Download 开启下载任务
func (task *Task) Download() {
	task.SetupFile()
	for _, layer := range task.Layers {
		task.downloadLayer(layer)
	}
}

// tileFetcher 瓦片加载器
func (task *Task) tileFetcher(mt maptile.Tile) {
	start := time.Now()
	//workers完成并清退
	defer func() {
		task.tileWG.Done()
		<-task.workers
	}()

	// 获取请求地址
	url := task.TileMap.GetTileURL(mt)
	resp, err := http.Get(url)
	if err != nil {
		log.Debugf("fetch :%s error, details: %s ~", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Debugf("fetch %v tile error, status code: %d ~", url, resp.StatusCode)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("read %v tile error ~ %s", mt, err)
		return
	}
	if len(body) == 0 {
		log.Debugf("nil tile %v ~", mt)
		return //zero byte tiles n
	}
	// tiledata
	td := Tile{
		T: mt,
		C: body,
	}

	if task.TileMap.Format == PBF {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err = zw.Write(body)
		if err != nil {
			log.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}
		td.C = buf.Bytes()
	}

	// enable savingpipe
	task.saveTile(td)
	BreakPointInst.SetSuccessed(mt)

	cost := time.Since(start).Milliseconds()
	log.Debugf("tile(z:%d, x:%d, y:%d), %dms , %.2f kb, %s ...\n", mt.Z, mt.X, mt.Y, cost, float32(len(body))/1024.0, url)
}

// SaveTile 保存瓦片
func (task *Task) saveTile(tile Tile) error {
	// defer task.wg.Done()
	err := saveToFiles(tile, task)
	if err != nil {
		log.Errorf("create %v tile file error ~ %s", tile.T, err)
	}
	return nil
}

// DownloadZoom 下载指定层级
func (task *Task) downloadLayer(layer Layer) {
	log.Infof("Task layer: %s starting", layer)
	bar := pb.New64(layer.Count).Prefix(fmt.Sprintf("Zoom %d : ", layer.Zoom)).Postfix("\n")
	bar.SetRefreshRate(time.Second)
	bar.Start()

	var tilelist = make(chan maptile.Tile, task.bufSize)

	go tilecover.CollectionChannel(layer.Collection, maptile.Zoom(layer.Zoom), tilelist)

	for tile := range tilelist {
		// 如果已经在成功列表里
		if BreakPointInst.IsSuccessed(tile) {
			log.Infoln("芜湖，该文件已下载，跳过")
			bar.Increment()
			continue
		}
		select {
		// 向队列发送数据
		case task.workers <- struct{}{}:
			bar.Increment()
			//设置请求发送间隔时间
			time.Sleep(time.Duration(task.timeDelay) * time.Millisecond)
			task.tileWG.Add(1)
			go task.tileFetcher(tile)
		case <-task.abort:
			log.Infof("Task %s got canceled.", task.Name)
		}
	}
	//等待该层结束
	task.tileWG.Wait()
	bar.FinishPrint(fmt.Sprintf("Task %s Zoom %d finished ~", task.ID, layer.Zoom))

}
