package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jaeyeom/sokoban/logic"
	"github.com/jaeyeom/sokoban/storage/gcloud"
	"github.com/jaeyeom/sugo/errors/must"
)

var (
	dataPath = flag.String("data_path", "./data/", "Path to data directory")
)

type Server struct {
	mu            sync.Mutex
	games         map[string]*logic.Game
	storages      map[string]logic.Storage
	sessions      map[string]*gameSession
	createStorage func(string) logic.Storage
	levels        []byte
	index         []byte
	gameApp       []byte
}

// gameSession serializes move, undo, and save for one game.
type gameSession struct {
	mu     sync.Mutex
	game   *logic.Game
	loaded bool
	ready  chan struct{}
}

func createStorage(gameID string) logic.Storage {
	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		bucket = "sokoban-jaeyeom"
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	storage := gcloud.NewStorage(gcloud.NewGCSObject(
		client, bucket, fmt.Sprintf("save-%s.json.gz", gameID)))
	delayedStorage := gcloud.NewDelayedStorage(
		context.Background(),
		storage,
		gcloud.NewUpdateChan(10*time.Second),
	)
	return delayedStorage
}

func (s *Server) sessionFor(gameID string) *gameSession {
	s.mu.Lock()
	if s.sessions == nil {
		s.sessions = make(map[string]*gameSession)
	}
	if sess, ok := s.sessions[gameID]; ok {
		s.mu.Unlock()
		<-sess.ready
		return sess
	}
	sess := &gameSession{ready: make(chan struct{})}
	s.sessions[gameID] = sess
	if s.games != nil {
		if game, ok := s.games[gameID]; ok {
			sess.game = game
			sess.loaded = true
			close(sess.ready)
			s.mu.Unlock()
			return sess
		}
	}
	s.mu.Unlock()

	st := s.createStorage(gameID)
	game := logic.NewGame(logic.NewBoardsFromString(string(s.levels)), st)

	s.mu.Lock()
	if s.games == nil {
		s.games = make(map[string]*logic.Game)
	}
	if s.storages == nil {
		s.storages = make(map[string]logic.Storage)
	}
	sess.game = game
	s.games[gameID] = game
	s.storages[gameID] = st
	close(sess.ready)
	s.mu.Unlock()
	return sess
}

func (s *Server) loadIfNeeded(sess *gameSession) error {
	if sess.loaded {
		return nil
	}
	err := sess.game.Load()
	if err != nil && !errors.Is(err, logic.ErrNotFound) {
		return err
	}
	sess.loaded = true
	return nil
}

func (s *Server) handleGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Path[len("/game/"):]
	sess := s.sessionFor(gameID)
	if r.Method == http.MethodGet {
		sess.mu.Lock()
		err := s.loadIfNeeded(sess)
		sess.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(s.index)
		return
	}
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess.mu.Lock()
	if err := s.loadIfNeeded(sess); err != nil {
		sess.mu.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := sess.game
	switch req.Action {
	case "right":
		game.Right()
	case "left":
		game.Left()
	case "up":
		game.Up()
	case "down":
		game.Down()
	case "undo":
		game.Undo()
	case "reset":
		game.ResetCurrentLevel()
	case "next":
		if game.CurrentBoard().IsSolved() {
			game.NextBoard()
		}
	case "":
		// Do nothing.
	}
	if err := game.Save(); err != nil {
		sess.mu.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(struct {
		Current int    `json:"current"`
		Grid    string `json:"grid"`
	}{
		Current: game.CurrentLevel(),
		Grid:    game.CurrentBoard().String(),
	})
	sess.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	flag.Parse()
	levels := must.Bytes(os.ReadFile(path.Join(*dataPath, "levels.txt")))
	tileset := must.Bytes(os.ReadFile(path.Join(*dataPath, "assets/tileset.png")))
	index := must.Bytes(os.ReadFile(path.Join(*dataPath, "index.html")))
	gameApp := must.Bytes(os.ReadFile(path.Join(*dataPath, "app/game.js")))
	bundle := must.Bytes(os.ReadFile(path.Join(*dataPath, "app/dist/bundle.js")))
	s := Server{
		games:         make(map[string]*logic.Game),
		storages:      make(map[string]logic.Storage),
		createStorage: createStorage,
		levels:        levels,
		index:         index,
		gameApp:       gameApp,
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Start the web server.
	http.HandleFunc("/assets/tileset.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(tileset)
	})
	http.HandleFunc("/app/game.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(gameApp)
	})
	http.HandleFunc("/app/dist/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(bundle)
	})
	http.HandleFunc("/game/", s.handleGame)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
