// Simplex noise in 2D, 3D and 4D

package simplex

import (
	"math"
	"math/rand"
)

// Inner class to speed upp gradient computations
// (array access is a lot slower than member access)
type grad struct {
	x, y, z, w float64
}

func (g *grad) dot2(x, y float64) float64 {
	return g.x*x + g.y*y
}
func (g *grad) dot3(x, y, z float64) float64 {
	return g.x*x + g.y*y + g.z*z
}
func (g *grad) dot4(x, y, z, w float64) float64 {
	return g.x*x + g.y*y + g.z*z + g.w*w
}

var grad3 = []grad{
	{+1, +1, +0, 0}, {-1, +1, +0, 0}, {+1, -1, +0, 0}, {-1, -1, +0, 0},
	{+1, +0, +1, 0}, {-1, +0, +1, 0}, {+1, +0, -1, 0}, {-1, +0, -1, 0},
	{+0, +1, +1, 0}, {+0, -1, +1, 0}, {+0, +1, -1, 0}, {+0, -1, -1, 0},
}

var grad4 = []grad{
	{+0, +1, +1, +1}, {+0, +1, +1, -1}, {+0, +1, -1, +1}, {+0, +1, -1, -1},
	{+0, -1, +1, +1}, {+0, -1, +1, -1}, {+0, -1, -1, +1}, {+0, -1, -1, -1},
	{+1, +0, +1, +1}, {+1, +0, +1, -1}, {+1, +0, -1, +1}, {+1, +0, -1, -1},
	{-1, +0, +1, +1}, {-1, +0, +1, -1}, {-1, +0, -1, +1}, {-1, +0, -1, -1},
	{+1, +1, +0, +1}, {+1, +1, +0, -1}, {+1, -1, +0, +1}, {+1, -1, +0, -1},
	{-1, +1, +0, +1}, {-1, +1, +0, -1}, {-1, -1, +0, +1}, {-1, -1, +0, -1},
	{+1, +1, +1, +0}, {+1, +1, -1, +0}, {+1, -1, +1, +0}, {+1, -1, -1, +0},
	{-1, +1, +1, +0}, {-1, +1, -1, +0}, {-1, -1, +1, +0}, {-1, -1, -1, +0},
}

// Skewing and unskewing factors for 2, 3, and 4 dimensions
var (
	F2 = 0.5 * (math.Sqrt(3.0) - 1.0)
	G2 = (3.0 - math.Sqrt(3.0)) / 6.0
	F3 = 1.0 / 3.0
	G3 = 1.0 / 6.0
	F4 = (math.Sqrt(5.0) - 1.0) / 4.0
	G4 = (5.0 - math.Sqrt(5.0)) / 20.0
)

// This method is a *lot* faster than using (int)Math.floor(x)
/*func fastfloor(x float64) int {
  int xi = (int)x;
  return x<xi ? xi-1 : xi;
}*/

type Octave struct {
	perm      [512]int
	permMod12 [512]int
}

func NewOctave(rnd *rand.Rand) *Octave {
	octave := &Octave{}
	p := rnd.Perm(256)

	for i := 0; i < 512; i++ {
		octave.perm[i] = p[i&255]
		octave.permMod12[i] = octave.perm[i] % 12
	}
	return octave
}

