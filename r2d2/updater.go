package r2d2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	repoOwner = "Victxrlarixs"
	repoName  = "r2d2-monitor"
	assetName = "r2d2-monitor-portable.exe"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckAndApplyUpdate checks for a new version on GitHub and applies it if found.
// This is designed to be called at startup.
func CheckAndApplyUpdate() {
	LogInfo("Checking for updates...")
	
	latest, err := getLatestRelease()
	if err != nil {
		LogError(err, "Failed to check for updates")
		return
	}

	// Compare versions (simple string compare for now, assuming vX.Y.Z format)
	if latest.TagName == Version {
		LogInfo("System is up to date.")
		return
	}

	fmt.Printf("\n[R2-D2] *Whistle* New version found: %s (Current: %s)\n", latest.TagName, Version)
	fmt.Println("[R2-D2] *Bleep bloop* Downloading update...")

	var downloadURL string
	for _, asset := range latest.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		LogInfo("No suitable asset found in latest release.")
		return
	}

	if err := downloadAndReplace(downloadURL); err != nil {
		LogError(err, "Failed to apply update")
		fmt.Printf("[R2-D2] *Sad bloop* Update failed: %v\n", err)
		return
	}

	fmt.Println("[R2-D2] *Joyful beep* Update complete! Restarting in 3 seconds...")
	time.Sleep(time.Second * 3)
	restartProcess()
}

func getLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadAndReplace(url string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	oldPath := exePath + ".old"
	newPath := exePath + ".tmp"

	// 1. Download to tmp
	out, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()

	// 2. Cleanup old if exists
	_ = os.Remove(oldPath)

	// 3. Rename current to .old
	// On Windows, you can rename a running executable!
	if err := os.Rename(exePath, oldPath); err != nil {
		return err
	}

	// 4. Rename tmp to current
	if err := os.Rename(newPath, exePath); err != nil {
		// Rollback if failed
		_ = os.Rename(oldPath, exePath)
		return err
	}

	return nil
}

func restartProcess() {
	exePath, _ := os.Executable()
	
	// Start the new process
	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error restarting: %v\n", err)
		os.Exit(1)
	}
	
	// Exit the current process
	os.Exit(0)
}

// CleanupOldVersion removes the .old file left over from a previous update.
func CleanupOldVersion() {
	exePath, _ := os.Executable()
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		_ = os.Remove(oldPath)
		LogInfo("Cleaned up old version binary.")
	}
}
