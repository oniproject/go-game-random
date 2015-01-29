package simplex

import (
	"math"
	"math/rand"
)

type Noise struct {
	octaves        []*Octave
	frequencys     []float64
	amplitudes     []float64
	largestFeature int
	persistence    float64
	seed           int64
}

// like       100,                0.1,                 5000
func NewNoise(largestFeature int, persistence float64, seed int64) *Noise {
	noise := &Noise{
		largestFeature: largestFeature,
		persistence:    persistence,
		seed:           seed,
	}

	// recieves a number (eg 128) and calculates wat power of 2 it's (eg 2^7)
	numberOfOctaves := int(math.Ceil(math.Log10(float64(largestFeature))) / math.Log10(2))

	rnd := rand.New(rand.NewSource(seed))

	for i := 0; i < numberOfOctaves; i++ {
		o := NewOctave(rnd.Int63())
		f := math.Pow(2, float64(i))
		a := math.Pow(persistence, float64(len(noise.octaves)-i+1))
		noise.octaves = append(noise.octaves, o)
		noise.frequencys = append(noise.frequencys, f)
		noise.amplitudes = append(noise.amplitudes, a)
	}

	return noise
}

func (n *Noise) Noise2D(x, y float64) (result float64) {
	for i, octave := range n.octaves {
		//f := n.frequencys[i]
		//a := n.amplitudes[i]
		f := math.Pow(2, float64(i))
		a := math.Pow(n.persistence, float64(len(n.octaves)-i+1))
		result += octave.Noise2D(x/f, y/f) * a
	}
	return
}

func (n *Noise) Noise3D(x, y, z float64) (result float64) {
	for i, octave := range n.octaves {
		f := math.Pow(2, float64(i))
		a := math.Pow(n.persistence, float64(len(n.octaves)-i+1))
		result += octave.Noise3D(float64(x)/f, float64(y)/f, float64(z)/f) * a
	}
	return
}

/*
func (n *Noise) Noise4D(x, y, z, w int) (result float64) {
	for i, octave := range n.octaves {
		f := math.Pow(2, float64(i))
		a := math.Pow(persistence, len(noise.octaves)-i+1)

		nx := x / n.frequencys[i]
		ny := y / n.frequencys[i]
		nz := z / n.frequencys[i]
		nw := w / n.frequencys[i]
		result += octave.Noise4D(fx, fy, fz, fw) * n.amplitudes[i]
	}
	return
}
*/
