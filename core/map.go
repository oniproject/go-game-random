package core

//import "fmt"

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
func (m *cellMap) At(x, y int) (v uint) {
	if !m.Check(x, y) {
		//fmt.Printf("Fail At %d/%d %d/%d\n", x, m.w, y, m.h)
		return
	}
	return m.cell[y][x]
}
func (m *cellMap) Set(x, y int, v uint) {
	if !m.Check(x, y) {
		//fmt.Printf("Fail Set %d/%d %d/%d\n", x, m.w, y, m.h)
		return
	}
	m.cell[y][x] = v
}

func (m *cellMap) Check(x, y int) (ok bool) {
	ok = x >= 0 && x < m.Width() && y >= 0 && y < m.Height()
	return
}

func Draw(dst Map, x, y int, src Map) {
	maxX, maxY := x+src.Width(), y+src.Height()
	if maxX > dst.Width() {
		maxX = dst.Width()
	}
	if maxY > dst.Height() {
		maxY = dst.Height()
	}

	startX, startY := x, y
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	for dstY := startX; dstY < maxY; dstY++ {
		for dstX := startY; dstX < maxX; dstX++ {
			srcX, srcY := dstX-x, dstY-y
			dst.Set(dstX, dstY, src.At(srcX, srcY))
		}
	}
}