// 2D simplex noise
func (octave *Octave) Noise2D(xin, yin float64) float64 {
	var n0, n1, n2 float64 // Noise contributions from the three corners

	// Skew the input space to determine which simplex cell we're in
	s := (xin + yin) * F2 // Hairy factor for 2D
	i := math.Floor(xin + s)
	j := math.Floor(yin + s)

	t := (i + j) * G2
	X0 := i - t // Unskew the cell origin back to (x,y) space
	Y0 := j - t
	x0 := xin - X0 // The x,y distances from the cell origin
	y0 := yin - Y0

	// For the 2D case, the simplex shape is an equilateral triangle.
	// Determine which simplex we are in.
	var i1, j1 int // Offsets for second (middle) corner of simplex in (i,j) coords
	if x0 > y0 {
		// lower triangle, XY order: (0,0)->(1,0)->(1,1)
		i1, j1 = 1, 0
	} else {
		// upper triangle, YX order: (0,0)->(0,1)->(1,1)
		i1, j1 = 0, 1
	}

	// A step of (1,0) in (i,j) means a step of (1-c,-c) in (x,y), and
	// a step of (0,1) in (i,j) means a step of (-c,1-c) in (x,y), where
	// c = (3-sqrt(3))/6
	x1 := x0 - float64(i1) + G2 // Offsets for middle corner in (x,y) unskewed coords
	y1 := y0 - float64(j1) + G2
	x2 := x0 - 1.0 + 2.0*G2 // Offsets for last corner in (x,y) unskewed coords
	y2 := y0 - 1.0 + 2.0*G2

	// Work out the hashed gradient indices of the three simplex corners
	ii, jj := int(i)&255, int(j)&255
	gi0 := octave.permMod12[ii+octave.perm[jj]]
	gi1 := octave.permMod12[ii+i1+octave.perm[jj+j1]]
	gi2 := octave.permMod12[ii+1+octave.perm[jj+1]]

	// Calculate the contribution from the three corners
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * grad3[gi0].dot2(x0, y0) // (x,y) of grad3 used for 2D gradient
	}
	t1 := 0.5 - x1*x1 - y1*y1
	if t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * grad3[gi1].dot2(x1, y1)
	}
	t2 := 0.5 - x2*x2 - y2*y2
	if t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * grad3[gi2].dot2(x2, y2)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to return values in the interval [-1,1].
	return 70.0 * (n0 + n1 + n2)
}

// 3D simplex noise
func (octave *Octave) Noise3D(xin, yin, zin float64) float64 {
	var n0, n1, n2, n3 float64 // Noise contributions from the four corners

	// Skew the input space to determine which simplex cell we're in
	s := (xin + yin + zin) * F3 // Very nice and simple skew factor for 3D
	i := math.Floor(xin + s)
	j := math.Floor(yin + s)
	k := math.Floor(zin + s)

	t := (i + j + k) * G3
	X0 := i - t // Unskew the cell origin back to (x,y,z) space
	Y0 := j - t
	Z0 := k - t
	x0 := xin - X0 // The x,y,z distances from the cell origin
	y0 := yin - Y0
	z0 := zin - Z0

	// For the 3D case, the simplex shape is a slightly irregular tetrahedron.
	// Determine which simplex we are in.
	var i1, j1, k1 int // Offsets for second corner of simplex in (i,j,k) coords
	var i2, j2, k2 int // Offsets for third corner of simplex in (i,j,k) coords
	if x0 >= y0 {
		if y0 >= z0 {
			// X Y Z order
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 1, 0
		} else if x0 >= z0 {
			// X Z Y order
			i1, j1, k2 = 1, 0, 0
			i2, j2, k2 = 1, 0, 1
		} else {
			// Z X Y order
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 1, 0, 1
		}
	} else { /* x0<y0*/
		if y0 < z0 {
			// Z Y X order
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 0, 1, 1
		} else if x0 < z0 {
			// Y Z X order
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 0, 1, 1
		} else {
			// Y X Z order
			i1, j1, k2 = 0, 1, 0
			i2, j2, k2 = 1, 1, 0
		}
	}
	// A step of (1,0,0) in (i,j,k) means a step of (1-c,-c,-c) in (x,y,z),
	// a step of (0,1,0) in (i,j,k) means a step of (-c,1-c,-c) in (x,y,z), and
	// a step of (0,0,1) in (i,j,k) means a step of (-c,-c,1-c) in (x,y,z), where
	// c = 1/6.
	x1 := x0 - float64(i1) + 1.0*G3 // Offsets for second corner in (x,y,z) coords
	y1 := y0 - float64(j1) + 1.0*G3
	z1 := z0 - float64(k1) + 1.0*G3
	x2 := x0 - float64(i2) + 2.0*G3 // Offsets for third corner in (x,y,z) coords
	y2 := y0 - float64(j2) + 2.0*G3
	z2 := z0 - float64(k2) + 2.0*G3
	x3 := x0 - 1.0 + 3.0*G3 // Offsets for last corner in (x,y,z) coords
	y3 := y0 - 1.0 + 3.0*G3
	z3 := z0 - 1.0 + 3.0*G3

	// Work out the hashed gradient indices of the four simplex corners
	ii, jj, kk := int(i)&255, int(j)&255, int(k)&255
	gi0 := octave.permMod12[ii+octave.perm[jj+octave.perm[kk]]]
	gi1 := octave.permMod12[ii+i1+octave.perm[jj+j1+octave.perm[kk+k1]]]
	gi2 := octave.permMod12[ii+i2+octave.perm[jj+j2+octave.perm[kk+k2]]]
	gi3 := octave.permMod12[ii+1+octave.perm[jj+1+octave.perm[kk+1]]]

	// Calculate the contribution from the four corners
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * grad3[gi0].dot3(x0, y0, z0)
	}
	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * grad3[gi1].dot3(x1, y1, z1)
	}
	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * grad3[gi2].dot3(x2, y2, z2)
	}
	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 >= 0 {
		t3 *= t3
		n3 = t3 * t3 * grad3[gi3].dot3(x3, y3, z3)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to stay just inside [-1,1]
	return 32.0 * (n0 + n1 + n2 + n3)
}

