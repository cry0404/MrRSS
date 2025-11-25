package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/version"
)

// HandleCheckUpdates checks for the latest version on GitHub.
func (h *Handler) HandleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentVersion := version.Version
	const githubAPI = "https://api.github.com/repos/WCY-dt/MrRSS/releases/latest"

	resp, err := http.Get(githubAPI)
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to check for updates",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API returned status: %d", resp.StatusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to fetch latest release",
		})
		return
	}

	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("Error decoding release info: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to parse release information",
		})
		return
	}

	// Remove 'v' prefix if present for comparison
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := compareVersions(latestVersion, currentVersion) > 0

	// Find the appropriate download URL based on platform
	var downloadURL string
	var assetName string
	var assetSize int64
	platform := runtime.GOOS
	arch := runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)

		// Match platform-specific installer/package with architecture
		// Asset naming convention: MrRSS-{version}-{platform}-{arch}-installer.{ext}
		platformArch := platform + "-" + arch

		if platform == "windows" {
			// For Windows, prefer installer.exe, fallback to .zip
			if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-installer.exe") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		} else if platform == "linux" {
			// For Linux, prefer .AppImage, fallback to .tar.gz
			if strings.Contains(name, platformArch) && strings.HasSuffix(name, ".appimage") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		} else if platform == "darwin" {
			// For macOS, use universal build (supports both arm64 and amd64)
			if strings.Contains(name, "darwin-universal") && strings.HasSuffix(name, ".dmg") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		}
	}

	response := map[string]interface{}{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
		"has_update":      hasUpdate,
		"platform":        platform,
		"arch":            arch,
	}

	if downloadURL != "" {
		response["download_url"] = downloadURL
		response["asset_name"] = assetName
		response["asset_size"] = assetSize
	}

	json.NewEncoder(w).Encode(response)
}

// compareVersions compares two semantic versions (e.g., "1.1.0" vs "1.0.0")
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}

// HandleVersion returns the current application version.
func (h *Handler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"version": version.Version,
	})
}

// HandleDownloadUpdate downloads the update file.
func (h *Handler) HandleDownloadUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DownloadURL string `json:"download_url"`
		AssetName   string `json:"asset_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate download URL is from the official GitHub repository releases
	const allowedURLPrefix = "https://github.com/WCY-dt/MrRSS/releases/download/"
	if !strings.HasPrefix(req.DownloadURL, allowedURLPrefix) {
		log.Printf("Invalid download URL attempted: %s", req.DownloadURL)
		http.Error(w, "Invalid download URL", http.StatusBadRequest)
		return
	}

	// Validate asset name to prevent path traversal
	if strings.Contains(req.AssetName, "..") || strings.Contains(req.AssetName, "/") || strings.Contains(req.AssetName, "\\") {
		log.Printf("Invalid asset name attempted: %s", req.AssetName)
		http.Error(w, "Invalid asset name", http.StatusBadRequest)
		return
	}

	// Create temp directory for download
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, req.AssetName)

	// Download the file
	log.Printf("Downloading update from: %s", req.DownloadURL)
	resp, err := http.Get(req.DownloadURL)
	if err != nil {
		log.Printf("Error downloading update: %v", err)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed with status: %d", resp.StatusCode)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Failed to create download file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Write the body to file with progress tracking
	totalSize := resp.ContentLength
	var bytesWritten int64

	// Create a buffer for efficient copying
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		nr, er := resp.Body.Read(buffer)
		if nr > 0 {
			nw, ew := out.Write(buffer[0:nr])
			if nw > 0 {
				bytesWritten += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	if err != nil {
		log.Printf("Error writing file: %v", err)
		os.Remove(filePath) // Clean up partial file
		http.Error(w, "Failed to write download file", http.StatusInternalServerError)
		return
	}

	// Ensure all data is flushed to disk
	if err := out.Sync(); err != nil {
		log.Printf("Error syncing file: %v", err)
		os.Remove(filePath) // Clean up
		http.Error(w, "Failed to save download file", http.StatusInternalServerError)
		return
	}

	// Verify the file size matches expected size
	if totalSize > 0 && bytesWritten != totalSize {
		log.Printf("Download incomplete: expected %d bytes, got %d bytes", totalSize, bytesWritten)
		os.Remove(filePath) // Clean up incomplete file
		http.Error(w, "Download incomplete", http.StatusInternalServerError)
		return
	}

	log.Printf("Update downloaded successfully to: %s (%.2f MB)", filePath, float64(bytesWritten)/(1024*1024))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"total_bytes":   totalSize,
		"bytes_written": bytesWritten,
	})
}

// HandleInstallUpdate triggers the installation of the downloaded update.
func (h *Handler) HandleInstallUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate file path is within temp directory to prevent path traversal
	tempDir := os.TempDir()
	cleanPath := filepath.Clean(req.FilePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(tempDir)) {
		log.Printf("Invalid file path attempted: %s", req.FilePath)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Validate file exists and is a regular file
	fileInfo, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		http.Error(w, "Update file not found", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("Error stating file: %v", err)
		http.Error(w, "Error accessing update file", http.StatusInternalServerError)
		return
	}
	if !fileInfo.Mode().IsRegular() {
		log.Printf("File is not a regular file: %s", cleanPath)
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	platform := runtime.GOOS
	log.Printf("Installing update from: %s on platform: %s", cleanPath, platform)

	// Helper function to schedule cleanup of installer file
	scheduleCleanup := func(filePath string, delay time.Duration) {
		go func() {
			time.Sleep(delay)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove installer: %v", err)
			} else {
				log.Printf("Successfully removed installer: %s", filePath)
			}
		}()
	}

	// Launch installer based on platform
	var cmd *exec.Cmd
	switch platform {
	case "windows":
		// Launch the installer - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".exe") {
			http.Error(w, "Invalid file type for Windows", http.StatusBadRequest)
			return
		}
		// Use start command with /B flag to launch in background
		// Format: start /B <executable_path>
		// The /B flag prevents creating a new window
		cmd = exec.Command("cmd.exe", "/C", "start", "/B", cleanPath)
		scheduleCleanup(cleanPath, 10*time.Second)
	case "linux":
		// Make AppImage executable and run it - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".appimage") {
			http.Error(w, "Invalid file type for Linux", http.StatusBadRequest)
			return
		}
		if err := os.Chmod(cleanPath, 0755); err != nil {
			log.Printf("Error making file executable: %v", err)
			http.Error(w, "Failed to prepare installer", http.StatusInternalServerError)
			return
		}
		cmd = exec.Command(cleanPath)
		scheduleCleanup(cleanPath, 10*time.Second)
	case "darwin":
		// Open the DMG file - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".dmg") {
			http.Error(w, "Invalid file type for macOS", http.StatusBadRequest)
			return
		}
		cmd = exec.Command("open", cleanPath)
		scheduleCleanup(cleanPath, 15*time.Second)
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	// Start the installer in the background
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting installer: %v", err)
		http.Error(w, "Failed to start installer", http.StatusInternalServerError)
		return
	}

	log.Printf("Installer started successfully, PID: %d", cmd.Process.Pid)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Installation started. Application will exit shortly.",
	})

	// Schedule graceful shutdown to allow the response to be sent
	// and give time for proper cleanup
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Initiating graceful shutdown for update installation...")
		// Note: In a production app, this should trigger the Wails shutdown handler
		// which will properly clean up resources. For now, we use os.Exit.
		os.Exit(0)
	}()
}
