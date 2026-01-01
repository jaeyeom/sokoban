// Package logic implements the main logic of Sokoban game.
package logic

import (
	"encoding/json"
	"strings"
)

// Tile represents a tile in the Sokoban game.
type Tile int

const (
	Empty          Tile = 0
	Player         Tile = 1
	Box            Tile = 2
	Wall           Tile = 3
	Target         Tile = 4
	PlayerOnTarget Tile = 5
	BoxOnTarget    Tile = 6
)

var (
	runesToTile = map[rune]Tile{
		'.': Empty,
		'O': Player,
		'X': Box,
		'#': Wall,
		'+': Target,
		'@': PlayerOnTarget,
		'*': BoxOnTarget,
	}

	tileToRunes = map[Tile]rune{
		Empty:          '.',
		Player:         'O',
		Box:            'X',
		Wall:           '#',
		Target:         '+',
		PlayerOnTarget: '@',
		BoxOnTarget:    '*',
	}
)

// NewTileFromRune returns a new tile from the given rune.
func NewTileFromRune(r rune) Tile {
	return runesToTile[r]
}

// String returns the string representation of the tile.
func (t Tile) String() string {
	return string(tileToRunes[t])
}

// IsPlayer returns true if the player is on the tile.
func (t Tile) IsPlayer() bool {
	return t&^Target == Player
}

// IsBox returns true if the box is on the tile.
func (t Tile) IsBox() bool {
	return t&^Target == Box
}

// IsWall returns true if the wall is on the tile.
func (t Tile) IsWall() bool {
	return t == Wall
}

// IsTarget returns true if the target is on the tile.
func (t Tile) IsTarget() bool {
	return t&Target != 0
}

// AddBox adds a box to the tile.
func (t Tile) AddBox() Tile {
	return t | Box
}

// RemoveBox removes a box from the tile.
func (t Tile) RemoveBox() Tile {
	return t & ^Box
}

// AddPlayer adds a player to the tile.
func (t Tile) AddPlayer() Tile {
	return t | Player
}

// RemovePlayer removes a player from the tile.
func (t Tile) RemovePlayer() Tile {
	return t & ^Player
}

// PlayerMoveResult represents the result of a player move.
type PlayerMoveResult int

const (
	UnknownPlayerMoveResult PlayerMoveResult = 0
	PlayerMoveFailed        PlayerMoveResult = 1
	PlayerMovedToOpen       PlayerMoveResult = 2
	PlayerPushedBox         PlayerMoveResult = 3
)

// String returns the string representation of the player move result.
func (p PlayerMoveResult) String() string {
	switch p {
	case PlayerMoveFailed:
		return "PlayerMoveFailed"
	case PlayerMovedToOpen:
		return "PlayerMovedToOpen"
	case PlayerPushedBox:
		return "PlayerPushedBox"
	}
	return "UnknownPlayerMoveResult"
}

var playerMoveEncoding = map[rune]PlayerMove{
	'n': {North, PlayerMovedToOpen},
	'e': {East, PlayerMovedToOpen},
	's': {South, PlayerMovedToOpen},
	'w': {West, PlayerMovedToOpen},
	'N': {North, PlayerPushedBox},
	'E': {East, PlayerPushedBox},
	'S': {South, PlayerPushedBox},
	'W': {West, PlayerPushedBox},
}

// PlayerMove represents a player move in the Sokoban game. It is used for
// recording the player moves and undoing them.
type PlayerMove struct {
	Dir    Vec2D
	Result PlayerMoveResult
}

func (m PlayerMove) String() string {
	for r, pm := range playerMoveEncoding {
		if pm == m {
			return string(r)
		}
	}
	return ""
}

// PlayerMoves represents a list of player moves.
type PlayerMoves []PlayerMove

// NewPlayerMovesFromString returns a new player moves from the given string.
func NewPlayerMovesFromString(s string) PlayerMoves {
	var pm PlayerMoves
	for _, r := range s {
		pm = append(pm, playerMoveEncoding[r])
	}
	return pm
}

// Pushed returns a new player moves with the given player move pushed.
func (pm PlayerMoves) Pushed(m PlayerMove) PlayerMoves {
	return append(pm, m)
}

// Popped returns the last player move and the rest of the player moves.
func (pm PlayerMoves) Popped() (PlayerMove, PlayerMoves) {
	if len(pm) == 0 {
		return PlayerMove{}, nil
	}
	return pm[len(pm)-1], pm[:len(pm)-1]
}

// String returns the string representation of the player moves.
func (pm PlayerMoves) String() string {
	var sb strings.Builder
	for _, m := range pm {
		sb.WriteString(m.String())
	}
	return sb.String()
}

// Board represents a Sokoban game board.
type Board struct {
	grid              *Grid[Tile]
	playerPos         Vec2D
	numBoxes          int
	numBoxesOnTargets int
	playerMoves       PlayerMoves
}

// NewBoardFromStringLines returns a new board from the given string lines.
// Please see NewBoardFromString for the format of the string.
func NewBoardFromStringLines(lines []string) *Board {
	NormalizeBoardStringLines(lines)
	rows, cols := len(lines), len(lines[0])
	grid := NewGrid[Tile](Vec2D{rows, cols}, nil)
	playerPos := Vec2D{-1, -1}
	numBoxes, numBoxesOnTargets := 0, 0
	for r, line := range lines {
		for c, t := range []rune(line) {
			tile := NewTileFromRune(t)
			grid.Set(Vec2D{r, c}, tile)
			if tile.IsPlayer() {
				playerPos = Vec2D{r, c}
			}
			if tile.IsBox() {
				numBoxes++
				if tile.IsTarget() {
					numBoxesOnTargets++
				}
			}
		}
	}
	return &Board{grid, playerPos, numBoxes, numBoxesOnTargets, nil}
}

