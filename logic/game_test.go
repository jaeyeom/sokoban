package logic

import (
	"fmt"
)

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
	// Level 1
	// #####
	// #OX+#
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
