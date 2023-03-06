package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

func saveToFiles(tile Tile, task *Task) error {
	rootdir := filepath.Base(task.File)
	dir := filepath.Join(rootdir, fmt.Sprintf(`%d`, tile.T.Z), fmt.Sprintf(`%d`, tile.T.X))
	os.MkdirAll(dir, os.ModePerm)
	fileName := filepath.Join(dir, fmt.Sprintf(`%d.%s`, tile.T.Y, task.TileMap.Format))
	return os.WriteFile(fileName, tile.C, os.ModePerm)
}

func loadCollection(path string) orb.Collection {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		log.Fatalf("unable to unmarshal feature: %v", err)
	}

	var collection orb.Collection
	for _, f := range fc.Features {
		collection = append(collection, f.Geometry)
	}

	return collection
}

type BitMap struct {
	bits []byte
	vmax uint
}
