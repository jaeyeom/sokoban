package main

import (
	"io"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jaeyeom/sokoban/logic"
)

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
