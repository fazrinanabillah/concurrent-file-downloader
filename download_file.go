package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadFile handles fetching a single file.
func DownloadFile(ctx context.Context, url, destDir string) DownloadResult {
	start := time.Now().UTC()
	fileName := filepath.Base(url)
	filePath := filepath.Join(destDir, fileName)

	// Prevent overwriting
	if _, err := os.Stat(filePath); err != nil {
		fileName = fmt.Sprintf("%d_%s", time.Now().UTC().Unix(), fileName)
		filePath = filepath.Join(destDir, fileName)
	}

	// create new HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return DownloadResult{URL: url, Error: err}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return DownloadResult{URL: url, Error: err}
	}

	defer resp.Body.Close()

	// Check for server response
	if resp.StatusCode != http.StatusOK {
		return DownloadResult{URL: url, Error: fmt.Errorf("server returned: %s", resp.Status)}
	}

	// Create the file on the local disk
	out, err := os.Create(filePath)
	if err != nil {
		return DownloadResult{URL: url, Error: err}
	}

	defer out.Close()

	// Stream the response body to the local file
	size, err := io.Copy(out, resp.Body)
	if err != nil {
		// Clean up the partial file on failure
		_ = os.Remove(filePath)
		return DownloadResult{URL: url, Error: err}
	}

	return DownloadResult{
		URL:      url,
		FileName: fileName,
		Size:     size,
		Duration: time.Since(start),
		Error:    nil,
	}
}
