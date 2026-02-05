package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ConcurrentDownloader manages the worker pool and orchestrates teh downloads
func ConcurrentDownloader(urls []string, destDir string, maxConcurrent int) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return ErrorCreatingDirectory
	}

	// If the entire batch takes longer than 1 minute, all active download will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Graceful shutdown (Handle Ctrl + C)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		slog.Warn("\nReceive interrupted signal, shutting down...")
		// Trigger cancellation for all goroutines
		cancel()
	}()

	results := make(chan DownloadResult, len(urls))
	// Concurrency limiter (Semaphore Pattern)
	// We use struct{} because an empty struct occupies 0 bytes of memory
	// We only care about the signal/blocking not the data value
	limiter := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	startTotal := time.Now()

	slog.Info("Starting concurrent downloads...", "total_files", len(urls), "max_concurrent", maxConcurrent)

	for _, url := range urls {
		wg.Add(1)
		go func(targetURL string) {
			defer wg.Done()

			// Acquire token from semaphore
			limiter <- struct{}{}

			// Release token when done
			defer func() { <-limiter }()

			// Check if context is cancelled before starting work
			if ctx.Err() != nil {
				results <- DownloadResult{URL: targetURL, Error: ctx.Err()}
				return
			}

			// Execute download logic
			results <- DownloadFile(ctx, targetURL, destDir)

		}(url)
	}

	// Start separate goroutine to wait for all workers to finish
	// Then close te result channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process Results
	var totalSize int64
	var errorCount int

	for result := range results {
		if result.Error != nil {
			slog.Error("Download failed", "url", result.URL, "error", result.Error)
			errorCount++
		} else {
			slog.Info("Download success", "file", result.FileName, "size_bytes", result.Size, "duration", result.Duration)
			totalSize += result.Size
		}
	}

	slog.Info("All tasks completed",
		"total_duration", time.Since(startTotal),
		"total_size_bytes", totalSize,
		"successful_downloads", len(urls)-errorCount,
		"failed_downloads", errorCount,
	)

	return nil
}