// 4D simplex noise, better simplex rank ordering method 2012-03-09

func (octave *Octave) Noise4D(x, y, z, w float64) float64 {
	var n0, n1, n2, n3, n4 float64 // Noise contributions from the five corners
	// Skew the (x,y,z,w) space to determine which cell of 24 simplices we're in
	s := (x + y + z + w) * F4 // Factor for 4D skewing
	i := math.Floor(x + s)
	j := math.Floor(y + s)
	k := math.Floor(z + s)
	l := math.Floor(w + s)

	t := (i + j + k + l) * G4 // Factor for 4D unskewing
	X0 := i - t               // Unskew the cell origin back to (x,y,z,w) space
	Y0 := j - t
	Z0 := k - t
	W0 := l - t
	x0 := x - X0 // The x,y,z,w distances from the cell origin
	y0 := y - Y0
	z0 := z - Z0
	w0 := w - W0

	// For the 4D case, the simplex is a 4D shape I won't even try to describe.
	// To find out which of the 24 possible simplices we're in, we need to
	// determine the magnitude ordering of x0, y0, z0 and w0.
	// Six pair-wise comparisons are performed between each possible pair
	// of the four coordinates, and the results are used to rank the numbers.
	var rankx, ranky, rankz, rankw int
	if x0 > y0 {
		rankx++
	} else {
		ranky++
	}
	if x0 > z0 {
		rankx++
	} else {
		rankz++
	}
	if x0 > w0 {
		rankx++
	} else {
		rankw++
	}
	if y0 > z0 {
		ranky++
	} else {
		rankz++
	}
	if y0 > w0 {
		ranky++
	} else {
		rankw++
	}
	if z0 > w0 {
		rankz++
	} else {
		rankw++
	}

	var i1, j1, k1, l1 int // The integer offsets for the second simplex corner
	var i2, j2, k2, l2 int // The integer offsets for the third simplex corner
	var i3, j3, k3, l3 int // The integer offsets for the fourth simplex corner
	// simplex[c] is a 4-vector with the numbers 0, 1, 2 and 3 in some order.
	// Many values of c will never occur, since e.g. x>y>z>w makes x<z, y<w and x<w
	// impossible. Only the 24 indices which have non-zero entries make any sense.
	// We use a thresholding to set the coordinates in turn from the largest magnitude.

	// Rank 3 denotes the largest coordinate.
	if rankx >= 3 {
		i1 = 1
	}
	if ranky >= 3 {
		j1 = 1
	}
	if rankz >= 3 {
		k1 = 1
	}
	if rankw >= 3 {
		l1 = 1
	}
	// Rank 2 denotes the second largest coordinate.
	if rankx >= 2 {
		i2 = 1
	}
	if ranky >= 2 {
		j2 = 1
	}
	if rankz >= 2 {
		k2 = 1
	}
	if rankw >= 2 {
		l2 = 1
	}
	// Rank 1 denotes the second smallest coordinate.
	if rankx >= 1 {
		i3 = 1
	}
	if ranky >= 1 {
		j3 = 1
	}
	if rankz >= 1 {
		k3 = 1
	}
	if rankw >= 1 {
		l3 = 1
	}

	// The fifth corner has all coordinate offsets = 1, so no need to compute that.
	x1 := x0 - float64(i1) + 1.0*G4 // Offsets for second corner in (x,y,z,w) coords
	y1 := y0 - float64(j1) + 1.0*G4
	z1 := z0 - float64(k1) + 1.0*G4
	w1 := w0 - float64(l1) + 1.0*G4
	x2 := x0 - float64(i2) + 2.0*G4 // Offsets for third corner in (x,y,z,w) coords
	y2 := y0 - float64(j2) + 2.0*G4
	z2 := z0 - float64(k2) + 2.0*G4
	w2 := w0 - float64(l2) + 2.0*G4
	x3 := x0 - float64(i3) + 3.0*G4 // Offsets for fourth corner in (x,y,z,w) coords
	y3 := y0 - float64(j3) + 3.0*G4
	z3 := z0 - float64(k3) + 3.0*G4
	w3 := w0 - float64(l3) + 3.0*G4
	x4 := x0 - 1.0 + 4.0*G4 // Offsets for last corner in (x,y,z,w) coords
	y4 := y0 - 1.0 + 4.0*G4
	z4 := z0 - 1.0 + 4.0*G4
	w4 := w0 - 1.0 + 4.0*G4

	// Work out the hashed gradient indices of the five simplex corners
	ii, jj, kk, ll := int(i)&255, int(j)&255, int(k)&255, int(l)&255
	gi0 := octave.perm[ii+octave.perm[jj+octave.perm[kk+octave.perm[ll]]]] % 32
	gi1 := octave.perm[ii+i1+octave.perm[jj+j1+octave.perm[kk+k1+octave.perm[ll+l1]]]] % 32
	gi2 := octave.perm[ii+i2+octave.perm[jj+j2+octave.perm[kk+k2+octave.perm[ll+l2]]]] % 32
	gi3 := octave.perm[ii+i3+octave.perm[jj+j3+octave.perm[kk+k3+octave.perm[ll+l3]]]] % 32
	gi4 := octave.perm[ii+1+octave.perm[jj+1+octave.perm[kk+1+octave.perm[ll+1]]]] % 32

	// Calculate the contribution from the five corners
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0 - w0*w0
	if t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * grad4[gi0].dot4(x0, y0, z0, w0)
	}
	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1 - w1*w1
	if t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * grad4[gi1].dot4(x1, y1, z1, w1)
	}
	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2 - w2*w2
	if t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * grad4[gi2].dot4(x2, y2, z2, w2)
	}
	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3 - w3*w3
	if t3 >= 0 {
		t3 *= t3
		n3 = t3 * t3 * grad4[gi3].dot4(x3, y3, z3, w3)
	}
	t4 := 0.6 - x4*x4 - y4*y4 - z4*z4 - w4*w4
	if t4 >= 0 {
		t4 *= t4
		n4 = t4 * t4 * grad4[gi4].dot4(x4, y4, z4, w4)
	}

	// Sum up and scale the result to cover the range [-1,1]
	return 27.0 * (n0 + n1 + n2 + n3 + n4)
}
