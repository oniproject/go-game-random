// +build ignore

package main

import (
	. "."
	"./core"
	. "./dla"
	. "./dungeon"
	. "./simplex"
	"flag"
	"github.com/seehuhn/mt19937"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"os/exec"
)

var outFile = flag.String("o", "simple.png", "output file name")
var viewer = flag.String("viewer", "feh", "image viewer")
var dlaDemo = flag.Bool("dla", false, "if true run DLA demo")
var noiseDemo = flag.String("noise", "simplex", "if any of [simplex] run noise demo")

var seed = flag.Int64("seed", 1, "random seed")

var dungeonDemo = flag.Bool("dungeon", false, "if true run dungeon demo")
var dungeonWidth = flag.Int("dw", 101, "dungeon width")
var dungeonHeight = flag.Int("dh", 101, "dungeon height")
var dungeonRoomMin = flag.Int("drMin", 3, "dungeon RoomMin")
var dungeonRoomMax = flag.Int("drMax", 9, "dungeon RoomMax")
var dungeonRoomPacked = flag.Bool("drP", false, "dungeon RoomPacked")
var dungeonLayout = flag.String("dLayout", "None", "dungeon Lauout [None, Cross, Box, Round]")
var dungeonCorridorLauout = flag.Int("dCorridor", CORRIDOR_Bent, "dungeon CorridorLayout [0..100]")
var dungeonDeadends = flag.Int("dDeadends", 100, "dungeon RemoveDeadends [0..100]")
var dungeonStairs = flag.Int("dStairs", 2, "dungeon AddStairs")

var drawer_config = &DrawerConfig{
	Fill:     0x000000,
	Grid:     0x333333,
	OpenGrid: 0xCCCCCC,
	Stairs:   0xCCCCCC,
	Arch:     0xFFFFFF,
	Door:     0xFFCCCC,

	Room:      0x330000,
	Corridor:  0x333333,
	Perimeter: 0x111111,

	Labels: 0xFF00FF,

	GridType: GRID_SQUARE,
	CellSize: 8,
}

func main() {
	flag.Parse()

	rnd := rand.New(mt19937.New())
	rnd.Seed(*seed)

	w, h := 101, 101
	switch {
	default:
		flag.PrintDefaults()
		return
	case *dungeonDemo:
		m := core.NewCellMap(w, h)

		dungeon := &Dungeon{
			DungeonWidth:   *dungeonWidth,
			DungeonHeight:  *dungeonHeight,
			DungeonLayout:  *dungeonLayout, // Cross, Box, Round
			RoomMin:        *dungeonRoomMin,
			RoomMax:        *dungeonRoomMax,
			RoomPacked:     *dungeonRoomPacked,
			CorridorLayout: *dungeonCorridorLauout,
			RemoveDeadends: *dungeonDeadends, // percentage
			AddStairs:      *dungeonStairs,   // count stairs
			Rand:           rnd,
		}

		dungeon.Create(m)

		drawer := NewDrawer(drawer_config)
		img := drawer.Draw(dungeon)

		saveImage(img)

	case *dlaDemo:
		m := core.NewCellMap(w*8, h*8)

		dim := 1
		rect := image.Rect(0, 0,
			(m.Width()+1)*dim,
			(m.Height()+1)*dim)
		ii := image.NewRGBA(rect)
		max := m.Width() * m.Height() / 4

		x := &DLA{
			OrthogonalAllowed: false,
			Rand:              rnd,
		}

		x.Run(m, m.Width()-3, 5, 0xCC0000, max)
		x.Run(m, 5, m.Height()/2, 0x0000CC, max*2)
		x.Run(m, 34, 34, 0x00CC00, max)
		x.Run(m, m.Height()-3, 5, 0x00CCCC, max/2)

		for y := 0; y < m.Height(); y++ {
			for x := 0; x < m.Width(); x++ {
				c := m.At(x, y)
				if c != 0 {
					src := &image.Uniform{rgba(c)}
					draw.Draw(ii, image.Rect(x*dim, y*dim, (x+1)*dim, (y+1)*dim), src, image.ZP, draw.Src)
				}
			}
		}

		saveImage(ii)

	case *noiseDemo == "simplex":
		rect := image.Rect(0, 0, 1500, 1500)
		ii := image.NewRGBA(rect)

		n := &Noise{
			LargestFeature: 800,
			Persistence:    0.75,
			Rand:           rnd,
		}
		n.Init()

		//red := rgba(0xFF0000)
		white := rgba(0xFFFFFF)
		black := rgba(0x000000)
		//green := rgba(0x0000FF)
		//n := NewOctave(5000)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				data := (n.Noise2D(float64(x), float64(y)) + 1) * 0.5
				/*if data > 1 {
					data = 1
				} else if data < 0 {
					data = 0
				}*/
				c := color.RGBA{
					R: uint8(data * 255),
					G: uint8(data * 255),
					B: uint8(data * 255),
					A: 255,
				}
				if data > 1 {
					c = white
				} else if data < 0 {
					c = black
				} /*else {
					switch {
					case data < 0.4:
						c = rgba(0x0000FF)
					case data < 0.55:
						c = rgba(0xFFFF00)
					case data < 0.8:
						c = rgba(0x00FF00)
					case data < 0.90:
						c = rgba(0xFFFFFF)
					}
				}*/
				ii.Set(x, y, c)

				//ii.SetGray(x, y, color.Gray{Y: uint8(data * 255)})
			}
		}
		saveImage(ii)
	}

	if *viewer != "" {
		exec.Command(*viewer, *outFile).Run()
	}
}

func saveImage(img image.Image) {
	file, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
}

func rgba(c uint) color.RGBA {
	return color.RGBA{
		uint8(c >> 16), // r
		uint8(c >> 8),  // g
		uint8(c >> 0),  // b
		0xff,           // a
	}
}
