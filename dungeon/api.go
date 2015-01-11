package dungeon

type Dungeon interface {
	Create(seed int64)
	Width() int
	Height() int
	At(x, y int) uint
	Doors() []*Door
	Stairs() []*Stair
}

func (d *dungeon) Width() int  { return d.n_cols }
func (d *dungeon) Height() int { return d.n_rows }
func (d *dungeon) At(x, y int) uint {
	check := x >= 0 && x < d.Width() && y >= 0 && y < d.Height()
	if !check {
		return 0
	}
	return d.cell[y][x]
}
func (d *dungeon) Doors() []*Door   { return d.door }
func (d *dungeon) Stairs() []*Stair { return d.stair }

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
