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
/*
  private static double dot(Grad g, double x, double y) {
    return g.x*x + g.y*y; }

  private static double dot(Grad g, double x, double y, double z) {
    return g.x*x + g.y*y + g.z*z; }

  private static double dot(Grad g, double x, double y, double z, double w) {
    return g.x*x + g.y*y + g.z*z + g.w*w; }
*/

type Octave struct {
	perm      [512]int
	permMod12 [512]int
}

func NewOctave(seed int64) *Octave {
	octave := &Octave{}
	p := rand.Perm(256)

	for i := 0; i < 512; i++ {
		octave.perm[i] = p[i&255] // ? 256
		octave.permMod12[i] = octave.perm[i] % 12
	}
	return octave
}

// 2D simplex noise
func (octave *Octave) Noise2D(xin, yin float64) float64 {
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

	var n0, n1, n2 float64 // Noise contributions from the three corners

	// Calculate the contribution from the three corners
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * grad3[gi0].dot2(x0, y0) // (x,y) of grad3 used for 2D gradient
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * grad3[gi1].dot2(x1, y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 < 0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * grad3[gi2].dot2(x2, y2)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to return values in the interval [-1,1].
	return 70.0 * (n0 + n1 + n2)
}

/*

  // 3D simplex noise
  public static double noise(double xin, double yin, double zin) {
    double n0, n1, n2, n3; // Noise contributions from the four corners
    // Skew the input space to determine which simplex cell we're in
    double s = (xin+yin+zin)*F3; // Very nice and simple skew factor for 3D
    int i = fastfloor(xin+s);
    int j = fastfloor(yin+s);
    int k = fastfloor(zin+s);
    double t = (i+j+k)*G3;
    double X0 = i-t; // Unskew the cell origin back to (x,y,z) space
    double Y0 = j-t;
    double Z0 = k-t;
    double x0 = xin-X0; // The x,y,z distances from the cell origin
    double y0 = yin-Y0;
    double z0 = zin-Z0;
    // For the 3D case, the simplex shape is a slightly irregular tetrahedron.
    // Determine which simplex we are in.
    int i1, j1, k1; // Offsets for second corner of simplex in (i,j,k) coords
    int i2, j2, k2; // Offsets for third corner of simplex in (i,j,k) coords
    if(x0>=y0) {
      if(y0>=z0)
        { i1=1; j1=0; k1=0; i2=1; j2=1; k2=0; } // X Y Z order
        else if(x0>=z0) { i1=1; j1=0; k1=0; i2=1; j2=0; k2=1; } // X Z Y order
        else { i1=0; j1=0; k1=1; i2=1; j2=0; k2=1; } // Z X Y order
      }
    else { // x0<y0
      if(y0<z0) { i1=0; j1=0; k1=1; i2=0; j2=1; k2=1; } // Z Y X order
      else if(x0<z0) { i1=0; j1=1; k1=0; i2=0; j2=1; k2=1; } // Y Z X order
      else { i1=0; j1=1; k1=0; i2=1; j2=1; k2=0; } // Y X Z order
    }
    // A step of (1,0,0) in (i,j,k) means a step of (1-c,-c,-c) in (x,y,z),
    // a step of (0,1,0) in (i,j,k) means a step of (-c,1-c,-c) in (x,y,z), and
    // a step of (0,0,1) in (i,j,k) means a step of (-c,-c,1-c) in (x,y,z), where
    // c = 1/6.
    double x1 = x0 - i1 + G3; // Offsets for second corner in (x,y,z) coords
    double y1 = y0 - j1 + G3;
    double z1 = z0 - k1 + G3;
    double x2 = x0 - i2 + 2.0*G3; // Offsets for third corner in (x,y,z) coords
    double y2 = y0 - j2 + 2.0*G3;
    double z2 = z0 - k2 + 2.0*G3;
    double x3 = x0 - 1.0 + 3.0*G3; // Offsets for last corner in (x,y,z) coords
    double y3 = y0 - 1.0 + 3.0*G3;
    double z3 = z0 - 1.0 + 3.0*G3;
    // Work out the hashed gradient indices of the four simplex corners
    int ii = i & 255;
    int jj = j & 255;
    int kk = k & 255;
    int gi0 = permMod12[ii+perm[jj+perm[kk]]];
    int gi1 = permMod12[ii+i1+perm[jj+j1+perm[kk+k1]]];
    int gi2 = permMod12[ii+i2+perm[jj+j2+perm[kk+k2]]];
    int gi3 = permMod12[ii+1+perm[jj+1+perm[kk+1]]];
    // Calculate the contribution from the four corners
    double t0 = 0.6 - x0*x0 - y0*y0 - z0*z0;
    if(t0<0) n0 = 0.0;
    else {
      t0 *= t0;
      n0 = t0 * t0 * dot(grad3[gi0], x0, y0, z0);
    }
    double t1 = 0.6 - x1*x1 - y1*y1 - z1*z1;
    if(t1<0) n1 = 0.0;
    else {
      t1 *= t1;
      n1 = t1 * t1 * dot(grad3[gi1], x1, y1, z1);
    }
    double t2 = 0.6 - x2*x2 - y2*y2 - z2*z2;
    if(t2<0) n2 = 0.0;
    else {
      t2 *= t2;
      n2 = t2 * t2 * dot(grad3[gi2], x2, y2, z2);
    }
    double t3 = 0.6 - x3*x3 - y3*y3 - z3*z3;
    if(t3<0) n3 = 0.0;
    else {
      t3 *= t3;
      n3 = t3 * t3 * dot(grad3[gi3], x3, y3, z3);
    }
    // Add contributions from each corner to get the final noise value.
    // The result is scaled to stay just inside [-1,1]
    return 32.0*(n0 + n1 + n2 + n3);
  }


  // 4D simplex noise, better simplex rank ordering method 2012-03-09
  public static double noise(double x, double y, double z, double w) {

    double n0, n1, n2, n3, n4; // Noise contributions from the five corners
    // Skew the (x,y,z,w) space to determine which cell of 24 simplices we're in
    double s = (x + y + z + w) * F4; // Factor for 4D skewing
    int i = fastfloor(x + s);
    int j = fastfloor(y + s);
    int k = fastfloor(z + s);
    int l = fastfloor(w + s);
    double t = (i + j + k + l) * G4; // Factor for 4D unskewing
    double X0 = i - t; // Unskew the cell origin back to (x,y,z,w) space
    double Y0 = j - t;
    double Z0 = k - t;
    double W0 = l - t;
    double x0 = x - X0;  // The x,y,z,w distances from the cell origin
    double y0 = y - Y0;
    double z0 = z - Z0;
    double w0 = w - W0;
    // For the 4D case, the simplex is a 4D shape I won't even try to describe.
    // To find out which of the 24 possible simplices we're in, we need to
    // determine the magnitude ordering of x0, y0, z0 and w0.
    // Six pair-wise comparisons are performed between each possible pair
    // of the four coordinates, and the results are used to rank the numbers.
    int rankx = 0;
    int ranky = 0;
    int rankz = 0;
    int rankw = 0;
    if(x0 > y0) rankx++; else ranky++;
    if(x0 > z0) rankx++; else rankz++;
    if(x0 > w0) rankx++; else rankw++;
    if(y0 > z0) ranky++; else rankz++;
    if(y0 > w0) ranky++; else rankw++;
    if(z0 > w0) rankz++; else rankw++;
    int i1, j1, k1, l1; // The integer offsets for the second simplex corner
    int i2, j2, k2, l2; // The integer offsets for the third simplex corner
    int i3, j3, k3, l3; // The integer offsets for the fourth simplex corner
    // simplex[c] is a 4-vector with the numbers 0, 1, 2 and 3 in some order.
    // Many values of c will never occur, since e.g. x>y>z>w makes x<z, y<w and x<w
    // impossible. Only the 24 indices which have non-zero entries make any sense.
    // We use a thresholding to set the coordinates in turn from the largest magnitude.
    // Rank 3 denotes the largest coordinate.
    i1 = rankx >= 3 ? 1 : 0;
    j1 = ranky >= 3 ? 1 : 0;
    k1 = rankz >= 3 ? 1 : 0;
    l1 = rankw >= 3 ? 1 : 0;
    // Rank 2 denotes the second largest coordinate.
    i2 = rankx >= 2 ? 1 : 0;
    j2 = ranky >= 2 ? 1 : 0;
    k2 = rankz >= 2 ? 1 : 0;
    l2 = rankw >= 2 ? 1 : 0;
    // Rank 1 denotes the second smallest coordinate.
    i3 = rankx >= 1 ? 1 : 0;
    j3 = ranky >= 1 ? 1 : 0;
    k3 = rankz >= 1 ? 1 : 0;
    l3 = rankw >= 1 ? 1 : 0;
    // The fifth corner has all coordinate offsets = 1, so no need to compute that.
    double x1 = x0 - i1 + G4; // Offsets for second corner in (x,y,z,w) coords
    double y1 = y0 - j1 + G4;
    double z1 = z0 - k1 + G4;
    double w1 = w0 - l1 + G4;
    double x2 = x0 - i2 + 2.0*G4; // Offsets for third corner in (x,y,z,w) coords
    double y2 = y0 - j2 + 2.0*G4;
    double z2 = z0 - k2 + 2.0*G4;
    double w2 = w0 - l2 + 2.0*G4;
    double x3 = x0 - i3 + 3.0*G4; // Offsets for fourth corner in (x,y,z,w) coords
    double y3 = y0 - j3 + 3.0*G4;
    double z3 = z0 - k3 + 3.0*G4;
    double w3 = w0 - l3 + 3.0*G4;
    double x4 = x0 - 1.0 + 4.0*G4; // Offsets for last corner in (x,y,z,w) coords
    double y4 = y0 - 1.0 + 4.0*G4;
    double z4 = z0 - 1.0 + 4.0*G4;
    double w4 = w0 - 1.0 + 4.0*G4;
    // Work out the hashed gradient indices of the five simplex corners
    int ii = i & 255;
    int jj = j & 255;
    int kk = k & 255;
    int ll = l & 255;
    int gi0 = perm[ii+perm[jj+perm[kk+perm[ll]]]] % 32;
    int gi1 = perm[ii+i1+perm[jj+j1+perm[kk+k1+perm[ll+l1]]]] % 32;
    int gi2 = perm[ii+i2+perm[jj+j2+perm[kk+k2+perm[ll+l2]]]] % 32;
    int gi3 = perm[ii+i3+perm[jj+j3+perm[kk+k3+perm[ll+l3]]]] % 32;
    int gi4 = perm[ii+1+perm[jj+1+perm[kk+1+perm[ll+1]]]] % 32;
    // Calculate the contribution from the five corners
    double t0 = 0.6 - x0*x0 - y0*y0 - z0*z0 - w0*w0;
    if(t0<0) n0 = 0.0;
    else {
      t0 *= t0;
      n0 = t0 * t0 * dot(grad4[gi0], x0, y0, z0, w0);
    }
   double t1 = 0.6 - x1*x1 - y1*y1 - z1*z1 - w1*w1;
    if(t1<0) n1 = 0.0;
    else {
      t1 *= t1;
      n1 = t1 * t1 * dot(grad4[gi1], x1, y1, z1, w1);
    }
   double t2 = 0.6 - x2*x2 - y2*y2 - z2*z2 - w2*w2;
    if(t2<0) n2 = 0.0;
    else {
      t2 *= t2;
      n2 = t2 * t2 * dot(grad4[gi2], x2, y2, z2, w2);
    }
   double t3 = 0.6 - x3*x3 - y3*y3 - z3*z3 - w3*w3;
    if(t3<0) n3 = 0.0;
    else {
      t3 *= t3;
      n3 = t3 * t3 * dot(grad4[gi3], x3, y3, z3, w3);
    }
   double t4 = 0.6 - x4*x4 - y4*y4 - z4*z4 - w4*w4;
    if(t4<0) n4 = 0.0;
    else {
      t4 *= t4;
      n4 = t4 * t4 * dot(grad4[gi4], x4, y4, z4, w4);
    }
    // Sum up and scale the result to cover the range [-1,1]
    return 27.0 * (n0 + n1 + n2 + n3 + n4);
  }
*/