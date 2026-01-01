package logic

import (
	"encoding/json"
	"errors"
	"strings"
)

func NewBoardsFromString(strBoards string) []*Board {
	var boards []*Board
	lines := append(strings.Split(strings.TrimSpace(strBoards), "\n"), "")
	start := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			if start == i {
				// Skip the empty board.
				start = i + 1
				continue
			}
			boards = append(boards, NewBoardFromStringLines(lines[start:i]))
			start = i + 1
		}
	}
	return boards
}

// GameState represents the state of a game.
type GameState struct {
	Current int      `json:"current"`
	Boards  []*Board `json:"boards",omitempty`
}

// InMemoryStorage represents an in-memory storage.
type InMemoryStorage struct {
	Stored []byte
}

// Save saves the game state.
func (ims *InMemoryStorage) Save(gs GameState) error {
	b, err := json.Marshal(&gs)
	if err != nil {
		return err
	}
	ims.Stored = b
	return nil
}

// Load loads the game state.
func (ims *InMemoryStorage) Load() (GameState, error) {
	var gs GameState
	err := json.Unmarshal(ims.Stored, &gs)
	return gs, err
}

// Storage represents a storage.
type Storage interface {
	Save(gs GameState) error
	Load() (GameState, error)
}

// Game represents a game with multiple boards.
type Game struct {
	current int
	boards  []*Board
	storage Storage
}

// NewGame returns a new game with the given boards.
func NewGame(boards []*Board, storage Storage) *Game {
	if storage == nil {
		storage = &InMemoryStorage{}
	}
	return &Game{0, boards, storage}
}

// CurrentLevel returns the current level number.
func (g *Game) CurrentLevel() int {
	return g.current + 1
}

// CurrentBoard returns the current board.
func (g *Game) CurrentBoard() *Board {
	return g.boards[g.current]
}

// NextBoard moves to the next board.
func (g *Game) NextBoard() {
	g.current = (g.current + 1) % len(g.boards)
}

// PrevBoard moves to the previous board.
func (g *Game) PrevBoard() {
	g.current = (g.current - 1 + len(g.boards)) % len(g.boards)
}

func (g *Game) handlePlayerMove() {
	if g.CurrentBoard().IsSolved() {
		g.NextBoard()
		g.CurrentBoard().Reset()
	}
}

// Left moves the player to the left.
func (g *Game) Left() {
	g.CurrentBoard().MovePlayer(West)
	g.handlePlayerMove()
}

// Right moves the player to the right.
func (g *Game) Right() {
	g.CurrentBoard().MovePlayer(East)
	g.handlePlayerMove()
}

// Up moves the player up.
func (g *Game) Up() {
	g.CurrentBoard().MovePlayer(North)
	g.handlePlayerMove()
}

// Down moves the player down.
func (g *Game) Down() {
	g.CurrentBoard().MovePlayer(South)
	g.handlePlayerMove()
}

// Undo undoes the last player move.
func (g *Game) Undo() {
	g.CurrentBoard().UndoPlayerMove()
}

// ResetCurrentLevel resets the current level.
func (g *Game) ResetCurrentLevel() {
	g.CurrentBoard().Reset()
}

// Save saves the game.
func (g *Game) Save() error {
	return g.storage.Save(GameState{
		Current: g.current,
		Boards:  g.boards,
	})
}

var ErrInvalidGameState = errors.New("invalid game state")

// Load loads the game.
func (g *Game) Load() error {
	loaded, err := g.storage.Load()
	if err != nil {
		return err
	}
	if len(loaded.Boards)-1 < loaded.Current {
		return ErrInvalidGameState
	}
	g.current = loaded.Current
	g.boards = loaded.Boards
	return nil
}
