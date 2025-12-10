// Package cache provides media caching functionality for anti-hotlinking support.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MediaCache handles caching of images and videos to work around anti-hotlinking
type MediaCache struct {
	cacheDir string
}

// NewMediaCache creates a new media cache instance
func NewMediaCache(cacheDir string) (*MediaCache, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &MediaCache{
		cacheDir: cacheDir,
	}, nil
}

// GetCachedPath returns the cached file path for a given URL
func (mc *MediaCache) GetCachedPath(url string) string {
	hash := hashURL(url)
	ext := getExtensionFromURL(url)
	return filepath.Join(mc.cacheDir, hash+ext)
}

// Exists checks if a media file is already cached
func (mc *MediaCache) Exists(url string) bool {
	path := mc.GetCachedPath(url)
	_, err := os.Stat(path)
	return err == nil
}

// Get retrieves cached media or downloads it if not cached
func (mc *MediaCache) Get(url, referer string) ([]byte, string, error) {
	// Check if already cached
	cachedPath := mc.GetCachedPath(url)
	if mc.Exists(url) {
		data, err := os.ReadFile(cachedPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read cached file: %w", err)
		}
		contentType := getContentTypeFromPath(cachedPath)
		return data, contentType, nil
	}

	// Download and cache
	data, contentType, err := mc.download(url, referer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download media: %w", err)
	}

	// Save to cache
	if err := os.WriteFile(cachedPath, data, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to cache media: %w", err)
	}

	return data, contentType, nil
}

// download fetches media from the given URL with proper headers
func (mc *MediaCache) download(url, referer string) ([]byte, string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to bypass anti-hotlinking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = getContentTypeFromPath(url)
	}

	return data, contentType, nil
}

// CleanupOldFiles removes cached files older than the specified age
func (mc *MediaCache) CleanupOldFiles(maxAgeDays int) (int, error) {
	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)
	count := 0

	entries, err := os.ReadDir(mc.cacheDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(mc.cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err == nil {
				count++
			}
		}
	}

	return count, nil
}

// GetCacheSize returns the total size of cached files in bytes
func (mc *MediaCache) GetCacheSize() (int64, error) {
	var totalSize int64

	entries, err := os.ReadDir(mc.cacheDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
	}

	return totalSize, nil
}

// CleanupBySize removes oldest files until cache is under the size limit
func (mc *MediaCache) CleanupBySize(maxSizeMB int) (int, error) {
	maxSize := int64(maxSizeMB) * 1024 * 1024
	currentSize, err := mc.GetCacheSize()
	if err != nil {
		return 0, err
	}

	if currentSize <= maxSize {
		return 0, nil
	}

	// Get all files with their modification times
	type fileInfo struct {
		path    string
		modTime time.Time
		size    int64
	}

	var files []fileInfo
	entries, err := os.ReadDir(mc.cacheDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileInfo{
			path:    filepath.Join(mc.cacheDir, entry.Name()),
			modTime: info.ModTime(),
			size:    info.Size(),
		})
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files until under limit
	count := 0
	for _, f := range files {
		if currentSize <= maxSize {
			break
		}

		if err := os.Remove(f.path); err == nil {
			currentSize -= f.size
			count++
		}
	}

	return count, nil
}

// hashURL creates a SHA256 hash of the URL for use as filename
func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}

// getExtensionFromURL extracts the file extension from URL
func getExtensionFromURL(url string) string {
	// Remove query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	ext := filepath.Ext(url)
	if ext == "" {
		// Try to guess from URL patterns
		if strings.Contains(url, "image") || strings.Contains(url, "img") {
			return ".jpg"
		}
		if strings.Contains(url, "video") || strings.Contains(url, "vid") {
			return ".mp4"
		}
		return ".bin"
	}

	return ext
}

// getContentTypeFromPath determines content type from file extension
func getContentTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
}
