package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestGameStateOmitsEmptyBoards(t *testing.T) {
	b, err := json.Marshal(GameState{Current: 2})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "boards") {
		t.Fatalf("empty boards were emitted: %s", b)
	}
}

func TestUndoWinningMoveStaysOnSolvedBoard(t *testing.T) {
	game := NewGame(NewBoardsFromString(`
		#####
		#OX+#
		#####

		#####
		#@X.#
		#...#
		#####`), &InMemoryStorage{})
	next := game.boards[1].String()
	game.boards[1].MovePlayer(East)
	dirtied := game.boards[1].String()
	if dirtied == next {
		t.Fatal("expected the next board to change")
	}

	game.Right()
	if game.CurrentLevel() != 1 {
		t.Fatalf("level = %d, want 1", game.CurrentLevel())
	}
	if !game.CurrentBoard().IsSolved() {
		t.Fatalf("board is not solved:\n%s", game.CurrentBoard())
	}
	if got := game.boards[1].String(); got != dirtied {
		t.Fatalf("next board was reset\ngot:\n%swant:\n%s", got, dirtied)
	}

	game.Undo()
	if game.CurrentLevel() != 1 {
		t.Fatalf("level after undo = %d, want 1", game.CurrentLevel())
	}
	if game.CurrentBoard().IsSolved() {
		t.Fatal("undo left the winning push in place")
	}
	want := NewBoardFromString(`
		#####
		#OX+#
		#####`).String()
	if got := game.CurrentBoard().String(); got != want {
		t.Fatalf("after undo\n%swant\n%s", got, want)
	}
	if got := game.boards[1].String(); got != dirtied {
		t.Fatalf("undo reset the next board\ngot:\n%swant:\n%s", got, dirtied)
	}
}

func ExampleNewBoardsFromString() {
	strBoards := `
		#####
		#OX+#
		#####

		#####
		#@X.#
		#...#
		#####`
	boards := NewBoardsFromString(strBoards)
	for i, board := range boards {
		fmt.Println("Board", i+1)
		fmt.Println(board)
	}
	// Output:
	// Board 1
	// #####
	// #OX+#
	// #####
	//
	// Board 2
	// #####
	// #@X.#
	// #...#
	// #####
}

func ExampleGame() {
	strBoards := `
		#####
		#OX+#
		#####

		#####
		#@X.#
		#...#
		#####`
	boards := NewBoardsFromString(strBoards)
	var storage InMemoryStorage
	game := NewGame(boards, &storage)
	printCurrent := func() {
		fmt.Println("Level", game.CurrentLevel())
		fmt.Println(game.CurrentBoard())
	}
	sequence := func(funcs ...func()) {
		for _, f := range funcs {
			f()
			printCurrent()
		}
	}
	save := func() {
		game.Save()
	}
	load := func() {
		game.Load()
	}
	printCurrent()
	sequence(
		game.Right,
		game.Undo,
		game.Right,
		game.NextBoard,
		game.Right,
		game.Right,
		game.Undo,
		game.Down,
		game.Right,
		save,
		game.Right,
		game.Up,
		game.Left,
		load,
		game.ResetCurrentLevel,
	)
	// Output:
	// Level 1
	// #####
	// #OX+#
	// #####
	//
	// Level 1
	// #####
	// #.O*#
	// #####
	//
	// Level 1
	// #####
	// #OX+#
	// #####
	//
	// Level 1
	// #####
	// #.O*#
	// #####
	//
	// Level 2
	// #####
	// #@X.#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #+OX#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #+OX#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #@X.#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #+X.#
	// #O..#
	// #####
	//
	// Level 2
	// #####
	// #+X.#
	// #.O.#
	// #####
	//
	// Level 2
	// #####
	// #+X.#
	// #.O.#
	// #####
	//
	// Level 2
	// #####
	// #+X.#
	// #..O#
	// #####
	//
	// Level 2
	// #####
	// #+XO#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #*O.#
	// #...#
	// #####
	//
	// Level 2
	// #####
	// #+X.#
	// #.O.#
	// #####
	//
	// Level 2
	// #####
	// #@X.#
	// #...#
	// #####
}
