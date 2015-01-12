package core

type ReadonlyMap interface {
	Width() int
	Height() int
	At(x, y int) uint
	Check(x, y int) bool
}

type Map interface {
	ReadonlyMap
	Set(x, y int, v uint)
}

type cellMap struct {
	cell [][]uint
	w, h int
}

func NewCellMap(w, h int) Map {
	m := &cellMap{
		cell: make([][]uint, h),
		w:    w, h: h,
	}

	cell := make([]uint, w*h)

	for i := range m.cell {
		m.cell[i], cell = cell[:w], cell[w:]
	}

	return m
}

func (m *cellMap) Width() int  { return m.w }
func (m *cellMap) Height() int { return m.h }
func (m *cellMap) At(x, y int) uint {
	if !m.Check(x, y) {
		return 0
	}
	return m.cell[y][x]
}
func (m *cellMap) Set(x, y int, v uint) {
	if !m.Check(x, y) {
		return
	}
	m.cell[y][x] = v
}

func (m *cellMap) Check(x, y int) bool {
	return x >= 0 && x < m.Width() && y >= 0 && y < m.Height()
}
