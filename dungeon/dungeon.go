package dungeon

import (
	"../core"
	"fmt"
	"math"
	"math/rand"
)

type DungeonConfig struct {
	DungeonWidth   int    //101,
	DungeonHeight  int    //101,
	DungeonLayout  string //"None", // Cross, Box, Round
	RoomMin        int    //3,
	RoomMax        int    //9,
	RoomPacked     bool
	CorridorLayout int //CORRIDOR_Bent,
	RemoveDeadends int //100, // percentage
	AddStairs      int //20,  // count stairs
}

// configuration

type mask [][]bool

var dungeon_layout = map[string]mask{
	"Box":   {{true, true, true}, {true, false, true}, {true, true, true}},
	"Cross": {{false, true, false}, {true, true, true}, {false, true, false}},
}

// corridor layout
const (
	CORRIDOR_Labyrinth = 0
	CORRIDOR_Bent      = 50
	CORRIDOR_Straight  = 100
)

// cell bits
const (
	NOTHING   = 0x00000000
	BLOCKED   = 0x00000001
	ROOM      = 0x00000002
	CORRIDOR  = 0x00000004
	_         = 0x00000008
	PERIMETER = 0x00000010
	ENTRANCE  = 0x00000020
	ROOM_ID   = 0x0000FFC0

	ARCH     = 0x00010000
	DOOR     = 0x00020000
	LOCKED   = 0x00040000
	TRAPPED  = 0x00080000
	SECRET   = 0x00100000
	PORTC    = 0x00200000
	STAIR_DN = 0x00400000
	STAIR_UP = 0x00800000

	LABEL = 0xFF000000

	OPENSPACE = ROOM | CORRIDOR
	DOORSPACE = ARCH | DOOR | LOCKED | TRAPPED | SECRET | PORTC
	ESPACE    = ENTRANCE | DOORSPACE | 0xFF000000
	STAIRS    = STAIR_DN | STAIR_UP

	BLOCK_ROOM = BLOCKED | ROOM
	BLOCK_CORR = BLOCKED | PERIMETER | CORRIDOR
	BLOCK_DOOR = BLOCKED | DOORSPACE
)

// directions

const (
	ZERO_DIR = 0
	NORTH    = 1
	SOUTH    = 2
	WEST     = 3
	EAST     = 4
)

var opposite = map[Dir]Dir{
	NORTH: SOUTH,
	SOUTH: NORTH,
	WEST:  EAST,
	EAST:  WEST,
}

type Dir uint8

func (dir *Dir) Opposite() Dir {
	switch *dir {
	case NORTH:
		return SOUTH
	case SOUTH:
		return NORTH
	case WEST:
		return EAST
	case EAST:
		return WEST
	}
	panic("fail direction")
	return ZERO_DIR
}

func (dir *Dir) di() int {
	switch *dir {
	case NORTH:
		return -1
	case SOUTH:
		return 1
	}
	return 0
}
func (dir *Dir) dj() int {
	switch *dir {
	case WEST:
		return -1
	case EAST:
		return 1
	}
	return 0
}

var dj_dirs = []Dir{NORTH, SOUTH, WEST, EAST}

type dungeon struct {
	*DungeonConfig

	n_rows int // 39,          # must be an odd number
	n_cols int // 39,          # must be an odd number

	// in create_dungeon
	n_i, n_j         int
	max_row, max_col int
	n_rooms          int

	// in init_cells
	//cell [][]uint

	room  map[uint]*Room
	stair []*Stair
	door  []*Door

	core.Map
	*rand.Rand
}

func NewDungeon(config *DungeonConfig) Dungeon {
	return &dungeon{
		DungeonConfig: config,
	}
}

