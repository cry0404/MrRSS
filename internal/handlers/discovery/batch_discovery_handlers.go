package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"MrRSS/internal/discovery"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

// HandleDiscoverAllFeeds discovers feeds from all subscriptions that haven't been discovered yet.
func HandleDiscoverAllFeeds(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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
	ctx, cancel := context.WithTimeout(context.Background(), core.BatchDiscoveryTimeout)
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
func HandleStartBatchDiscovery(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if a discovery is already running
	h.DiscoveryMu.Lock()
	if h.BatchDiscoveryState != nil && h.BatchDiscoveryState.IsRunning {
		h.DiscoveryMu.Unlock()
		http.Error(w, "Batch discovery already in progress", http.StatusConflict)
		return
	}

	// Initialize state
	h.BatchDiscoveryState = &core.DiscoveryState{
		IsRunning:  true,
		IsComplete: false,
		Progress: discovery.Progress{
			Stage:   "starting",
			Message: "Starting batch discovery",
		},
	}
	h.DiscoveryMu.Unlock()

	// Get all feeds
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		h.DiscoveryMu.Lock()
		h.BatchDiscoveryState.IsRunning = false
		h.BatchDiscoveryState.IsComplete = true
		h.BatchDiscoveryState.Error = err.Error()
		h.DiscoveryMu.Unlock()
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
		h.DiscoveryMu.Lock()
		h.BatchDiscoveryState.IsRunning = false
		h.BatchDiscoveryState.IsComplete = true
		h.BatchDiscoveryState.Progress.Message = "All feeds have already been discovered"
		h.DiscoveryMu.Unlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "complete",
			"message": "All feeds have already been discovered",
		})
		return
	}

	// Update initial state with total count
	h.DiscoveryMu.Lock()
	h.BatchDiscoveryState.Progress.Total = len(feedsToDiscover)
	h.DiscoveryMu.Unlock()

	// Start discovery in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), core.BatchDiscoveryTimeout)
		defer cancel()

		allDiscovered := make(map[string][]discovery.DiscoveredBlog)
		discoveredCount := 0

		log.Printf("Starting background batch discovery for %d feeds", len(feedsToDiscover))

		for i, feed := range feedsToDiscover {
			select {
			case <-ctx.Done():
				log.Println("Batch discovery cancelled: timeout")
				h.DiscoveryMu.Lock()
				h.BatchDiscoveryState.IsRunning = false
				h.BatchDiscoveryState.IsComplete = true
				h.BatchDiscoveryState.Error = "Discovery timeout"
				h.DiscoveryMu.Unlock()
				return
			default:
			}

			// Update progress
			h.DiscoveryMu.Lock()
			if h.BatchDiscoveryState != nil {
				h.BatchDiscoveryState.Progress = discovery.Progress{
					Stage:      "processing_feed",
					Message:    fmt.Sprintf("Processing feed %d of %d", i+1, len(feedsToDiscover)),
					Detail:     feed.Title,
					Current:    i + 1,
					Total:      len(feedsToDiscover),
					FeedName:   feed.Title,
					FoundCount: discoveredCount,
				}
			}
			h.DiscoveryMu.Unlock()

			log.Printf("Discovering from feed: %s (%s)", feed.Title, feed.URL)

			// Create a per-feed progress callback
			feedProgressCb := func(progress discovery.Progress) {
				h.DiscoveryMu.Lock()
				if h.BatchDiscoveryState != nil {
					progress.FeedName = feed.Title
					progress.FoundCount = discoveredCount
					progress.Current = i + 1
					progress.Total = len(feedsToDiscover)
					h.BatchDiscoveryState.Progress = progress
				}
				h.DiscoveryMu.Unlock()
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
			h.DiscoveryMu.Lock()
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
			h.DiscoveryMu.Unlock()

			// Mark the feed as discovered
			if err := h.DB.MarkFeedDiscovered(feed.ID); err != nil {
				log.Printf("Error marking feed as discovered: %v", err)
			}
		}

		log.Printf("Batch discovery complete: discovered %d feeds from %d sources", discoveredCount, len(feedsToDiscover))

		// Update final state
		h.DiscoveryMu.Lock()
		if h.BatchDiscoveryState != nil {
			h.BatchDiscoveryState.IsRunning = false
			h.BatchDiscoveryState.IsComplete = true
			h.BatchDiscoveryState.Progress.Stage = "complete"
			h.BatchDiscoveryState.Progress.Message = fmt.Sprintf("Found %d feeds from %d sources", discoveredCount, len(feedsToDiscover))
			h.BatchDiscoveryState.Progress.FoundCount = discoveredCount
			// Store feeds as a slice for the response
			var allFeedsSlice []discovery.DiscoveredBlog
			for _, blogs := range allDiscovered {
				allFeedsSlice = append(allFeedsSlice, blogs...)
			}
			h.BatchDiscoveryState.Feeds = allFeedsSlice
		}
		h.DiscoveryMu.Unlock()
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
		"total":  len(feedsToDiscover),
	})
}

// HandleGetBatchDiscoveryProgress returns the current progress of batch discovery.
func HandleGetBatchDiscoveryProgress(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.DiscoveryMu.RLock()
	state := h.BatchDiscoveryState
	h.DiscoveryMu.RUnlock()

	if state == nil {
		json.NewEncoder(w).Encode(&core.DiscoveryState{
			IsRunning:  false,
			IsComplete: false,
		})
		return
	}

	json.NewEncoder(w).Encode(state)
}

// HandleClearBatchDiscovery clears the batch discovery state.
func HandleClearBatchDiscovery(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.DiscoveryMu.Lock()
	h.BatchDiscoveryState = nil
	h.DiscoveryMu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
