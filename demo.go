package main

import (
	"fmt"
	"time"
)

func main() {
	// Simulate download size
	totalSize := 11.5 * 1024 * 1024 // 11.5 MB in bytes
	downloaded := 0.0

	// Simulate download speed (2.8 MB/s)
	speed := 2.8 * 1024 * 1024 // 2.8 MB in bytes

	// Progress bar length
	barLength := 40

	for downloaded < totalSize {
		// Simulate downloading
		time.Sleep(500 * time.Millisecond) // Simulate download time
		downloaded += speed / 2            // Update downloaded size

		// Calculate percentage
		percent := downloaded / totalSize * 100

		// Calculate the length of the filled part of the progress bar
		filledLength := int(float64(barLength) * downloaded / totalSize)

		// Create the progress bar string
		bar := fmt.Sprintf("\r%s %6.1f%% | amberj-httpserver | %6.1fMB/%6.1fMB %.1fMB/s %ds",
			createBar(filledLength, barLength),
			percent,
			downloaded/(1024*1024),
			totalSize/(1024*1024),
			speed/(1024*1024),
			int(totalSize-downloaded)/int(speed),
		)

		// Print the progress bar
		fmt.Print(bar)

		// Break if download is complete
		if downloaded >= totalSize {
			break
		}
	}

	// Print a newline when the download is complete
	fmt.Println()
}

// createBar creates the progress bar string.
func createBar(filledLength, totalLength int) string {
	bar := ""
	for i := 0; i < totalLength; i++ {
		if i < filledLength {
			bar += "━"
		} else {
			bar += " "
		}
	}
	return bar
}