// recalc
func (dungeon *dungeon) Create(seed int64, m core.Map) {
	dungeon.Map = m
	dungeon.Rand = rand.New(rand.NewSource(seed))

	dungeon.room = make(map[uint]*Room)

	dungeon.n_i = dungeon.Width() / 2
	dungeon.n_j = dungeon.Height() / 2
	dungeon.n_rows = dungeon.n_i * 2
	dungeon.n_cols = dungeon.n_j * 2
	dungeon.max_row = dungeon.n_rows - 1
	dungeon.max_col = dungeon.n_cols - 1
	dungeon.n_rooms = 0

	dungeon.init_cells()
	dungeon.emplace_rooms()
	dungeon.open_rooms()
	dungeon.label_rooms()
	dungeon.corridors()
	dungeon.emplace_stairs(dungeon.AddStairs)
	dungeon.clean_dungeon()
}

func (dungeon *dungeon) init_cells() {
	if mask, ok := dungeon_layout[dungeon.DungeonLayout]; ok {
		dungeon.mask_cells(mask)
	} else if dungeon.DungeonLayout == "Round" {
		dungeon.round_mask()
	}
}

func (dungeon *dungeon) mask_cells(mask [][]bool) {
	r_x := int(float64(len(mask)) / float64(dungeon.n_rows+1))
	c_x := int(float64(len(mask[0])) / float64(dungeon.n_cols+1))

	for r := 0; r < dungeon.n_rows; r++ {
		for c := 0; c < dungeon.n_cols; c++ {
			if !mask[r*r_x][c*c_x] {
				dungeon.Set(c, r, BLOCKED)
			}
		}
	}
}

func (dungeon *dungeon) round_mask() {
	center_r := dungeon.n_rows / 2
	center_c := dungeon.n_cols / 2

	for r := 0; r < dungeon.n_rows; r++ {
		for c := 0; c < dungeon.n_cols; c++ {
			x := math.Pow(float64(c-center_c), 2.0)
			y := math.Pow(float64(r-center_r), 2.0)
			d := math.Sqrt(x + y)
			if d > float64(center_c) {
				dungeon.Set(c, r, BLOCKED)
			}
		}
	}
}

func (dungeon *dungeon) emplace_rooms() {
	if dungeon.RoomPacked {
		dungeon.pack_rooms()
	} else {
		dungeon.scatter_rooms()
	}
}

func (dungeon *dungeon) pack_rooms() {
	for i := 0; i < dungeon.n_i; i++ {
		r := (i * 2) + 1
		for j := 0; j < dungeon.n_j; j++ {
			c := (j * 2) + 1

			if dungeon.At(c, r)&ROOM != 0 {
				continue
			}
			if (i == 0 || j == 0) && dungeon.Intn(2) == 1 {
				continue
			}

			// TODO
			proto := map[string]int{"i": i, "j": j}
			dungeon.emplace_room(proto)
		}
	}
}

func (dungeon *dungeon) scatter_rooms() {
	n_rooms := dungeon.alloc_rooms()
	for i := 0; i < n_rooms; i++ {
		dungeon.emplace_room(map[string]int{})
	}
}

func (dungeon *dungeon) alloc_rooms() (n_rooms int) {
	dungeon_area := dungeon.n_cols * dungeon.n_rows
	room_area := dungeon.RoomMax * dungeon.RoomMax
	n_rooms = dungeon_area / room_area
	return
}

