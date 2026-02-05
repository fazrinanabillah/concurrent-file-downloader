package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadFile handles the logic for fetching a file from a URL and saving it locally
func DownloadFile(url, destDir string) error {
	fileName := filepath.Base(url)
	filePath := filepath.Join(destDir, fileName)

	// Create the local file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer out.Close()

	slog.Info("Downloading from", "url", url)
	start := time.Now().UTC()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Check for server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	// Stream the body content to the local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file content: %w", err)
	}

	slog.Info("Download finished", "duration", time.Since(start))

	return nil
}

func SequentialDownloader(urls []string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to creating directory: %w", err)
	}

	start := time.Now().UTC()
	for _, url := range urls {
		if err := DownloadFile(url, destDir); err != nil {
			slog.Error("Error downloading", "url", url)
			continue
		}
	}

	slog.Info("All downloads finished", "duration", time.Since(start))
	return nil
}

func main() {
	urls := []string{
		"https://www.w3.org/WAI/WCAG21/Techniques/pdf/img/table-word.jpg",
		"https://go.dev/images/go-logo-blue.svg",
	}

	if err := SequentialDownloader(urls, "./"); err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info("All tasks completed successfully")
}
