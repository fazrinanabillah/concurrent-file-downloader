package main

import "time"

// DownloadResult holds the outcome of a single download operation.
type DownloadResult struct {
	URL      string
	FileName string
	Size     int64
	Duration time.Duration
	Error    error
}