func (dungeon *dungeon) emplace_room(proto map[string]int) {
	if dungeon.n_rooms == 999 {
		return
	}

	// room position and size
	proto = dungeon.set_room(proto)

	// room boundaries
	r1 := proto["i"]*2 + 1
	c1 := proto["j"]*2 + 1
	r2 := (proto["i"]+proto["height"])*2 - 1
	c2 := (proto["j"]+proto["width"])*2 - 1

	if r1 < 1 || r2 > dungeon.max_row {
		return
	}
	if c1 < 1 || c2 > dungeon.max_col {
		return
	}

	// check for collisions with existing rooms
	hit, blocked := dungeon.sound_room(r1, c1, r2, c2)
	if blocked || len(hit) != 0 {
		return
	}

	dungeon.n_rooms++
	room_id := uint(dungeon.n_rooms)

	// emplace room
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			cell := dungeon.At(c, r)
			if cell&ENTRANCE != 0 {
				cell = cell &^ ESPACE
			} else if cell&PERIMETER != 0 {
				cell = cell &^ PERIMETER
			}
			dungeon.Set(c, r, cell|ROOM|(room_id<<6))
		}
	}

	height := (r2 - r1 + 1) * 10
	width := (c2 - c1 + 1) * 10

	room_data := &Room{
		id: uint(room_id), row: r1, col: c1,
		North: r1, South: r2, West: c1, East: c2,
		height: height, width: width, area: (height * width),
	}
	dungeon.room[room_id] = room_data

	// block corridors from room boundary
	// check for door openings from adjacent rooms
	for r := r1 - 1; r <= r2+1; r++ {
		if r == dungeon.n_rows {
			continue
		}
		if dungeon.At(c1-1, r)&(ROOM|ENTRANCE) == 0 {
			cell := dungeon.At(c1-1, r)
			dungeon.Set(c1-1, r, cell|PERIMETER)
		}
		if c2+1 == dungeon.n_cols {
			continue
		}
		if dungeon.At(c2+1, r)&(ROOM|ENTRANCE) == 0 {
			cell := dungeon.At(c2+1, r)
			dungeon.Set(c2+1, r, cell|PERIMETER)
		}
	}
	for c := c1 - 1; c <= c2+1; c++ {
		if c == dungeon.n_cols {
			continue
		}
		if dungeon.At(c, r1-1)&(ROOM|ENTRANCE) == 0 {
			cell := dungeon.At(c, r1-1)
			dungeon.Set(c, r1-1, cell|PERIMETER)
		}
		if r2+1 == dungeon.n_rows {
			continue
		}
		if dungeon.At(c, r2+1)&(ROOM|ENTRANCE) == 0 {
			cell := dungeon.At(c, r2+1)
			dungeon.Set(c, r2+1, cell|PERIMETER)
		}
	}
}

// room position and size
func (dungeon *dungeon) set_room(proto map[string]int) map[string]int {
	max := dungeon.RoomMax
	min := dungeon.RoomMin
	base := (min + 1) / 2
	radix := (max-min)/2 + 1

	if _, ok := proto["height"]; !ok {
		r := radix
		if i, ok := proto["i"]; ok {
			a := dungeon.n_i - base - i
			if a < 0 {
				a = 0
			}
			if a < radix {
				r = a
			}
		}
		if r != 0 {
			r = dungeon.Intn(r)
		}
		proto["height"] = r + base
	}

	if _, ok := proto["width"]; !ok {
		r := radix
		if j, ok := proto["j"]; ok {
			a := dungeon.n_j - base - j
			if a < 0 {
				a = 0
			}
			if a < radix {
				r = a
			}
		}
		if r != 0 {
			r = dungeon.Intn(r)
		}
		proto["width"] = r + base
	}

	if _, ok := proto["i"]; !ok {
		proto["i"] = dungeon.Intn(dungeon.n_i) - proto["height"]
	}
	if _, ok := proto["j"]; !ok {
		proto["j"] = dungeon.Intn(dungeon.n_j) - proto["width"]
	}

	return proto
}

func (dungeon *dungeon) sound_room(r1, c1, r2, c2 int) (hit map[uint]int, blocked bool) {
	hit = make(map[uint]int)

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if dungeon.At(c, r)&BLOCKED != 0 {
				blocked = true
				return
			}
			if dungeon.At(c, r)&ROOM != 0 {
				id := (dungeon.At(c, r) & ROOM_ID) >> 6
				hit[id] += 1
			}
		}
	}
	return
}

// emplace openings for doors and corridors
func (dungeon *dungeon) open_rooms() {
	connects := make(map[string]bool)
	for id := uint(1); id <= uint(dungeon.n_rooms); id++ {
		dungeon.open_room(dungeon.room[id], connects)
	}
}

