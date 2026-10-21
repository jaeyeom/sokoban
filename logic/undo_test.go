package logic

import (
	"encoding/json"
	"math/rand"
	"testing"
)

func TestUndoRestoresSuccessfulMoves(t *testing.T) {
	board := NewBoardFromString(`
		#####
		#.O.#
		#.XX#
		#...#
		#*++#
		#####`)
	dirs := []Vec2D{North, East, South, West}
	rng := rand.New(rand.NewSource(1))
	history := []string{board.String()}
	counts := []int{board.numBoxesOnTargets}
	for i := 0; i < 80; i++ {
		if board.MovePlayer(dirs[rng.Intn(len(dirs))]) == PlayerMoveFailed {
			continue
		}
		history = append(history, board.String())
		counts = append(counts, board.numBoxesOnTargets)
	}
	if len(history) < 2 {
		t.Fatal("expected some successful moves")
	}
	for i := len(history) - 2; i >= 0; i-- {
		board.UndoPlayerMove()
		if got := board.String(); got != history[i] {
			t.Fatalf("undo to step %d\ngot:\n%swant:\n%s", i, got, history[i])
		}
		if board.numBoxesOnTargets != counts[i] {
			t.Fatalf("step %d boxes on targets = %d, want %d", i, board.numBoxesOnTargets, counts[i])
		}
	}
}

func TestUndoPushOntoTarget(t *testing.T) {
	board := NewBoardFromString(`
		#####
		#O..#
		#X..#
		#+..#
		#####`)
	start := board.String()
	startOnTargets := board.numBoxesOnTargets
	if got := board.MovePlayer(South); got != PlayerPushedBox {
		t.Fatalf("push = %v\n%s", got, board)
	}
	if board.numBoxesOnTargets != startOnTargets+1 {
		t.Fatalf("boxes on targets = %d, want %d", board.numBoxesOnTargets, startOnTargets+1)
	}
	board.UndoPlayerMove()
	board.UndoPlayerMove()
	if board.String() != start {
		t.Fatalf("after undo\n%s\nwant\n%s", board, start)
	}
	if board.numBoxesOnTargets != startOnTargets {
		t.Fatalf("boxes on targets = %d, want %d", board.numBoxesOnTargets, startOnTargets)
	}
}

func TestSaveLoadKeepsUndoHistory(t *testing.T) {
	board := NewBoardFromString(`
		#####
		#.O.#
		#.XX#
		#...#
		#*++#
		#####`)
	start := board.String()
	board.MovePlayer(East)
	board.MovePlayer(South)
	raw, err := json.Marshal(board)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Board
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatal(err)
	}
	loaded.UndoPlayerMove()
	loaded.UndoPlayerMove()
	if loaded.String() != start {
		t.Fatalf("after save/load undo\n%s\nwant\n%s", loaded.String(), start)
	}
	if loaded.playerMoves.String() != "" {
		t.Fatalf("moves = %q", loaded.playerMoves.String())
	}
}
