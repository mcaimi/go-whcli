package gui

import (
	"sync"

	"github.com/mcaimi/go-whcli/pkg/common"
)

// AppState holds the GUI's current search parameters. It is read and
// written from both the UI goroutine and background API-fetch goroutines,
// so access is guarded by a mutex.
type AppState struct {
	mu     sync.Mutex
	params common.QueryParameters
}

// NewAppState returns a fresh AppState seeded with the same defaults as the
// CLI's "search"/"random" commands (cmd/whcli/cli.go).
func NewAppState() *AppState {
	return &AppState{
		params: common.QueryParameters{
			Resolution: "1920x1080",
			Purity:     "100",
			Kind:       "111",
			Page:       1,
		},
	}
}

// Params returns a copy of the current query parameters.
func (s *AppState) Params() common.QueryParameters {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.params
}

// SetParams replaces the current query parameters.
func (s *AppState) SetParams(p common.QueryParameters) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = p
}
