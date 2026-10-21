package gcloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"cloud.google.com/go/storage"
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

type memoryStorage struct {
	saved   *logic.GameState
	loadErr error
	saveErr error
}

func (m *memoryStorage) Save(gs logic.GameState) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	cp := gs
	m.saved = &cp
	return nil
}

func (m *memoryStorage) Load() (logic.GameState, error) {
	if m.loadErr != nil {
		return logic.GameState{}, m.loadErr
	}
	if m.saved == nil {
		return logic.GameState{}, logic.ErrNotFound
	}
	return *m.saved, nil
}

func TestSaveFlushesSnapshotBeforeReturn(t *testing.T) {
	underlying := &memoryStorage{}
	delayed := NewDelayedStorage(context.Background(), underlying, make(chan UpdateRequest))
	board := logic.NewBoardFromString("#####\n#O..#\n#####")
	board.MovePlayer(logic.East)
	snapshot := board.String()
	if err := delayed.Save(logic.GameState{Current: 3, Boards: []*logic.Board{board}}); err != nil {
		t.Fatal(err)
	}
	board.MovePlayer(logic.East)
	if board.String() == snapshot {
		t.Fatal("board did not change after save")
	}

	other := NewDelayedStorage(context.Background(), underlying, make(chan UpdateRequest))
	loaded, err := other.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Current != 3 {
		t.Fatalf("current = %d, want 3", loaded.Current)
	}
	if len(loaded.Boards) != 1 || loaded.Boards[0].String() != snapshot {
		got := ""
		if len(loaded.Boards) == 1 {
			got = loaded.Boards[0].String()
		}
		t.Fatalf("loaded board\n%swant\n%s", got, snapshot)
	}
	if loaded.Boards[0].UndoPlayerMove().Dir != logic.East {
		t.Fatal("saved history was not the move from before Save")
	}
}

func TestDelayedLoadReturnsMissingObject(t *testing.T) {
	delayed := NewDelayedStorage(context.Background(), &memoryStorage{}, make(chan UpdateRequest))
	_, err := delayed.Load()
	if !errors.Is(err, logic.ErrNotFound) {
		t.Fatalf("Load err = %v, want ErrNotFound", err)
	}
}

func TestDelayedLoadReturnsOtherErrors(t *testing.T) {
	boom := errors.New("gcs down")
	delayed := NewDelayedStorage(context.Background(), &memoryStorage{loadErr: boom}, make(chan UpdateRequest))
	_, err := delayed.Load()
	if !errors.Is(err, boom) {
		t.Fatalf("Load err = %v, want gcs down", err)
	}
}

func TestSaveReturnsFlushError(t *testing.T) {
	boom := errors.New("write failed")
	delayed := NewDelayedStorage(context.Background(), &memoryStorage{saveErr: boom}, make(chan UpdateRequest))
	err := delayed.Save(logic.GameState{Current: 1})
	if !errors.Is(err, boom) {
		t.Fatalf("Save err = %v, want write failed", err)
	}
}

type gcsStub struct {
	readErr error
}

func (g gcsStub) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return nil, g.readErr
}

func (g gcsStub) NewWriter(ctx context.Context) io.WriteCloser {
	return nopWriteCloser{io.Discard}
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func TestStorageLoadMissingObject(t *testing.T) {
	_, err := NewStorage(gcsStub{readErr: storage.ErrObjectNotExist}).Load()
	if !errors.Is(err, logic.ErrNotFound) {
		t.Fatalf("Load err = %v, want ErrNotFound", err)
	}
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
	if storage.SaveCount != 2 {
		t.Errorf("SaveCount = %d, want 2", storage.SaveCount)
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