// emplace openings for doors and corridors
func (dungeon *dungeon) open_room(room *Room, connects map[string]bool) {
	list := dungeon.door_sills(room)
	if len(list) == 0 {
		return
	}
	n_opens := dungeon.alloc_opens(room)

	for i := 0; i < n_opens; i++ {
	Start:
		if len(list) == 0 {
			break
		}
		iii := dungeon.Intn(len(list))
		sill := list[iii]
		list = append(list[:iii], list[iii+1:]...)

		door_r := sill.door_r
		door_c := sill.door_c
		door_cell := dungeon.At(door_c, door_r)
		if door_cell&DOORSPACE != 0 {
			goto Start
		}
		out_id := sill.out_id
		if out_id != 0 {
			min, max := minmax(int(room.id), int(out_id))
			connect := fmt.Sprintf("%d,%d", min, max)
			if connects[connect] {
				goto Start
			}
			connects[connect] = true
		}

		open_r := sill.sill_r
		open_c := sill.sill_c
		open_dir := sill.dir

		// open door
		for x := 0; x < 3; x++ {
			r := open_r + open_dir.di()*x
			c := open_c + open_dir.dj()*x

			cell := dungeon.At(c, r)
			dungeon.Set(c, r, (cell&^PERIMETER)|ENTRANCE)
		}

		door := &Door{
			row:    door_r,
			col:    door_c,
			out_id: out_id,
		}

		cell := dungeon.At(door_c, door_r)

		switch dungeon.door_type() {
		case ARCH:
			cell |= ARCH
			door.key = "arch"
			door.t = "Archway"
		case DOOR:
			cell |= DOOR
			cell |= 'o' << 24
			door.key = "open"
			door.t = "Unlocked Door"
		case LOCKED:
			cell |= LOCKED
			cell |= 'x' << 24
			door.key = "lock"
			door.t = "Locked Door"
		case TRAPPED:
			cell |= TRAPPED
			cell |= 't' << 24
			door.key = "trap"
			door.t = "Trapped Door"
		case SECRET:
			cell |= SECRET
			cell |= 's' << 24
			door.key = "secret"
			door.t = "Secret Door"
		case PORTC:
			cell |= PORTC
			cell |= '#' << 24
			door.key = "portc"
			door.t = "Portcullis"
		}

		dungeon.Set(door_c, door_r, cell)

		if room.door == nil {
			room.door = make(map[Dir][]*Door)
		}
		room.door[open_dir] = append(room.door[open_dir], door)
	}
}

// allocate number of opens
func (dungeon *dungeon) alloc_opens(room *Room) (n_opens int) {
	h := float64(room.South-room.North)/2.0 + 1.0
	w := float64(room.East-room.West)/2.0 + 1.0
	flumph := int(math.Sqrt(w * h))
	n_opens = flumph + dungeon.Intn(flumph)
	return
}

// list available sills
func (dungeon *dungeon) door_sills(room *Room) []*Sill {
	list := []*Sill{}

	if room.North >= 3 {
		for c := room.West; c <= room.East; c += 2 {
			sill := dungeon.check_sill(room, room.North, c, NORTH)
			if sill != nil {
				list = append(list, sill)
			}
		}
	}
	if room.South <= dungeon.n_rows-3 {
		for c := room.West; c <= room.East; c += 2 {
			sill := dungeon.check_sill(room, room.South, c, SOUTH)
			if sill != nil {
				list = append(list, sill)
			}
		}
	}

	if room.West >= 3 {
		for r := room.North; r <= room.South; r += 2 {
			sill := dungeon.check_sill(room, r, room.West, WEST)
			if sill != nil {
				list = append(list, sill)
			}
		}
	}
	if room.East <= dungeon.n_rows-3 {
		for r := room.North; r <= room.South; r += 2 {
			sill := dungeon.check_sill(room, r, room.East, EAST)
			if sill != nil {
				list = append(list, sill)
			}
		}
	}

	// shuffle
	for i := range list {
		j := dungeon.Intn(i + 1)
		list[i], list[j] = list[j], list[i]
	}
	/*
		n := len(list)
		for i := 0; i < n; i++ {
			j := i + dungeon.Intn(n-1)
			list[i], list[j] = list[j], list[i]
		}
	*/
	return list
}

type Sill struct {
	sill_r, sill_c int
	dir            Dir
	door_r, door_c int
	out_id         uint
}

