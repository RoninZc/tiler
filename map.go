package main

import (
	"strconv"
	"strings"

	"github.com/paulmach/orb/maptile"
)

// TileMap 瓦片地图类型
type TileMap struct {
	ID          int
	Name        string
	Description string
	Min         int
	Max         int
	Format      string
	JSON        string
	URL         string
	Token       string
}

// TileURL 获取瓦片URL
func (m *TileMap) GetTileURL(t maptile.Tile) string {
	url := strings.Replace(m.URL, "{x}", strconv.Itoa(int(t.X)), -1)
	url = strings.Replace(url, "{y}", strconv.Itoa(int(t.Y)), -1)
	url = strings.Replace(url, "{z}", strconv.Itoa(int(t.Z)), -1)
	return url
}
