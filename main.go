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

	fmt.Println("Downloading from: ", url)
	start := time.Now().UTC()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Check for server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s", resp.Status)
	}

	// Stream the body content to the local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Download finished in %s\n", time.Since(start))

	return nil
}

func main() {
	url := "https://www.w3.org/WAI/WCAG21/Techniques/pdf/img/table-word.jpg"

	if err := DownloadFile(url, "./"); err != nil {
		slog.Error(err.Error())
		return
	}

	fmt.Println("All tasks completed successfully")
}
