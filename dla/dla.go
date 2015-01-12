// Diffusion-limited aggregation
// see http://www.roguebasin.com/index.php?title=Diffusion-limited_aggregation

package dla

import (
	"../core"
	"math/rand"
)

const (
	MAGIC = 5
)

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type DLA struct {
	OrthogonalAllowed bool
	*rand.Rand

	// variable used to track the percentage of the Map filled
	allocatedBlocks int

	m core.Map
}

// rootX, rootY - is where the gcxth start from. Such as center of the Map
// orthogonalAllowed - Orthogonal movement allowed? If not? it carves a wider cooridor on diagonal
func (dla *DLA) Run(m core.Map, rootX, rootY int, cell uint, max int) {
	dla.m = m
	dla.allocatedBlocks = 0

	builderSpawned := false
	builderMoveDirection := 0
	// this is how lonb corridors can be
	stepped := 0

	var cx, cy int

	// quit when an eight of the Map is filled
	//for allocatedBlocks < m.Width()*m.Height()/8 {
	for dla.allocatedBlocks < max {
		if !builderSpawned {
			// spawn at random position
			cx = 2 + dla.Intn(m.Width()-2)
			cy = 2 + dla.Intn(m.Height()-2)
			// see if builder is ontop of root
			if abs(rootX-cx) == 0 && abs(rootY-cy) == 0 {
				// builder was spawned too close to root,
				// clear that floor and respawn
				dla.simpleAlloc(cx, cy, cell)
			} else {
				builderSpawned = true
				builderMoveDirection = dla.Intn(8)
				stepped = 0
			}
			continue
		}

		// builder already spawned and knows it's direction, move builder
		switch {
		case builderMoveDirection == 0 && cy >= 0: // North
			cy--
			stepped++
		case builderMoveDirection == 1 && cx < m.Height(): // East
			cx++
			stepped++
		case builderMoveDirection == 2 && cy < m.Width(): // South
			cy++
			stepped++
		case builderMoveDirection == 3 && cx >= 0: // West
			cx++
			stepped++
		case builderMoveDirection == 4 && cx < m.Height() && cy >= 0: // Northeast
			cy--
			cx++
			stepped++
		case builderMoveDirection == 5 && cx < m.Height() && cy < m.Width(): // Southeast
			cy++
			cx++
			stepped++
		case builderMoveDirection == 6 && cx >= 0 && cy < m.Height(): // Southwest
			cy++
			cx--
			stepped++
		case builderMoveDirection == 7 && cx >= 0 && cy >= 0: // Northwest
			cy--
			cx--
			stepped++
		}

		// ensure that the builder is touching an existing spot
		if !(m.Check(cx, cy) && stepped <= MAGIC) {
			builderSpawned = false
			continue
		}

		switch {
		case m.At(cx+1, cy) == cell: // East
			dla.simpleAlloc(cx, cy, cell)
		case m.At(cx-1, cy) == cell: // West
			dla.simpleAlloc(cx, cy, cell)
		case m.At(cx, cy+1) == cell: // South
			dla.simpleAlloc(cx, cy, cell)
		case m.At(cx, cy-1) == cell: // North
			dla.simpleAlloc(cx, cy, cell)
		case m.At(cx+1, cy-1) == cell: // Northeast
			dla.orthoAlloc(cx, cy, cell, +1)
		case m.At(cx+1, cy+1) == cell: // Southeast
			dla.orthoAlloc(cx, cy, cell, +1)
		case m.At(cx-1, cy+1) == cell: // Southwest
			dla.orthoAlloc(cx, cy, cell, -1)
		case m.At(cx-1, cy-1) == cell: // Northwest
			dla.orthoAlloc(cx, cy, cell, -1)
		}
	}
}

func (dla *DLA) simpleAlloc(cx, cy int, cell uint) {
	if dla.m.At(cx, cy) != cell {
		dla.m.Set(cx, cy, cell)
		dla.allocatedBlocks++
	}
}
func (dla *DLA) orthoAlloc(cx, cy int, cell uint, sx int) {
	if dla.m.At(cx, cy) != cell {
		dla.m.Set(cx, cy, cell)
		dla.allocatedBlocks++
		if !dla.OrthogonalAllowed {
			dla.m.Set(cx+sx, cy, cell)
			dla.allocatedBlocks++
		}
	}
}