func (dungeon *dungeon) check_sill(room *Room, sill_r, sill_c int, dir Dir) *Sill {
	door_r := sill_r + dir.di()
	door_c := sill_c + dir.dj()

	door_cell := dungeon.At(door_c, door_r)
	if door_cell&PERIMETER == 0 {
		return nil
	}
	if door_cell&BLOCK_DOOR != 0 {
		return nil
	}
	out_r := door_r + dir.di()
	out_c := door_c + dir.dj()
	out_cell := dungeon.At(out_c, out_r)
	if out_cell&BLOCKED != 0 {
		return nil
	}

	out_id := uint(0)
	if out_cell&ROOM != 0 {
		out_id = (out_cell & ROOM_ID) >> 6
		if out_id == room.id {
			return nil
		}
	}

	return &Sill{
		sill_r: sill_r,
		sill_c: sill_c,
		dir:    dir,
		door_r: door_r,
		door_c: door_c,
		out_id: out_id,
	}
}

// random door type
func (dungeon *dungeon) door_type() int {
	i := dungeon.Intn(110)
	switch {
	case i < 15:
		return ARCH
	case i < 60:
		return DOOR
	case i < 75:
		return LOCKED
	case i < 90:
		return TRAPPED
	case i < 100:
		return SECRET
	}
	return PORTC
}

func (dungeon *dungeon) label_rooms() {
	for id := uint(1); id <= uint(dungeon.n_rooms); id++ {
		room := dungeon.room[id]
		label := fmt.Sprint(room.id)
		label_r := (room.North + room.South) / 2
		label_c := (room.West+room.East-len(label))/2 + 1
		for c, char := range label {
			cell := dungeon.At(label_c+c, label_r)
			dungeon.Set(label_c+c, label_r, cell|uint(char)<<24)
		}
	}
}

// generate corridors
func (dungeon *dungeon) corridors() {
	for i := 1; i < dungeon.n_i; i++ {
		r := i*2 + 1
		for j := 1; j < dungeon.n_j; j++ {
			c := j*2 + 1
			if dungeon.At(c, r)&CORRIDOR != 0 {
				continue
			}
			dungeon.tunnel(i, j, ZERO_DIR)
		}
	}
}

//  recursively tunnel
func (dungeon *dungeon) tunnel(i, j int, last_dir Dir) {
	dirs := dungeon.tunnel_directions(last_dir)
	for _, dir := range dirs {
		if dungeon.open_tunnel(i, j, dir) {
			next_i := i + dir.di()
			next_j := j + dir.dj()
			dungeon.tunnel(next_i, next_j, dir)
		}
	}
}

func (dungeon *dungeon) tunnel_directions(last_dir Dir) (dirs []Dir) {
	dirs = make([]Dir, len(dj_dirs))
	for i, v := range dungeon.Perm(len(dj_dirs)) {
		dirs[v] = dj_dirs[i]
	}

	p := dungeon.CorridorLayout

	if last_dir != ZERO_DIR && p != 0 {
		if dungeon.Intn(100) < p {
			dirs = append([]Dir{last_dir}, dirs...)
		}
	}
	return
}

func (dungeon *dungeon) open_tunnel(i, j int, dir Dir) (ok bool) {
	this_r := i*2 + 1
	this_c := j*2 + 1
	next_r := (i+dir.di())*2 + 1
	next_c := (j+dir.dj())*2 + 1
	mid_r := (this_r + next_r) / 2
	mid_c := (this_c + next_c) / 2

	if dungeon.sound_tunnel(mid_r, mid_c, next_r, next_c) {
		ok = dungeon.delve_tunnel(this_r, this_c, next_r, next_c)
	}
	return
}

func minmax(a, b int) (int, int) {
	if a > b {
		return b, a
	}
	return a, b
}

// don't open blocked cells, room perimeters, or other corridors
func (dungeon *dungeon) sound_tunnel(mid_r, mid_c, next_r, next_c int) (ok bool) {
	if next_r < 0 || next_r > dungeon.n_rows {
		return
	}
	if next_c < 0 || next_c > dungeon.n_cols {
		return
	}

	r1, r2 := minmax(mid_r, next_r)
	c1, c2 := minmax(mid_c, next_c)

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if dungeon.At(c, r)&BLOCK_CORR != 0 {
				return
			}
		}
	}
	return true
}

