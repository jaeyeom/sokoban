package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jaeyeom/sokoban/logic"
)

func postGame(s *Server, id, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/game/"+id, strings.NewReader(body))
	s.handleGame(w, req)
	return w
}

func TestNextLevelAfterSolve(t *testing.T) {
	levels := []byte(`
#####
#OX+#
#####

#####
#@X.#
#...#
#####
`)
	var storage logic.InMemoryStorage
	s := &Server{
		games: map[string]*logic.Game{
			"1": logic.NewGame(logic.NewBoardsFromString(string(levels)), &storage),
		},
		storages: map[string]logic.Storage{"1": &storage},
	}
	stayed := postGame(s, "1", `{"action":"next"}`)
	if stayed.Code != http.StatusOK {
		t.Fatalf("unsolved next status = %d", stayed.Code)
	}
	if s.games["1"].CurrentLevel() != 1 {
		t.Fatalf("unsolved next level = %d, want 1", s.games["1"].CurrentLevel())
	}

	solved := postGame(s, "1", `{"action":"right"}`)
	if solved.Code != http.StatusOK {
		t.Fatalf("solve status = %d body %s", solved.Code, solved.Body.String())
	}
	if s.games["1"].CurrentLevel() != 1 {
		t.Fatalf("solved level = %d, want 1", s.games["1"].CurrentLevel())
	}
	if !strings.Contains(solved.Body.String(), "#.O*#") {
		t.Fatalf("solved grid = %s", solved.Body.String())
	}

	advanced := postGame(s, "1", `{"action":"next"}`)
	if advanced.Code != http.StatusOK {
		t.Fatalf("next status = %d body %s", advanced.Code, advanced.Body.String())
	}
	if s.games["1"].CurrentLevel() != 2 {
		t.Fatalf("level = %d, want 2", s.games["1"].CurrentLevel())
	}
	if !strings.Contains(advanced.Body.String(), "#@X.#") {
		t.Fatalf("next grid = %s", advanced.Body.String())
	}
}

type loadErrStorage struct {
	err error
}

func (s loadErrStorage) Save(gs logic.GameState) error { return nil }

func (s loadErrStorage) Load() (logic.GameState, error) {
	return logic.GameState{}, s.err
}

func TestLoadErrorIsReturned(t *testing.T) {
	loads := 0
	s := &Server{
		games:    map[string]*logic.Game{},
		storages: map[string]logic.Storage{},
		createStorage: func(string) logic.Storage {
			return loadCountStorage{loads: &loads, err: errors.New("gcs down")}
		},
		levels: []byte("#####\n#O..#\n#####\n"),
	}
	first := postGame(s, "missing", `{"action":"right"}`)
	if first.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body %s", first.Code, first.Body.String())
	}
	second := postGame(s, "missing", `{"action":"right"}`)
	if second.Code != http.StatusInternalServerError {
		t.Fatalf("retry status = %d, want 500; body %s", second.Code, second.Body.String())
	}
	if loads != 2 {
		t.Fatalf("loads = %d, want 2", loads)
	}
}

type loadCountStorage struct {
	loads *int
	err   error
}

func (s loadCountStorage) Save(gs logic.GameState) error { return nil }

func (s loadCountStorage) Load() (logic.GameState, error) {
	*s.loads++
	return logic.GameState{}, s.err
}

func TestMissingSaveStartsNewGame(t *testing.T) {
	s := &Server{
		games:    map[string]*logic.Game{},
		storages: map[string]logic.Storage{},
		createStorage: func(string) logic.Storage {
			return loadErrStorage{err: logic.ErrNotFound}
		},
		levels: []byte("#####\n#O..#\n#####\n"),
	}
	w := postGame(s, "new", `{"action":"right"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "#.O.#") {
		t.Fatalf("grid = %s", w.Body.String())
	}
	if s.games["new"].CurrentLevel() != 1 {
		t.Fatalf("level = %d", s.games["new"].CurrentLevel())
	}
}

func TestOverlappingMoveAndUndo(t *testing.T) {
	const layout = "#####\n#O..#\n#####\n"
	var storage logic.InMemoryStorage
	game := logic.NewGame(logic.NewBoardsFromString(layout), &storage)
	s := &Server{
		games:    map[string]*logic.Game{"g": game},
		storages: map[string]logic.Storage{"g": &storage},
	}
	start := logic.NewBoardFromString(layout).String()
	movedBoard := logic.NewBoardFromString(layout)
	movedBoard.MovePlayer(logic.East)
	moved := movedBoard.String()

	var wg sync.WaitGroup
	release := make(chan struct{})
	for _, action := range []string{`{"action":"right"}`, `{"action":"undo"}`} {
		wg.Add(1)
		go func(action string) {
			defer wg.Done()
			<-release
			w := postGame(s, "g", action)
			if w.Code != http.StatusOK {
				t.Errorf("status = %d body %s", w.Code, w.Body.String())
			}
		}(action)
	}
	close(release)
	wg.Wait()

	got := s.games["g"].CurrentBoard().String()
	if got != start && got != moved {
		t.Fatalf("board\n%swant either\n%s\nor\n%s", got, start, moved)
	}
}

func Test_handleGame(t *testing.T) {
	tcs := []struct {
		method               string
		body                 string
		expectedCurrentBoard string
	}{
		{"GET", "", `
#####
#.O.#
#.XX#
#...#
#*++#
#####
`},
		{"POST", `{"action":"right"}`, `
#####
#..O#
#.XX#
#...#
#*++#
#####
`},
		{"POST", `{"action":"down"}`, `
#####
#...#
#.XO#
#..X#
#*++#
#####
`},
		{"POST", `{"action":"left"}`, `
#####
#...#
#XO.#
#..X#
#*++#
#####
`},
		{"POST", `{"action":"up"}`, `
#####
#.O.#
#X..#
#..X#
#*++#
#####
`},
		{"POST", `{"action":"undo"}`, `
#####
#...#
#XO.#
#..X#
#*++#
#####
`},
		{"POST", `{"action":"undo"}`, `
#####
#...#
#.XO#
#..X#
#*++#
#####
`},
	}
	b, err := os.ReadFile("../../data/levels.txt")
	if err != nil {
		log.Fatal(err)
	}
	var storage logic.InMemoryStorage
	s := Server{
		games: map[string]*logic.Game{
			"1": logic.NewGame(logic.NewBoardsFromString(string(b)), &storage),
		},
		storages: map[string]logic.Storage{
			"1": &storage,
		},
	}
	s.games["1"].Save()
	for i, tc := range tcs {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, "/game/1", strings.NewReader(tc.body))
		s.handleGame(w, req)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		newStorage := logic.InMemoryStorage{Stored: body}
		newGame := logic.NewGame(nil, &newStorage)
		newGame.Load()
		expected := strings.TrimSpace(tc.expectedCurrentBoard) + "\n"
		if got := s.games["1"].CurrentBoard().String(); got != expected {
			t.Errorf("%d. CurrentBoard() = %q; want %q", i, got, expected)
		}
	}
}
