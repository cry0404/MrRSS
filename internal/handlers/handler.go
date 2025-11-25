// Package handlers contains the HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package handlers

import (
	"sync"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/translation"
)

// Discovery timeout constants
const (
	// SingleFeedDiscoveryTimeout is the timeout for discovering feeds from a single source
	SingleFeedDiscoveryTimeout = 90 * time.Second
	// BatchDiscoveryTimeout is the timeout for discovering feeds from all sources
	BatchDiscoveryTimeout = 5 * time.Minute
)

// DiscoveryState represents the current state of a discovery operation
type DiscoveryState struct {
	IsRunning  bool                       `json:"is_running"`
	Progress   discovery.Progress         `json:"progress"`
	Feeds      []discovery.DiscoveredBlog `json:"feeds,omitempty"`
	Error      string                     `json:"error,omitempty"`
	IsComplete bool                       `json:"is_complete"`
}

// Handler holds all dependencies for HTTP handlers.
type Handler struct {
	DB               *database.DB
	Fetcher          *feed.Fetcher
	Translator       translation.Translator
	DiscoveryService *discovery.Service

	// Discovery state tracking for polling-based progress
	discoveryMu          sync.RWMutex
	singleDiscoveryState *DiscoveryState
	batchDiscoveryState  *DiscoveryState
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Handler {
	return &Handler{
		DB:               db,
		Fetcher:          fetcher,
		Translator:       translator,
		DiscoveryService: discovery.NewService(),
	}
}
