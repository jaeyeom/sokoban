package gcloud

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
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
		if errors.Is(err, storage.ErrObjectNotExist) {
			return logic.GameState{}, logic.ErrNotFound
		}
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
	hasCache   bool
	loadErr    error
	updateChan chan UpdateRequest
}

// UpdateRequest represents an update request.
type UpdateRequest struct {
	Save   *logic.GameState
	Load   *logic.GameState
	Flush  bool
	Done   chan struct{}
	result *updateResult
}

type updateResult struct {
	err error
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

// NewDelayedStorage creates a storage that snapshots on Save and flushes that
// snapshot before Save returns. updateChan still flushes a snapshot that has
// not reached storage yet. Use NewUpdateChan to create one.
func NewDelayedStorage(ctx context.Context, storage logic.Storage, updateChan chan UpdateRequest) *DelayedStorage {
	cache := &logic.InMemoryStorage{}
	ds := &DelayedStorage{
		storage:    storage,
		cache:      cache,
		updateChan: updateChan,
	}
	gs, err := storage.Load()
	if err == nil {
		if saveErr := cache.Save(gs); saveErr != nil {
			ds.loadErr = saveErr
		} else {
			ds.hasCache = true
		}
	} else if !errors.Is(err, logic.ErrNotFound) {
		ds.loadErr = err
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
			var err error
			if req.Save != nil {
				err = ds.saveSnapshot(*req.Save)
			}
			if err == nil && req.Load != nil {
				err = ds.loadInto(req.Load)
			}
			if err == nil && req.Flush {
				err = ds.flush()
			}
			if req.result != nil {
				req.result.err = err
			}
			if req.Done != nil {
				close(req.Done)
			}
		}
	}
}

func (ds *DelayedStorage) saveSnapshot(gs logic.GameState) error {
	if ds.loadErr != nil {
		return ds.loadErr
	}
	if err := ds.cache.Save(gs); err != nil {
		return err
	}
	ds.hasCache = true
	ds.dirty = true
	return ds.flush()
}

func (ds *DelayedStorage) flush() error {
	if !ds.dirty {
		return nil
	}
	gs, err := ds.cache.Load()
	if err != nil {
		return err
	}
	if err := ds.storage.Save(gs); err != nil {
		return err
	}
	ds.dirty = false
	return nil
}

func (ds *DelayedStorage) loadInto(dst *logic.GameState) error {
	if ds.loadErr != nil {
		return ds.loadErr
	}
	if !ds.hasCache {
		return logic.ErrNotFound
	}
	gs, err := ds.cache.Load()
	if err != nil {
		return err
	}
	*dst = gs
	return nil
}

func (ds *DelayedStorage) Save(gs logic.GameState) error {
	done := make(chan struct{})
	result := &updateResult{}
	ds.updateChan <- UpdateRequest{Save: &gs, Done: done, result: result}
	<-done
	return result.err
}

func (ds *DelayedStorage) Load() (logic.GameState, error) {
	done := make(chan struct{})
	result := &updateResult{}
	req := UpdateRequest{Load: &logic.GameState{}, Done: done, result: result}
	ds.updateChan <- req
	<-done
	if result.err != nil {
		return logic.GameState{}, result.err
	}
	return *req.Load, nil
}
