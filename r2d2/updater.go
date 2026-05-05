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

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpdateState represents the current progress of a background update.
type UpdateState struct {
	Percentage float64
	Status     string
	Done       bool
	Error      error
	URL        string
}

// FetchLatestReleaseInfo checks GitHub for the latest release metadata.
func FetchLatestReleaseInfo() (*GithubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// DownloadWithProgress performs the download and calls progressCb with the percentage.
func DownloadWithProgress(url string, progressCb func(UpdateState)) {
	exePath, err := os.Executable()
	if err != nil {
		progressCb(UpdateState{Error: err})
		return
	}

	newPath := exePath + ".tmp"
	out, err := os.Create(newPath)
	if err != nil {
		progressCb(UpdateState{Error: err})
		return
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		progressCb(UpdateState{Error: err})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progressCb(UpdateState{Error: fmt.Errorf("bad status: %s", resp.Status)})
		return
	}

	size := resp.ContentLength
	buffer := make([]byte, 32*1024)
	var downloaded int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				progressCb(UpdateState{Error: writeErr})
				return
			}
			downloaded += int64(n)
			if size > 0 {
				progressCb(UpdateState{
					Percentage: float64(downloaded) / float64(size),
					Status:     fmt.Sprintf("Downloading: %.0f%%", (float64(downloaded)/float64(size))*100),
				})
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			progressCb(UpdateState{Error: err})
			return
		}
	}

	out.Close()
	
	// Atomic Swap
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		progressCb(UpdateState{Error: err})
		return
	}
	if err := os.Rename(newPath, exePath); err != nil {
		_ = os.Rename(oldPath, exePath)
		progressCb(UpdateState{Error: err})
		return
	}

	progressCb(UpdateState{Done: true, Status: "Update complete! Restarting..."})
}

// RestartApp triggers a self-restart.
func RestartApp() {
	exePath, _ := os.Executable()
	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Start()
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

// Legacy function for backward compatibility if needed, though we will use the new async flow.
func CheckAndApplyUpdate() {
	// (Keeping signature but it will be unused in the new TUI flow)
}
