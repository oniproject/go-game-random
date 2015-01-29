// +build ignore

package main

import (
	. "."
	"./core"
	. "./dla"
	. "./dungeon"
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
	w, h := 101, 101
	{
		m := core.NewCellMap(w, h)

		dungeon := &Dungeon{
			DungeonWidth:   101,
			DungeonHeight:  101,
			DungeonLayout:  "None", // Cross, Box, Round
			RoomMin:        3,
			RoomMax:        9,
			RoomPacked:     false,
			CorridorLayout: CORRIDOR_Bent,
			RemoveDeadends: 100, // percentage
			AddStairs:      2,   // count stairs
			Rand:           rand.New(mt19937.New()),
		}

		dungeon.Seed(1)
		dungeon.Create(m)

		drawer := NewDrawer(drawer_config)
		img := drawer.Draw(dungeon)

		saveImage(img, "simple.png")
	}

	if false {
		m := core.NewCellMap(w*8, h*8)

		dim := 1
		rect := image.Rect(0, 0,
			(m.Width()+1)*dim,
			(m.Height()+1)*dim)
		ii := image.NewRGBA(rect)
		max := m.Width() * m.Height() / 4

		x := &DLA{
			OrthogonalAllowed: false,
			Rand:              rand.New(mt19937.New()),
		}
		x.Seed(1)

		log.Println("start 1")
		x.Run(m, m.Width()-3, 5, 0xCC0000, max)
		log.Println("start 2")
		x.Run(m, 5, m.Height()/2, 0x0000CC, max*2)
		log.Println("start 3")
		x.Run(m, 34, 34, 0x00CC00, max)
		log.Println("start 4")
		x.Run(m, m.Height()-3, 5, 0x00CCCC, max/2)
		log.Println("finish")

		for y := 0; y < m.Height(); y++ {
			for x := 0; x < m.Width(); x++ {
				c := m.At(x, y)
				if c != 0 {
					src := &image.Uniform{rgba(c)}
					draw.Draw(ii, image.Rect(x*dim, y*dim, (x+1)*dim, (y+1)*dim), src, image.ZP, draw.Src)
				}
			}
		}
		log.Println("end draw")

		saveImage(ii, "simple.png")
	}

	exec.Command("feh", "simple.png").Run()
}

func saveImage(img image.Image, fname string) {
	file, err := os.Create(fname)
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
