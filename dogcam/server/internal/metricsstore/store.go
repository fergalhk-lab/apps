package metricsstore

import (
	"sync"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
)

type Store struct {
	mu      sync.RWMutex
	payload *dogcampb.MetricsPayload
}

func New() *Store {
	return &Store{}
}

func (s *Store) Update(p *dogcampb.MetricsPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payload = p
}

func (s *Store) Get() *dogcampb.MetricsPayload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.payload
}
