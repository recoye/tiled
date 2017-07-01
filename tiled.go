package tiled

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

const (
	FHFLAG = 0x80000000
	FVFLAG = 0x40000000
	FDFLAG = 0x20000000
	RHFLAG = 0x10000000
)

type TiledImage struct {
	XMLName xml.Name `xml:"image"`
	Width   uint     `xml:"width,attr"`
	Height  uint     `xml:"height,attr"`
	Source  string   `xml:"source,attr"`
}
type TiledLayer struct {
	XMLName xml.Name  `xml:"layer"`
	Name    string    `xml:"name,attr"`
	Width   uint32    `xml:"width,attr"`
	Height  uint32    `xml:"height,attr"`
	Data    TiledData `xml:"data"`
	cells   []uint32
}
type TiledData struct {
	XMLName     xml.Name `xml:"data"`
	Encoding    string   `xml:"encoding,attr"`
	Compression string   `xml:"compression,attr"`
	Data        string   `xml:",innerxml"`
}
type TileCell struct {
	XMLName xml.Name `xml:"tile"`
	Id      uint32   `xml:"id,attr"`
	Terrain string   `xml:"terrain,attr"`
}
type TerrainTypes struct {
	XMLName xml.Name `xml:"terrain"`
	Name    string   `xml:"name,attr"`
	tile    int      `xml:"tile,attr"`
}
type TiledSet struct {
	XMLName     xml.Name       `xml:"tileset"`
	FirstGrid   int            `xml:"firstgrid,attr"`
	Name        string         `xml:"name,attr"`
	TiledWidth  int            `xml:"tiledwidth,attr"`
	TiledHeight int            `xml:"tiledheight,attr"`
	Space       int            `xml:"space,attr"`
	Margin      int            `xml:"margin,attr"`
	TileCount   int            `xml:"tilecount,attr"`
	Columns     int            `xml:columns,attr"`
	Image       TiledImage     `xml:"image"`
	TileCells   []TileCell     `xml:"tile"`
	Terrain     []TerrainTypes `xml:"terraintypes>terrain"`
}
type TiledMap struct {
	XMLName     xml.Name      `xml:"map"`
	version     string        `xml:"version,attr"`
	Width       int           `xml:"width,attr"`
	Height      int           `xml:"height,attr"`
	renderOrder string        `xml:"renderorder,attr"`
	TiledSets   []TiledSet    `xml:"tileset"`
	Layers      []*TiledLayer `xml:"layer"`
}
type Tiled struct {
	Width       uint
	Height      uint
	TiledWidht  uint
	TiledHeight uint
	TiledMap    *TiledMap
}

func NewTiled(fileName string) (*Tiled, error) {
	tiled := &Tiled{}
	return tiled.Init(fileName)
}

func (this *TiledLayer) Init() error {
	data, err := this.Data.Decode()
	if err != nil {
		return err
	}
	this.cells = make([]uint32, this.Width*this.Height)
	j := 0

	flag := (uint32)(FHFLAG | FVFLAG | FDFLAG | RHFLAG)
	flag = ^flag
	size := this.Width * this.Height * 4
	var i uint32
	for i = 0; i < size-3; i += 4 {
		gid := (uint32(data[i]) | (uint32(data[i+1]) << 8) | (uint32(data[i+2]) << 16) | (uint32(data[i+3]) << 24))
		this.cells[j] = gid & flag
		j++
	}
	// log.Println(this.cells)
	return nil
}

func (this *TiledLayer) GetCells() []uint32 {
	return this.cells
}
func (this *TiledLayer) GetCell(idx uint32) uint32 {
	return this.cells[idx]
}

func (this *TiledData) Decode() ([]byte, error) {
	var data string = strings.TrimSpace(this.Data)
	var out bytes.Buffer
	var err error
	var in []byte
	switch this.Encoding {
	case "base64":
		in, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, err
		}
	default:
		in = []byte(data)
	}
	switch this.Compression {
	case "zlib":
		var r io.ReadCloser
		b := bytes.NewReader(in)
		r, err = zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		io.Copy(&out, r)

		o := out.Bytes()

		return o, err
	default:
		return in, nil
	}

}

func (this *Tiled) Init(fileName string) (*Tiled, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	// var tileMap TiledMap
	// err = xml.Unmarshal(content, &tileMap)
	this.TiledMap = &TiledMap{}
	err = xml.Unmarshal(content, this.TiledMap)
	if err != nil {
		return nil, err
	}
	for _, layer := range this.TiledMap.Layers {
		if e := layer.Init(); e != nil {
			log.Println(layer.GetCells())
			return nil, e
		}
	}
	// log.Println(tileMap.Tiledsets)
	// log.Println(this.TiledMap.Layers)
	return this, nil
}

func (this *Tiled) GetTerrain(terrain string) ([]bool, error) {
	var block []bool = make([]bool, this.TiledMap.Width*this.TiledMap.Height)
	for _, tile := range this.TiledMap.TiledSets {
		ter := 0
		for _, key := range tile.Terrain {
			if strings.Compare(key.Name, terrain) == 0 {
				break
			}
			ter++
		}
		var i uint32
		for _, key := range tile.TileCells {
			if strings.Contains(","+key.Terrain+",", ","+strconv.Itoa(ter)+",") {
				for _, layer := range this.TiledMap.Layers {
					for i = 0; i < layer.Width*layer.Height; i++ {
						if key.Id == layer.GetCell(i) {
							block[i] = true
						}
					} // for i < this.width * this.height
				} // for layer
			} // if string.contains
		} // for tile.ileCells
	}
	return block, nil
}