func (dungeon *dungeon) delve_tunnel(this_r, this_c, next_r, next_c int) bool {
	r1, r2 := minmax(this_r, next_r)
	c1, c2 := minmax(this_c, next_c)

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			cell := dungeon.At(c, r)
			dungeon.Set(c, r, (cell&^ENTRANCE)|CORRIDOR)
		}
	}
	return true
}

func (dungeon *dungeon) emplace_stairs(n int) {
	if n <= 0 {
		return
	}
	list := dungeon.stair_ends()
	if len(list) == 0 {
		return
	}

	for i := 0; i < n; i++ {
		if len(list) == 0 {
			break
		}

		iii := dungeon.Intn(len(list))
		stair := list[iii]
		list = append(list[:iii], list[iii+1:]...)

		r := stair.row
		c := stair.col
		t := dungeon.Intn(2)
		if i < 2 {
			t = i
		}

		cell := dungeon.At(c, r)
		if t == 0 {
			cell |= STAIR_DN
			cell |= 'd' << 24
			stair.key = "down"
		} else {
			cell |= STAIR_UP
			cell |= 'u' << 24
			stair.key = "up"
		}
		dungeon.Set(c, r, cell)

		dungeon.stair = append(dungeon.stair, stair)
	}
}

type Tunnel struct {
	walled   [][]int
	corridor [][]int

	stair []int
	next  []int

	open []int

	close   [][]int
	recurse []int
}

var stair_end = map[Dir]Tunnel{
	NORTH: {
		walled:   [][]int{{1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}, {0, 1}, {1, 1}},
		corridor: [][]int{{0, 0}, {1, 0}, {2, 0}},
		stair:    []int{0, 0},
		next:     []int{1, 0},
	},
	SOUTH: {
		walled:   [][]int{{-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}},
		corridor: [][]int{{0, 0}, {-1, 0}, {-2, 0}},
		stair:    []int{0, 0},
		next:     []int{-1, 0},
	},
	WEST: {
		walled:   [][]int{{-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}},
		corridor: [][]int{{0, 0}, {0, 1}, {0, 2}},
		stair:    []int{0, 0},
		next:     []int{0, 1},
	},
	EAST: {
		walled:   [][]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, 1}, {1, 1}, {1, 0}, {1, -1}},
		corridor: [][]int{{0, 0}, {0, -1}, {0, -2}},
		stair:    []int{0, 0},
		next:     []int{0, -1},
	},
}

// list available ends
func (dungeon *dungeon) stair_ends() (list []*Stair) {
	//ROW:
	for i := 0; i < dungeon.n_i; i++ {
		r := i*2 + 1
	COL:
		for j := 0; j < dungeon.n_j; j++ {
			c := j*2 + 1

			if dungeon.At(c, r) != CORRIDOR {
				continue
			}
			if dungeon.At(c, r)&STAIRS != 0 {
				continue
			}

			for dir := range stair_end {
				if dungeon.check_tunnel(r, c, stair_end[dir]) {
					end := &Stair{row: r, col: c}
					n := stair_end[dir].next
					end.next_row = end.row + n[0]
					end.next_col = end.col + n[1]
					list = append(list, end)
					continue COL
				}
			}
		}
	}
	return list
}

// final clean-up
func (dungeon *dungeon) clean_dungeon() {
	// remove deadend corridors
	dungeon.collapse_tunnels(dungeon.RemoveDeadends, close_end)

	dungeon.fix_doors()
	dungeon.empty_blocks()
}

func (dungeon *dungeon) collapse_tunnels(p int, xc map[Dir]Tunnel) {
	if p == 0 {
		return
	}

	all := p == 100

	for i := 0; i < dungeon.n_i; i++ {
		r := i*2 + 1
		for j := 0; j < dungeon.n_j; j++ {
			c := j*2 + 1

			if dungeon.At(c, r)&OPENSPACE == 0 {
				continue
			}
			if dungeon.At(c, r)&STAIRS != 0 {
				continue
			}
			if !(all || dungeon.Intn(100) < p) {
				continue
			}

			dungeon.collapse(r, c, xc)
		}
	}
}

