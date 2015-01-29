package dungeon

//import "../core"

/*
type Dungeon interface {
	core.ReadonlyMap
	Create(seed int64, m core.Map)
	Doors() []*Door
	Stairs() []*Stair
}*/

func (d *Dungeon) Doors() []*Door   { return d.door }
func (d *Dungeon) Stairs() []*Stair { return d.stair }

type Door struct {
	row, col int
	key      string
	t        string
	out_id   uint
}

func (d *Door) X() int       { return d.col }
func (d *Door) Y() int       { return d.row }
func (d *Door) Key() string  { return d.key }
func (d *Door) Type() string { return d.t }

type Stair struct {
	row, col           int
	key                string
	next_row, next_col int
}

func (s *Stair) X() int       { return s.col }
func (s *Stair) Y() int       { return s.row }
func (s *Stair) NextX() int   { return s.next_col }
func (s *Stair) NextY() int   { return s.next_row }
func (s *Stair) IsDown() bool { return s.key == "down" }

type Room struct {
	South int
	North int
	East  int
	West  int

	id   uint
	door map[Dir][]*Door

	row, col int
	height   int
	width    int
	area     int
}
