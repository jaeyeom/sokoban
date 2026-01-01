// Test for Sokoban game logic package.
package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestBoard_NewBoardFromString(t *testing.T) {
	strBoard := `
            #####
            #.O.#
            #.XX#
            #...#
            #*++#
            #####`
	board := NewBoardFromString(strBoard)
	if board == nil {
		t.Errorf("NewBoardFromString(%q) = nil; want non-nil", strBoard)
	}
}

func ExampleNewBoardFromString() {
	strBoard := `
	    #####
	    #.O.#
	    #.XX#
	    #...#
	    #*++#
	    #####`
	board := NewBoardFromString(strBoard)
	fmt.Println(board)
	// Output:
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
}

func ExampleBoard_MovePlayer() {
	strBoard := `
	    #####
	    #.O.#
	    #.XX#
	    #...#
	    #*++#
	    #####`
	board := NewBoardFromString(strBoard)
	movePlayer := func(dir Vec2D) {
		fmt.Printf("%v %v\n", DirectionName(dir), board.MovePlayer(dir))
		fmt.Println(board)
		if board.IsSolved() {
			fmt.Println("Solved!")
		}
	}
	undo := func() {
		move := board.UndoPlayerMove()
		fmt.Printf("Undo %v %v\n", DirectionName(move.Dir), move.Result)
		fmt.Println(board)
	}
	reset := func() {
		board.Reset()
		fmt.Println("Reset")
		fmt.Println(board)
	}
	fmt.Println(board)
	undo()
	movePlayer(West)
	undo()
	movePlayer(West)
	movePlayer(South)
	reset()
	movePlayer(East)
	movePlayer(East)
	movePlayer(South)
	movePlayer(South)
	movePlayer(South)
	movePlayer(West)
	movePlayer(South)
	movePlayer(North)
	movePlayer(West)
	movePlayer(North)
	movePlayer(North)
	movePlayer(East)
	movePlayer(South)
	movePlayer(South)
	// Output:
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// Undo  UnknownPlayerMoveResult
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// West PlayerMovedToOpen
	// #####
	// #O..#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// Undo West PlayerMovedToOpen
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// West PlayerMovedToOpen
	// #####
	// #O..#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// South PlayerMovedToOpen
	// #####
	// #...#
	// #OXX#
	// #...#
	// #*++#
	// #####
	//
	// Reset
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// East PlayerMovedToOpen
	// #####
	// #..O#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// East PlayerMoveFailed
	// #####
	// #..O#
	// #.XX#
	// #...#
	// #*++#
	// #####
	//
	// South PlayerPushedBox
	// #####
	// #...#
	// #.XO#
	// #..X#
	// #*++#
	// #####
	//
	// South PlayerPushedBox
	// #####
	// #...#
	// #.X.#
	// #..O#
	// #*+*#
	// #####
	//
	// South PlayerMoveFailed
	// #####
	// #...#
	// #.X.#
	// #..O#
	// #*+*#
	// #####
	//
	// West PlayerMovedToOpen
	// #####
	// #...#
	// #.X.#
	// #.O.#
	// #*+*#
	// #####
	//
	// South PlayerMovedToOpen
	// #####
	// #...#
	// #.X.#
	// #...#
	// #*@*#
	// #####
	//
	// North PlayerMovedToOpen
	// #####
	// #...#
	// #.X.#
	// #.O.#
	// #*+*#
	// #####
	//
	// West PlayerMovedToOpen
	// #####
	// #...#
	// #.X.#
	// #O..#
	// #*+*#
	// #####
	//
	// North PlayerMovedToOpen
	// #####
	// #...#
	// #OX.#
	// #...#
	// #*+*#
	// #####
	//
	// North PlayerMovedToOpen
	// #####
	// #O..#
	// #.X.#
	// #...#
	// #*+*#
	// #####
	//
	// East PlayerMovedToOpen
	// #####
	// #.O.#
	// #.X.#
	// #...#
	// #*+*#
	// #####
	//
	// South PlayerPushedBox
	// #####
	// #...#
	// #.O.#
	// #.X.#
	// #*+*#
	// #####
	//
	// South PlayerPushedBox
	// #####
	// #...#
	// #...#
	// #.O.#
	// #***#
	// #####
	//
	// Solved!
}

func Example_marshalJSON() {
	strBoard := `
	    #####
	    #.O.#
	    #.XX#
	    #...#
	    #*++#
	    #####`
	board := NewBoardFromString(strBoard)
	board.UndoPlayerMove()
	board.MovePlayer(West)
	board.UndoPlayerMove()
	board.MovePlayer(East)
	board.MovePlayer(East)
	board.MovePlayer(South)

	b, _ := json.MarshalIndent(board, "", "  ")
	fmt.Println(string(b))
	// Output:
	// {
	//   "grid": "#####\n#...#\n#.XO#\n#..X#\n#*++#\n#####\n",
	//   "playerMoves": "eS"
	// }
}

func ExampleNewPlayerMovesFromString() {
	str := "nEsSw"
	pms := NewPlayerMovesFromString(str)
	fmt.Println(pms)
	// Output:
	// nEsSw
}

func Example_unmarshalJSON() {
	strBoard := `
	    #####
	    #.O.#
	    #.XX#
	    #...#
	    #*++#
	    #####`
	board := NewBoardFromString(strBoard)
	board.UndoPlayerMove()
	board.MovePlayer(West)
	board.UndoPlayerMove()
	board.MovePlayer(East)
	board.MovePlayer(East)
	board.MovePlayer(South)

	b, _ := json.Marshal(board)
	var newBoard Board
	_ = json.Unmarshal(b, &newBoard)
	fmt.Println(&newBoard)
	newBoard.UndoPlayerMove()
	newBoard.UndoPlayerMove()
	fmt.Println("Undo 2 times")
	fmt.Println(&newBoard)
	// Output:
	// #####
	// #...#
	// #.XO#
	// #..X#
	// #*++#
	// #####
	//
	// Undo 2 times
	// #####
	// #.O.#
	// #.XX#
	// #...#
	// #*++#
	// #####
}

func ExampleNormalizeBoardStringLines() {
	// Handles non-recatangular boards.
	strBoard := `
....#####
....#...#
....#X..#
..###..X##
..#..X.X.#
###.#.##.#...######
#...#.##.#####..++#
#.X..X..........++#
#####.###.#O##..++#
....#.....#########
....#######
`
	lines := strings.Split(strings.TrimSpace(strBoard), "\n")
	NormalizeBoardStringLines(lines)
	fmt.Println(strings.Join(lines, "\n"))
	// Output:
	// ....#####..........
	// ....#...#..........
	// ....#X..#..........
	// ..###..X##.........
	// ..#..X.X.#.........
	// ###.#.##.#...######
	// #...#.##.#####..++#
	// #.X..X..........++#
	// #####.###.#O##..++#
	// ....#.....#########
	// ....#######........
}