var close_end = map[Dir]Tunnel{
	NORTH: {
		walled:  [][]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}},
		close:   [][]int{{0, 0}},
		recurse: []int{-1, 0},
	},
	SOUTH: {
		walled:  [][]int{{0, -1}, {-1, -1}, {-1, 0}, {-1, 1}, {0, 1}},
		close:   [][]int{{0, 0}},
		recurse: []int{1, 0},
	},
	WEST: {
		walled:  [][]int{{-1, 0}, {-1, 1}, {0, 1}, {1, 1}, {1, 0}},
		close:   [][]int{{0, 0}},
		recurse: []int{0, -1},
	},
	EAST: {
		walled:  [][]int{{-1, 0}, {-1, -1}, {0, -1}, {1, -1}, {1, 0}},
		close:   [][]int{{0, 0}},
		recurse: []int{0, 1},
	},
}

func (dungeon *dungeon) collapse(r, c int, xc map[Dir]Tunnel) {
	if !dungeon.checkPos(r, c) || dungeon.At(c, r)&OPENSPACE == 0 {
		return
	}

	for dir := range xc {
		if dungeon.check_tunnel(r, c, xc[dir]) {
			for _, p := range xc[dir].close {
				dungeon.Set(c+p[1], r+p[0], NOTHING)
			}
			p := xc[dir].open
			if len(p) != 0 {
				cell := dungeon.At(c+p[1], r+p[0])
				dungeon.Set(c+p[1], r+p[0], cell|CORRIDOR)
			}
			p = xc[dir].recurse
			if len(p) != 0 {
				dungeon.collapse(r+p[0], c+p[1], xc)
			}
		}
	}
}
func (dungeon *dungeon) checkPos(r, c int) bool {
	rr := r >= 0 && r < dungeon.n_rows
	cc := c >= 0 && c < dungeon.n_cols
	return rr && cc
}

func (dungeon *dungeon) check_tunnel(r, c int, check Tunnel) (ok bool) {
	for _, p := range check.corridor {
		rr := r + p[0]
		cc := c + p[1]
		if !dungeon.checkPos(rr, cc) {
			continue
		}
		if dungeon.At(cc, rr) != CORRIDOR {
			return
		}
	}
	for _, p := range check.walled {
		rr := r + p[0]
		cc := c + p[1]
		if !dungeon.checkPos(rr, cc) {
			continue
		}
		if dungeon.At(cc, rr)&OPENSPACE != 0 {
			return
		}
	}
	return true
}

// fix door lists
func (dungeon *dungeon) fix_doors() {
	w, h := dungeon.n_cols, dungeon.n_rows

	fixed := make([][]bool, h)

	bools := make([]bool, w*h)
	for i := range fixed {
		fixed[i], bools = bools[:w], bools[w:]
	}

	for _, room := range dungeon.room {
		for dir := range room.door {
			shiny := []*Door{}
			for _, door := range room.door[dir] {
				door_r := door.row
				door_c := door.col
				door_cell := dungeon.At(door_c, door_r)
				if door_cell&OPENSPACE == 0 {
					continue
				}
				if fixed[door_r][door_c] {
					shiny = append(shiny, door)
				} else {
					out_id := door.out_id
					if out_id != 0 {
						out_dir := opposite[dir]
						dungeon.room[out_id].door[out_dir] = append(dungeon.room[out_id].door[out_dir], door)
					}
					shiny = append(shiny, door)
					fixed[door_r][door_c] = true
				}
			}
			if len(shiny) != 0 {
				room.door[dir] = shiny
				dungeon.door = append(dungeon.door, shiny...)
			} else {
				delete(room.door, dir)
			}
		}
	}
}

func (dungeon *dungeon) empty_blocks() {
	for r := 0; r < dungeon.n_rows; r++ {
		for c := 0; c < dungeon.n_cols; c++ {
			if dungeon.At(c, r)&BLOCKED != 0 {
				dungeon.Set(c, r, NOTHING)
			}
		}
	}
}
