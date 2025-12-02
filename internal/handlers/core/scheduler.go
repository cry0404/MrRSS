package core

import (
	"context"
	"log"
	"strconv"
	"time"
)

// StartBackgroundScheduler starts the background scheduler for auto-updates and cleanup.
func (h *Handler) StartBackgroundScheduler(ctx context.Context) {
	// Run initial cleanup only if auto_cleanup is enabled
	go func() {
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		if autoCleanup == "true" {
			log.Println("Running initial article cleanup...")
			count, err := h.DB.CleanupOldArticles()
			if err != nil {
				log.Printf("Error during initial cleanup: %v", err)
			} else {
				log.Printf("Initial cleanup: removed %d old articles", count)
			}
		}
	}()

	for {
		intervalStr, err := h.DB.GetSetting("update_interval")
		interval := 10
		if err == nil {
			if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
				interval = i
			}
		}

		log.Printf("Next auto-update in %d minutes", interval)

		select {
		case <-ctx.Done():
			log.Println("Stopping background scheduler")
			return
		case <-time.After(time.Duration(interval) * time.Minute):
			h.Fetcher.FetchAll(ctx)
			// Run cleanup after fetching new articles only if auto_cleanup is enabled
			go func() {
				autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
				if autoCleanup == "true" {
					count, err := h.DB.CleanupOldArticles()
					if err != nil {
						log.Printf("Error during automatic cleanup: %v", err)
					} else if count > 0 {
						log.Printf("Automatic cleanup: removed %d old articles", count)
					}
				}
			}()
		}
	}
}
