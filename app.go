package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	startupFile string
	watcher     *fsnotify.Watcher
	watchMap    map[string]bool
}

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Body    string `json:"body"`
}

// UpdateCheckResult represents the result of an update check
type UpdateCheckResult struct {
	HasUpdate   bool   `json:"hasUpdate"`
	Version     string `json:"version"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		watchMap: make(map[string]bool),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	// close watcher cleanly when app exits
	if a.watcher != nil {
		a.watcher.Close()
		a.watcher = nil
	}
}

// StartWatching starts watching a file for changes
func (a *App) StartWatching(path string) error {
	if a.watchMap[path] {
		return nil
	}

	if a.watcher == nil {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		a.watcher = w

		debounceMap := make(map[string]*time.Timer)

		// start background goroutine to listen for events
		go func() {
			for {
				select {
				case event, ok := <-a.watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Write) {
						// debounce — wait 300ms after last event
						// reset timer if another event fires
						if timer, exists := debounceMap[event.Name]; exists {
							timer.Stop()
						}
						path := event.Name
						debounceMap[path] = time.AfterFunc(300*time.Millisecond, func() {
							runtime.EventsEmit(a.ctx, "fileChanged", path)
							delete(debounceMap, path)
						})
					}
				case err, ok := <-a.watcher.Errors:
					if !ok {
						return
					}
					_ = err
				}
			}
		}()
	}

	// add file to watcher
	err := a.watcher.Add(path)
	if err != nil {
		return err
	}

	// mark as watched
	a.watchMap[path] = true
	return nil
}

// StopWatching stops watching a file for changes
func (a *App) StopWatching(path string) error {
	// not watching this file? skip
	if !a.watchMap[path] {
		return nil
	}

	// remove from watcher
	if a.watcher != nil {
		err := a.watcher.Remove(path)
		if err != nil {
			return err
		}
	}

	// remove from watchMap
	delete(a.watchMap, path)

	// if no more files watched → close watcher entirely
	if len(a.watchMap) == 0 && a.watcher != nil {
		a.watcher.Close()
		a.watcher = nil
	}

	return nil
}

func (a *App) GetStartupFile() string {
	return a.startupFile
}

// OpenFile opens a native file dialog and returns the selected file path
func (a *App) OpenFile() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown Files",
				Pattern:     "*.md;*.markdown;*.txt;*.mdx",
			},
		},
	})
	if err != nil {
		return ""
	}
	return path
}

// OpenImportFile opens a native file dialog for importing documents
func (a *App) OpenImportFile() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Document",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PDF Files",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return ""
	}
	return path
}

// ReadFile reads a file from disk and returns its content as text
func (a *App) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReadFileBytes reads a file and returns raw bytes (for binary files like PDF)
func (a *App) ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// SaveFile saves content to an existing file path
func (a *App) SaveFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// SaveFileAs opens a native save dialog and saves content to chosen path
func (a *App) SaveFileAs(content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Markdown File",
		DefaultFilename: "untitled.md",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown Files",
				Pattern:     "*.md;*.markdown",
			},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

// CheckForUpdates checks the GitHub API for the latest release
func (a *App) CheckForUpdates(currentVersion string) UpdateCheckResult {
	const repoURL = "https://api.github.com/repositories/1281269179/releases/latest"

	println("Checking for updates. Current version:", currentVersion)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(repoURL)
	if err != nil {
		println("Error fetching GitHub API:", err.Error())
		return UpdateCheckResult{HasUpdate: false}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		println("GitHub API returned status:", resp.StatusCode)
		return UpdateCheckResult{HasUpdate: false}
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		println("Error decoding GitHub response:", err.Error())
		return UpdateCheckResult{HasUpdate: false}
	}

	println("Latest version from GitHub:", release.TagName)

	// Compare versions (strip 'v' prefix if present)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	if current != latest {
		println("Update available! Current:", current, "Latest:", latest)
		return UpdateCheckResult{
			HasUpdate:   true,
			Version:     release.TagName,
			Url:         release.HtmlUrl,
			Description: release.Body,
		}
	}

	println("No update available. Current:", current, "Latest:", latest)
	return UpdateCheckResult{HasUpdate: false}
}

// ConvertToMarkdownFile converts a file on disk to markdown
func (a *App) ConvertToMarkdownFile(filePath string) ConversionResult {
	return ConvertToMarkdown(filePath)
}

// ConvertFileContent converts uploaded file bytes to markdown
// Called from the frontend when the user imports a file
func (a *App) ConvertFileContent(fileName string, content []byte) ConversionResult {
	return ConvertFileContent(fileName, content)
}
