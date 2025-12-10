package core

import (
	"context"
	"log"
	"strconv"
	"time"

	"MrRSS/internal/cache"
	"MrRSS/internal/utils"
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

		// Run initial media cache cleanup if enabled
		mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
		if mediaCacheEnabled == "true" {
			log.Println("Running initial media cache cleanup...")
			h.cleanupMediaCache()
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

				// Run media cache cleanup if enabled
				mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
				if mediaCacheEnabled == "true" {
					h.cleanupMediaCache()
				}
			}()
		}
	}
}

// cleanupMediaCache performs media cache cleanup based on settings
func (h *Handler) cleanupMediaCache() {
	cacheDir, err := utils.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		return
	}

	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		return
	}

	// Get settings
	maxAgeDaysStr, _ := h.DB.GetSetting("media_cache_max_age_days")
	maxSizeMBStr, _ := h.DB.GetSetting("media_cache_max_size_mb")

	maxAgeDays, err := strconv.Atoi(maxAgeDaysStr)
	if err != nil || maxAgeDays <= 0 {
		maxAgeDays = 7 // Default
	}

	maxSizeMB, err := strconv.Atoi(maxSizeMBStr)
	if err != nil || maxSizeMB <= 0 {
		maxSizeMB = 100 // Default
	}

	// Cleanup by age
	ageCount, err := mediaCache.CleanupOldFiles(maxAgeDays)
	if err != nil {
		log.Printf("Failed to cleanup old media files: %v", err)
	} else if ageCount > 0 {
		log.Printf("Media cache cleanup: removed %d old files", ageCount)
	}

	// Cleanup by size
	sizeCount, err := mediaCache.CleanupBySize(maxSizeMB)
	if err != nil {
		log.Printf("Failed to cleanup media files by size: %v", err)
	} else if sizeCount > 0 {
		log.Printf("Media cache cleanup: removed %d files to stay under size limit", sizeCount)
	}
}
