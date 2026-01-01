package logic

// Vec2D is a 2D vector of int.
type Vec2D [2]int

// Add returns the sum of two vectors.
func (v Vec2D) Add(w Vec2D) Vec2D {
	return Vec2D{v[0] + w[0], v[1] + w[1]}
}

// Opposite returns the opposite of the vector.
func (v Vec2D) Opposite() Vec2D {
	return Vec2D{-v[0], -v[1]}
}

// Grid is a 2D grid of any type.
type Grid[T any] struct {
	size Vec2D
	data []T
}

// NewGrid returns a new grid of the given size and data. If data is nil, it
// will be initialized with the zero value of T.
func NewGrid[T any](size Vec2D, data []T) *Grid[T] {
	if data == nil {
		data = make([]T, size[0]*size[1])
	}
	return &Grid[T]{size, data}
}

// Size returns the size of the grid.
func (g Grid[T]) Size() Vec2D {
	return g.size
}

// Get returns the value at the given position.
func (g Grid[T]) Get(pos Vec2D) T {
	return g.data[pos[0]*g.size[1]+pos[1]]
}

// Set sets the value at the given position.
func (g Grid[T]) Set(pos Vec2D, val T) {
	g.data[pos[0]*g.size[1]+pos[1]] = val
}
