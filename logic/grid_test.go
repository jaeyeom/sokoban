package logic

import "testing"

func TestGrid_NewGrid_empty(t *testing.T) {
	grid := NewGrid[int](Vec2D{3, 3}, nil)
	tcs := []struct {
		pos Vec2D
		val int
	}{
		{Vec2D{0, 0}, 0},
		{Vec2D{0, 1}, 0},
		{Vec2D{0, 2}, 0},
		{Vec2D{1, 0}, 0},
		{Vec2D{1, 1}, 0},
		{Vec2D{1, 2}, 0},
		{Vec2D{2, 0}, 0},
		{Vec2D{2, 1}, 0},
		{Vec2D{2, 2}, 0},
	}
	for _, tc := range tcs {
		if got := grid.Get(tc.pos); got != tc.val {
			t.Errorf("grid.Get(%v) = %v; want %v", tc.pos, got, tc.val)
		}
	}
}

func TestGrid_Get(t *testing.T) {
	grid := NewGrid(Vec2D{3, 3}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	tcs := []struct {
		pos Vec2D
		val int
	}{
		{Vec2D{0, 0}, 1},
		{Vec2D{0, 1}, 2},
		{Vec2D{0, 2}, 3},
		{Vec2D{1, 0}, 4},
		{Vec2D{1, 1}, 5},
		{Vec2D{1, 2}, 6},
		{Vec2D{2, 0}, 7},
		{Vec2D{2, 1}, 8},
		{Vec2D{2, 2}, 9},
	}
	for _, tc := range tcs {
		if got := grid.Get(tc.pos); got != tc.val {
			t.Errorf("grid.Get(%v) = %v; want %v", tc.pos, got, tc.val)
		}
	}
}

func TestGrid_Set(t *testing.T) {
	grid := NewGrid(Vec2D{3, 3}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	tcs := []struct {
		pos Vec2D
		val int
	}{
		{Vec2D{0, 0}, 10},
		{Vec2D{0, 1}, 20},
		{Vec2D{0, 2}, 30},
		{Vec2D{1, 0}, 40},
		{Vec2D{1, 1}, 50},
		{Vec2D{1, 2}, 60},
		{Vec2D{2, 0}, 70},
		{Vec2D{2, 1}, 80},
		{Vec2D{2, 2}, 90},
	}
	for _, tc := range tcs {
		grid.Set(tc.pos, tc.val)
		if got := grid.Get(tc.pos); got != tc.val {
			t.Errorf("grid.Get(%v) = %v; want %v", tc.pos, got, tc.val)
		}
	}
}
