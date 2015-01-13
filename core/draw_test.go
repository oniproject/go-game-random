package core

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test(t *testing.T) {
	Convey("Map", t, func() {
		src := &cellMap{
			w: 3, h: 3,
			cell: [][]uint{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
		}

		Convey("Operations", func() {
			var (
				v  uint
				ok bool
			)

			Convey("WH", func() {
				So(src.Width(), ShouldEqual, 3)
				So(src.Height(), ShouldEqual, 3)
			})
			Convey("At&Set", func() {
				v, ok = src.At(1, 1)
				So(v, ShouldEqual, 5)
				So(ok, ShouldBeTrue)

				So(src.Set(1, 1, 999), ShouldBeTrue)
				v, ok = src.At(1, 1)
				So(v, ShouldEqual, 999)
				So(ok, ShouldBeTrue)
			})
			Convey("Check", func() {
				So(src.Check(0, 0), ShouldBeTrue)
				So(src.Check(1, 1), ShouldBeTrue)
				So(src.Check(-1, 0), ShouldBeFalse)
				So(src.Check(0, -1), ShouldBeFalse)
				So(src.Check(src.Width(), 0), ShouldBeFalse)
				So(src.Check(0, src.Height()), ShouldBeFalse)
			})
		})

		Convey("Draw", func() {
			dst := &cellMap{
				w: 5, h: 5,
				cell: [][]uint{
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
				},
			}

			Convey("Basic", func() {
				Draw(dst, 1, 1, src)
				So(dst.cell, ShouldResemble, [][]uint{
					{0, 0, 0, 0, 0},
					{0, 1, 2, 3, 0},
					{0, 4, 5, 6, 0},
					{0, 7, 8, 9, 0},
					{0, 0, 0, 0, 0},
				})
			})

			Convey("TopLeft 0", func() {
				Draw(dst, 0, 0, src)
				So(dst.cell, ShouldResemble, [][]uint{
					{1, 2, 3, 0, 0},
					{4, 5, 6, 0, 0},
					{7, 8, 9, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
				})
			})

			Convey("TopLeft -1", func() {
				Draw(dst, -1, -1, src)
				So(dst.cell, ShouldResemble, [][]uint{
					{5, 6, 0, 0, 0},
					{8, 9, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
				})
			})

			Convey("BottomRigth 0 (2)", func() {
				Draw(dst, 2, 2, src)
				So(dst.cell, ShouldResemble, [][]uint{
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 1, 2, 3},
					{0, 0, 4, 5, 6},
					{0, 0, 7, 8, 9},
				})
			})

			Convey("BottomRigth -1 (3)", func() {
				Draw(dst, 3, 3, src)
				So(dst.cell, ShouldResemble, [][]uint{
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0},
					{0, 0, 0, 1, 2},
					{0, 0, 0, 4, 5},
				})
			})
		})
	})
}
