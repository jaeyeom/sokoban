package gcloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/jaeyeom/sokoban/logic"
)

type InMemoryGCSObject struct {
	Stored map[string][]byte
	Path   string
}

func (im *InMemoryGCSObject) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(im.Stored[im.Path])), nil
}

func (im *InMemoryGCSObject) NewWriter(ctx context.Context) io.WriteCloser {
	return &InMemoryGCSWriter{im: im, w: &bytes.Buffer{}}
}

type InMemoryGCSWriter struct {
	im *InMemoryGCSObject
	w  *bytes.Buffer
}

func (imw *InMemoryGCSWriter) Write(p []byte) (n int, err error) {
	return imw.w.Write(p)
}

func (imw *InMemoryGCSWriter) Close() error {
	imw.im.Stored[imw.im.Path] = imw.w.Bytes()
	return nil
}

func ExampleNewStorage() {
	storage := NewStorage(&InMemoryGCSObject{
		Stored: map[string][]byte{},
		Path:   "gs://sokoban/game1.json",
	})
	err := storage.Save(logic.GameState{
		Current: 0,
		Boards: []*logic.Board{
			logic.NewBoardFromStringLines([]string{
				"#####",
				"#.O.#",
				"#.XX#",
				"#*++#",
				"#####",
			}),
		},
	})
	if err != nil {
		panic(err)
	}
	gs2, err := storage.Load()
	if err != nil {
		panic(err)
	}
	fmt.Println("Current Level:", gs2.Current)
	for _, b := range gs2.Boards {
		fmt.Println(b.String())
	}
	// Output:
	// Current Level: 0
	// #####
	// #.O.#
	// #.XX#
	// #*++#
	// #####
}

type CountingStorage struct {
	SaveCount int
	LoadCount int
}

func (cs *CountingStorage) Save(gs logic.GameState) error {
	cs.SaveCount++
	return nil
}

func (cs *CountingStorage) Load() (logic.GameState, error) {
	cs.LoadCount++
	return logic.GameState{}, nil
}

func TestDelayedStorage(t *testing.T) {
	storage := &CountingStorage{}
	storage.Save(logic.GameState{})
	if storage.SaveCount != 1 {
		t.Errorf("SaveCount = %d, want 1", storage.SaveCount)
	}
	updateChan := make(chan UpdateRequest)
	delayedStorage := NewDelayedStorage(context.Background(), storage, updateChan)
	delayedStorage.Save(logic.GameState{})
	if storage.SaveCount != 1 {
		t.Errorf("SaveCount = %d, want 1", storage.SaveCount)
	}
	if storage.LoadCount != 1 {
		t.Errorf("LoadCount = %d, want 1", storage.LoadCount)
	}
	done := make(chan struct{})
	updateChan <- UpdateRequest{Flush: true, Done: done}
	<-done
	if storage.SaveCount != 2 {
		t.Errorf("SaveCount = %d, want 2", storage.SaveCount)
	}
	done = make(chan struct{})
	updateChan <- UpdateRequest{Flush: true, Done: done}
	<-done
	if storage.SaveCount != 2 {
		t.Errorf("SaveCount = %d, want 2", storage.SaveCount)
	}
}
