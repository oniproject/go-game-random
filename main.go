// +build ignore

package main

import (
	. "."
	. "./dungeon"
	"image/png"
	"log"
	"os"
	"os/exec"
)

var dungeon_config = &DungeonConfig{
	Width:          101,
	Height:         101,
	DungeonLayout:  "None", // Cross, Box, Round
	RoomMin:        3,
	RoomMax:        9,
	RoomPacked:     false,
	CorridorLayout: CORRIDOR_Bent,
	RemoveDeadends: 100, // percentage
	AddStairs:      2,   // count stairs
}

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
	CellSize: 18,
}

func main() {
	dungeon := NewDungeon(dungeon_config)
	dungeon.Create(1)

	drawer := NewDrawer(drawer_config)
	img := drawer.Draw(dungeon)

	file, err := os.Create("simple.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}

	exec.Command("feh", "simple.png").Run()
}
