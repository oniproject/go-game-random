package simplex

import (
	"math"
	"math/rand"
)

type Noise struct {
	LargestFeature int
	Persistence    float64
	Rand           *rand.Rand

	octaves    []*Octave
	frequencys []float64
	amplitudes []float64
}

func (n *Noise) Init() {
	// recieves a number (eg 128) and calculates wat power of 2 it's (eg 2^7)
	numberOfOctaves := int(math.Ceil(math.Log10(float64(n.LargestFeature))) / math.Log10(2))

	for i := 0; i < numberOfOctaves; i++ {
		o := NewOctave(n.Rand)
		f := math.Pow(2, float64(i))
		a := math.Pow(n.Persistence, float64(numberOfOctaves-i))
		n.octaves = append(n.octaves, o)
		n.frequencys = append(n.frequencys, f)
		n.amplitudes = append(n.amplitudes, a)
	}
}

func (n *Noise) Noise2D(x, y float64) (result float64) {
	for i, octave := range n.octaves {
		f, a := n.frequencys[i], n.amplitudes[i]
		result += octave.Noise2D(x/f, y/f) * a
	}
	return
}

func (n *Noise) Noise3D(x, y, z float64) (result float64) {
	for i, octave := range n.octaves {
		f, a := n.frequencys[i], n.amplitudes[i]
		result += octave.Noise3D(float64(x)/f, float64(y)/f, float64(z)/f) * a
	}
	return
}

func (n *Noise) Noise4D(x, y, z, w float64) (result float64) {
	for i, octave := range n.octaves {
		f, a := n.frequencys[i], n.amplitudes[i]
		result += octave.Noise4D(x/f, y/f, z/f, w/f) * a
	}
	return
}
