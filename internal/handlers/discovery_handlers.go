package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"MrRSS/internal/discovery"
	"MrRSS/internal/models"
)

// HandleDiscoverBlogs discovers blogs from a feed's friend links.
func (h *Handler) HandleDiscoverBlogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FeedID int64 `json:"feed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the specific feed by ID
	targetFeed, err := h.DB.GetFeedByID(req.FeedID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Feed not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool) // Continue with empty set
	}

	// Discover blogs with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("Starting blog discovery for feed: %s (%s)", targetFeed.Title, targetFeed.URL)
	discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, targetFeed.URL)
	if err != nil {
		log.Printf("Error discovering blogs: %v", err)
		http.Error(w, fmt.Sprintf("Failed to discover blogs: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter out already-subscribed feeds
	filtered := make([]discovery.DiscoveredBlog, 0)
	for _, blog := range discovered {
		if !subscribedURLs[blog.RSSFeed] {
			filtered = append(filtered, blog)
		} else {
			log.Printf("Filtering out already-subscribed feed: %s (%s)", blog.Name, blog.RSSFeed)
		}
	}

	// Mark the feed as discovered
	if err := h.DB.MarkFeedDiscovered(req.FeedID); err != nil {
		log.Printf("Error marking feed as discovered: %v", err)
	}

	log.Printf("Discovered %d blogs, %d after filtering", len(discovered), len(filtered))
	json.NewEncoder(w).Encode(filtered)
}

// HandleStartSingleDiscovery starts a single feed discovery in the background.
func (h *Handler) HandleStartSingleDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FeedID int64 `json:"feed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if a discovery is already running
	h.discoveryMu.Lock()
	if h.singleDiscoveryState != nil && h.singleDiscoveryState.IsRunning {
		h.discoveryMu.Unlock()
		http.Error(w, "Discovery already in progress", http.StatusConflict)
		return
	}

	// Initialize state
	h.singleDiscoveryState = &DiscoveryState{
		IsRunning:  true,
		IsComplete: false,
		Progress: discovery.Progress{
			Stage:   "starting",
			Message: "Starting discovery",
		},
	}
	h.discoveryMu.Unlock()

	// Get the specific feed by ID
	targetFeed, err := h.DB.GetFeedByID(req.FeedID)
	if err != nil {
		h.discoveryMu.Lock()
		h.singleDiscoveryState.IsRunning = false
		h.singleDiscoveryState.IsComplete = true
		h.singleDiscoveryState.Error = "Feed not found"
		h.discoveryMu.Unlock()
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Start discovery in background
	go func() {
		// Create a progress callback that updates the state
		progressCb := func(progress discovery.Progress) {
			h.discoveryMu.Lock()
			if h.singleDiscoveryState != nil {
				h.singleDiscoveryState.Progress = progress
			}
			h.discoveryMu.Unlock()
		}

		ctx, cancel := context.WithTimeout(context.Background(), SingleFeedDiscoveryTimeout)
		defer cancel()

		log.Printf("Starting background discovery for feed: %s (%s)", targetFeed.Title, targetFeed.URL)
		discovered, err := h.DiscoveryService.DiscoverFromFeedWithProgress(ctx, targetFeed.URL, progressCb)

		h.discoveryMu.Lock()
		defer h.discoveryMu.Unlock()

		if h.singleDiscoveryState == nil {
			return
		}

		h.singleDiscoveryState.IsRunning = false
		h.singleDiscoveryState.IsComplete = true

		if err != nil {
			log.Printf("Error discovering blogs: %v", err)
			h.singleDiscoveryState.Error = err.Error()
			return
		}

		// Filter out already-subscribed feeds
		filtered := make([]discovery.DiscoveredBlog, 0)
		for _, blog := range discovered {
			if !subscribedURLs[blog.RSSFeed] {
				filtered = append(filtered, blog)
			}
		}

		h.singleDiscoveryState.Feeds = filtered

		// Mark the feed as discovered
		if err := h.DB.MarkFeedDiscovered(req.FeedID); err != nil {
			log.Printf("Error marking feed as discovered: %v", err)
		}

		log.Printf("Discovery complete: found %d blogs", len(filtered))
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// HandleGetSingleDiscoveryProgress returns the current progress of single feed discovery.
func (h *Handler) HandleGetSingleDiscoveryProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.discoveryMu.RLock()
	state := h.singleDiscoveryState
	h.discoveryMu.RUnlock()

	if state == nil {
		json.NewEncoder(w).Encode(&DiscoveryState{
			IsRunning:  false,
			IsComplete: false,
		})
		return
	}

	json.NewEncoder(w).Encode(state)
}

// HandleClearSingleDiscovery clears the single feed discovery state.
func (h *Handler) HandleClearSingleDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.discoveryMu.Lock()
	h.singleDiscoveryState = nil
	h.discoveryMu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

// HandleDiscoverAllFeeds discovers feeds from all subscriptions that haven't been discovered yet.
func (h *Handler) HandleDiscoverAllFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all feeds
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool) // Continue with empty set
	}

	// Filter feeds that haven't been discovered yet
	var feedsToDiscover []models.Feed
	for _, feed := range feeds {
		if !feed.DiscoveryCompleted {
			feedsToDiscover = append(feedsToDiscover, feed)
		}
	}

	if len(feedsToDiscover) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "All feeds have already been discovered",
			"discovered_from": 0,
			"feeds_found":     0,
		})
		return
	}

	// Discover feeds with timeout
	ctx, cancel := context.WithTimeout(context.Background(), BatchDiscoveryTimeout)
	defer cancel()

	allDiscovered := make(map[string][]discovery.DiscoveredBlog)
	discoveredCount := 0

	log.Printf("Starting batch discovery for %d feeds", len(feedsToDiscover))

