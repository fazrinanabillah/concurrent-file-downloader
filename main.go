package main

import (
	"log/slog"
	"os"
)

func main() {
	urls := []string{
		"https://file-examples.com/storage/feb18787c869840dc97160c/2017/10/file_example_JPG_1MB.jpg",
		"https://www.w3.org/WAI/WCAG21/Techniques/pdf/img/table-word.jpg",
		"https://go.dev/images/go-logo-blue.svg",
	}

	// Test case: duplicate URL to verify collision handling
	urls = append(urls, "https://go.dev/images/go-logo-blue.svg")

	// Run the downloader with concurrency limit of 3
	if err := ConcurrentDownloader(urls, "./downloads", 3); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}

	slog.Info("All tasks completed successfully")
}
