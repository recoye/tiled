package main

import (
	"bytes"
	"encoding/binary"
	"github.com/recoye/tiled"
	"log"
	"reflect"
)

func main() {
	v := 255 << 32
	var k int
	k = 10
	var b []byte
	b = []byte{0x01, 0x00, 0x00, 0x01}

	d := b[0] | (b[1] << 8) | (b[2] << 16) | (b[3] << 24)

	// log.Println("...;", l)

	var x int32
	r := bytes.NewBuffer(b)
	binary.Read(r, binary.LittleEndian, &x)

	log.Println("...d", d, x, b[3]<<2)

	log.Println("...", v, reflect.TypeOf(v), k&v, k&(^v))
	tiledMap, err := tiled.NewTiled("map1.tmx")
	if err != nil {
		log.Println(err, "ss")
	}
	// log.Println(tiledMap.TiledMap.TiledSets)
	tiledMap.GetTerrain("water")
}