discoveryLoop:
	for _, feed := range feedsToDiscover {
		select {
		case <-ctx.Done():
			log.Println("Batch discovery cancelled: timeout")
			break discoveryLoop
		default:
		}

		log.Printf("Discovering from feed: %s (%s)", feed.Title, feed.URL)
		discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, feed.URL)
		if err != nil {
			log.Printf("Error discovering from feed %s: %v", feed.Title, err)
			continue
		}

		// Filter out already-subscribed feeds
		filtered := make([]discovery.DiscoveredBlog, 0)
		for _, blog := range discovered {
			if !subscribedURLs[blog.RSSFeed] {
				filtered = append(filtered, blog)
			}
		}

		if len(filtered) > 0 {
			allDiscovered[feed.Title] = filtered
			discoveredCount += len(filtered)
		}

		// Mark the feed as discovered
		if err := h.DB.MarkFeedDiscovered(feed.ID); err != nil {
			log.Printf("Error marking feed as discovered: %v", err)
		}
	}

	log.Printf("Batch discovery complete: discovered %d feeds from %d sources", discoveredCount, len(feedsToDiscover))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"discovered_from": len(feedsToDiscover),
		"feeds_found":     discoveredCount,
		"feeds":           allDiscovered,
	})
}

// HandleStartBatchDiscovery starts batch discovery in the background.
func (h *Handler) HandleStartBatchDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if a discovery is already running
	h.discoveryMu.Lock()
	if h.batchDiscoveryState != nil && h.batchDiscoveryState.IsRunning {
		h.discoveryMu.Unlock()
		http.Error(w, "Batch discovery already in progress", http.StatusConflict)
		return
	}

	// Initialize state
	h.batchDiscoveryState = &DiscoveryState{
		IsRunning:  true,
		IsComplete: false,
		Progress: discovery.Progress{
			Stage:   "starting",
			Message: "Starting batch discovery",
		},
	}
	h.discoveryMu.Unlock()

	// Get all feeds
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		h.discoveryMu.Lock()
		h.batchDiscoveryState.IsRunning = false
		h.batchDiscoveryState.IsComplete = true
		h.batchDiscoveryState.Error = err.Error()
		h.discoveryMu.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Filter feeds that haven't been discovered yet
	var feedsToDiscover []models.Feed
	for _, feed := range feeds {
		if !feed.DiscoveryCompleted {
			feedsToDiscover = append(feedsToDiscover, feed)
		}
	}

	if len(feedsToDiscover) == 0 {
		h.discoveryMu.Lock()
		h.batchDiscoveryState.IsRunning = false
		h.batchDiscoveryState.IsComplete = true
		h.batchDiscoveryState.Progress.Message = "All feeds have already been discovered"
		h.discoveryMu.Unlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "complete",
			"message": "All feeds have already been discovered",
		})
		return
	}

	// Update initial state with total count
	h.discoveryMu.Lock()
	h.batchDiscoveryState.Progress.Total = len(feedsToDiscover)
	h.discoveryMu.Unlock()

	// Start discovery in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), BatchDiscoveryTimeout)
		defer cancel()

		allDiscovered := make(map[string][]discovery.DiscoveredBlog)
		discoveredCount := 0

		log.Printf("Starting background batch discovery for %d feeds", len(feedsToDiscover))

		for i, feed := range feedsToDiscover {
			select {
			case <-ctx.Done():
				log.Println("Batch discovery cancelled: timeout")
				h.discoveryMu.Lock()
				h.batchDiscoveryState.IsRunning = false
				h.batchDiscoveryState.IsComplete = true
				h.batchDiscoveryState.Error = "Discovery timeout"
				h.discoveryMu.Unlock()
				return
			default:
			}

			// Update progress
			h.discoveryMu.Lock()
			if h.batchDiscoveryState != nil {
				h.batchDiscoveryState.Progress = discovery.Progress{
					Stage:      "processing_feed",
					Message:    fmt.Sprintf("Processing feed %d of %d", i+1, len(feedsToDiscover)),
					Detail:     feed.Title,
					Current:    i + 1,
					Total:      len(feedsToDiscover),
					FeedName:   feed.Title,
					FoundCount: discoveredCount,
				}
			}
			h.discoveryMu.Unlock()

			log.Printf("Discovering from feed: %s (%s)", feed.Title, feed.URL)

			// Create a per-feed progress callback
			feedProgressCb := func(progress discovery.Progress) {
				h.discoveryMu.Lock()
				if h.batchDiscoveryState != nil {
					progress.FeedName = feed.Title
					progress.FoundCount = discoveredCount
					progress.Current = i + 1
					progress.Total = len(feedsToDiscover)
					h.batchDiscoveryState.Progress = progress
				}
				h.discoveryMu.Unlock()
			}

			discovered, err := h.DiscoveryService.DiscoverFromFeedWithProgress(ctx, feed.URL, feedProgressCb)
			if err != nil {
				log.Printf("Error discovering from feed %s: %v", feed.Title, err)
				if err := h.DB.MarkFeedDiscovered(feed.ID); err != nil {
					log.Printf("Error marking feed as discovered: %v", err)
				}
				continue
			}

			// Filter out already-subscribed feeds
			h.discoveryMu.Lock()
			filtered := make([]discovery.DiscoveredBlog, 0)
			for _, blog := range discovered {
				if !subscribedURLs[blog.RSSFeed] {
					filtered = append(filtered, blog)
					subscribedURLs[blog.RSSFeed] = true
				}
			}

			if len(filtered) > 0 {
				allDiscovered[feed.Title] = filtered
				discoveredCount += len(filtered)
			}
			h.discoveryMu.Unlock()

			// Mark the feed as discovered
			if err := h.DB.MarkFeedDiscovered(feed.ID); err != nil {
				log.Printf("Error marking feed as discovered: %v", err)
			}
		}

		log.Printf("Batch discovery complete: discovered %d feeds from %d sources", discoveredCount, len(feedsToDiscover))

		// Update final state
		h.discoveryMu.Lock()
		if h.batchDiscoveryState != nil {
			h.batchDiscoveryState.IsRunning = false
			h.batchDiscoveryState.IsComplete = true
			h.batchDiscoveryState.Progress.Stage = "complete"
			h.batchDiscoveryState.Progress.Message = fmt.Sprintf("Found %d feeds from %d sources", discoveredCount, len(feedsToDiscover))
			h.batchDiscoveryState.Progress.FoundCount = discoveredCount
			// Store feeds as a slice for the response
			var allFeedsSlice []discovery.DiscoveredBlog
			for _, blogs := range allDiscovered {
				allFeedsSlice = append(allFeedsSlice, blogs...)
			}
			h.batchDiscoveryState.Feeds = allFeedsSlice
		}
		h.discoveryMu.Unlock()
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
		"total":  len(feedsToDiscover),
	})
}

// HandleGetBatchDiscoveryProgress returns the current progress of batch discovery.
func (h *Handler) HandleGetBatchDiscoveryProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.discoveryMu.RLock()
	state := h.batchDiscoveryState
	h.discoveryMu.RUnlock()

	if state == nil {
		json.NewEncoder(w).Encode(&DiscoveryState{
			IsRunning:  false,
			IsComplete: false,
		})
		return
	}

	json.NewEncoder(w).Encode(state)
}

// HandleClearBatchDiscovery clears the batch discovery state.
func (h *Handler) HandleClearBatchDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.discoveryMu.Lock()
	h.batchDiscoveryState = nil
	h.discoveryMu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