// NewBoardFromString returns a new board from the given string. The string
// should represent the board in the following format:
//
//	#####
//	#.O.#
//	#.XX#
//	#...#
//	#*++#
//	#####
//
// If the player is on the target, use '@' instead of 'O'. If the box is on the
// target, use '*' instead of 'X'.
func NewBoardFromString(s string) *Board {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return NewBoardFromStringLines(lines)
}

// String returns the string representation of the board.
func (b *Board) String() string {
	var sb strings.Builder
	size := b.grid.Size()
	for r := 0; r < size[0]; r++ {
		for c := 0; c < size[1]; c++ {
			sb.WriteString(b.grid.Get(Vec2D{r, c}).String())
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

var (
	North = Vec2D{-1, 0}
	East  = Vec2D{0, 1}
	South = Vec2D{1, 0}
	West  = Vec2D{0, -1}

	directionNames = map[Vec2D]string{
		North: "North",
		East:  "East",
		South: "South",
		West:  "West",
	}
)

// DirectionName returns the name of the direction or empty string.
func DirectionName(dir Vec2D) string {
	return directionNames[dir]
}

func (b *Board) moveBox(pos, dir Vec2D) bool {
	nextPos := pos.Add(dir)
	t := b.grid.Get(nextPos)
	if t.IsWall() || t.IsBox() {
		return false
	}
	posTile, nextPosTile := b.grid.Get(pos), b.grid.Get(nextPos)
	if posTile.IsTarget() {
		b.numBoxesOnTargets--
	}
	if nextPosTile.IsTarget() {
		b.numBoxesOnTargets++
	}
	b.grid.Set(pos, posTile.RemoveBox())
	b.grid.Set(nextPos, nextPosTile.AddBox())
	return true
}

func (b *Board) setPlayerPos(pos Vec2D) {
	b.grid.Set(b.playerPos, b.grid.Get(b.playerPos).RemovePlayer())
	b.grid.Set(pos, b.grid.Get(pos).AddPlayer())
	b.playerPos = pos
}

func (b *Board) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Grid        string `json:"grid"`
		PlayerMoves string `json:"playerMoves"`
	}{
		Grid:        b.String(),
		PlayerMoves: b.playerMoves.String(),
	})
}

func (b *Board) UnmarshalJSON(data []byte) error {
	var board struct {
		Grid        string `json:"grid"`
		PlayerMoves string `json:"playerMoves"`
	}
	if err := json.Unmarshal(data, &board); err != nil {
		return err
	}
	newBoard := NewBoardFromString(board.Grid)
	newBoard.playerMoves = NewPlayerMovesFromString(board.PlayerMoves)
	*b = *newBoard
	return nil
}

// MovePlayer moves the player in the given direction. It returns the result of
// the move. If the player moved to an open tile, the result is
// PlayerMovedToOpen. If the player pushed a box, the result is PlayerPushedBox.
// Otherwise the result is PlayerMoveFailed.
func (b *Board) MovePlayer(dir Vec2D) (result PlayerMoveResult) {
	pos := b.playerPos
	nextPos := pos.Add(dir)
	t := b.grid.Get(nextPos)
	if t.IsWall() {
		return PlayerMoveFailed
	}
	if t.IsBox() {
		if !b.moveBox(nextPos, dir) {
			return PlayerMoveFailed
		}
		result = PlayerPushedBox
	} else {
		result = PlayerMovedToOpen
	}
	b.setPlayerPos(nextPos)
	b.playerMoves = b.playerMoves.Pushed(PlayerMove{dir, result})
	return result
}

// UndoPlayerMove undoes the last player move. It returns move that was undone.
// For example, if the player pushed a box to the north right before calling
// this function, it will return PlayerMove{North, PlayerPushedBox} and the
// player and the box will move back to the south. If there's no moves to undo,
// it will return zero value of PlayerMove type.
func (b *Board) UndoPlayerMove() (move PlayerMove) {
	if len(b.playerMoves) == 0 {
		return PlayerMove{}
	}
	move, b.playerMoves = b.playerMoves.Popped()
	oppositeDir := move.Dir.Opposite()
	originalPos := b.playerPos
	undoPos := originalPos.Add(oppositeDir)
	b.setPlayerPos(undoPos)
	if move.Result == PlayerPushedBox {
		boxPos := originalPos.Add(move.Dir)
		b.moveBox(boxPos, oppositeDir)
	}

	return move
}

// IsSolved returns true if all the boxes are on the target.
func (b *Board) IsSolved() bool {
	return b.numBoxes-b.numBoxesOnTargets == 0
}

// Reset
func (b *Board) Reset() {
	n := len(b.playerMoves)
	for i := 0; i < n; i++ {
		b.UndoPlayerMove()
	}
}

func NormalizeBoardStringLines(lines []string) {
	var cols int
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
		if cols < len(lines[i]) {
			cols = len(lines[i])
		}
	}
	for i, line := range lines {
		lines[i] = line + strings.Repeat(".", cols-len(line))
	}
}
