package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
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
	games         map[string]*logic.Game
	storages      map[string]logic.Storage
	createStorage func(string) logic.Storage
	levels        []byte
	index         []byte
	gameApp       []byte
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

func (s *Server) handleGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Path[len("/game/"):]
	if _, ok := s.games[gameID]; !ok {
		s.storages[gameID] = s.createStorage(gameID)
		s.games[gameID] = logic.NewGame(logic.NewBoardsFromString(string(s.levels)), s.storages[gameID])
		s.games[gameID].Load()
	}
	game := s.games[gameID]
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.Write(s.index)
	} else if r.Method == http.MethodPost {
		var req struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
		case "":
			// Do nothing.
		}
		if err := game.Save(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		gameState := struct {
			Current int    `json:"current"`
			Grid    string `json:"grid"`
		}{
			Current: game.CurrentLevel(),
			Grid:    game.CurrentBoard().String(),
		}
		b, err := json.Marshal(gameState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	}
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
