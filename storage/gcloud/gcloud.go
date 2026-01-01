package gcloud

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jaeyeom/sokoban/logic"
)

// ReadWriteFactory represents a factory that creates a reader and a writer.
type ReadWriteFactory interface {
	NewReader(ctx context.Context) (io.ReadCloser, error)
	NewWriter(ctx context.Context) io.WriteCloser
}

// Storage represents a storage.
type Storage struct {
	rwf ReadWriteFactory
}

// NewStorage creates a new storage.
func NewStorage(rwf ReadWriteFactory) *Storage {
	return &Storage{rwf: rwf}
}

// Save saves the game state.
func (s *Storage) Save(gs logic.GameState) error {
	b, err := json.Marshal(&gs)
	if err != nil {
		return err
	}
	ctx := context.Background()
	w := s.rwf.NewWriter(ctx)
	defer w.Close()
	ww := gzip.NewWriter(w)
	defer ww.Close()
	n, err := ww.Write(b)
	log.Printf("wrote %d bytes", n)
	return err
}

// Load loads the game state.
func (s *Storage) Load() (logic.GameState, error) {
	ctx := context.Background()
	r, err := s.rwf.NewReader(ctx)
	if err != nil {
		return logic.GameState{}, err
	}
	defer r.Close()
	rr, err := gzip.NewReader(r)
	if err != nil {
		return logic.GameState{}, err
	}
	defer rr.Close()
	var gs logic.GameState
	err = json.NewDecoder(rr).Decode(&gs)
	return gs, err
}

// GCSObject represents a Google Cloud Storage object.
type GCSObject struct {
	client *storage.Client
	bucket string
	object string
}

// NewGCSObject creates a new GCSObject.
func NewGCSObject(client *storage.Client, bucket, object string) *GCSObject {
	return &GCSObject{client: client, bucket: bucket, object: object}
}

// NewReader creates a new reader from the object.
func (gcs *GCSObject) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return gcs.client.Bucket(gcs.bucket).Object(gcs.object).NewReader(ctx)
}

// NewWriter creates a new writer to the object.
func (gcs *GCSObject) NewWriter(ctx context.Context) io.WriteCloser {
	log.Printf("writing to %s/%s", gcs.bucket, gcs.object)
	return gcs.client.Bucket(gcs.bucket).Object(gcs.object).NewWriter(ctx)
}

// DelayedStorage represents a delayed storage.
type DelayedStorage struct {
	storage    logic.Storage
	cache      logic.Storage
	dirty      bool
	updateChan chan UpdateRequest
}

// UpdateRequest represents an update request.
type UpdateRequest struct {
	Save  *logic.GameState
	Load  *logic.GameState
	Flush bool
	Done  chan struct{}
}

// NewUpdateChan creates a new update channel that sends a flush request
func NewUpdateChan(duration time.Duration) chan UpdateRequest {
	c := make(chan UpdateRequest)
	go func() {
		for {
			time.Sleep(duration)
			c <- UpdateRequest{Flush: true}
		}
	}()
	return c
}

// NewDelayedStorage creates a new delayed storage. It saves the game state to
// the storage when it is flushed. Use NewUpdateChan to create an update channel
// that flushes the storage periodically.
func NewDelayedStorage(ctx context.Context, storage logic.Storage, updateChan chan UpdateRequest) *DelayedStorage {
	cache := &logic.InMemoryStorage{}
	gs, err := storage.Load()
	if err == nil {
		cache.Save(gs)
	}
	ds := &DelayedStorage{
		storage:    storage,
		cache:      cache,
		updateChan: updateChan,
	}
	go ds.run(ctx)
	return ds
}

func (ds *DelayedStorage) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ds.updateChan:
			if req.Save != nil {
				ds.dirty = true
				ds.cache.Save(*req.Save)
			}
			if req.Load != nil {
				gs, err := ds.cache.Load()
				if err == nil {
					*req.Load = gs
				}
			}
			if req.Flush {
				if ds.dirty {
					gs, _ := ds.cache.Load()
					ds.storage.Save(gs)
					ds.dirty = false
				}
			}
			if req.Done != nil {
				close(req.Done)
			}
		}
	}
}

func (ds *DelayedStorage) Save(gs logic.GameState) error {
	done := make(chan struct{})
	ds.updateChan <- UpdateRequest{Save: &gs, Done: done}
	<-done
	return nil
}

func (ds *DelayedStorage) Load() (logic.GameState, error) {
	done := make(chan struct{})
	req := UpdateRequest{Load: &logic.GameState{}, Done: done}
	ds.updateChan <- req
	<-done
	return *req.Load, nil
}
