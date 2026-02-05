# Go Concurrent Downloader
A personal project exploring Go's concurrency patterns. This downloader implements a worker pool to manage resources efficiently, handles graceful shutdowns via Context, and follows the standard Go project layout for better maintainability.

## How to Run

You can run the application directly using `go run`. Make sure to point to the entry file in the `cmd` directory.

```bash
go run cmd/app/main.go