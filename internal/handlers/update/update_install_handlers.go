package update

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
)

// HandleInstallUpdate triggers the installation of the downloaded update.
func HandleInstallUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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
