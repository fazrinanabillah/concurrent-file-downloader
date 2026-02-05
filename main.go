package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
		_ = os.Remove(filePath)
		return err
	}

	defer resp.Body.Close()

	// Check for server response
	if resp.StatusCode != http.StatusOK {
		_ = os.Remove(filePath)
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

type Result struct {
	URL      string
	FileName string
	Size     int64
	Duration time.Duration
	Error    error
}

func ConcurrentDownloader(urls []string, destDir string, maxConcurrent int) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	results := make(chan Result)
	limiter := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			limiter <- struct{}{}
			defer func() { <-limiter }()

			start := time.Now().UTC()
			fileName := filepath.Base(url)
			filePath := filepath.Join(destDir, fileName)

			out, err := os.Create(filePath)
			if err != nil {
				results <- Result{URL: url, Error: err}
				return
			}
			defer out.Close()

			resp, err := http.Get(url)
			if err != nil {
				_ = os.Remove(filePath)
				results <- Result{URL: url, Error: err}
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				_ = os.Remove(filePath)
				results <- Result{URL: url, Error: fmt.Errorf("status: %s", resp.Status)}
				return
			}

			size, err := io.Copy(out, resp.Body)
			if err != nil {
				_ = os.Remove(filePath)
				results <- Result{URL: url, Error: err}
				return
			}

			timeSince := time.Since(start)
			results <- Result{URL: url, FileName: fileName, Size: size, Duration: timeSince, Error: nil}

		}(url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalSize int64
	var errors []error
	start := time.Now().UTC()

	for result := range results {
		if result.Error != nil {
			fmt.Printf("Error downloading %s: %s\n", result.URL, result.Error.Error())
			errors = append(errors, result.Error)
		} else {
			totalSize += result.Size
			fmt.Printf("Downloaded %s (%d bytes) in %s\n", result.FileName, result.Size, result.Duration)
		}
	}

	startedSince := time.Since(start)
	fmt.Printf("All downloads completed in %s, Total: %d bytes\n", startedSince, totalSize)

	if len(errors) > 0 {
		return fmt.Errorf("error downloading: %+v", errors)
	}

	return nil
}

func main() {
	urls := []string{
		"https://file-examples.com/storage/feb18787c869840dc97160c/2017/10/file_example_JPG_1MB.jpg",
		"https://www.w3.org/WAI/WCAG21/Techniques/pdf/img/table-word.jpg",
		"https://go.dev/images/go-logo-blue.svg",
	}

	if err := ConcurrentDownloader(urls, "./", 3); err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info("All tasks completed successfully")
}
