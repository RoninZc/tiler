package main

import (
	"github.com/paulmach/orb"

	"github.com/paulmach/orb/maptile"
)

// TileSize 默认瓦片大小
const TileSize = 256

// ZoomMin 最小级别
const ZoomMin = 0

// ZoomMax 最大级别
const ZoomMax = 20

// Tile 自定义瓦片存储
type Tile struct {
	T maptile.Tile
	C []byte
}

// Layer 级别&瓦片数
type Layer struct {
	URL        string
	Zoom       int
	Count      int64
	Collection orb.Collection
}

// Constants representing TileFormat types
const (
	GZIP string = "gzip" // encoding = gzip
	ZLIB        = "zlib" // encoding = deflate
	PNG         = "png"
	JPG         = "jpg"
	PBF         = "pbf"
	WEBP        = "webp"
)
